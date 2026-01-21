# Comparison Service

> Manages vehicle comparison lists for users.

---

## Overview

The Comparison Service allows users to add vehicles to a comparison list and view them side-by-side. This helps buyers make informed decisions by comparing specifications, features, and pricing.

**Responsibilities:**
- Add/remove vehicles from comparison
- Limit comparison to maximum 4 vehicles
- Persist comparison list (per user or session)
- Generate comparison data

---

## Database Tables

### Comparison Items Table

```sql
CREATE TABLE comparison_items (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    listing_id      UUID NOT NULL REFERENCES car_listings(id) ON DELETE CASCADE,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(user_id, listing_id)
);

CREATE INDEX idx_comparison_items_user_id ON comparison_items(user_id);
CREATE INDEX idx_comparison_items_listing_id ON comparison_items(listing_id);
```

---

## Field Descriptions

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Comparison item identifier |
| `user_id` | UUID | User who added the item |
| `listing_id` | UUID | Listing being compared |
| `created_at` | TIMESTAMP | When added to comparison |

---

## API Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| `GET` | `/api/comparison` | Get comparison list | Yes |
| `POST` | `/api/comparison/:listingId` | Add to comparison | Yes |
| `DELETE` | `/api/comparison/:listingId` | Remove from comparison | Yes |
| `DELETE` | `/api/comparison` | Clear comparison list | Yes |
| `GET` | `/api/comparison/check/:listingId` | Check if in comparison | Yes |
| `GET` | `/api/comparison/compare` | Get comparison data | Yes |

---

## Request/Response Examples

### GET /api/comparison

Get all items in comparison list.

**Response:**
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": "listing-uuid-1",
        "title": "Mercedes-Benz S-Class S500",
        "make": "Mercedes-Benz",
        "model": "S-Class S500",
        "year": 2023,
        "price": 85000000,
        "mileage": 12000,
        "image": "https://storage.exotics.lk/listings/1.jpg"
      },
      {
        "id": "listing-uuid-2",
        "title": "BMW 7 Series 740i",
        "make": "BMW",
        "model": "7 Series 740i",
        "year": 2022,
        "price": 72000000,
        "mileage": 18000,
        "image": "https://storage.exotics.lk/listings/2.jpg"
      }
    ],
    "count": 2,
    "maxItems": 4
  }
}
```

### POST /api/comparison/:listingId

Add a listing to comparison.

**Response (Success):**
```json
{
  "success": true,
  "message": "Added to comparison",
  "data": {
    "count": 3,
    "maxItems": 4
  }
}
```

**Response (Max Reached):**
```json
{
  "success": false,
  "error": {
    "code": "LIMIT_EXCEEDED",
    "message": "You can only compare up to 4 vehicles at once"
  }
}
```

### GET /api/comparison/compare

Get detailed comparison data for all items.

**Response:**
```json
{
  "success": true,
  "data": {
    "vehicles": [
      {
        "id": "listing-uuid-1",
        "title": "Mercedes-Benz S-Class S500",
        "image": "https://storage.exotics.lk/listings/1.jpg",
        "specs": {
          "make": "Mercedes-Benz",
          "model": "S-Class S500",
          "year": 2023,
          "price": 85000000,
          "mileage": 12000,
          "transmission": "Automatic",
          "fuelType": "Petrol",
          "bodyType": "Sedan",
          "color": "Midnight Black",
          "engineSize": "4.0L",
          "doors": 4,
          "seats": 5
        },
        "features": [
          "Leather Seats",
          "Sunroof",
          "Navigation System",
          "360° Camera"
        ],
        "healthScore": 94,
        "sellerRating": 4.8
      },
      {
        "id": "listing-uuid-2",
        "title": "BMW 7 Series 740i",
        "image": "https://storage.exotics.lk/listings/2.jpg",
        "specs": {
          "make": "BMW",
          "model": "7 Series 740i",
          "year": 2022,
          "price": 72000000,
          "mileage": 18000,
          "transmission": "Automatic",
          "fuelType": "Petrol",
          "bodyType": "Sedan",
          "color": "Alpine White",
          "engineSize": "3.0L",
          "doors": 4,
          "seats": 5
        },
        "features": [
          "Leather Seats",
          "Sunroof",
          "Navigation System",
          "Parking Sensors"
        ],
        "healthScore": 88,
        "sellerRating": 4.5
      }
    ],
    "comparison": {
      "priceRange": {
        "lowest": 72000000,
        "highest": 85000000,
        "difference": 13000000
      },
      "yearRange": {
        "oldest": 2022,
        "newest": 2023
      },
      "mileageRange": {
        "lowest": 12000,
        "highest": 18000
      },
      "commonFeatures": [
        "Leather Seats",
        "Sunroof",
        "Navigation System"
      ],
      "uniqueFeatures": {
        "listing-uuid-1": ["360° Camera"],
        "listing-uuid-2": ["Parking Sensors"]
      }
    }
  }
}
```

---

## Business Logic

### Adding to Comparison

```javascript
const MAX_COMPARISON_ITEMS = 4;

async function addToComparison(userId, listingId) {
  // Check current count
  const count = await db.query(
    'SELECT COUNT(*) FROM comparison_items WHERE user_id = $1',
    [userId]
  );
  
  if (parseInt(count.rows[0].count) >= MAX_COMPARISON_ITEMS) {
    throw new LimitExceededError(`You can only compare up to ${MAX_COMPARISON_ITEMS} vehicles at once`);
  }
  
  // Check if listing exists and is active
  const listing = await db.query(
    'SELECT id FROM car_listings WHERE id = $1 AND status = $2',
    [listingId, 'active']
  );
  
  if (!listing.rows[0]) {
    throw new NotFoundError('Listing not found or inactive');
  }
  
  // Check if already in comparison
  const existing = await db.query(
    'SELECT id FROM comparison_items WHERE user_id = $1 AND listing_id = $2',
    [userId, listingId]
  );
  
  if (existing.rows[0]) {
    throw new ConflictError('Listing is already in comparison');
  }
  
  // Add to comparison
  await db.query(
    'INSERT INTO comparison_items (user_id, listing_id) VALUES ($1, $2)',
    [userId, listingId]
  );
  
  const newCount = parseInt(count.rows[0].count) + 1;
  
  return { count: newCount, maxItems: MAX_COMPARISON_ITEMS };
}
```

### Getting Comparison Data

```javascript
async function getComparisonData(userId) {
  const items = await db.query(`
    SELECT 
      cl.*,
      (SELECT image_url FROM listing_images WHERE listing_id = cl.id AND is_cover = TRUE LIMIT 1) as image,
      (SELECT json_agg(feature_name) FROM listing_features WHERE listing_id = cl.id) as features,
      (SELECT AVG(rating) FROM reviews WHERE seller_id = cl.user_id) as seller_rating
    FROM comparison_items ci
    JOIN car_listings cl ON ci.listing_id = cl.id
    WHERE ci.user_id = $1
    ORDER BY ci.created_at
  `, [userId]);
  
  if (items.rows.length === 0) {
    return { vehicles: [], comparison: null };
  }
  
  // Build comparison analysis
  const vehicles = items.rows;
  const prices = vehicles.map(v => v.price);
  const years = vehicles.map(v => v.year);
  const mileages = vehicles.map(v => v.mileage);
  
  // Find common and unique features
  const allFeatures = vehicles.map(v => v.features || []);
  const commonFeatures = allFeatures.reduce((common, features) => 
    common.filter(f => features.includes(f)), allFeatures[0] || []);
  
  const uniqueFeatures = {};
  vehicles.forEach(v => {
    uniqueFeatures[v.id] = (v.features || []).filter(f => !commonFeatures.includes(f));
  });
  
  return {
    vehicles: vehicles.map(v => ({
      id: v.id,
      title: v.title,
      image: v.image,
      specs: {
        make: v.make,
        model: v.model,
        year: v.year,
        price: v.price,
        mileage: v.mileage,
        transmission: v.transmission,
        fuelType: v.fuel_type,
        bodyType: v.body_type,
        color: v.color,
        engineSize: v.engine_size,
        doors: v.doors,
        seats: v.seats
      },
      features: v.features || [],
      healthScore: v.health_score,
      sellerRating: v.seller_rating ? parseFloat(v.seller_rating).toFixed(1) : null
    })),
    comparison: {
      priceRange: {
        lowest: Math.min(...prices),
        highest: Math.max(...prices),
        difference: Math.max(...prices) - Math.min(...prices)
      },
      yearRange: {
        oldest: Math.min(...years),
        newest: Math.max(...years)
      },
      mileageRange: {
        lowest: Math.min(...mileages),
        highest: Math.max(...mileages)
      },
      commonFeatures,
      uniqueFeatures
    }
  };
}
```

---

## Frontend Integration

The frontend currently uses React Context with localStorage. To integrate with backend:

```typescript
// Update ComparisonContext to use API
async function addToComparison(car: Car) {
  if (comparisonList.length >= MAX_COMPARISON_ITEMS) {
    toast.error(`You can only compare up to ${MAX_COMPARISON_ITEMS} vehicles at once.`);
    return;
  }

  try {
    await api.post(`/comparison/${car.id}`);
    setComparisonList(prev => [...prev, car]);
    toast.success(`${car.title} added to comparison.`);
  } catch (error) {
    toast.error(error.message);
  }
}
```

---

## Error Responses

```json
// 400 - Limit Exceeded
{
  "success": false,
  "error": {
    "code": "LIMIT_EXCEEDED",
    "message": "You can only compare up to 4 vehicles at once"
  }
}

// 409 - Already in Comparison
{
  "success": false,
  "error": {
    "code": "CONFLICT",
    "message": "Listing is already in comparison"
  }
}

// 404 - Listing Not Found
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Listing not found or inactive"
  }
}
```

---

## Related Services

- **Listings Service** - Provides listing data for comparison

