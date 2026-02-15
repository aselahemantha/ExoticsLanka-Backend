# Exotics Lanka - Backend Services Documentation

> Complete backend service documentation for the Exotics Lanka luxury car marketplace.

---

## 📁 Service Documentation

| # | Service | File | Description |
|---|---------|------|-------------|
| 01 | **Listings Service** | [01-listings-service.md](./01-listings-service.md) | Car listings CRUD, search, filters, health scores |
| 02 | **Favorites Service** | [02-favorites-service.md](./02-favorites-service.md) | User wishlists, save/unsave vehicles |
| 03 | **Messaging Service** | [03-messaging-service.md](./03-messaging-service.md) | Buyer-seller conversations, real-time chat |
| 04 | **Reviews Service** | [04-reviews-service.md](./04-reviews-service.md) | Seller ratings, review management |
| 05 | **Saved Searches Service** | [05-saved-searches-service.md](./05-saved-searches-service.md) | Search presets, alert notifications |
| 06 | **Reports Service** | [06-reports-service.md](./06-reports-service.md) | Listing reports, content moderation |
| 07 | **Contact Service** | [07-contact-service.md](./07-contact-service.md) | Contact form, support tickets |
| 08 | **Analytics Service** | [08-analytics-service.md](./08-analytics-service.md) | Dealer dashboard, metrics, insights |
| 09 | **Comparison Service** | [09-comparison-service.md](./09-comparison-service.md) | Vehicle comparison lists |
| 10 | **Image Service** | [10-image-service.md](./10-image-service.md) | Image uploads, processing, storage |
| 11 | **Notification Service** | [11-notification-service.md](./11-notification-service.md) | Email, SMS, push notifications |

---

## 🏗️ Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         Frontend (React)                         │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      API Gateway / Express                       │
│                   (Authentication, Rate Limiting)                │
└─────────────────────────────────────────────────────────────────┘
                                │
        ┌───────────────────────┼───────────────────────┐
        ▼                       ▼                       ▼
┌───────────────┐     ┌───────────────┐     ┌───────────────┐
│   Listings    │     │   Messaging   │     │   Analytics   │
│   Service     │     │   Service     │     │   Service     │
└───────────────┘     └───────────────┘     └───────────────┘
        │                       │                       │
        └───────────────────────┼───────────────────────┘
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      PostgreSQL Database                         │
└─────────────────────────────────────────────────────────────────┘
                                │
        ┌───────────────────────┼───────────────────────┐
        ▼                       ▼                       ▼
┌───────────────┐     ┌───────────────┐     ┌───────────────┐
│   AWS S3 /    │     │    Redis      │     │   SendGrid    │
│   Cloudinary  │     │   (Cache)     │     │   (Email)     │
└───────────────┘     └───────────────┘     └───────────────┘
```

---

## 📊 Database Tables Summary

| Table | Service | Description |
|-------|---------|-------------|
| `users` | Auth | User accounts |
| `car_brands` | Listings | Brand catalog |
| `car_listings` | Listings | Vehicle listings |
| `listing_images` | Listings/Image | Listing photos |
| `listing_features` | Listings | Car features |
| `favorites` | Favorites | User wishlists |
| `conversations` | Messaging | Chat threads |
| `messages` | Messaging | Individual messages |
| `reviews` | Reviews | Seller reviews |
| `review_helpful_votes` | Reviews | Helpful votes |
| `review_photos` | Reviews | Review images |
| `saved_searches` | Saved Searches | Search presets |
| `listing_reports` | Reports | Moderation queue |
| `contact_inquiries` | Contact | Support tickets |
| `listing_views` | Analytics | View tracking |
| `dealer_analytics` | Analytics | Daily aggregates |
| `comparison_items` | Comparison | Compare lists |

---

## 🔌 API Endpoints Summary

### Public Endpoints (No Auth)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/listings` | GET | Browse listings |
| `/api/listings/:id` | GET | View listing |
| `/api/brands` | GET | Get brands |
| `/api/reviews/seller/:id` | GET | Seller reviews |
| `/api/contact` | POST | Contact form |

### Authenticated Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/listings` | POST | Create listing |
| `/api/favorites` | GET/POST/DELETE | Manage favorites |
| `/api/conversations` | GET/POST | Messaging |
| `/api/reviews` | POST | Write review |
| `/api/searches` | GET/POST | Saved searches |
| `/api/comparison` | GET/POST/DELETE | Compare vehicles |

### Dealer Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/analytics/overview` | GET | Dashboard stats |
| `/api/analytics/funnel` | GET | Conversion funnel |
| `/api/analytics/insights` | GET | AI insights |

### Admin Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/reports` | GET/PUT | Moderate reports |
| `/api/contact` | GET/PUT | Handle inquiries |
| `/api/brands` | POST/PUT/DELETE | Manage brands |

---

## 🔧 Technology Stack

| Layer | Technology |
|-------|------------|
| **Runtime** | Node.js 18+ |
| **Framework** | Express.js |
| **Database** | PostgreSQL 14+ |
| **Cache** | Redis |
| **File Storage** | AWS S3 / Cloudinary |
| **Email** | SendGrid |
| **SMS** | Twilio (optional) |
| **Queue** | Bull (Redis-based) |
| **Real-time** | Socket.io (optional) |

---

## 🚀 Getting Started

1. **Clone and install dependencies:**
   ```bash
   cd backend
   npm install
   ```

2. **Set up environment variables:**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Run database migrations:**
   ```bash
   npm run migrate
   ```

4. **Seed initial data:**
   ```bash
   npm run seed
   ```

5. **Start development server:**
   ```bash
   npm run dev
   ```

---

## 📝 Environment Variables

```env
# Database
DATABASE_URL=postgresql://user:pass@localhost:5432/exotics_lanka

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRES_IN=7d

# AWS S3
AWS_ACCESS_KEY=xxx
AWS_SECRET_KEY=xxx
AWS_REGION=ap-south-1
S3_BUCKET=exotics-lanka

# Email (SendGrid)
SENDGRID_API_KEY=xxx

# Redis
REDIS_URL=redis://localhost:6379

# App
PORT=3001
NODE_ENV=development
FRONTEND_URL=http://localhost:5173
```

---

## 📚 Additional Resources

- [API Documentation](../api/README.md) - Full API reference
- [Database Schema](../database/schema.sql) - Complete SQL schema
- [Deployment Guide](../deployment/README.md) - Production deployment

---

*Generated for Exotics Lanka Car Marketplace*

