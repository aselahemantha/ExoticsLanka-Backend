# API Services & Endpoints

This document outlines the microservices in the ExoticsLanka backend and their relevant API endpoints. Most endpoints require an authentication token (`Authorization: Bearer <token>`).

## 1. Auth Service
Port: `8081`

- `POST /api/auth/register` - Register a new user
- `POST /api/auth/login` - Login to get access tokens
- `GET /api/auth/me` - Get current user info (Protected)
- `POST /api/auth/refresh` - Refresh access token
- `POST /api/auth/change-password` - Change user password
- `POST /api/auth/logout` - Logout user
- `POST /api/auth/forgot-password` - Request password reset

## 2. User/Profile Service
Port: `8082`

- `GET /api/users/me` - Get current user profile
- `PATCH /api/users/me` - Update user profile
- `POST /api/users/me/verification` - Request seller/dealer verification
- `DELETE /api/users/me` - Delete user account

## 3. Listings Service
Port: `8083`

- `POST /api/listings` - Create a new listing
- `GET /api/listings` - Get all listings (supports search & filter via query params)
- `GET /api/listings/{id}` - Get listing by ID
- `PATCH /api/listings/{id}` - Update listing
- `DELETE /api/listings/{id}` - Delete listing
- `GET /api/listings/featured` - Get featured listings
- `GET /api/listings/trending` - Get trending listings
- `GET /api/brands` - Get available car brands
- `GET /api/models?brand={brand}` - Get models for a given brand
- `POST /api/listings/{id}/view` - Increment listing view counter
- `GET /api/users/me/listings` - Get the user's created listings
- `POST /api/listings/{id}/favorite` - Add a listing to favorites
- `DELETE /api/listings/{id}/favorite` - Remove a listing from favorites
- `GET /api/users/me/favorites` - Get all of the user's favorited listings
- `POST /api/listings/{id}/reports` - Submit a report for moderation

## 4. Reviews Service
Port: `8085`

- `POST /api/reviews` - Create a new review
- `GET /api/reviews/seller/{seller_id}` - Get reviews for a seller
- `GET /api/reviews/seller/{seller_id}/stats` - Get seller review statistics
- `POST /api/reviews/{review_id}/helpful` - Mark review as helpful
- `POST /api/reviews/{review_id}/response` - Respond to a review (Seller)
- `POST /api/reviews/{review_id}/photos` - Add photos to a review
- `DELETE /api/reviews/{review_id}` - Delete a review

## 5. Messaging Service
Port: `8086`

- `POST /api/conversations` - Create a new conversation
- `GET /api/conversations` - Get all user conversations
- `GET /api/messages/unread-count` - Get total unread messages count
- `POST /api/conversations/{conversation_id}/messages` - Reply/Send a message in a conversation
- `PUT /api/conversations/{conversation_id}/read` - Mark a conversation as read
- `GET /api/conversations/{conversation_id}` - Get conversation details and messages

## 6. Saved Searches Service
Port: `8087`

- `POST /api/searches` - Create a saved search
- `GET /api/searches` - Get all saved searches
- `POST /api/searches/{search_id}/run` - Run a specific saved search
- `POST /api/searches/{search_id}/check` - Check for new matches for a search
- `GET /api/searches/new-matches` - Get overview of all new matches

## 7. Contact Service
Port: `8089`

- `POST /api/contact` - Submit a contact inquiry
- `GET /api/contact` - Get all inquiries (Admin)
- `GET /api/contact/{inquiry_id}` - Get inquiry details (Admin)
- `PUT /api/contact/{inquiry_id}` - Respond to an inquiry (Admin)
- `GET /api/contact/stats` - Get inquiry stats (Admin)

## 8. Analytics Service
Port: `8090`

- `POST /api/analytics/track` - Track an event (view, contact_view, etc.)
- `POST /api/analytics/jobs/aggregate` - Trigger analytics aggregation (Admin/Dealer)
- `GET /api/analytics/overview` - Get dashboard overview
- `GET /api/analytics/insights` - Get analytics insights
- `GET /api/analytics/inventory` - Get inventory performance metrics

## 9. Comparison Service
Port: `8091`

- `POST /api/comparison/{listing_id}` - Add a listing to comparison
- `GET /api/comparison` - Get the current comparison list
- `GET /api/comparison/check/{listing_id}` - Check if listing is in comparison
- `GET /api/comparison/compare` - Compare view for listings
- `DELETE /api/comparison/{listing_id}` - Remove a listing from comparison
- `DELETE /api/comparison` - Clear comparison list

## 10. Image Service
Handles image uploads and management.
- `POST /images/upload` *(Route format varies)* - Upload a listing image
- `PUT /images/reorder` *(Route format varies)* - Reorder listing images
- `DELETE /images/{imageId}` *(Route format varies)* - Delete a listing image
- `POST /images/avatar` *(Route format varies)* - Upload user avatar profile picture

## 11. Notification Service
Handles user notification preferences and delivery.
- `GET /notifications/preferences` *(Route format varies)* - Get notification preferences
- `PUT /notifications/preferences` *(Route format varies)* - Update notification preferences
- `POST /notifications/send` *(Route format varies)* - Send an internal system notification
