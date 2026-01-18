# Reviews Service

> Manages seller reviews, ratings, and feedback system.

---

## Overview

The Reviews Service handles the rating and review system for sellers. Buyers can leave reviews for sellers after interacting with them, and sellers can respond to reviews. The service also supports helpful votes and review photos.

**Responsibilities:**
- Create, update, delete reviews
- Calculate seller average ratings
- Rating distribution statistics
- Helpful vote system
- Seller response handling
- Review photo attachments
- Prevent duplicate reviews

---

## Database Tables

### 1. Reviews Table

```sql
CREATE TABLE reviews (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    listing_id          UUID REFERENCES car_listings(id) ON DELETE SET NULL,
    seller_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    buyer_id            UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Review Content
    rating              INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    title               VARCHAR(255),
    comment             TEXT,
    
    -- Metadata
    verified_purchase   BOOLEAN DEFAULT FALSE,
    helpful_count       INT DEFAULT 0,
    
    -- Seller Response
    seller_response     TEXT,
    seller_response_at  TIMESTAMP,
    
    -- Timestamps
    created_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Prevent duplicate reviews from same buyer for same listing
    UNIQUE(listing_id, buyer_id)
);

-- Indexes
CREATE INDEX idx_reviews_seller_id ON reviews(seller_id);
CREATE INDEX idx_reviews_buyer_id ON reviews(buyer_id);
CREATE INDEX idx_reviews_listing_id ON reviews(listing_id);
CREATE INDEX idx_reviews_rating ON reviews(seller_id, rating);
CREATE INDEX idx_reviews_created_at ON reviews(seller_id, created_at DESC);
CREATE INDEX idx_reviews_helpful ON reviews(helpful_count DESC);
```

### 2. Review Helpful Votes Table

```sql
CREATE TABLE review_helpful_votes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    review_id       UUID NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- One vote per user per review
    UNIQUE(review_id, user_id)
);

CREATE INDEX idx_review_votes_review_id ON review_helpful_votes(review_id);
CREATE INDEX idx_review_votes_user_id ON review_helpful_votes(user_id);
```

### 3. Review Photos Table

```sql
CREATE TABLE review_photos (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    review_id       UUID NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    photo_url       TEXT NOT NULL,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_review_photos_review_id ON review_photos(review_id);
```

---

## Field Descriptions

### Reviews

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Review identifier |
| `listing_id` | UUID | Related listing (nullable if deleted) |
| `seller_id` | UUID | The reviewed seller |
| `buyer_id` | UUID | The review author |
| `rating` | INT | Star rating (1-5) |
| `title` | VARCHAR(255) | Review title/headline |
| `comment` | TEXT | Detailed review text |
| `verified_purchase` | BOOLEAN | Whether buyer completed a transaction |
| `helpful_count` | INT | Number of helpful votes |
| `seller_response` | TEXT | Seller's reply to the review |
| `seller_response_at` | TIMESTAMP | When seller responded |
| `created_at` | TIMESTAMP | When review was submitted |
| `updated_at` | TIMESTAMP | Last update time |

---

## API Endpoints

### Reviews

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| `GET` | `/api/reviews/listing/:listingId` | Get reviews for listing | No |
| `GET` | `/api/reviews/seller/:sellerId` | Get reviews for seller | No |
| `GET` | `/api/reviews/seller/:sellerId/stats` | Get seller rating stats | No |
| `POST` | `/api/reviews` | Create review | Yes |
| `PUT` | `/api/reviews/:id` | Update review | Yes (Author) |
| `DELETE` | `/api/reviews/:id` | Delete review | Yes (Author/Admin) |

### Interactions

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| `POST` | `/api/reviews/:id/helpful` | Toggle helpful vote | Yes |
| `POST` | `/api/reviews/:id/response` | Add seller response | Yes (Seller) |
| `DELETE` | `/api/reviews/:id/response` | Remove seller response | Yes (Seller) |

### Photos

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| `POST` | `/api/reviews/:id/photos` | Upload review photos | Yes (Author) |
| `DELETE` | `/api/reviews/:id/photos/:photoId` | Delete photo | Yes (Author) |

---

## Request/Response Examples

### GET /api/reviews/seller/:sellerId

Get all reviews for a seller.

**Query Parameters:**
```
?page=1
&limit=10
&sortBy=recent|helpful|rating_high|rating_low
&rating=5  (filter by specific rating)
```

**Response:**
```json
{
  "success": true,
  "data": {
    "reviews": [
      {
        "id": "review-uuid-1",
        "rating": 5,
        "title": "Excellent experience!",
        "comment": "Very professional dealer. The car was exactly as described and the process was smooth.",
        "verifiedPurchase": true,
        "helpfulCount": 12,
        "hasVotedHelpful": false,
        "buyer": {
          "id": "buyer-uuid",
          "name": "John D.",
          "avatar": "https://storage.exotics.lk/avatars/john.jpg"
        },
        "listing": {
          "id": "listing-uuid",
          "title": "Mercedes-Benz S-Class S500"
        },
        "photos": [
          "https://storage.exotics.lk/reviews/photo1.jpg",
          "https://storage.exotics.lk/reviews/photo2.jpg"
        ],
        "sellerResponse": {
          "comment": "Thank you for your kind words! It was a pleasure doing business with you.",
          "respondedAt": "2024-01-16T14:00:00Z"
        },
        "createdAt": "2024-01-15T10:30:00Z",
        "updatedAt": "2024-01-15T10:30:00Z"
      },
      {
        "id": "review-uuid-2",
        "rating": 4,
        "title": "Good overall",
        "comment": "Nice car, but delivery took longer than expected.",
        "verifiedPurchase": false,
        "helpfulCount": 3,
        "hasVotedHelpful": true,
        "buyer": {
          "id": "buyer-uuid-2",
          "name": "Sarah M.",
          "avatar": null
        },
        "listing": null,
        "photos": [],
        "sellerResponse": null,
        "createdAt": "2024-01-10T08:15:00Z",
        "updatedAt": "2024-01-10T08:15:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 24,
      "totalPages": 3
    }
  }
}
```

### GET /api/reviews/seller/:sellerId/stats

Get rating statistics for a seller.

**Response:**
```json
{
  "success": true,
  "data": {
    "averageRating": 4.8,
    "totalReviews": 24,
    "verifiedReviews": 18,
    "distribution": {
      "5": 18,
      "4": 4,
      "3": 1,
      "2": 1,
      "1": 0
    },
    "percentages": {
      "5": 75,
      "4": 16.67,
      "3": 4.17,
      "2": 4.17,
      "1": 0
    },
    "recentTrend": {
      "last30Days": {
        "average": 4.9,
        "count": 5
      },
      "last90Days": {
        "average": 4.7,
        "count": 12
      }
    }
  }
}
```

### POST /api/reviews

Create a new review.

**Request Body:**
```json
{
  "listingId": "listing-uuid",
  "sellerId": "seller-uuid",
  "rating": 5,
  "title": "Excellent experience!",
  "comment": "Very professional dealer. The car was exactly as described and the process was smooth. Would highly recommend!"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Review submitted successfully",
  "data": {
    "id": "review-uuid-new",
    "rating": 5,
    "title": "Excellent experience!",
    "createdAt": "2024-01-16T10:00:00Z"
  }
}
```

### PUT /api/reviews/:id

Update an existing review.

**Request Body:**
```json
{
  "rating": 4,
  "title": "Updated: Good experience",
  "comment": "Updated my review after follow-up service..."
}
```

**Response:**
```json
{
  "success": true,
  "message": "Review updated",
  "data": {
    "id": "review-uuid",
    "rating": 4,
    "updatedAt": "2024-01-17T08:00:00Z"
  }
}
```

### POST /api/reviews/:id/helpful

Toggle helpful vote on a review.

**Response (Vote Added):**
```json
{
  "success": true,
  "message": "Marked as helpful",
  "data": {
    "helpfulCount": 13,
    "hasVoted": true
  }
}
```

**Response (Vote Removed):**
```json
{
  "success": true,
  "message": "Helpful vote removed",
  "data": {
    "helpfulCount": 12,
    "hasVoted": false
  }
}
```

### POST /api/reviews/:id/response

Add seller response to a review.

**Request Body:**
```json
{
  "comment": "Thank you for your feedback. We appreciate your business and are glad you had a positive experience!"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Response added",
  "data": {
    "sellerResponse": "Thank you for your feedback...",
    "respondedAt": "2024-01-16T14:00:00Z"
  }
}
```

### POST /api/reviews/:id/photos

Upload photos to a review.

**Request:** Multipart form data with images

**Response:**
```json
{
  "success": true,
  "message": "Photos uploaded",
  "data": {
    "photos": [
      {
        "id": "photo-uuid-1",
        "url": "https://storage.exotics.lk/reviews/photo1.jpg"
      },
      {
        "id": "photo-uuid-2",
        "url": "https://storage.exotics.lk/reviews/photo2.jpg"
      }
    ]
  }
}
```

---

## Business Logic

### Creating a Review

```javascript
async function createReview(buyerId, data) {
  const { listingId, sellerId, rating, title, comment } = data;
  
  // Verify seller exists
  const seller = await db.query(
    'SELECT id, role FROM users WHERE id = $1',
    [sellerId]
  );
  
  if (!seller.rows[0]) {
    throw new NotFoundError('Seller not found');
  }
  
  // Cannot review yourself
  if (buyerId === sellerId) {
    throw new BadRequestError('You cannot review yourself');
  }
  
  // Check for existing review on this listing
  if (listingId) {
    const existing = await db.query(
      'SELECT id FROM reviews WHERE listing_id = $1 AND buyer_id = $2',
      [listingId, buyerId]
    );
    
    if (existing.rows[0]) {
      throw new ConflictError('You have already reviewed this listing');
    }
  }
  
  // Check if buyer has interacted with seller (verified purchase)
  const hasInteraction = await db.query(`
    SELECT 1 FROM conversations 
    WHERE buyer_id = $1 AND seller_id = $2
    LIMIT 1
  `, [buyerId, sellerId]);
  
  const verifiedPurchase = hasInteraction.rows.length > 0;
  
  // Create review
  const review = await db.query(`
    INSERT INTO reviews (listing_id, seller_id, buyer_id, rating, title, comment, verified_purchase)
    VALUES ($1, $2, $3, $4, $5, $6, $7)
    RETURNING *
  `, [listingId, sellerId, buyerId, rating, title, comment, verifiedPurchase]);
  
  // Notify seller
  await notificationService.sendNewReviewNotification(sellerId, review.rows[0]);
  
  return review.rows[0];
}
```

### Calculating Seller Stats

```javascript
async function getSellerStats(sellerId) {
  const stats = await db.query(`
    SELECT 
      AVG(rating)::NUMERIC(3,2) as average_rating,
      COUNT(*) as total_reviews,
      COUNT(*) FILTER (WHERE verified_purchase = TRUE) as verified_reviews,
      COUNT(*) FILTER (WHERE rating = 5) as five_star,
      COUNT(*) FILTER (WHERE rating = 4) as four_star,
      COUNT(*) FILTER (WHERE rating = 3) as three_star,
      COUNT(*) FILTER (WHERE rating = 2) as two_star,
      COUNT(*) FILTER (WHERE rating = 1) as one_star
    FROM reviews
    WHERE seller_id = $1
  `, [sellerId]);
  
  const result = stats.rows[0];
  const total = parseInt(result.total_reviews);
  
  // Calculate percentages
  const distribution = {
    5: parseInt(result.five_star),
    4: parseInt(result.four_star),
    3: parseInt(result.three_star),
    2: parseInt(result.two_star),
    1: parseInt(result.one_star)
  };
  
  const percentages = {};
  for (const [stars, count] of Object.entries(distribution)) {
    percentages[stars] = total > 0 ? ((count / total) * 100).toFixed(2) : 0;
  }
  
  // Get recent trends
  const recentTrend = await db.query(`
    SELECT 
      AVG(rating) FILTER (WHERE created_at > NOW() - INTERVAL '30 days')::NUMERIC(3,2) as avg_30d,
      COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '30 days') as count_30d,
      AVG(rating) FILTER (WHERE created_at > NOW() - INTERVAL '90 days')::NUMERIC(3,2) as avg_90d,
      COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '90 days') as count_90d
    FROM reviews
    WHERE seller_id = $1
  `, [sellerId]);
  
  return {
    averageRating: parseFloat(result.average_rating) || 0,
    totalReviews: total,
    verifiedReviews: parseInt(result.verified_reviews),
    distribution,
    percentages,
    recentTrend: {
      last30Days: {
        average: parseFloat(recentTrend.rows[0].avg_30d) || 0,
        count: parseInt(recentTrend.rows[0].count_30d)
      },
      last90Days: {
        average: parseFloat(recentTrend.rows[0].avg_90d) || 0,
        count: parseInt(recentTrend.rows[0].count_90d)
      }
    }
  };
}
```

### Toggle Helpful Vote

```javascript
async function toggleHelpful(reviewId, userId) {
  // Check if user has already voted
  const existing = await db.query(
    'SELECT id FROM review_helpful_votes WHERE review_id = $1 AND user_id = $2',
    [reviewId, userId]
  );
  
  let hasVoted;
  
  if (existing.rows[0]) {
    // Remove vote
    await db.query(
      'DELETE FROM review_helpful_votes WHERE review_id = $1 AND user_id = $2',
      [reviewId, userId]
    );
    
    await db.query(
      'UPDATE reviews SET helpful_count = GREATEST(helpful_count - 1, 0) WHERE id = $1',
      [reviewId]
    );
    
    hasVoted = false;
  } else {
    // Add vote
    await db.query(
      'INSERT INTO review_helpful_votes (review_id, user_id) VALUES ($1, $2)',
      [reviewId, userId]
    );
    
    await db.query(
      'UPDATE reviews SET helpful_count = helpful_count + 1 WHERE id = $1',
      [reviewId]
    );
    
    hasVoted = true;
  }
  
  // Get updated count
  const review = await db.query(
    'SELECT helpful_count FROM reviews WHERE id = $1',
    [reviewId]
  );
  
  return {
    helpfulCount: review.rows[0].helpful_count,
    hasVoted
  };
}
```

### Add Seller Response

```javascript
async function addSellerResponse(reviewId, sellerId, comment) {
  // Verify review exists and belongs to this seller
  const review = await db.query(
    'SELECT id, seller_id, seller_response FROM reviews WHERE id = $1',
    [reviewId]
  );
  
  if (!review.rows[0]) {
    throw new NotFoundError('Review not found');
  }
  
  if (review.rows[0].seller_id !== sellerId) {
    throw new ForbiddenError('You can only respond to your own reviews');
  }
  
  if (review.rows[0].seller_response) {
    throw new ConflictError('You have already responded to this review');
  }
  
  // Add response
  const updated = await db.query(`
    UPDATE reviews 
    SET seller_response = $1, seller_response_at = NOW(), updated_at = NOW()
    WHERE id = $2
    RETURNING seller_response, seller_response_at
  `, [comment, reviewId]);
  
  return updated.rows[0];
}
```

---

## Validation Rules

```javascript
const reviewValidation = {
  rating: {
    required: true,
    min: 1,
    max: 5,
    type: 'integer'
  },
  title: {
    maxLength: 255
  },
  comment: {
    maxLength: 2000
  },
  photos: {
    maxCount: 5,
    maxSizeBytes: 5 * 1024 * 1024 // 5MB per photo
  }
};

const responseValidation = {
  comment: {
    required: true,
    minLength: 10,
    maxLength: 1000
  }
};
```

---

## Error Responses

```json
// 404 - Review Not Found
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Review not found"
  }
}

// 409 - Already Reviewed
{
  "success": false,
  "error": {
    "code": "CONFLICT",
    "message": "You have already reviewed this listing"
  }
}

// 403 - Not Owner
{
  "success": false,
  "error": {
    "code": "FORBIDDEN",
    "message": "You can only edit your own reviews"
  }
}

// 400 - Self Review
{
  "success": false,
  "error": {
    "code": "BAD_REQUEST",
    "message": "You cannot review yourself"
  }
}

// 409 - Already Responded
{
  "success": false,
  "error": {
    "code": "CONFLICT",
    "message": "You have already responded to this review"
  }
}
```

---

## Related Services

- **User Service** - Provides buyer/seller information
- **Listings Service** - Provides listing context
- **Notification Service** - Sends notifications for new reviews and responses
- **Image Service** - Handles review photo uploads
- **Analytics Service** - Tracks review sentiment for dealers

