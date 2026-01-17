# Favorites Service

> Manages user wishlists and saved vehicles.

---

## Overview

The Favorites Service allows users to save listings to their wishlist for later viewing. It tracks user preferences and helps with engagement metrics.

**Responsibilities:**
- Add/remove listings from favorites
- Retrieve user's favorite listings
- Update listing favorite counts
- Support bulk operations (clear all)
- Check if listing is favorited

---

## Database Tables

### Favorites Table

```sql
CREATE TABLE favorites (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    listing_id      UUID NOT NULL REFERENCES car_listings(id) ON DELETE CASCADE,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Prevent duplicate favorites
    UNIQUE(user_id, listing_id)
);

-- Indexes for efficient queries
CREATE INDEX idx_favorites_user_id ON favorites(user_id);
CREATE INDEX idx_favorites_listing_id ON favorites(listing_id);
CREATE INDEX idx_favorites_created_at ON favorites(user_id, created_at DESC);
```

---

## Field Descriptions

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Unique identifier |
| `user_id` | UUID | User who favorited the listing |
| `listing_id` | UUID | The favorited listing |
| `created_at` | TIMESTAMP | When the favorite was added |

---

## API Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `GET` | `/api/favorites` | Get user's favorites | Yes |
| `POST` | `/api/favorites/:listingId` | Add to favorites | Yes |
| `DELETE` | `/api/favorites/:listingId` | Remove from favorites | Yes |
| `DELETE` | `/api/favorites` | Clear all favorites | Yes |
| `GET` | `/api/favorites/check/:listingId` | Check if favorited | Yes |
| `GET` | `/api/favorites/count` | Get favorites count | Yes |

---

## Request/Response Examples

### GET /api/favorites

Get all favorited listings for the authenticated user.

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
    "favorites": [
      {
        "id": "fav-uuid-1",
        "createdAt": "2024-01-15T10:30:00Z",
        "listing": {
          "id": "listing-uuid-1",
          "title": "Mercedes-Benz S-Class S500",
          "make": "Mercedes-Benz",
          "model": "S-Class S500",
          "year": 2023,
          "price": 85000000,
          "mileage": 12000,
          "location": "Colombo 07",
          "status": "active",
          "coverImage": "https://storage.exotics.lk/listings/cover.jpg",
          "healthScore": 94,
          "views": 1250,
          "daysListed": 5
        }
      },
      {
        "id": "fav-uuid-2",
        "createdAt": "2024-01-14T15:45:00Z",
        "listing": {
          "id": "listing-uuid-2",
          "title": "BMW 7 Series 740i",
          "make": "BMW",
          "model": "7 Series 740i",
          "year": 2022,
          "price": 72000000,
          "mileage": 18000,
          "location": "Negombo",
          "status": "active",
          "coverImage": "https://storage.exotics.lk/listings/cover2.jpg",
          "healthScore": 88,
          "views": 980,
          "daysListed": 12
        }
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 8,
      "totalPages": 1
    }
  }
}
```

### POST /api/favorites/:listingId

Add a listing to favorites.

**URL Parameter:**
- `listingId` - UUID of the listing to favorite

**Response (Success):**
```json
{
  "success": true,
  "message": "Added to favorites",
  "data": {
    "id": "fav-uuid-new",
    "listingId": "listing-uuid",
    "createdAt": "2024-01-16T10:00:00Z"
  }
}
```

**Response (Already Favorited):**
```json
{
  "success": false,
  "error": {
    "code": "ALREADY_EXISTS",
    "message": "Listing is already in favorites"
  }
}
```

### DELETE /api/favorites/:listingId

Remove a listing from favorites.

**URL Parameter:**
- `listingId` - UUID of the listing to unfavorite

**Response:**
```json
{
  "success": true,
  "message": "Removed from favorites"
}
```

### DELETE /api/favorites

Clear all favorites for the user.

**Response:**
```json
{
  "success": true,
  "message": "All favorites cleared",
  "data": {
    "removedCount": 8
  }
}
```

### GET /api/favorites/check/:listingId

Check if a specific listing is in user's favorites.

**Response:**
```json
{
  "success": true,
  "data": {
    "isFavorited": true,
    "favoritedAt": "2024-01-15T10:30:00Z"
  }
}
```

### GET /api/favorites/count

Get the total count of user's favorites.

**Response:**
```json
{
  "success": true,
  "data": {
    "count": 8
  }
}
```

---

## Business Logic

### Adding to Favorites

```javascript
async function addToFavorites(userId, listingId) {
  // Check if listing exists and is active
  const listing = await db.query(
    'SELECT id, status FROM car_listings WHERE id = $1',
    [listingId]
  );
  
  if (!listing.rows[0]) {
    throw new NotFoundError('Listing not found');
  }
  
  if (listing.rows[0].status !== 'active') {
    throw new BadRequestError('Cannot favorite inactive listing');
  }
  
  // Check if already favorited
  const existing = await db.query(
    'SELECT id FROM favorites WHERE user_id = $1 AND listing_id = $2',
    [userId, listingId]
  );
  
  if (existing.rows[0]) {
    throw new ConflictError('Listing is already in favorites');
  }
  
  // Add to favorites
  const favorite = await db.query(
    'INSERT INTO favorites (user_id, listing_id) VALUES ($1, $2) RETURNING *',
    [userId, listingId]
  );
  
  // Update listing favorites count
  await db.query(
    'UPDATE car_listings SET favorites_count = favorites_count + 1 WHERE id = $1',
    [listingId]
  );
  
  return favorite.rows[0];
}
```

### Removing from Favorites

```javascript
async function removeFromFavorites(userId, listingId) {
  const result = await db.query(
    'DELETE FROM favorites WHERE user_id = $1 AND listing_id = $2 RETURNING listing_id',
    [userId, listingId]
  );
  
  if (result.rowCount === 0) {
    throw new NotFoundError('Favorite not found');
  }
  
  // Update listing favorites count
  await db.query(
    'UPDATE car_listings SET favorites_count = GREATEST(favorites_count - 1, 0) WHERE id = $1',
    [listingId]
  );
  
  return true;
}
```

### Clearing All Favorites

```javascript
async function clearAllFavorites(userId) {
  // Get all listing IDs before deleting
  const favorites = await db.query(
    'SELECT listing_id FROM favorites WHERE user_id = $1',
    [userId]
  );
  
  const listingIds = favorites.rows.map(f => f.listing_id);
  
  // Delete all favorites
  const result = await db.query(
    'DELETE FROM favorites WHERE user_id = $1',
    [userId]
  );
  
  // Update favorites count for all affected listings
  if (listingIds.length > 0) {
    await db.query(
      'UPDATE car_listings SET favorites_count = GREATEST(favorites_count - 1, 0) WHERE id = ANY($1)',
      [listingIds]
    );
  }
  
  return result.rowCount;
}
```

### Getting Favorites with Listing Details

```javascript
async function getUserFavorites(userId, page = 1, limit = 20) {
  const offset = (page - 1) * limit;
  
  const favorites = await db.query(`
    SELECT 
      f.id,
      f.created_at,
      cl.id as listing_id,
      cl.title,
      cl.make,
      cl.model,
      cl.year,
      cl.price,
      cl.mileage,
      cl.location,
      cl.status,
      cl.health_score,
      cl.views,
      cl.days_listed,
      (SELECT image_url FROM listing_images WHERE listing_id = cl.id AND is_cover = TRUE LIMIT 1) as cover_image
    FROM favorites f
    JOIN car_listings cl ON f.listing_id = cl.id
    WHERE f.user_id = $1
    ORDER BY f.created_at DESC
    LIMIT $2 OFFSET $3
  `, [userId, limit, offset]);
  
  const total = await db.query(
    'SELECT COUNT(*) FROM favorites WHERE user_id = $1',
    [userId]
  );
  
  return {
    favorites: favorites.rows,
    pagination: {
      page,
      limit,
      total: parseInt(total.rows[0].count),
      totalPages: Math.ceil(parseInt(total.rows[0].count) / limit)
    }
  };
}
```

---

## Frontend Integration

The frontend uses a React Context to manage favorites state:

```typescript
interface FavoritesContextType {
  favorites: string[]; // Array of listing IDs
  addToFavorites: (listingId: string) => Promise<void>;
  removeFromFavorites: (listingId: string) => Promise<void>;
  isFavorited: (listingId: string) => boolean;
  clearFavorites: () => Promise<void>;
  favoritesCount: number;
}
```

### Optimistic Updates

For better UX, implement optimistic updates:

```javascript
async function toggleFavorite(listingId) {
  const isFavorited = favorites.includes(listingId);
  
  // Optimistic update
  if (isFavorited) {
    setFavorites(prev => prev.filter(id => id !== listingId));
  } else {
    setFavorites(prev => [...prev, listingId]);
  }
  
  try {
    if (isFavorited) {
      await api.delete(`/favorites/${listingId}`);
    } else {
      await api.post(`/favorites/${listingId}`);
    }
  } catch (error) {
    // Revert on error
    if (isFavorited) {
      setFavorites(prev => [...prev, listingId]);
    } else {
      setFavorites(prev => prev.filter(id => id !== listingId));
    }
    throw error;
  }
}
```

---

## Error Responses

```json
// 404 - Listing Not Found
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Listing not found"
  }
}

// 409 - Already Favorited
{
  "success": false,
  "error": {
    "code": "ALREADY_EXISTS",
    "message": "Listing is already in favorites"
  }
}

// 400 - Cannot Favorite Inactive Listing
{
  "success": false,
  "error": {
    "code": "BAD_REQUEST",
    "message": "Cannot favorite inactive listing"
  }
}

// 401 - Unauthorized
{
  "success": false,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Authentication required"
  }
}
```

---

## Related Services

- **Listings Service** - Provides listing data and updates favorites_count
- **Analytics Service** - Tracks favorite actions for dealer analytics
- **Notification Service** - Optional: Notify sellers of new favorites

