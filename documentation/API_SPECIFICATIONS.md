# 📡 API Specifications - Exotics Lanka

**Version:** 1.0  
**Base URL:** `https://api.exoticslanka.com/v1`  
**Authentication:** JWT Bearer Token  

---

## 🔐 **AUTHENTICATION APIs**

### **1. Register User**
```http
POST /api/auth/register
Content-Type: application/json

Request:
{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "fullName": "John Doe",
  "role": "buyer" | "seller",
  "phone": "+94771234567"
}

Response: 201 Created
{
  "success": true,
  "data": {
    "userId": "usr_123",
    "email": "user@example.com",
    "role": "buyer",
    "verificationSent": true
  },
  "message": "Registration successful. Please verify your email."
}
```

### **2. Login User**
```http
POST /api/auth/login
Content-Type: application/json

Request:
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}

Response: 200 OK
{
  "success": true,
  "data": {
    "user": {
      "id": "usr_123",
      "email": "user@example.com",
      "role": "buyer",
      "fullName": "John Doe"
    },
    "tokens": {
      "accessToken": "eyJhbGciOiJIUzI1NiIs...",
      "refreshToken": "eyJhbGciOiJIUzI1NiIs...",
      "expiresIn": 3600
    }
  }
}
```

### **3. Refresh Token**
```http
POST /api/auth/refresh
Content-Type: application/json

Request:
{
  "refreshToken": "eyJhbGciOiJIUzI1NiIs..."
}

Response: 200 OK
{
  "success": true,
  "data": {
    "accessToken": "eyJhbGciOiJIUzI1NiIs...",
    "expiresIn": 3600
  }
}
```

---

## 🚗 **LISTING APIs**

### **1. Get All Listings**
```http
GET /api/listings?page=1&limit=20&status=active
Authorization: Bearer {token}

Response: 200 OK
{
  "success": true,
  "data": {
    "listings": [
      {
        "id": "lst_123",
        "title": "2020 Porsche 911 Carrera",
        "description": "Pristine condition...",
        "price": 85000000,
        "currency": "LKR",
        "year": 2020,
        "make": "Porsche",
        "model": "911 Carrera",
        "mileage": 15000,
        "fuelType": "Petrol",
        "transmission": "Automatic",
        "location": "Colombo",
        "condition": "Used",
        "status": "active",
        "images": [
          {
            "id": "img_1",
            "url": "https://cdn.exotics.lk/...",
            "thumbnail": "https://cdn.exotics.lk/.../thumb",
            "order": 1,
            "isPrimary": true
          }
        ],
        "seller": {
          "id": "usr_456",
          "name": "Premium Auto Gallery",
          "rating": 4.8,
          "verified": true
        },
        "views": 1250,
        "favorites": 45,
        "createdAt": "2025-01-05T10:30:00Z",
        "updatedAt": "2025-01-06T08:15:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 150,
      "totalPages": 8
    }
  }
}
```

### **2. Get Single Listing**
```http
GET /api/listings/:id
Authorization: Bearer {token}

Response: 200 OK
{
  "success": true,
  "data": {
    "id": "lst_123",
    "title": "2020 Porsche 911 Carrera",
    "description": "Full description...",
    "price": 85000000,
    "currency": "LKR",
    "year": 2020,
    "make": "Porsche",
    "model": "911 Carrera",
    "mileage": 15000,
    "fuelType": "Petrol",
    "transmission": "Automatic",
    "bodyType": "Coupe",
    "color": "Midnight Black",
    "location": "Colombo",
    "condition": "Used",
    "status": "active",
    "features": [
      "Leather Interior",
      "Sunroof",
      "Navigation System",
      "Backup Camera"
    ],
    "images": [...],
    "videos": [
      {
        "id": "vid_1",
        "url": "https://cdn.exotics.lk/videos/...",
        "thumbnail": "https://cdn.exotics.lk/videos/.../thumb",
        "duration": 120
      }
    ],
    "seller": {
      "id": "usr_456",
      "name": "Premium Auto Gallery",
      "phone": "+94771234567",
      "email": "contact@premiumauto.lk",
      "rating": 4.8,
      "reviewCount": 125,
      "verified": true,
      "memberSince": "2023-01-15"
    },
    "stats": {
      "views": 1250,
      "favorites": 45,
      "messages": 18
    },
    "createdAt": "2025-01-05T10:30:00Z",
    "updatedAt": "2025-01-06T08:15:00Z"
  }
}
```

### **3. Create Listing**
```http
POST /api/listings
Authorization: Bearer {token}
Content-Type: application/json

Request:
{
  "title": "2020 Porsche 911 Carrera",
  "description": "Full description...",
  "price": 85000000,
  "currency": "LKR",
  "year": 2020,
  "make": "Porsche",
  "model": "911 Carrera",
  "mileage": 15000,
  "fuelType": "Petrol",
  "transmission": "Automatic",
  "bodyType": "Coupe",
  "color": "Midnight Black",
  "location": "Colombo",
  "condition": "Used",
  "features": [
    "Leather Interior",
    "Sunroof",
    "Navigation System"
  ],
  "status": "draft"
}

Response: 201 Created
{
  "success": true,
  "data": {
    "id": "lst_123",
    "title": "2020 Porsche 911 Carrera",
    "status": "draft",
    "createdAt": "2025-01-06T10:30:00Z"
  },
  "message": "Listing created successfully. Upload images to publish."
}
```

### **4. Upload Listing Images**
```http
POST /api/listings/:id/images
Authorization: Bearer {token}
Content-Type: multipart/form-data

Request:
- images: [File, File, File]

Response: 200 OK
{
  "success": true,
  "data": {
    "images": [
      {
        "id": "img_1",
        "url": "https://cdn.exotics.lk/...",
        "thumbnail": "https://cdn.exotics.lk/.../thumb",
        "order": 1
      },
      {
        "id": "img_2",
        "url": "https://cdn.exotics.lk/...",
        "thumbnail": "https://cdn.exotics.lk/.../thumb",
        "order": 2
      }
    ]
  },
  "message": "Images uploaded successfully"
}
```

### **5. Search Listings**
```http
GET /api/listings/search?q=porsche&make=Porsche&minPrice=50000000&maxPrice=100000000&year=2020&location=Colombo&page=1&limit=20
Authorization: Bearer {token}

Response: 200 OK
{
  "success": true,
  "data": {
    "results": [...],
    "filters": {
      "applied": {
        "make": "Porsche",
        "priceRange": [50000000, 100000000],
        "year": 2020,
        "location": "Colombo"
      },
      "available": {
        "makes": ["Porsche", "Ferrari", "Lamborghini"],
        "years": [2024, 2023, 2022, 2021, 2020],
        "locations": ["Colombo", "Galle", "Kandy"]
      }
    },
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 12,
      "totalPages": 1
    }
  }
}
```

---

## ❤️ **FAVORITES APIs**

### **1. Get User Favorites**
```http
GET /api/favorites
Authorization: Bearer {token}

Response: 200 OK
{
  "success": true,
  "data": {
    "favorites": [
      {
        "id": "fav_123",
        "listing": {
          "id": "lst_123",
          "title": "2020 Porsche 911 Carrera",
          "price": 85000000,
          "image": "https://cdn.exotics.lk/...",
          "status": "active"
        },
        "addedAt": "2025-01-05T14:20:00Z"
      }
    ],
    "count": 15
  }
}
```

### **2. Add to Favorites**
```http
POST /api/favorites/:listingId
Authorization: Bearer {token}

Response: 201 Created
{
  "success": true,
  "data": {
    "id": "fav_123",
    "listingId": "lst_123",
    "userId": "usr_123",
    "addedAt": "2025-01-06T10:30:00Z"
  },
  "message": "Added to favorites"
}
```

### **3. Remove from Favorites**
```http
DELETE /api/favorites/:listingId
Authorization: Bearer {token}

Response: 200 OK
{
  "success": true,
  "message": "Removed from favorites"
}
```

---

## 💬 **MESSAGING APIs**

### **1. Get Conversations**
```http
GET /api/messages/conversations
Authorization: Bearer {token}

Response: 200 OK
{
  "success": true,
  "data": {
    "conversations": [
      {
        "id": "conv_123",
        "listing": {
          "id": "lst_123",
          "title": "2020 Porsche 911 Carrera",
          "image": "https://cdn.exotics.lk/..."
        },
        "participant": {
          "id": "usr_456",
          "name": "Premium Auto Gallery",
          "avatar": "https://cdn.exotics.lk/avatars/..."
        },
        "lastMessage": {
          "content": "Is this still available?",
          "sentAt": "2025-01-06T09:15:00Z",
          "read": false
        },
        "unreadCount": 2,
        "updatedAt": "2025-01-06T09:15:00Z"
      }
    ]
  }
}
```

### **2. Get Messages**
```http
GET /api/messages/:conversationId?page=1&limit=50
Authorization: Bearer {token}

Response: 200 OK
{
  "success": true,
  "data": {
    "messages": [
      {
        "id": "msg_123",
        "conversationId": "conv_123",
        "sender": {
          "id": "usr_123",
          "name": "John Doe"
        },
        "content": "Is this still available?",
        "readAt": null,
        "createdAt": "2025-01-06T09:15:00Z"
      },
      {
        "id": "msg_124",
        "conversationId": "conv_123",
        "sender": {
          "id": "usr_456",
          "name": "Premium Auto Gallery"
        },
        "content": "Yes, it's available!",
        "readAt": "2025-01-06T09:20:00Z",
        "createdAt": "2025-01-06T09:18:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 50,
      "hasMore": false
    }
  }
}
```

### **3. Send Message**
```http
POST /api/messages/:conversationId
Authorization: Bearer {token}
Content-Type: application/json

Request:
{
  "content": "Is this still available?"
}

Response: 201 Created
{
  "success": true,
  "data": {
    "id": "msg_123",
    "conversationId": "conv_123",
    "content": "Is this still available?",
    "createdAt": "2025-01-06T09:15:00Z"
  }
}
```

---

## ⭐ **REVIEW APIs**

### **1. Get Seller Reviews**
```http
GET /api/reviews/seller/:sellerId?page=1&limit=10
Authorization: Bearer {token}

Response: 200 OK
{
  "success": true,
  "data": {
    "reviews": [
      {
        "id": "rev_123",
        "listing": {
          "id": "lst_123",
          "title": "2020 Porsche 911 Carrera"
        },
        "reviewer": {
          "id": "usr_789",
          "name": "Jane Smith",
          "verifiedBuyer": true
        },
        "rating": 5,
        "title": "Excellent service!",
        "content": "Very professional seller...",
        "response": {
          "content": "Thank you for your review!",
          "createdAt": "2025-01-06T10:00:00Z"
        },
        "helpful": 12,
        "createdAt": "2025-01-05T15:30:00Z"
      }
    ],
    "summary": {
      "averageRating": 4.8,
      "totalReviews": 125,
      "distribution": {
        "5": 90,
        "4": 25,
        "3": 8,
        "2": 2,
        "1": 0
      }
    },
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 125,
      "totalPages": 13
    }
  }
}
```

### **2. Create Review**
```http
POST /api/reviews
Authorization: Bearer {token}
Content-Type: application/json

Request:
{
  "listingId": "lst_123",
  "sellerId": "usr_456",
  "rating": 5,
  "title": "Excellent service!",
  "content": "Very professional seller..."
}

Response: 201 Created
{
  "success": true,
  "data": {
    "id": "rev_123",
    "rating": 5,
    "createdAt": "2025-01-06T10:30:00Z"
  },
  "message": "Review submitted successfully"
}
```

---

## 🔍 **SEARCH & ALERTS APIs**

### **1. Save Search**
```http
POST /api/search/saved
Authorization: Bearer {token}
Content-Type: application/json

Request:
{
  "name": "Porsche under 100M",
  "filters": {
    "make": "Porsche",
    "maxPrice": 100000000,
    "location": "Colombo"
  }
}

Response: 201 Created
{
  "success": true,
  "data": {
    "id": "search_123",
    "name": "Porsche under 100M",
    "filters": {...},
    "createdAt": "2025-01-06T10:30:00Z"
  }
}
```

### **2. Create Search Alert**
```http
POST /api/search/alerts
Authorization: Bearer {token}
Content-Type: application/json

Request:
{
  "savedSearchId": "search_123",
  "frequency": "daily" | "weekly" | "instant",
  "notificationChannels": ["email", "push"]
}

Response: 201 Created
{
  "success": true,
  "data": {
    "id": "alert_123",
    "savedSearchId": "search_123",
    "frequency": "daily",
    "active": true,
    "createdAt": "2025-01-06T10:30:00Z"
  },
  "message": "Alert created successfully"
}
```

---

## 📊 **ANALYTICS APIs**

### **1. Get Dashboard Stats (Seller)**
```http
GET /api/analytics/dashboard
Authorization: Bearer {token}

Response: 200 OK
{
  "success": true,
  "data": {
    "listings": {
      "total": 12,
      "active": 8,
      "sold": 3,
      "draft": 1
    },
    "performance": {
      "totalViews": 5420,
      "totalFavorites": 234,
      "totalMessages": 156
    },
    "topListings": [
      {
        "id": "lst_123",
        "title": "2020 Porsche 911 Carrera",
        "views": 1250,
        "favorites": 45
      }
    ],
    "recentActivity": [
      {
        "type": "view",
        "listing": "2020 Porsche 911 Carrera",
        "timestamp": "2025-01-06T09:45:00Z"
      }
    ]
  }
}
```

---

## 🔄 **ERROR RESPONSES**

### **Standard Error Format:**
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": {
      "field": "email",
      "reason": "Email already exists"
    }
  }
}
```

### **Common Error Codes:**
- `400` - Bad Request (validation error)
- `401` - Unauthorized (invalid/missing token)
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found (resource doesn't exist)
- `409` - Conflict (duplicate resource)
- `422` - Unprocessable Entity (validation failed)
- `429` - Too Many Requests (rate limit exceeded)
- `500` - Internal Server Error

---

## 🔑 **AUTHENTICATION**

All API requests (except auth endpoints) require a JWT token:

```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

Token expires in 1 hour. Use refresh token to get new access token.

---

## 🚀 **RATE LIMITING**

- **Authenticated:** 1000 requests/hour
- **Unauthenticated:** 100 requests/hour
- **Search:** 50 requests/minute

Headers:
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1641024000
```

---

**📡 Complete API specifications ready for implementation!**

