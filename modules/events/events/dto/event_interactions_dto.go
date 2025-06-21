package dto

import (
	"time"
)

type AddCommentInput struct {
	Content string `json:"content" validate:"required,min=1,max=500"`
}

type EditCommentInput struct {
	Content string `json:"content" validate:"required,min=1,max=500"`
}

type FavoriteResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	EventID   string    `json:"event_id"`
	CreatedAt time.Time `json:"created_at"`
	Event     struct {
		ID        string    `json:"id"`
		Title     string    `json:"title"`
		StartTime time.Time `json:"start_time"`
		EventType string    `json:"event_type"`
	} `json:"event"`
}

type CommentResponse struct {
	ID        string            `json:"id"`
	UserID    string            `json:"user_id"`
	EventID   string            `json:"event_id"`
	Content   string            `json:"content"`
	ParentID  *string           `json:"parent_id,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	User      CommentUser       `json:"user"`
	Replies   []CommentResponse `json:"replies,omitempty"`
}

type CommentUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

type VoteResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	EventID   string    `json:"event_id"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type EventInteractionStats struct {
	Upvotes     int64 `json:"upvotes"`
	Downvotes   int64 `json:"downvotes"`
	Comments    int64 `json:"comments"`
	Favorites   int64 `json:"favorites"`
	IsFavorited bool  `json:"is_favorited"`
	UserVote    *int  `json:"user_vote"`
}

type UserInteractionsResponse struct {
	Favorites []FavoriteResponse `json:"favorites"`
	Comments  []CommentResponse  `json:"comments"`
	Votes     []VoteResponse     `json:"votes"`
}
