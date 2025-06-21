package service

import (
	"errors"
	"fmt"
	"html"
	"regexp"
	"strings"
	"ticket-zetu-api/modules/events/models/events"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	MaxCommentLength          = 500
	VoteTypeUp       VoteType = "up"
	VoteTypeDown     VoteType = "down"
)

type VoteType string

// ContentFilter defines rules for content validation
type ContentFilter struct {
	profanityRegex *regexp.Regexp
	scriptRegex    *regexp.Regexp
}

func NewContentFilter() *ContentFilter {
	return &ContentFilter{
		profanityRegex: regexp.MustCompile(`(?i)\b(fuck|shit|damn|asshole|bastard|bitch|cock|dick|porn|sex|xxx)\b`),
		scriptRegex:    regexp.MustCompile(`(?i)<\s*script|javascript:|\bon\w+=`),
	}
}

// ValidateContent checks content for length, profanity, and scripts
func (f *ContentFilter) ValidateContent(content string, maxLength int) error {
	// Check length
	if utf8.RuneCountInString(content) > maxLength {
		return fmt.Errorf("content exceeds maximum length of %d characters", maxLength)
	}

	// Check for empty content
	if strings.TrimSpace(content) == "" {
		return errors.New("content cannot be empty")
	}

	// Check for profanity
	if f.profanityRegex.MatchString(content) {
		return errors.New("content contains inappropriate language")
	}

	// Check for script tags or dangerous attributes
	if f.scriptRegex.MatchString(content) {
		return errors.New("content contains potential script injection")
	}

	return nil
}

// AddComment handles adding a new top-level comment
func (s *eventService) AddComment(userID, eventID, content string) (*events.Comment, error) {
	if userID == "" || eventID == "" || content == "" {
		return nil, errors.New("userID, eventID, and content are required")
	}

	// Validate and sanitize content
	if err := s.contentFilter.ValidateContent(content, MaxCommentLength); err != nil {
		return nil, err
	}
	sanitizedContent := html.EscapeString(strings.TrimSpace(content))

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	comment := events.Comment{
		ID:        uuid.New().String(),
		UserID:    userID,
		EventID:   eventID,
		Content:   sanitizedContent,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := tx.Create(&comment).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	// Fetch event and organizer details
	var event struct {
		Title       string `json:"title"`
		OrganizerID string `json:"organizer_id"`
	}
	if err := tx.Table("events").Where("id = ?", eventID).First(&event).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to get event details: %w", err)
	}

	var organizer struct {
		CreatedBy string `json:"created_by"`
		Name      string `json:"name"`
	}
	if err := tx.Table("organizers").Where("id = ?", event.OrganizerID).First(&organizer).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to get organizer details: %w", err)
	}

	var commentAuthor struct {
		Username string `json:"username"`
	}
	if err := tx.Table("user_profiles").Where("id = ?", userID).First(&commentAuthor).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to get comment author details: %w", err)
	}

	// Send notifications after successful comment creation
	metadata := map[string]interface{}{
		"event_id":           eventID,
		"event_title":        event.Title,
		"comment_id":         comment.ID,
		"commenter_id":       userID,
		"commenter_username": commentAuthor.Username,
		"comment_text":       sanitizedContent,
	}

	// Notify commenter
	s.sendNotification(
		"comment",
		"Commented on "+event.Title,
		fmt.Sprintf("You have successfully commented on %s: %s", event.Title, sanitizedContent),
		userID,
		eventID,
		[]string{userID},
		metadata,
	)

	// Notify organizer owner
	if organizer.CreatedBy != "" && organizer.CreatedBy != userID {
		metadata["action"] = "new_comment"
		s.sendNotification(
			"new_comment",
			"New Comment on "+event.Title,
			fmt.Sprintf("%s has commented on your event, %s: %s", commentAuthor.Username, event.Title, sanitizedContent),
			userID,
			eventID,
			[]string{organizer.CreatedBy},
			metadata,
		)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &comment, nil
}

// AddReply handles adding a reply as a comment with parent_id
func (s *eventService) AddReply(userID, eventID, commentID, content string) (*events.Comment, error) {
	if userID == "" || eventID == "" || commentID == "" || content == "" {
		return nil, errors.New("userID, eventID, commentID, and content are required")
	}

	// Validate and sanitize content
	if err := s.contentFilter.ValidateContent(content, MaxCommentLength); err != nil {
		return nil, err
	}
	sanitizedContent := html.EscapeString(strings.TrimSpace(content))

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Validate parent comment exists and belongs to the event
	var parentComment events.Comment
	if err := tx.Where("id = ? AND event_id = ?", commentID, eventID).First(&parentComment).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("parent comment not found or doesn't belong to this event")
		}
		return nil, fmt.Errorf("failed to validate parent comment: %w", err)
	}

	reply := events.Comment{
		ID:        uuid.New().String(),
		UserID:    userID,
		EventID:   eventID,
		Content:   sanitizedContent,
		ParentID:  &commentID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := tx.Create(&reply).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create reply: %w", err)
	}

	// Fetch event and organizer details
	var event struct {
		Title       string `json:"title"`
		OrganizerID string `json:"organizer_id"`
	}
	if err := tx.Table("events").Where("id = ?", eventID).First(&event).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to get event details: %w", err)
	}

	var organizer struct {
		CreatedBy string `json:"created_by"`
		Name      string `json:"name"`
	}
	if err := tx.Table("organizers").Where("id = ?", event.OrganizerID).First(&organizer).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to get organizer details: %w", err)
	}

	var replyAuthor struct {
		Username string `json:"username"`
	}
	if err := tx.Table("user_profiles").Where("id = ?", userID).First(&replyAuthor).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to get reply author details: %w", err)
	}

	// Send notifications after successful reply creation
	metadata := map[string]interface{}{
		"event_id":         eventID,
		"event_title":      event.Title,
		"comment_id":       commentID,
		"reply_id":         reply.ID,
		"replier_id":       userID,
		"replier_username": replyAuthor.Username,
		"reply_text":       sanitizedContent,
	}

	// Notify replier
	s.sendNotification(
		"reply",
		"Replied to a comment on "+event.Title,
		fmt.Sprintf("You have successfully replied to a comment on %s: %s", event.Title, sanitizedContent),
		userID,
		eventID,
		[]string{userID},
		metadata,
	)

	// Notify parent comment author
	if parentComment.UserID != userID {
		metadata["action"] = "new_reply"
		s.sendNotification(
			"new_reply",
			"New Reply to Your Comment on "+event.Title,
			fmt.Sprintf("%s has replied to your comment on %s: %s", replyAuthor.Username, event.Title, sanitizedContent),
			userID,
			eventID,
			[]string{parentComment.UserID},
			metadata,
		)
	}

	// Notify organizer owner if different from replier
	if organizer.CreatedBy != "" && organizer.CreatedBy != userID {
		metadata["action"] = "new_reply"
		s.sendNotification(
			"new_reply",
			"New Reply on "+event.Title,
			fmt.Sprintf("%s has replied to a comment on your event, %s: %s", replyAuthor.Username, event.Title, sanitizedContent),
			userID,
			eventID,
			[]string{organizer.CreatedBy},
			metadata,
		)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &reply, nil
}

// EditComment handles updating both comments and replies
func (s *eventService) EditComment(userID, commentID, newContent string) (*events.Comment, error) {
	if userID == "" || commentID == "" || newContent == "" {
		return nil, errors.New("userID, commentID, and newContent are required")
	}

	// Validate and sanitize content
	if err := s.contentFilter.ValidateContent(newContent, MaxCommentLength); err != nil {
		return nil, err
	}
	sanitizedContent := html.EscapeString(strings.TrimSpace(newContent))

	var comment events.Comment
	if err := s.db.Where("id = ? AND user_id = ?", commentID, userID).First(&comment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("comment not found or not owned by user")
		}
		return nil, fmt.Errorf("failed to find comment: %w", err)
	}

	// Check edit window
	if time.Since(comment.CreatedAt) > 10*time.Minute {
		return nil, errors.New("comment can only be edited within 10 minutes of creation")
	}

	comment.Content = sanitizedContent
	comment.UpdatedAt = time.Now()

	if err := s.db.Save(&comment).Error; err != nil {
		return nil, fmt.Errorf("failed to update comment: %w", err)
	}

	return &comment, nil
}

// DeleteComment handles deleting a comment and its replies
func (s *eventService) DeleteComment(userID, commentID string) error {
	if userID == "" || commentID == "" {
		return errors.New("userID and commentID are required")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var comment events.Comment
		if err := tx.Where("id = ? AND user_id = ?", commentID, userID).First(&comment).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("comment not found or not owned by user")
			}
			return fmt.Errorf("failed to find comment: %w", err)
		}

		// Delete the comment and all its replies
		if err := tx.Where("id = ? OR parent_id = ?", commentID, commentID).Delete(&events.Comment{}).Error; err != nil {
			return fmt.Errorf("failed to delete comment: %w", err)
		}

		return nil
	})
}
