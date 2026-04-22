# API Endpoints Guide

All API endpoints are served from the modular monolith on a single port (default: `8080`). All routes are prefixed with `/api`.

## Base URL
`http://localhost:8080/api`

## Authentication
Most endpoints require a Bearer token in the `Authorization` header.
`Authorization: Bearer <access_token>`

---

### 1. Auth Module
- `POST /auth/register` - New user registration
- `POST /auth/login` - Returns `accessToken` and `refreshToken`
- `GET /auth/me` - Current user information
- `POST /auth/refresh` - Refresh access token
- `POST /auth/logout` - Revoke session
- `POST /auth/forgot-password` - Trigger recovery email

### 2. Listings Module
- `GET /listings` - Search and filter listings
- `GET /listings/:id` - Detailed listing view
- `POST /listings` - Create new listing (Protected)
- `PATCH /listings/:id` - Update listing (Owner)
- `DELETE /listings/:id` - Remove listing (Owner)
- `GET /brands` - List available makes
- `GET /models` - List models for a make

### 3. Messaging Module
- `GET /conversations` - List user chats
- `POST /conversations` - Start new chat on a listing
- `POST /conversations/:id/messages` - Send a message
- `PUT /conversations/:id/read` - Mark as read

### 4. Image Module
- `POST /listings/:id/images` - Upload images
- `PUT /listings/:id/images/reorder` - Set display order
- `PUT /users/me/avatar` - Update profile picture

### 5. Favorites & Comparison
- `POST /favorites/:listingId` - Add to watchlist
- `GET /favorites` - List watchlisted items
- `POST /comparison/:listingId` - Add to compare list
- `GET /comparison/compare` - Side-by-side view

### 6. Admin & Support
- `POST /contact` - Public support inquiry
- `GET /contact` - List inquiries (Admin)
- `POST /reports` - Report a listing for moderation
- `GET /reports` - View reports (Admin)

---

> [!TIP]
> For a full list of request payloads and automated tests, refer to [request.http](../request.http) or the Postman collection in the root directory.
