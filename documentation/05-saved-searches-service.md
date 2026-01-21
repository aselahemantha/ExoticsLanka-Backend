# Saved Searches Service

> Manages user search presets and alert notifications.

---

## Overview

The Saved Searches Service allows users to save their search criteria and receive notifications when new listings match their preferences. This helps users stay updated on new inventory without manually searching.

**Responsibilities:**
- Save search filter presets
- Track matching listings
- Check for new matches
- Send alert notifications
- Manage alert preferences

---

## Database Tables

### Saved Searches Table

```sql
CREATE TABLE saved_searches (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name                VARCHAR(100) NOT NULL,
    
    -- Search Filters (stored as JSON)
    filters             JSONB NOT NULL,
    
    -- Alert Settings
    alert_enabled       BOOLEAN DEFAULT TRUE,
    alert_frequency     VARCHAR(20) DEFAULT 'daily' 
                        CHECK (alert_frequency IN ('instant', 'daily', 'weekly')),
    
    -- Match Tracking
    last_checked        TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_notified       TIMESTAMP,
    new_matches_count   INT DEFAULT 0,
    total_matches       INT DEFAULT 0,
    
    -- Timestamps
    created_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_saved_searches_user_id ON saved_searches(user_id);
CREATE INDEX idx_saved_searches_alert ON saved_searches(alert_enabled, alert_frequency);
CREATE INDEX idx_saved_searches_last_checked ON saved_searches(last_checked);
```

---

## Field Descriptions

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Saved search identifier |
| `user_id` | UUID | Owner of the saved search |
| `name` | VARCHAR(100) | User-given name for the search |
| `filters` | JSONB | Search filter criteria |
| `alert_enabled` | BOOLEAN | Whether to send notifications |
| `alert_frequency` | VARCHAR(20) | instant, daily, or weekly |
| `last_checked` | TIMESTAMP | Last time matches were checked |
| `last_notified` | TIMESTAMP | Last time user was notified |
| `new_matches_count` | INT | New listings since last check |
| `total_matches` | INT | Total matching listings |
| `created_at` | TIMESTAMP | When search was saved |
| `updated_at` | TIMESTAMP | Last update time |

### Filters Schema

```typescript
interface SearchFilters {
  searchQuery: string;           // Free text search
  brands: string[];              // e.g., ["BMW", "Mercedes-Benz"]
  priceRange: [number, number];  // [min, max] in LKR
  yearRange: [number, number];   // [min, max] year
  mileageRange: [number, number];// [min, max] km
  fuelTypes: string[];           // e.g., ["Petrol", "Hybrid"]
  transmissions: string[];       // e.g., ["Automatic"]
  bodyTypes: string[];           // e.g., ["Sedan", "SUV"]
  locations: string[];           // e.g., ["Colombo", "Kandy"]
  condition: string[];           // e.g., ["Used - Excellent"]
}
```

**Example filters value:**
```json
{
  "searchQuery": "",
  "brands": ["BMW", "Mercedes-Benz"],
  "priceRange": [30000000, 80000000],
  "yearRange": [2020, 2024],
  "mileageRange": [0, 50000],
  "fuelTypes": ["Petrol", "Hybrid"],
  "transmissions": ["Automatic"],
  "bodyTypes": [],
  "locations": ["Colombo"],
  "condition": []
}
```

---

## API Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| `GET` | `/api/searches` | Get user's saved searches | Yes |
| `GET` | `/api/searches/:id` | Get single saved search | Yes |
| `POST` | `/api/searches` | Save new search | Yes |
| `PUT` | `/api/searches/:id` | Update saved search | Yes |
| `DELETE` | `/api/searches/:id` | Delete saved search | Yes |
| `POST` | `/api/searches/:id/check` | Check for new matches | Yes |
| `PUT` | `/api/searches/:id/alerts` | Update alert settings | Yes |
| `GET` | `/api/searches/new-matches` | Get total new matches | Yes |
| `POST` | `/api/searches/:id/run` | Run search and get results | Yes |

---

## Request/Response Examples

### GET /api/searches

Get all saved searches for the user.

**Response:**
```json
{
  "success": true,
  "data": {
    "searches": [
      {
        "id": "search-uuid-1",
        "name": "Luxury SUVs under 80M",
        "filters": {
          "searchQuery": "",
          "brands": ["Land Rover", "Mercedes-Benz"],
          "priceRange": [30000000, 80000000],
          "yearRange": [2020, 2024],
          "mileageRange": [0, 50000],
          "fuelTypes": [],
          "transmissions": ["Automatic"],
          "bodyTypes": ["SUV"],
          "locations": [],
          "condition": []
        },
        "alertEnabled": true,
        "alertFrequency": "daily",
        "lastChecked": "2024-01-16T08:00:00Z",
        "newMatchesCount": 3,
        "totalMatches": 12,
        "createdAt": "2024-01-10T10:00:00Z"
      },
      {
        "id": "search-uuid-2",
        "name": "Toyota Hybrids in Colombo",
        "filters": {
          "searchQuery": "",
          "brands": ["Toyota"],
          "priceRange": [5000000, 20000000],
          "yearRange": [2018, 2024],
          "mileageRange": [0, 100000],
          "fuelTypes": ["Hybrid"],
          "transmissions": [],
          "bodyTypes": [],
          "locations": ["Colombo"],
          "condition": []
        },
        "alertEnabled": true,
        "alertFrequency": "instant",
        "lastChecked": "2024-01-16T10:30:00Z",
        "newMatchesCount": 0,
        "totalMatches": 8,
        "createdAt": "2024-01-05T14:00:00Z"
      }
    ],
    "totalNewMatches": 3
  }
}
```

### GET /api/searches/:id

Get a single saved search with details.

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "search-uuid-1",
    "name": "Luxury SUVs under 80M",
    "filters": {
      "searchQuery": "",
      "brands": ["Land Rover", "Mercedes-Benz"],
      "priceRange": [30000000, 80000000],
      "yearRange": [2020, 2024],
      "mileageRange": [0, 50000],
      "fuelTypes": [],
      "transmissions": ["Automatic"],
      "bodyTypes": ["SUV"],
      "locations": [],
      "condition": []
    },
    "alertEnabled": true,
    "alertFrequency": "daily",
    "lastChecked": "2024-01-16T08:00:00Z",
    "lastNotified": "2024-01-15T08:00:00Z",
    "newMatchesCount": 3,
    "totalMatches": 12,
    "createdAt": "2024-01-10T10:00:00Z",
    "updatedAt": "2024-01-10T10:00:00Z"
  }
}
```

### POST /api/searches

Save a new search.

**Request Body:**
```json
{
  "name": "Luxury SUVs under 80M",
  "filters": {
    "searchQuery": "",
    "brands": ["Land Rover", "Mercedes-Benz"],
    "priceRange": [30000000, 80000000],
    "yearRange": [2020, 2024],
    "mileageRange": [0, 50000],
    "fuelTypes": [],
    "transmissions": ["Automatic"],
    "bodyTypes": ["SUV"],
    "locations": [],
    "condition": []
  },
  "alertEnabled": true,
  "alertFrequency": "daily"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Search saved successfully",
  "data": {
    "id": "search-uuid-new",
    "name": "Luxury SUVs under 80M",
    "totalMatches": 12,
    "createdAt": "2024-01-16T11:00:00Z"
  }
}
```

### PUT /api/searches/:id

Update a saved search.

**Request Body:**
```json
{
  "name": "Updated: Luxury SUVs",
  "filters": {
    "searchQuery": "",
    "brands": ["Land Rover", "Mercedes-Benz", "BMW"],
    "priceRange": [30000000, 100000000],
    "yearRange": [2020, 2024],
    "mileageRange": [0, 50000],
    "fuelTypes": [],
    "transmissions": ["Automatic"],
    "bodyTypes": ["SUV"],
    "locations": [],
    "condition": []
  }
}
```

**Response:**
```json
{
  "success": true,
  "message": "Search updated",
  "data": {
    "id": "search-uuid",
    "totalMatches": 18,
    "updatedAt": "2024-01-16T12:00:00Z"
  }
}
```

### PUT /api/searches/:id/alerts

Update alert settings only.

**Request Body:**
```json
{
  "alertEnabled": true,
  "alertFrequency": "weekly"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Alert settings updated"
}
```

### POST /api/searches/:id/check

Check for new matches and mark as checked.

**Response:**
```json
{
  "success": true,
  "data": {
    "newMatches": 3,
    "totalMatches": 15,
    "newListings": [
      {
        "id": "listing-uuid-1",
        "title": "Land Rover Defender P400",
        "price": 68000000,
        "createdAt": "2024-01-16T09:00:00Z"
      },
      {
        "id": "listing-uuid-2",
        "title": "Mercedes-Benz GLE 450",
        "price": 72000000,
        "createdAt": "2024-01-16T07:30:00Z"
      },
      {
        "id": "listing-uuid-3",
        "title": "Mercedes-Benz GLS 580",
        "price": 78000000,
        "createdAt": "2024-01-15T18:00:00Z"
      }
    ]
  }
}
```

### POST /api/searches/:id/run

Run the saved search and return full results.

**Query Parameters:**
```
?page=1
&limit=20
&sortBy=date_desc
```

**Response:**
```json
{
  "success": true,
  "data": {
    "listings": [
      {
        "id": "listing-uuid-1",
        "title": "Land Rover Defender P400",
        "make": "Land Rover",
        "model": "Defender P400",
        "year": 2022,
        "price": 68000000,
        "mileage": 25000,
        "location": "Kandy",
        "coverImage": "https://storage.exotics.lk/listings/cover.jpg",
        "healthScore": 82,
        "isNew": true,
        "createdAt": "2024-01-16T09:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 15,
      "totalPages": 1
    }
  }
}
```

### GET /api/searches/new-matches

Get total count of new matches across all saved searches.

**Response:**
```json
{
  "success": true,
  "data": {
    "totalNewMatches": 5,
    "bySearch": [
      { "id": "search-uuid-1", "name": "Luxury SUVs", "newMatches": 3 },
      { "id": "search-uuid-2", "name": "Toyota Hybrids", "newMatches": 2 }
    ]
  }
}
```

---

## Business Logic

### Saving a Search

```javascript
async function saveSearch(userId, data) {
  const { name, filters, alertEnabled = true, alertFrequency = 'daily' } = data;
  
  // Validate filters
  validateFilters(filters);
  
  // Count current matches
  const totalMatches = await countMatchingListings(filters);
  
  // Create saved search
  const search = await db.query(`
    INSERT INTO saved_searches (
      user_id, name, filters, alert_enabled, alert_frequency, 
      total_matches, last_checked
    )
    VALUES ($1, $2, $3, $4, $5, $6, NOW())
    RETURNING *
  `, [userId, name, JSON.stringify(filters), alertEnabled, alertFrequency, totalMatches]);
  
  return search.rows[0];
}
```

### Counting Matching Listings

```javascript
async function countMatchingListings(filters) {
  let query = `SELECT COUNT(*) FROM car_listings WHERE status = 'active'`;
  const params = [];
  let paramIndex = 1;
  
  // Search query (full-text search)
  if (filters.searchQuery) {
    query += ` AND to_tsvector('english', title || ' ' || make || ' ' || model) @@ plainto_tsquery($${paramIndex})`;
    params.push(filters.searchQuery);
    paramIndex++;
  }
  
  // Brands
  if (filters.brands?.length > 0) {
    query += ` AND make = ANY($${paramIndex})`;
    params.push(filters.brands);
    paramIndex++;
  }
  
  // Price range
  if (filters.priceRange) {
    query += ` AND price >= $${paramIndex} AND price <= $${paramIndex + 1}`;
    params.push(filters.priceRange[0], filters.priceRange[1]);
    paramIndex += 2;
  }
  
  // Year range
  if (filters.yearRange) {
    query += ` AND year >= $${paramIndex} AND year <= $${paramIndex + 1}`;
    params.push(filters.yearRange[0], filters.yearRange[1]);
    paramIndex += 2;
  }
  
  // Mileage range
  if (filters.mileageRange) {
    query += ` AND mileage >= $${paramIndex} AND mileage <= $${paramIndex + 1}`;
    params.push(filters.mileageRange[0], filters.mileageRange[1]);
    paramIndex += 2;
  }
  
  // Fuel types
  if (filters.fuelTypes?.length > 0) {
    query += ` AND fuel_type = ANY($${paramIndex})`;
    params.push(filters.fuelTypes);
    paramIndex++;
  }
  
  // Transmissions
  if (filters.transmissions?.length > 0) {
    query += ` AND transmission = ANY($${paramIndex})`;
    params.push(filters.transmissions);
    paramIndex++;
  }
  
  // Body types
  if (filters.bodyTypes?.length > 0) {
    query += ` AND body_type = ANY($${paramIndex})`;
    params.push(filters.bodyTypes);
    paramIndex++;
  }
  
  // Locations
  if (filters.locations?.length > 0) {
    query += ` AND location = ANY($${paramIndex})`;
    params.push(filters.locations);
    paramIndex++;
  }
  
  // Condition
  if (filters.condition?.length > 0) {
    query += ` AND condition = ANY($${paramIndex})`;
    params.push(filters.condition);
    paramIndex++;
  }
  
  const result = await db.query(query, params);
  return parseInt(result.rows[0].count);
}
```

### Checking for New Matches

```javascript
async function checkNewMatches(searchId, userId) {
  // Get saved search
  const search = await db.query(
    'SELECT * FROM saved_searches WHERE id = $1 AND user_id = $2',
    [searchId, userId]
  );
  
  if (!search.rows[0]) {
    throw new NotFoundError('Saved search not found');
  }
  
  const { filters, last_checked, total_matches } = search.rows[0];
  
  // Count current matches
  const currentMatches = await countMatchingListings(filters);
  
  // Get new listings since last check
  const newListings = await getNewListingsSince(filters, last_checked);
  
  // Update saved search
  await db.query(`
    UPDATE saved_searches 
    SET 
      last_checked = NOW(),
      new_matches_count = $1,
      total_matches = $2
    WHERE id = $3
  `, [newListings.length, currentMatches, searchId]);
  
  return {
    newMatches: newListings.length,
    totalMatches: currentMatches,
    newListings
  };
}

async function getNewListingsSince(filters, sinceDate) {
  // Build the same query as countMatchingListings but with date filter and returning listings
  let query = `
    SELECT id, title, make, model, year, price, mileage, location, created_at,
           (SELECT image_url FROM listing_images WHERE listing_id = cl.id AND is_cover = TRUE LIMIT 1) as cover_image
    FROM car_listings cl
    WHERE status = 'active' AND created_at > $1
  `;
  const params = [sinceDate];
  // ... add same filters as countMatchingListings
  
  query += ` ORDER BY created_at DESC LIMIT 10`;
  
  const result = await db.query(query, params);
  return result.rows;
}
```

---

## Background Jobs

### Check Alerts (Hourly)

```javascript
// Run every hour
async function checkSearchAlerts() {
  // Get searches due for checking based on frequency
  const searches = await db.query(`
    SELECT ss.*, u.email, u.name as user_name
    FROM saved_searches ss
    JOIN users u ON ss.user_id = u.id
    WHERE ss.alert_enabled = TRUE
      AND (
        (ss.alert_frequency = 'instant' AND ss.last_checked < NOW() - INTERVAL '1 hour')
        OR (ss.alert_frequency = 'daily' AND ss.last_checked < NOW() - INTERVAL '1 day')
        OR (ss.alert_frequency = 'weekly' AND ss.last_checked < NOW() - INTERVAL '7 days')
      )
  `);
  
  for (const search of searches.rows) {
    const newListings = await getNewListingsSince(search.filters, search.last_checked);
    
    if (newListings.length > 0) {
      // Update search
      await db.query(`
        UPDATE saved_searches 
        SET last_checked = NOW(), new_matches_count = $1, total_matches = total_matches + $1
        WHERE id = $2
      `, [newListings.length, search.id]);
      
      // Send notification
      await notificationService.sendSearchAlert({
        email: search.email,
        userName: search.user_name,
        searchName: search.name,
        newCount: newListings.length,
        listings: newListings
      });
    } else {
      // Just update last_checked
      await db.query(
        'UPDATE saved_searches SET last_checked = NOW() WHERE id = $1',
        [search.id]
      );
    }
  }
}
```

### Send Daily Digest (Daily at 8 AM)

```javascript
// Run at 8:00 AM daily
async function sendDailyDigest() {
  // Get users with daily alerts that have new matches
  const digests = await db.query(`
    SELECT 
      u.id as user_id,
      u.email,
      u.name,
      json_agg(json_build_object(
        'id', ss.id,
        'name', ss.name,
        'newMatches', ss.new_matches_count
      )) as searches
    FROM users u
    JOIN saved_searches ss ON ss.user_id = u.id
    WHERE ss.alert_enabled = TRUE
      AND ss.alert_frequency = 'daily'
      AND ss.new_matches_count > 0
    GROUP BY u.id, u.email, u.name
  `);
  
  for (const digest of digests.rows) {
    await notificationService.sendDailyDigest({
      email: digest.email,
      userName: digest.name,
      searches: digest.searches
    });
    
    // Reset new matches count
    await db.query(`
      UPDATE saved_searches 
      SET new_matches_count = 0, last_notified = NOW()
      WHERE user_id = $1 AND alert_frequency = 'daily'
    `, [digest.user_id]);
  }
}
```

---

## Validation Rules

```javascript
const savedSearchValidation = {
  name: {
    required: true,
    minLength: 1,
    maxLength: 100
  },
  filters: {
    required: true,
    type: 'object'
  },
  alertFrequency: {
    enum: ['instant', 'daily', 'weekly']
  }
};

const filtersValidation = {
  searchQuery: { maxLength: 200 },
  brands: { type: 'array', maxItems: 20 },
  priceRange: { type: 'array', length: 2, items: { type: 'number', min: 0 } },
  yearRange: { type: 'array', length: 2, items: { type: 'integer', min: 1900, max: 2030 } },
  mileageRange: { type: 'array', length: 2, items: { type: 'integer', min: 0 } },
  fuelTypes: { type: 'array', maxItems: 10 },
  transmissions: { type: 'array', maxItems: 5 },
  bodyTypes: { type: 'array', maxItems: 10 },
  locations: { type: 'array', maxItems: 20 },
  condition: { type: 'array', maxItems: 5 }
};
```

---

## Error Responses

```json
// 404 - Search Not Found
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Saved search not found"
  }
}

// 400 - Invalid Filters
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid filters",
    "details": [
      { "field": "priceRange", "message": "Price range must have exactly 2 values" }
    ]
  }
}

// 403 - Not Owner
{
  "success": false,
  "error": {
    "code": "FORBIDDEN",
    "message": "You can only access your own saved searches"
  }
}
```

---

## Related Services

- **Listings Service** - Provides listing data for matching
- **Notification Service** - Sends alert emails
- **User Service** - Provides user preferences

