# Organizer API Documentation

This document describes the API endpoints for managing organizers and their images.

## Base URL
All endpoints start with: `/api/v1`

## Authentication
All endpoints require authentication. Include a valid JWT token in the `Authorization` header.

## Organizer Endpoints

### Get All Organizers
**GET** `/organizers`

Returns a list of all organizers the user has permission to view.

**Permissions Required**: `view:organizers`

**Response**:
```json
[
  {
    "id": "uuid",
    "name": "string",
    "contact_person": "string",
    "email": "string",
    "phone": "string",
    "company_name": "string",
    "tax_id": "string",
    "bank_account_info": "string",
    "image_url": "string",
    "commission_rate": 0.0,
    "balance": 0.0,
    "status": "active|inactive",
    "is_flagged": false,
    "is_banned": false,
    "created_by": "uuid",
    "created_by_user": {
      "id": "uuid",
      "email": "string",
      "username": "string",
      "first_name": "string",
      "last_name": "string"
    }
  }
]
```

### Get My Organizer
**GET** `/organizers/my-organization`

Returns the organizer record for the currently authenticated user.

**Response**: Same structure as Get All Organizers

### Get Single Organizer
**GET** `/organizers/:id`

Returns details for a specific organizer.

**Parameters**:
- `id`: Organizer UUID

**Permissions Required**: `view:organizers`

**Response**: Same structure as Get All Organizers

### Search Organizers
**GET** `/organizers/search`

Search organizers with pagination.

**Query Parameters**:
- `search`: Search term (optional)
- `created_by`: Filter by creator UUID (optional)
- `page`: Page number (default: 1)
- `page_size`: Items per page (default: 10)

**Response**:
```json
{
  "data": [/* organizer objects */],
  "pagination": {
    "total": 0,
    "page": 1,
    "page_size": 10,
    "total_pages": 1
  }
}
```

### Create Organizer
**POST** `/organizers`

Create a new organizer.

**Permissions Required**: `create:organizers`

**Request Body**:
```json
{
  "name": "string (required, 2-255 chars)",
  "contact_person": "string (required, 2-255 chars)",
  "email": "string (required, valid email)",
  "phone": "string (optional, max 50 chars)",
  "company_name": "string (optional, max 255 chars)",
  "tax_id": "string (optional, max 100 chars)",
  "bank_account_info": "string (optional)",
  "commission_rate": "number (0-100)",
  "balance": "number (>=0)",
  "notes": "string (optional)"
}
```

**Response**: Success message

### Update Organizer
**PUT** `/organizers/:id`

Update an existing organizer.

**Permissions Required**: `update:organizers`

**Request Body**: Same as Create Organizer

**Response**: Success message

### Delete Organizer
**DELETE** `/organizers/:id`

Delete an organizer. Organizer must be inactive first.

**Permissions Required**: `delete:organizers`

**Response**: Success message

### Deactivate Organizer
**PATCH** `/organizers/:id/deactivate`

Deactivate an organizer.

**Permissions Required**: `update:organizers`

**Response**: Success message

### Toggle Organizer Status
**PATCH** `/organizers/:id/toggle-status`

Toggle between active/inactive status.

**Permissions Required**: `update:organizers`

**Response**: Success message

### Flag Organizer
**PATCH** `/organizers/:id/flag`

Toggle flag status on an organizer.

**Permissions Required**: `update:organizers`

**Response**: Success message

### Ban Organizer
**PATCH** `/organizers/:id/ban`

Toggle ban status on an organizer.

**Permissions Required**: `update:organizers`

**Response**: Success message

## Organization Image Endpoints

### Add Organization Image
**POST** `/organizations/:organization_id/image`

Upload an image for an organization.

**Permissions Required**: `update:organizations`

**Request**:
- Content-Type: `multipart/form-data`
- Form field: `image` (single image file)

**File Requirements**:
- Max size: 10MB
- Allowed types: JPEG, PNG, GIF, WebP

**Response**:
```json
{
  "image_url": "string"
}
```

### Delete Organization Image
**DELETE** `/organizations/:organization_id/image`

Remove an organization's image.

**Permissions Required**: `update:organizations`

**Response**: Success message

## Error Responses
All errors follow this format:
```json
{
  "error": "string",
  "message": "string",
  "status": "error"
}
```

Common error statuses:
- 400 Bad Request (validation errors)
- 401 Unauthorized (missing/invalid token)
- 403 Forbidden (missing permissions)
- 404 Not Found (resource doesn't exist)
- 409 Conflict (duplicate email)
- 500 Internal Server Error
```

This README provides:
1. Clear endpoint documentation
2. Required permissions for each endpoint
3. Request/response formats
4. Validation requirements
5. Error handling information
6. Organization image upload specifics

