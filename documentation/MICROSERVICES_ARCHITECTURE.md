# 🏗️ Exotics Lanka - Microservices Architecture

**Document Version:** 1.0  
**Date:** January 6, 2026  
**Architecture Type:** Microservices + Serverless Functions  

---

## 📋 **TABLE OF CONTENTS**

1. [Architecture Overview](#architecture-overview)
2. [Microservices Breakdown](#microservices-breakdown)
3. [Micro Functions](#micro-functions)
4. [API Endpoints](#api-endpoints)
5. [Database Schema](#database-schema)
6. [Service Communication](#service-communication)
7. [Technology Stack](#technology-stack)
8. [Deployment Strategy](#deployment-strategy)

---

## 🎯 **ARCHITECTURE OVERVIEW**

### **Current State:**
- ✅ Frontend: React + TypeScript (Complete)
- ⏳ Backend: Not implemented (using localStorage)
- ⏳ Database: Not implemented
- ⏳ Authentication: Mock implementation

### **Target State:**
- ✅ Frontend: React + TypeScript (Deployed)
- 🎯 Backend: 8 Microservices + Serverless Functions
- 🎯 Database: PostgreSQL + Redis + S3
- 🎯 Authentication: JWT + OAuth

### **Architecture Pattern:**
```
┌─────────────────────────────────────────────────────┐
│                    API Gateway                       │
│              (Kong / AWS API Gateway)                │
└─────────────────────────────────────────────────────┘
                         │
        ┌────────────────┼────────────────┐
        │                │                │
┌───────▼──────┐  ┌──────▼──────┐  ┌────▼─────┐
│   Service 1   │  │  Service 2  │  │Service 3 │
│ Authentication│  │   Listings  │  │  Users   │
└───────────────┘  └─────────────┘  └──────────┘
        │                │                │
        └────────────────┼────────────────┘
                         │
                ┌────────▼─────────┐
                │  Message Queue   │
                │  (RabbitMQ/SQS)  │
                └──────────────────┘
```

---

## 🔧 **MICROSERVICES BREAKDOWN**

### **1. Authentication Service** 🔐

**Responsibility:** User authentication, authorization, session management

**Micro Functions:**
- `registerUser()` - Create new user account
- `loginUser()` - Authenticate user credentials
- `logoutUser()` - Invalidate session/token
- `refreshToken()` - Renew access token
- `verifyEmail()` - Email verification
- `resetPassword()` - Password reset flow
- `changePassword()` - Update user password
- `validateToken()` - JWT token validation
- `getAuthSession()` - Get current session
- `revokeToken()` - Revoke specific token

**Database Tables:**
- `users` (id, email, password_hash, role, status)
- `sessions` (id, user_id, token, expires_at)
- `verification_tokens` (id, user_id, token, type, expires_at)

**Tech Stack:**
- Node.js / Go
- JWT + bcrypt
- Redis (session storage)
- PostgreSQL (user data)

---

### **2. User Profile Service** 👤

**Responsibility:** User profile management, preferences, settings

**Micro Functions:**
- `getUserProfile()` - Get user profile details
- `updateProfile()` - Update user information
- `uploadAvatar()` - Upload profile picture
- `getSellerProfile()` - Get seller-specific data
- `updateSellerInfo()` - Update seller information
- `getPreferences()` - Get user preferences
- `updatePreferences()` - Update user settings
- `getUserStats()` - Get user statistics
- `deactivateAccount()` - Deactivate user account
- `deleteAccount()` - Permanently delete account

**Database Tables:**
- `user_profiles` (id, user_id, full_name, phone, location, bio)
- `seller_profiles` (id, user_id, business_name, rating, verified)
- `user_preferences` (id, user_id, notifications, language, theme)

**Tech Stack:**
- Node.js / Python
- AWS S3 (image storage)
- PostgreSQL

---

### **3. Listing Service** 🚗

**Responsibility:** Car listing CRUD operations, search, filtering

**Micro Functions:**
- `createListing()` - Create new car listing
- `updateListing()` - Update listing details
- `deleteListing()` - Delete listing
- `getListing()` - Get single listing
- `getListings()` - Get all listings (paginated)
- `searchListings()` - Search with filters
- `getMyListings()` - Get user's listings
- `publishListing()` - Change status to active
- `unpublishListing()` - Change status to draft
- `markAsSold()` - Mark listing as sold
- `uploadImages()` - Upload car images
- `reorderImages()` - Change image order
- `deleteImage()` - Delete car image
- `getFeaturedListings()` - Get featured cars
- `incrementViews()` - Track listing views

**Database Tables:**
- `listings` (id, user_id, title, description, price, year, make, model, mileage, fuel_type, transmission, location, condition, status, views, created_at)
- `listing_images` (id, listing_id, url, order, is_primary)
- `listing_features` (id, listing_id, feature_name)
- `listing_views` (id, listing_id, user_id, viewed_at, ip_address)

**Tech Stack:**
- Node.js / Python
- Elasticsearch (search)
- AWS S3 (images)
- PostgreSQL
- Redis (cache)

---

### **4. Favorites Service** ❤️

**Responsibility:** User favorites/wishlist management

**Micro Functions:**
- `addToFavorites()` - Add listing to favorites
- `removeFromFavorites()` - Remove from favorites
- `getFavorites()` - Get user's favorites
- `isFavorite()` - Check if listing is favorited
- `clearAllFavorites()` - Remove all favorites
- `shareWishlist()` - Generate shareable link
- `getSharedWishlist()` - View shared wishlist
- `getFavoriteStats()` - Get favorite statistics

**Database Tables:**
- `favorites` (id, user_id, listing_id, created_at)
- `shared_wishlists` (id, user_id, token, expires_at)

**Tech Stack:**
- Node.js / Go
- Redis (fast access)
- PostgreSQL

---

### **5. Messaging Service** 💬

**Responsibility:** Buyer-seller communication, chat, notifications

**Micro Functions:**
- `sendMessage()` - Send a message
- `getConversations()` - Get all conversations
- `getMessages()` - Get messages in conversation
- `markAsRead()` - Mark messages as read
- `deleteConversation()` - Delete conversation
- `getUnreadCount()` - Get unread message count
- `searchMessages()` - Search in messages
- `blockUser()` - Block a user
- `reportConversation()` - Report abuse
- `quickReply()` - Use quick reply template

**Database Tables:**
- `conversations` (id, listing_id, buyer_id, seller_id, last_message_at)
- `messages` (id, conversation_id, sender_id, content, read_at, created_at)
- `blocked_users` (id, user_id, blocked_user_id)

**Tech Stack:**
- Node.js + Socket.io
- MongoDB / PostgreSQL
- Redis (pub/sub)
- WebSocket (real-time)

---

### **6. Review & Rating Service** ⭐

**Responsibility:** Seller ratings, reviews, reputation management

**Micro Functions:**
- `createReview()` - Write a review
- `updateReview()` - Edit review
- `deleteReview()` - Delete review
- `getReviews()` - Get reviews for seller
- `getAverageRating()` - Calculate average rating
- `respondToReview()` - Seller response
- `reportReview()` - Report inappropriate review
- `voteHelpful()` - Mark review as helpful
- `getReviewStats()` - Get review statistics
- `verifyBuyer()` - Verify buyer badge

**Database Tables:**
- `reviews` (id, listing_id, seller_id, reviewer_id, rating, title, content, verified_buyer, created_at)
- `review_responses` (id, review_id, seller_id, content, created_at)
- `review_votes` (id, review_id, user_id, vote_type)
- `review_reports` (id, review_id, reporter_id, reason)

**Tech Stack:**
- Node.js / Python
- PostgreSQL
- Redis (cache ratings)

---

### **7. Search & Filter Service** 🔍

**Responsibility:** Advanced search, saved searches, alerts

**Micro Functions:**
- `performSearch()` - Execute search query
- `applyFilters()` - Apply multiple filters
- `getSuggestions()` - Auto-suggest results
- `saveSearch()` - Save search criteria
- `getSavedSearches()` - Get user's saved searches
- `deleteSavedSearch()` - Delete saved search
- `createAlert()` - Create search alert
- `getMatchingListings()` - Find new matches
- `sendAlerts()` - Send notification for matches
- `updateAlertFrequency()` - Change alert frequency

**Database Tables:**
- `saved_searches` (id, user_id, name, filters, created_at)
- `search_alerts` (id, saved_search_id, user_id, frequency, last_sent_at, active)
- `alert_matches` (id, alert_id, listing_id, notified_at)

**Tech Stack:**
- Node.js / Python
- Elasticsearch / Algolia
- PostgreSQL
- Cron Jobs / Lambda

---

### **8. Comparison Service** 🔄

**Responsibility:** Vehicle comparison functionality

**Micro Functions:**
- `addToComparison()` - Add vehicle to comparison
- `removeFromComparison()` - Remove vehicle
- `getComparisonList()` - Get vehicles to compare
- `compareVehicles()` - Generate comparison data
- `clearComparison()` - Clear all vehicles
- `shareComparison()` - Generate shareable link
- `getComparisonCount()` - Get current count

**Database Tables:**
- `comparisons` (id, user_id, listing_ids, created_at)
- `shared_comparisons` (id, user_id, listing_ids, token, expires_at)

**Tech Stack:**
- Node.js / Go
- Redis (temporary storage)
- PostgreSQL

---

### **9. Analytics Service** 📊

**Responsibility:** Tracking, analytics, reporting

**Micro Functions:**
- `trackPageView()` - Track page views
- `trackUserAction()` - Track user actions
- `trackListingView()` - Track listing views
- `getDashboardStats()` - Get seller dashboard stats
- `getPerformanceMetrics()` - Get listing performance
- `generateReport()` - Generate analytics report
- `trackConversion()` - Track conversion events
- `getPopularListings()` - Get trending listings
- `getUserBehavior()` - Analyze user behavior

**Database Tables:**
- `page_views` (id, user_id, page, timestamp)
- `user_actions` (id, user_id, action_type, metadata, timestamp)
- `listing_analytics` (id, listing_id, views, favorites, messages, date)

**Tech Stack:**
- Node.js / Python
- ClickHouse / BigQuery
- Redis (real-time)
- Kafka (event streaming)

---

### **10. Notification Service** 🔔

**Responsibility:** Email, SMS, push notifications

**Micro Functions:**
- `sendEmail()` - Send email notification
- `sendSMS()` - Send SMS notification
- `sendPushNotification()` - Send push notification
- `sendWelcomeEmail()` - Welcome new users
- `sendVerificationEmail()` - Email verification
- `sendPasswordReset()` - Password reset email
- `sendMessageNotification()` - New message alert
- `sendAlertNotification()` - Search alert notification
- `getNotificationHistory()` - Get sent notifications
- `updateNotificationPreferences()` - Update preferences

**Database Tables:**
- `notifications` (id, user_id, type, channel, status, sent_at)
- `notification_preferences` (id, user_id, email_enabled, sms_enabled, push_enabled)
- `email_templates` (id, name, subject, body)

**Tech Stack:**
- Node.js
- SendGrid / AWS SES (email)
- Twilio (SMS)
- Firebase (push)
- Queue (Bull / SQS)

---

### **11. Payment Service** 💳

**Responsibility:** Payment processing, subscriptions, featured listings

**Micro Functions:**
- `createPaymentIntent()` - Initialize payment
- `processPayment()` - Process payment
- `refundPayment()` - Refund transaction
- `getPaymentHistory()` - Get transaction history
- `createSubscription()` - Create subscription plan
- `cancelSubscription()` - Cancel subscription
- `upgradeListing()` - Feature listing (paid)
- `getInvoice()` - Generate invoice
- `validatePayment()` - Verify payment status

**Database Tables:**
- `payments` (id, user_id, amount, status, payment_method, created_at)
- `subscriptions` (id, user_id, plan_id, status, starts_at, ends_at)
- `featured_listings` (id, listing_id, payment_id, starts_at, ends_at)
- `invoices` (id, payment_id, invoice_number, pdf_url)

**Tech Stack:**
- Node.js
- Stripe / PayPal
- PostgreSQL
- AWS S3 (invoices)

---

### **12. Media Service** 📸

**Responsibility:** Image upload, processing, video hosting

**Micro Functions:**
- `uploadImage()` - Upload single image
- `uploadMultipleImages()` - Batch upload
- `resizeImage()` - Generate thumbnails
- `optimizeImage()` - Compress images
- `deleteImage()` - Remove image
- `uploadVideo()` - Upload video
- `generateThumbnail()` - Video thumbnail
- `streamVideo()` - Video streaming
- `get360View()` - 360° view images

**Database Tables:**
- `media_files` (id, user_id, filename, url, type, size, created_at)
- `image_variants` (id, media_file_id, variant_type, url)

**Tech Stack:**
- Node.js / Python
- AWS S3 / Cloudinary
- Sharp / Pillow (image processing)
- FFmpeg (video processing)
- CDN (CloudFront)

---

### **13. Report & Moderation Service** 🛡️

**Responsibility:** Content moderation, reporting, spam detection

**Micro Functions:**
- `reportListing()` - Report inappropriate listing
- `reportUser()` - Report user
- `reportReview()` - Report review
- `getModerationQueue()` - Get items to review
- `approveContent()` - Approve content
- `rejectContent()` - Reject/remove content
- `banUser()` - Ban user account
- `detectSpam()` - Auto spam detection
- `getReportedItems()` - Get all reports

**Database Tables:**
- `reports` (id, reporter_id, reported_type, reported_id, reason, status, created_at)
- `moderation_actions` (id, moderator_id, action_type, target_id, created_at)
- `banned_users` (id, user_id, reason, banned_until)

**Tech Stack:**
- Node.js / Python
- PostgreSQL
- ML Model (spam detection)
- Redis (cache)

---

## ⚡ **SERVERLESS MICRO FUNCTIONS**

### **Edge Functions (Cloudflare Workers / Lambda@Edge):**

1. **`/api/functions/image-resize`**
   - Auto-resize images on-the-fly
   - Cache at CDN edge

2. **`/api/functions/geo-redirect`**
   - Redirect based on user location
   - Localized content

3. **`/api/functions/rate-limit`**
   - API rate limiting
   - DDoS protection

4. **`/api/functions/seo-metadata`**
   - Dynamic OG image generation
   - Social media previews

### **Scheduled Functions (Cron Jobs):**

1. **`cleanupExpiredSessions()`**
   - Run: Every hour
   - Clean expired JWT tokens

2. **`sendSearchAlerts()`**
   - Run: Daily at 9 AM
   - Send matching listing alerts

3. **`generateDailyReports()`**
   - Run: Daily at midnight
   - Generate analytics reports

4. **`backupDatabase()`**
   - Run: Daily at 2 AM
   - Backup databases to S3

5. **`cleanupOldMedia()`**
   - Run: Weekly
   - Remove unused media files

---

## 🔌 **API ENDPOINTS**

### **Authentication Service** (`/api/auth`)

```
POST   /api/auth/register          - Register new user
POST   /api/auth/login             - Login user
POST   /api/auth/logout            - Logout user
POST   /api/auth/refresh           - Refresh access token
POST   /api/auth/verify-email      - Verify email
POST   /api/auth/forgot-password   - Request password reset
POST   /api/auth/reset-password    - Reset password
GET    /api/auth/session           - Get current session
```

### **User Profile Service** (`/api/users`)

```
GET    /api/users/:id              - Get user profile
PUT    /api/users/:id              - Update profile
POST   /api/users/:id/avatar       - Upload avatar
GET    /api/users/:id/seller       - Get seller profile
PUT    /api/users/:id/seller       - Update seller info
GET    /api/users/:id/preferences  - Get preferences
PUT    /api/users/:id/preferences  - Update preferences
DELETE /api/users/:id              - Delete account
```

### **Listing Service** (`/api/listings`)

```
GET    /api/listings               - Get all listings
GET    /api/listings/:id           - Get single listing
POST   /api/listings               - Create listing
PUT    /api/listings/:id           - Update listing
DELETE /api/listings/:id           - Delete listing
GET    /api/listings/search        - Search listings
GET    /api/listings/featured      - Get featured listings
POST   /api/listings/:id/images    - Upload images
PUT    /api/listings/:id/images    - Reorder images
DELETE /api/listings/:id/images/:imageId - Delete image
POST   /api/listings/:id/publish   - Publish listing
POST   /api/listings/:id/sold      - Mark as sold
```

### **Favorites Service** (`/api/favorites`)

```
GET    /api/favorites              - Get user favorites
POST   /api/favorites/:listingId   - Add to favorites
DELETE /api/favorites/:listingId   - Remove from favorites
DELETE /api/favorites              - Clear all favorites
POST   /api/favorites/share        - Share wishlist
GET    /api/favorites/shared/:token - Get shared wishlist
```

### **Messaging Service** (`/api/messages`)

```
GET    /api/messages/conversations - Get conversations
GET    /api/messages/:conversationId - Get messages
POST   /api/messages/:conversationId - Send message
PUT    /api/messages/:messageId/read - Mark as read
DELETE /api/messages/:conversationId - Delete conversation
GET    /api/messages/unread        - Get unread count
```

### **Review Service** (`/api/reviews`)

```
GET    /api/reviews/seller/:sellerId - Get seller reviews
GET    /api/reviews/listing/:listingId - Get listing reviews
POST   /api/reviews                - Create review
PUT    /api/reviews/:id            - Update review
DELETE /api/reviews/:id            - Delete review
POST   /api/reviews/:id/response   - Seller response
POST   /api/reviews/:id/vote       - Vote helpful
POST   /api/reviews/:id/report     - Report review
```

### **Search Service** (`/api/search`)

```
GET    /api/search                 - Search listings
GET    /api/search/suggestions     - Auto-suggestions
GET    /api/search/saved           - Get saved searches
POST   /api/search/saved           - Save search
DELETE /api/search/saved/:id       - Delete saved search
GET    /api/search/alerts          - Get search alerts
POST   /api/search/alerts          - Create alert
PUT    /api/search/alerts/:id      - Update alert
DELETE /api/search/alerts/:id      - Delete alert
```

### **Comparison Service** (`/api/compare`)

```
GET    /api/compare                - Get comparison list
POST   /api/compare/:listingId     - Add to comparison
DELETE /api/compare/:listingId     - Remove from comparison
DELETE /api/compare                - Clear comparison
POST   /api/compare/share          - Share comparison
GET    /api/compare/shared/:token  - Get shared comparison
```

---

## 💾 **DATABASE SCHEMA**

### **PostgreSQL Tables:**

```sql
-- Users & Authentication
users (id, email, password_hash, role, status, created_at, updated_at)
sessions (id, user_id, token, expires_at)
verification_tokens (id, user_id, token, type, expires_at)

-- User Profiles
user_profiles (id, user_id, full_name, phone, location, bio, avatar_url)
seller_profiles (id, user_id, business_name, rating, review_count, verified)
user_preferences (id, user_id, email_notifications, sms_notifications, language)

-- Listings
listings (id, user_id, title, description, price, year, make, model, mileage, 
          fuel_type, transmission, location, condition, status, views, created_at)
listing_images (id, listing_id, url, thumbnail_url, order, is_primary)
listing_features (id, listing_id, feature_name)

-- Favorites
favorites (id, user_id, listing_id, created_at)

-- Messages
conversations (id, listing_id, buyer_id, seller_id, last_message_at)
messages (id, conversation_id, sender_id, content, read_at, created_at)

-- Reviews
reviews (id, listing_id, seller_id, reviewer_id, rating, title, content, 
         verified_buyer, created_at)
review_responses (id, review_id, seller_id, content, created_at)
review_votes (id, review_id, user_id, vote_type)

-- Search & Alerts
saved_searches (id, user_id, name, filters, created_at)
search_alerts (id, saved_search_id, frequency, last_sent_at, active)

-- Analytics
listing_analytics (id, listing_id, date, views, favorites, messages)
user_actions (id, user_id, action_type, metadata, timestamp)

-- Payments
payments (id, user_id, amount, currency, status, payment_method, created_at)
subscriptions (id, user_id, plan_id, status, starts_at, ends_at)
```

### **Redis Cache Structure:**

```
// User sessions
session:{token} -> user_id, role, expires_at

// Listing cache
listing:{id} -> listing_data (TTL: 5 min)

// Search results
search:{hash} -> results (TTL: 2 min)

// Favorites quick access
favorites:{user_id} -> [listing_ids]

// Comparison list
compare:{user_id} -> [listing_ids]

// Rate limiting
rate_limit:{ip}:{endpoint} -> count (TTL: 1 min)
```

---

## 🔄 **SERVICE COMMUNICATION**

### **1. Synchronous Communication (REST API):**
- Frontend → API Gateway → Microservices
- HTTP/HTTPS
- JSON payload

### **2. Asynchronous Communication (Message Queue):**
- RabbitMQ / AWS SQS
- Event-driven architecture

**Events:**
```
listing.created -> Trigger search indexing
listing.updated -> Update search index
user.registered -> Send welcome email
message.sent -> Send push notification
payment.completed -> Activate featured listing
review.created -> Update seller rating
```

### **3. Real-time Communication (WebSocket):**
- Socket.io / AWS AppSync
- Real-time messaging
- Live notifications

---

## 🛠️ **TECHNOLOGY STACK**

### **Frontend:**
- ✅ React + TypeScript
- ✅ TailwindCSS
- ✅ React Router
- ✅ React Query (data fetching)

### **Backend Services:**
- Node.js + Express (main services)
- Go (high-performance services)
- Python (ML/data processing)

### **Databases:**
- PostgreSQL (primary database)
- MongoDB (messages, logs)
- Redis (cache, sessions)
- Elasticsearch (search)

### **Message Queue:**
- RabbitMQ / AWS SQS
- Apache Kafka (analytics)

### **Storage:**
- AWS S3 (images, files)
- CloudFront CDN

### **Authentication:**
- JWT tokens
- OAuth 2.0 (social login)
- bcrypt (password hashing)

### **DevOps:**
- Docker + Kubernetes
- GitHub Actions (CI/CD)
- AWS / Google Cloud / Azure
- Nginx (reverse proxy)
- Prometheus + Grafana (monitoring)

---

## 🚀 **DEPLOYMENT STRATEGY**

### **Infrastructure:**

```
┌─────────────────────────────────────────────┐
│           Load Balancer (Nginx)             │
└─────────────────────────────────────────────┘
                    │
        ┌───────────┼───────────┐
        │           │           │
  ┌─────▼────┐ ┌───▼────┐ ┌───▼────┐
  │ Instance │ │Instance│ │Instance│
  │    1     │ │   2    │ │   3    │
  └──────────┘ └────────┘ └────────┘
```

### **Deployment Options:**

**Option 1: AWS**
- ECS/EKS (containers)
- RDS (PostgreSQL)
- ElastiCache (Redis)
- S3 + CloudFront
- API Gateway
- Lambda (serverless)

**Option 2: Google Cloud**
- Cloud Run (containers)
- Cloud SQL (PostgreSQL)
- Memorystore (Redis)
- Cloud Storage + CDN
- Cloud Functions

**Option 3: Kubernetes (Self-hosted)**
- K8s cluster
- Helm charts
- Istio (service mesh)
- Self-managed databases

### **CI/CD Pipeline:**

```
1. Code Push → GitHub
2. Run Tests → GitHub Actions
3. Build Docker Images
4. Push to Container Registry
5. Deploy to Staging
6. Run E2E Tests
7. Deploy to Production
8. Monitor & Alert
```

---

## 📊 **SCALABILITY CONSIDERATIONS**

### **Horizontal Scaling:**
- Load balancer distributes traffic
- Multiple instances per service
- Auto-scaling based on CPU/memory

### **Database Scaling:**
- Read replicas (PostgreSQL)
- Database sharding (if needed)
- Connection pooling

### **Caching Strategy:**
- Redis cache layer
- CDN for static assets
- API response caching

### **Performance Optimization:**
- Database indexing
- Query optimization
- Lazy loading
- Pagination

---

## 🔒 **SECURITY CONSIDERATIONS**

- JWT authentication
- HTTPS only
- Rate limiting
- Input validation
- SQL injection prevention
- XSS protection
- CSRF tokens
- API key management
- Role-based access control (RBAC)
- Data encryption at rest
- Regular security audits

---

## 📈 **MONITORING & LOGGING**

- Application metrics (Prometheus)
- Logging (ELK Stack)
- Error tracking (Sentry)
- Uptime monitoring (Pingdom)
- APM (New Relic / Datadog)
- Custom dashboards (Grafana)

---

## ✅ **IMPLEMENTATION PHASES**

### **Phase 1: Core Services (MVP)**
1. Authentication Service
2. User Profile Service
3. Listing Service
4. Basic Search

### **Phase 2: Engagement Features**
5. Favorites Service
6. Messaging Service
7. Advanced Search

### **Phase 3: Trust & Growth**
8. Review Service
9. Payment Service
10. Analytics Service

### **Phase 4: Optimization**
11. Media Service
12. Notification Service
13. Moderation Service

---

**🎯 Ready to build a scalable backend!**

