# Listings Service

> Manages all car listing operations including CRUD, search, filtering, and metrics.

---

## Overview

The Listings Service is the core service of the Exotics Lanka platform, handling all vehicle listing operations from creation to deletion, including advanced search capabilities and performance tracking.

**Responsibilities:**
- Create, read, update, delete car listings
- Advanced search with multiple filters
- Pagination and sorting
- Featured/trending listing management
- Health score calculation
- View tracking and analytics

---

## Database Tables

### 1. Car Listings (Primary Table)

```sql
CREATE TABLE car_listings (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Basic Information
    title               VARCHAR(255) NOT NULL,
    make                VARCHAR(100) NOT NULL,
    model               VARCHAR(100) NOT NULL,
    year                INT NOT NULL CHECK (year >= 1900 AND year <= 2030),
    price               DECIMAL(15, 2) NOT NULL CHECK (price > 0),
    mileage             INT NOT NULL CHECK (mileage >= 0),
    condition           VARCHAR(50) NOT NULL,
    
    -- Specifications
    transmission        VARCHAR(50),
    fuel_type           VARCHAR(50),
    body_type           VARCHAR(50),
    color               VARCHAR(50),
    doors               INT CHECK (doors >= 1 AND doors <= 6),
    seats               INT CHECK (seats >= 1 AND seats <= 12),
    engine_size         VARCHAR(20),
    drivetrain          VARCHAR(20),
    
    -- Description & Contact
    description         TEXT,
    location            VARCHAR(255) NOT NULL,
    contact_phone       VARCHAR(20),
    contact_email       VARCHAR(255),
    
    -- Status & Metrics
    status              VARCHAR(20) DEFAULT 'pending' 
                        CHECK (status IN ('draft', 'pending', 'active', 'sold', 'expired', 'rejected')),
    health_score        INT DEFAULT 50 CHECK (health_score >= 0 AND health_score <= 100),
    views               INT DEFAULT 0,
    favorites_count     INT DEFAULT 0,
    days_listed         INT DEFAULT 0,
    
    -- Flags
    is_new              BOOLEAN DEFAULT FALSE,
    is_featured         BOOLEAN DEFAULT FALSE,
    is_verified         BOOLEAN DEFAULT FALSE,
    trending            BOOLEAN DEFAULT FALSE,
    
    -- Pricing Intelligence
    market_avg_price    DECIMAL(15, 2),
    price_alert         TEXT,
    
    -- Timestamps
    created_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    published_at        TIMESTAMP,
    expires_at          TIMESTAMP
);

-- Indexes
CREATE INDEX idx_listings_user_id ON car_listings(user_id);
CREATE INDEX idx_listings_status ON car_listings(status);
CREATE INDEX idx_listings_make ON car_listings(make);
CREATE INDEX idx_listings_year ON car_listings(year);
CREATE INDEX idx_listings_price ON car_listings(price);
CREATE INDEX idx_listings_location ON car_listings(location);
CREATE INDEX idx_listings_created_at ON car_listings(created_at DESC);
CREATE INDEX idx_listings_health_score ON car_listings(health_score DESC);

-- Full-text search
CREATE INDEX idx_listings_search ON car_listings 
USING GIN (to_tsvector('english', title || ' ' || make || ' ' || model || ' ' || COALESCE(description, '')));
```

### 2. Listing Images

```sql
CREATE TABLE listing_images (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    listing_id      UUID NOT NULL REFERENCES car_listings(id) ON DELETE CASCADE,
    image_url       TEXT NOT NULL,
    thumbnail_url   TEXT,
    is_cover        BOOLEAN DEFAULT FALSE,
    sort_order      INT DEFAULT 0,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_listing_images_listing_id ON listing_images(listing_id);
CREATE INDEX idx_listing_images_cover ON listing_images(listing_id, is_cover) WHERE is_cover = TRUE;
```

### 3. Listing Features

```sql
CREATE TABLE listing_features (
    id              SERIAL PRIMARY KEY,
    listing_id      UUID NOT NULL REFERENCES car_listings(id) ON DELETE CASCADE,
    feature_name    VARCHAR(100) NOT NULL,
    
    UNIQUE(listing_id, feature_name)
);

CREATE INDEX idx_listing_features_listing_id ON listing_features(listing_id);
```

### 4. Car Brands (Lookup Table)

```sql
CREATE TABLE car_brands (
    id              SERIAL PRIMARY KEY,
    name            VARCHAR(100) NOT NULL UNIQUE,
    logo_url        TEXT,
    listing_count   INT DEFAULT 0,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Sample Data
INSERT INTO car_brands (name) VALUES 
('Mercedes-Benz'), ('BMW'), ('Porsche'), ('Land Rover'), 
('Toyota'), ('Honda'), ('Lexus'), ('Audi'), ('Ferrari'), ('Lamborghini');
```

---

## Field Descriptions

### Car Listings Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | Auto | Unique identifier |
| `user_id` | UUID | Yes | Owner (seller/dealer) reference |
| `title` | VARCHAR(255) | Yes | Display title (e.g., "Mercedes-Benz S-Class S500") |
| `make` | VARCHAR(100) | Yes | Car manufacturer |
| `model` | VARCHAR(100) | Yes | Car model name |
| `year` | INT | Yes | Manufacturing year (1900-2030) |
| `price` | DECIMAL | Yes | Price in LKR |
| `mileage` | INT | Yes | Odometer reading in km |
| `condition` | VARCHAR(50) | Yes | Vehicle condition |
| `transmission` | VARCHAR(50) | No | Automatic, Manual, CVT, Semi-Automatic |
| `fuel_type` | VARCHAR(50) | No | Petrol, Diesel, Hybrid, Electric |
| `body_type` | VARCHAR(50) | No | Sedan, SUV, Coupe, Convertible, etc. |
| `color` | VARCHAR(50) | No | Exterior color |
| `doors` | INT | No | Number of doors (1-6) |
| `seats` | INT | No | Seating capacity (1-12) |
| `engine_size` | VARCHAR(20) | No | Engine displacement (e.g., "3.0L") |
| `drivetrain` | VARCHAR(20) | No | FWD, RWD, AWD, 4WD |
| `description` | TEXT | No | Detailed description |
| `location` | VARCHAR(255) | Yes | City/area location |
| `contact_phone` | VARCHAR(20) | No | Contact phone number |
| `contact_email` | VARCHAR(255) | No | Contact email |
| `status` | VARCHAR(20) | Auto | draft, pending, active, sold, expired, rejected |
| `health_score` | INT | Auto | Listing quality score (0-100) |
| `views` | INT | Auto | Total view count |
| `favorites_count` | INT | Auto | Number of favorites |
| `days_listed` | INT | Auto | Days since published |
| `is_new` | BOOLEAN | Auto | New arrival flag (< 7 days) |
| `is_featured` | BOOLEAN | Admin | Featured/promoted listing |
| `is_verified` | BOOLEAN | Admin | Verified by platform |
| `trending` | BOOLEAN | Auto | Currently trending |
| `market_avg_price` | DECIMAL | Auto | Market average for comparison |
| `price_alert` | TEXT | Auto | Price warning message |

### Condition Options

```javascript
const conditions = [
  "Brand New",
  "Used - Like New",
  "Used - Excellent", 
  "Used - Good",
  "Used - Fair"
];
```

### Available Features

```javascript
const availableFeatures = [
  "Leather Seats",
  "Sunroof",
  "Navigation System",
  "Parking Sensors",
  "Backup Camera",
  "Cruise Control",
  "Heated Seats",
  "Ventilated Seats",
  "Bluetooth",
  "Apple CarPlay",
  "Android Auto",
  "Premium Sound System",
  "Alloy Wheels",
  "Keyless Entry",
  "Push Start",
  "Lane Departure Warning",
  "Blind Spot Monitor",
  "Adaptive Cruise Control",
  "360° Camera",
  "Head-Up Display"
];
```

---

## API Endpoints

### Listings CRUD

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `GET` | `/api/listings` | Get all listings with filters | No |
| `GET` | `/api/listings/:id` | Get single listing by ID | No |
| `POST` | `/api/listings` | Create new listing | Yes (Seller/Dealer) |
| `PUT` | `/api/listings/:id` | Update listing | Yes (Owner) |
| `DELETE` | `/api/listings/:id` | Delete listing | Yes (Owner) |
| `PUT` | `/api/listings/:id/status` | Update listing status | Yes (Owner/Admin) |

### Special Queries

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `GET` | `/api/listings/featured` | Get featured listings | No |
| `GET` | `/api/listings/trending` | Get trending listings | No |
| `GET` | `/api/listings/user/:userId` | Get user's listings | No |
| `GET` | `/api/listings/similar/:id` | Get similar listings | No |
| `POST` | `/api/listings/:id/view` | Record a view | No |

### Brands

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `GET` | `/api/brands` | Get all brands with counts | No |
| `GET` | `/api/brands/:id` | Get brand details | No |
| `POST` | `/api/brands` | Create brand | Yes (Admin) |
| `PUT` | `/api/brands/:id` | Update brand | Yes (Admin) |
| `DELETE` | `/api/brands/:id` | Delete brand | Yes (Admin) |

---

## Request/Response Examples

### GET /api/listings

**Query Parameters:**
```
?search=BMW
&brands=Mercedes-Benz,BMW
&minPrice=10000000
&maxPrice=50000000
&minYear=2018
&maxYear=2024
&minMileage=0
&maxMileage=100000
&fuelTypes=Petrol,Hybrid
&transmissions=Automatic
&bodyTypes=Sedan,SUV
&locations=Colombo,Kandy
&conditions=Used - Excellent,Used - Like New
&status=active
&sortBy=price_asc
&page=1
&limit=20
```

**Sort Options:**
- `price_asc` - Price low to high
- `price_desc` - Price high to low
- `date_desc` - Newest first
- `date_asc` - Oldest first
- `views_desc` - Most viewed
- `health_desc` - Best health score

**Response:**
```json
{
  "success": true,
  "data": {
    "listings": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "title": "Mercedes-Benz S-Class S500",
        "make": "Mercedes-Benz",
        "model": "S-Class S500",
        "year": 2023,
        "price": 85000000,
        "mileage": 12000,
        "condition": "Used - Like New",
        "transmission": "Automatic",
        "fuelType": "Petrol",
        "bodyType": "Sedan",
        "color": "Midnight Black",
        "location": "Colombo 07",
        "status": "active",
        "healthScore": 94,
        "views": 1250,
        "favoritesCount": 84,
        "daysListed": 5,
        "isNew": true,
        "isVerified": true,
        "coverImage": "https://storage.exotics.lk/listings/cover.jpg",
        "imagesCount": 6,
        "user": {
          "id": "user-uuid",
          "name": "Premium Auto Gallery",
          "role": "dealer"
        },
        "createdAt": "2024-01-15T10:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 156,
      "totalPages": 8
    },
    "filters": {
      "brands": ["Mercedes-Benz", "BMW"],
      "priceRange": [10000000, 50000000]
    }
  }
}
```

### GET /api/listings/:id

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Mercedes-Benz S-Class S500",
    "make": "Mercedes-Benz",
    "model": "S-Class S500",
    "year": 2023,
    "price": 85000000,
    "mileage": 12000,
    "condition": "Used - Like New",
    "transmission": "Automatic",
    "fuelType": "Petrol",
    "bodyType": "Sedan",
    "color": "Midnight Black",
    "doors": 4,
    "seats": 5,
    "engineSize": "4.0L",
    "drivetrain": "RWD",
    "description": "This exceptional Mercedes-Benz S-Class represents...",
    "location": "Colombo 07",
    "contactPhone": "+94771234567",
    "contactEmail": "dealer@example.com",
    "status": "active",
    "healthScore": 94,
    "views": 1250,
    "favoritesCount": 84,
    "daysListed": 5,
    "isNew": true,
    "isFeatured": false,
    "isVerified": true,
    "trending": false,
    "marketAvgPrice": 82000000,
    "priceAlert": null,
    "images": [
      {
        "id": "img-uuid-1",
        "imageUrl": "https://storage.exotics.lk/listings/full/1.jpg",
        "thumbnailUrl": "https://storage.exotics.lk/listings/thumb/1.jpg",
        "isCover": true,
        "sortOrder": 0
      },
      {
        "id": "img-uuid-2",
        "imageUrl": "https://storage.exotics.lk/listings/full/2.jpg",
        "thumbnailUrl": "https://storage.exotics.lk/listings/thumb/2.jpg",
        "isCover": false,
        "sortOrder": 1
      }
    ],
    "features": [
      "Leather Seats",
      "Sunroof",
      "Navigation System",
      "Parking Sensors",
      "Premium Sound System"
    ],
    "user": {
      "id": "user-uuid",
      "name": "Premium Auto Gallery",
      "email": "dealer@example.com",
      "phone": "+94771234567",
      "role": "dealer",
      "verified": true,
      "avatarUrl": "https://storage.exotics.lk/avatars/dealer.jpg",
      "rating": 4.8,
      "reviewCount": 24
    },
    "createdAt": "2024-01-15T10:30:00Z",
    "updatedAt": "2024-01-16T08:15:00Z",
    "publishedAt": "2024-01-15T10:35:00Z"
  }
}
```

### POST /api/listings

**Request Body:**
```json
{
  "title": "Mercedes-Benz S-Class S500",
  "make": "Mercedes-Benz",
  "model": "S-Class S500",
  "year": 2023,
  "price": 85000000,
  "mileage": 12000,
  "condition": "Used - Like New",
  "transmission": "Automatic",
  "fuelType": "Petrol",
  "bodyType": "Sedan",
  "color": "Midnight Black",
  "doors": 4,
  "seats": 5,
  "engineSize": "4.0L",
  "drivetrain": "RWD",
  "description": "This exceptional Mercedes-Benz S-Class represents...",
  "location": "Colombo 07",
  "contactPhone": "+94771234567",
  "contactEmail": "dealer@example.com",
  "features": [
    "Leather Seats",
    "Sunroof",
    "Navigation System"
  ]
}
```

**Response:**
```json
{
  "success": true,
  "message": "Listing created successfully",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "pending"
  }
}
```

### PUT /api/listings/:id

**Request Body (Partial Update):**
```json
{
  "price": 82000000,
  "description": "Updated description...",
  "features": ["Leather Seats", "Sunroof", "Navigation System", "360° Camera"]
}
```

### PUT /api/listings/:id/status

**Request Body:**
```json
{
  "status": "sold"
}
```

### POST /api/listings/:id/view

**Request Body:**
```json
{
  "sessionId": "session-uuid",
  "referrer": "https://google.com"
}
```

---

## Business Logic

### Health Score Calculation

The health score (0-100) measures listing quality and affects search ranking.

```javascript
function calculateHealthScore(listing) {
  let score = 50; // Base score
  
  // Images (+20 max)
  const imageCount = listing.images?.length || 0;
  score += Math.min(imageCount * 4, 20);
  
  // Description length (+10 max)
  const descLength = listing.description?.length || 0;
  if (descLength > 200) score += 10;
  else if (descLength > 100) score += 5;
  
  // Features (+10 max)
  const featureCount = listing.features?.length || 0;
  score += Math.min(featureCount, 10);
  
  // Price competitiveness (+10)
  if (listing.marketAvgPrice && listing.price <= listing.marketAvgPrice) {
    score += 10;
  }
  
  // Completeness (+10)
  const optionalFields = ['transmission', 'fuelType', 'color', 'bodyType', 'doors', 'seats'];
  const filledCount = optionalFields.filter(f => listing[f]).length;
  score += Math.round((filledCount / optionalFields.length) * 10);
  
  return Math.min(score, 100);
}
```

### Trending Detection

Listings become trending based on engagement metrics:

```javascript
async function updateTrendingStatus() {
  // Get listings with high recent engagement
  const trending = await db.query(`
    WITH recent_stats AS (
      SELECT 
        listing_id,
        COUNT(*) as recent_views,
        COUNT(DISTINCT user_id) as unique_viewers
      FROM listing_views
      WHERE created_at > NOW() - INTERVAL '7 days'
      GROUP BY listing_id
    )
    SELECT cl.id
    FROM car_listings cl
    JOIN recent_stats rs ON cl.id = rs.listing_id
    WHERE cl.status = 'active'
      AND rs.recent_views > 100
      AND rs.unique_viewers > 50
    ORDER BY rs.recent_views DESC
    LIMIT 20
  `);
  
  // Update trending flags
  await db.query(`UPDATE car_listings SET trending = FALSE WHERE trending = TRUE`);
  await db.query(`UPDATE car_listings SET trending = TRUE WHERE id = ANY($1)`, [trending.map(t => t.id)]);
}
```

### New Arrival Flag

Listings are marked as "new" for the first 7 days:

```javascript
function isNewArrival(publishedAt) {
  const daysSincePublished = Math.floor(
    (Date.now() - new Date(publishedAt).getTime()) / (1000 * 60 * 60 * 24)
  );
  return daysSincePublished <= 7;
}
```

---

## Validation Rules

```javascript
const listingValidation = {
  title: {
    required: true,
    minLength: 10,
    maxLength: 255
  },
  make: {
    required: true,
    maxLength: 100
  },
  model: {
    required: true,
    maxLength: 100
  },
  year: {
    required: true,
    min: 1900,
    max: new Date().getFullYear() + 1
  },
  price: {
    required: true,
    min: 100000, // 100K LKR minimum
    max: 1000000000 // 1B LKR maximum
  },
  mileage: {
    required: true,
    min: 0,
    max: 1000000
  },
  condition: {
    required: true,
    enum: ['Brand New', 'Used - Like New', 'Used - Excellent', 'Used - Good', 'Used - Fair']
  },
  location: {
    required: true,
    maxLength: 255
  },
  description: {
    maxLength: 5000
  },
  images: {
    maxCount: 15,
    maxSizeBytes: 10 * 1024 * 1024 // 10MB per image
  }
};
```

---

## Background Jobs

### Update Days Listed (Daily)

```javascript
// Run at 00:00 daily
async function updateDaysListed() {
  await db.query(`
    UPDATE car_listings 
    SET days_listed = days_listed + 1,
        is_new = CASE WHEN days_listed < 7 THEN TRUE ELSE FALSE END
    WHERE status = 'active'
  `);
}
```

### Expire Listings (Daily)

```javascript
// Run at 00:00 daily
async function expireListings() {
  await db.query(`
    UPDATE car_listings 
    SET status = 'expired' 
    WHERE status = 'active' 
      AND expires_at IS NOT NULL 
      AND expires_at < NOW()
  `);
}
```

### Update Brand Counts (Hourly)

```javascript
// Run every hour
async function updateBrandCounts() {
  await db.query(`
    UPDATE car_brands b
    SET listing_count = (
      SELECT COUNT(*) 
      FROM car_listings cl 
      WHERE cl.make = b.name AND cl.status = 'active'
    )
  `);
}
```

### Recalculate Health Scores (Daily)

```javascript
// Run at 02:00 daily
async function recalculateHealthScores() {
  const listings = await db.query(`
    SELECT id, description, 
           (SELECT COUNT(*) FROM listing_images WHERE listing_id = cl.id) as image_count,
           (SELECT COUNT(*) FROM listing_features WHERE listing_id = cl.id) as feature_count,
           transmission, fuel_type, color, body_type, doors, seats,
           price, market_avg_price
    FROM car_listings cl
    WHERE status = 'active'
  `);
  
  for (const listing of listings.rows) {
    const score = calculateHealthScore(listing);
    await db.query(`UPDATE car_listings SET health_score = $1 WHERE id = $2`, [score, listing.id]);
  }
}
```

---

## Error Responses

```json
// 400 Bad Request
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": [
      { "field": "price", "message": "Price must be greater than 0" },
      { "field": "year", "message": "Year must be between 1900 and 2025" }
    ]
  }
}

// 404 Not Found
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Listing not found"
  }
}

// 403 Forbidden
{
  "success": false,
  "error": {
    "code": "FORBIDDEN",
    "message": "You don't have permission to edit this listing"
  }
}
```

---

## Related Services

- **Image Service** - Handles image uploads for listings
- **Favorites Service** - Manages user favorites (updates favorites_count)
- **Analytics Service** - Tracks listing performance metrics
- **Search Service** - Advanced search functionality

