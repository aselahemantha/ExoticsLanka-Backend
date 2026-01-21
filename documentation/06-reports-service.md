# Reports Service

> Handles listing reports and content moderation.

---

## Overview

The Reports Service allows users to report inappropriate, fraudulent, or incorrect listings. It provides a moderation queue for administrators to review and take action on reported content.

**Responsibilities:**
- Accept listing reports from users
- Track report status
- Provide admin moderation queue
- Resolve/dismiss reports
- Take action on listings (remove, warn, etc.)

---

## Database Tables

### Listing Reports Table

```sql
CREATE TABLE listing_reports (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    listing_id      UUID NOT NULL REFERENCES car_listings(id) ON DELETE CASCADE,
    reporter_id     UUID REFERENCES users(id) ON DELETE SET NULL,
    
    -- Report Details
    reason          VARCHAR(50) NOT NULL 
                    CHECK (reason IN ('misleading', 'duplicate', 'sold', 'spam', 'inappropriate', 'other')),
    details         TEXT,
    
    -- Moderation
    status          VARCHAR(20) DEFAULT 'pending' 
                    CHECK (status IN ('pending', 'reviewing', 'resolved', 'dismissed')),
    admin_notes     TEXT,
    action_taken    VARCHAR(50),
    resolved_at     TIMESTAMP,
    resolved_by     UUID REFERENCES users(id),
    
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_listing_reports_listing_id ON listing_reports(listing_id);
CREATE INDEX idx_listing_reports_status ON listing_reports(status);
CREATE INDEX idx_listing_reports_reason ON listing_reports(reason);
CREATE INDEX idx_listing_reports_created_at ON listing_reports(created_at DESC);
CREATE INDEX idx_listing_reports_reporter ON listing_reports(reporter_id);
```

---

## Field Descriptions

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Report identifier |
| `listing_id` | UUID | Reported listing |
| `reporter_id` | UUID | User who submitted report (nullable) |
| `reason` | VARCHAR(50) | Category of the report |
| `details` | TEXT | Additional details from reporter |
| `status` | VARCHAR(20) | pending, reviewing, resolved, dismissed |
| `admin_notes` | TEXT | Internal notes from moderator |
| `action_taken` | VARCHAR(50) | What action was taken |
| `resolved_at` | TIMESTAMP | When report was resolved |
| `resolved_by` | UUID | Admin who resolved |
| `created_at` | TIMESTAMP | When report was submitted |

### Report Reasons

| Reason | Label | Description |
|--------|-------|-------------|
| `misleading` | Misleading information | Price, mileage, or other details are incorrect |
| `duplicate` | Duplicate listing | This vehicle is already listed |
| `sold` | Already sold | Vehicle is no longer available |
| `spam` | Spam or scam | Suspicious or fraudulent listing |
| `inappropriate` | Inappropriate content | Contains offensive material |
| `other` | Other | Something else is wrong |

### Report Statuses

| Status | Description |
|--------|-------------|
| `pending` | Awaiting review |
| `reviewing` | Under active review |
| `resolved` | Action taken, report closed |
| `dismissed` | No action needed, report closed |

### Actions Taken

| Action | Description |
|--------|-------------|
| `listing_removed` | Listing was deleted |
| `listing_edited` | Listing was modified by admin |
| `user_warned` | User received a warning |
| `user_suspended` | User account suspended |
| `no_violation` | No violation found |

---

## API Endpoints

### Public Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| `POST` | `/api/reports` | Submit listing report | Yes |

### Admin Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| `GET` | `/api/reports` | Get all reports | Yes (Admin) |
| `GET` | `/api/reports/:id` | Get report details | Yes (Admin) |
| `PUT` | `/api/reports/:id` | Update report status | Yes (Admin) |
| `GET` | `/api/reports/stats` | Get report statistics | Yes (Admin) |
| `GET` | `/api/reports/listing/:listingId` | Get reports for listing | Yes (Admin) |

---

## Request/Response Examples

### POST /api/reports

Submit a report for a listing.

**Request Body:**
```json
{
  "listingId": "listing-uuid",
  "reason": "misleading",
  "details": "The mileage shown is 12,000 km but the car clearly has more wear. The photos also don't match the actual vehicle I went to see."
}
```

**Response:**
```json
{
  "success": true,
  "message": "Report submitted. We'll review this listing within 24 hours.",
  "data": {
    "id": "report-uuid",
    "status": "pending",
    "createdAt": "2024-01-16T10:00:00Z"
  }
}
```

### GET /api/reports (Admin)

Get all reports with filtering.

**Query Parameters:**
```
?status=pending
&reason=misleading
&page=1
&limit=20
&sortBy=date_desc
```

**Response:**
```json
{
  "success": true,
  "data": {
    "reports": [
      {
        "id": "report-uuid-1",
        "reason": "misleading",
        "reasonLabel": "Misleading information",
        "details": "The mileage shown is incorrect...",
        "status": "pending",
        "createdAt": "2024-01-16T10:00:00Z",
        "listing": {
          "id": "listing-uuid",
          "title": "Toyota Aqua G",
          "price": 8500000,
          "status": "active",
          "coverImage": "https://storage.exotics.lk/listings/cover.jpg",
          "user": {
            "id": "seller-uuid",
            "name": "John Seller",
            "email": "seller@example.com"
          }
        },
        "reporter": {
          "id": "reporter-uuid",
          "name": "Jane Reporter",
          "email": "reporter@example.com"
        }
      },
      {
        "id": "report-uuid-2",
        "reason": "sold",
        "reasonLabel": "Already sold",
        "details": "I contacted the seller and the car was sold 2 weeks ago.",
        "status": "pending",
        "createdAt": "2024-01-15T14:30:00Z",
        "listing": {
          "id": "listing-uuid-2",
          "title": "Honda Civic Turbo",
          "price": 16800000,
          "status": "active",
          "coverImage": "https://storage.exotics.lk/listings/cover2.jpg",
          "user": {
            "id": "seller-uuid-2",
            "name": "Mike Seller",
            "email": "mike@example.com"
          }
        },
        "reporter": {
          "id": "reporter-uuid-2",
          "name": "Bob Reporter",
          "email": "bob@example.com"
        }
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 45,
      "totalPages": 3
    }
  }
}
```

### GET /api/reports/:id (Admin)

Get detailed report information.

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "report-uuid-1",
    "reason": "misleading",
    "reasonLabel": "Misleading information",
    "details": "The mileage shown is incorrect. When I went to see the car, the odometer showed 85,000 km not 12,000 km as listed.",
    "status": "pending",
    "adminNotes": null,
    "actionTaken": null,
    "resolvedAt": null,
    "createdAt": "2024-01-16T10:00:00Z",
    "listing": {
      "id": "listing-uuid",
      "title": "Toyota Aqua G",
      "make": "Toyota",
      "model": "Aqua G",
      "year": 2017,
      "price": 8500000,
      "mileage": 12000,
      "location": "Kurunegala",
      "status": "active",
      "healthScore": 72,
      "views": 3200,
      "createdAt": "2024-01-12T08:00:00Z",
      "images": [
        "https://storage.exotics.lk/listings/1.jpg",
        "https://storage.exotics.lk/listings/2.jpg"
      ],
      "user": {
        "id": "seller-uuid",
        "name": "John Seller",
        "email": "seller@example.com",
        "phone": "+94771234567",
        "role": "seller",
        "createdAt": "2023-06-15T00:00:00Z",
        "listingsCount": 5,
        "reportedCount": 2
      }
    },
    "reporter": {
      "id": "reporter-uuid",
      "name": "Jane Reporter",
      "email": "reporter@example.com",
      "createdAt": "2023-08-20T00:00:00Z",
      "reportsSubmitted": 3
    },
    "resolvedBy": null,
    "relatedReports": [
      {
        "id": "report-uuid-old",
        "reason": "spam",
        "status": "dismissed",
        "createdAt": "2024-01-10T12:00:00Z"
      }
    ]
  }
}
```

### PUT /api/reports/:id (Admin)

Update report status and take action.

**Request Body:**
```json
{
  "status": "resolved",
  "adminNotes": "Verified that mileage was incorrect. Listing has been removed and seller warned.",
  "actionTaken": "listing_removed"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Report resolved",
  "data": {
    "id": "report-uuid",
    "status": "resolved",
    "actionTaken": "listing_removed",
    "resolvedAt": "2024-01-16T14:00:00Z",
    "resolvedBy": {
      "id": "admin-uuid",
      "name": "Admin User"
    }
  }
}
```

### GET /api/reports/stats (Admin)

Get report statistics.

**Query Parameters:**
```
?period=30d  (7d, 30d, 90d, all)
```

**Response:**
```json
{
  "success": true,
  "data": {
    "period": "30d",
    "totalReports": 156,
    "byStatus": {
      "pending": 23,
      "reviewing": 5,
      "resolved": 98,
      "dismissed": 30
    },
    "byReason": {
      "misleading": 45,
      "duplicate": 12,
      "sold": 38,
      "spam": 28,
      "inappropriate": 8,
      "other": 25
    },
    "byAction": {
      "listing_removed": 42,
      "listing_edited": 15,
      "user_warned": 28,
      "user_suspended": 3,
      "no_violation": 40
    },
    "avgResolutionTimeHours": 18.5,
    "topReportedSellers": [
      { "id": "seller-uuid-1", "name": "John Seller", "reportCount": 8 },
      { "id": "seller-uuid-2", "name": "Mike Seller", "reportCount": 5 }
    ]
  }
}
```

---

## Business Logic

### Submitting a Report

```javascript
async function submitReport(reporterId, data) {
  const { listingId, reason, details } = data;
  
  // Verify listing exists
  const listing = await db.query(
    'SELECT id, user_id, status FROM car_listings WHERE id = $1',
    [listingId]
  );
  
  if (!listing.rows[0]) {
    throw new NotFoundError('Listing not found');
  }
  
  // Cannot report own listing
  if (listing.rows[0].user_id === reporterId) {
    throw new BadRequestError('You cannot report your own listing');
  }
  
  // Check for duplicate recent report
  const recentReport = await db.query(`
    SELECT id FROM listing_reports 
    WHERE listing_id = $1 AND reporter_id = $2 
    AND created_at > NOW() - INTERVAL '24 hours'
  `, [listingId, reporterId]);
  
  if (recentReport.rows[0]) {
    throw new ConflictError('You have already reported this listing recently');
  }
  
  // Create report
  const report = await db.query(`
    INSERT INTO listing_reports (listing_id, reporter_id, reason, details)
    VALUES ($1, $2, $3, $4)
    RETURNING *
  `, [listingId, reporterId, reason, details]);
  
  // Notify admins
  await notificationService.notifyAdminsNewReport(report.rows[0]);
  
  // Check if listing has multiple reports - auto-flag
  const reportCount = await db.query(
    'SELECT COUNT(*) FROM listing_reports WHERE listing_id = $1 AND status = $2',
    [listingId, 'pending']
  );
  
  if (parseInt(reportCount.rows[0].count) >= 3) {
    // Auto-flag listing for urgent review
    await notificationService.notifyAdminsUrgentReport(listingId);
  }
  
  return report.rows[0];
}
```

### Resolving a Report

```javascript
async function resolveReport(reportId, adminId, data) {
  const { status, adminNotes, actionTaken } = data;
  
  // Get report
  const report = await db.query(
    'SELECT * FROM listing_reports WHERE id = $1',
    [reportId]
  );
  
  if (!report.rows[0]) {
    throw new NotFoundError('Report not found');
  }
  
  // Update report
  const updated = await db.query(`
    UPDATE listing_reports 
    SET 
      status = $1,
      admin_notes = $2,
      action_taken = $3,
      resolved_at = NOW(),
      resolved_by = $4
    WHERE id = $5
    RETURNING *
  `, [status, adminNotes, actionTaken, adminId, reportId]);
  
  // Take action on listing if needed
  if (actionTaken === 'listing_removed') {
    await db.query(
      'UPDATE car_listings SET status = $1 WHERE id = $2',
      ['rejected', report.rows[0].listing_id]
    );
    
    // Notify seller
    await notificationService.notifyListingRemoved(report.rows[0].listing_id);
  }
  
  if (actionTaken === 'user_warned') {
    // Send warning to seller
    const listing = await db.query(
      'SELECT user_id FROM car_listings WHERE id = $1',
      [report.rows[0].listing_id]
    );
    await notificationService.sendUserWarning(listing.rows[0].user_id, report.rows[0]);
  }
  
  if (actionTaken === 'user_suspended') {
    const listing = await db.query(
      'SELECT user_id FROM car_listings WHERE id = $1',
      [report.rows[0].listing_id]
    );
    await db.query(
      'UPDATE users SET status = $1 WHERE id = $2',
      ['suspended', listing.rows[0].user_id]
    );
  }
  
  return updated.rows[0];
}
```

### Getting Moderation Queue

```javascript
async function getReports(filters, page = 1, limit = 20) {
  const offset = (page - 1) * limit;
  let query = `
    SELECT 
      r.id,
      r.reason,
      r.details,
      r.status,
      r.admin_notes,
      r.action_taken,
      r.resolved_at,
      r.created_at,
      json_build_object(
        'id', cl.id,
        'title', cl.title,
        'price', cl.price,
        'status', cl.status,
        'coverImage', (SELECT image_url FROM listing_images WHERE listing_id = cl.id AND is_cover = TRUE LIMIT 1),
        'user', json_build_object('id', s.id, 'name', s.name, 'email', s.email)
      ) as listing,
      json_build_object('id', rep.id, 'name', rep.name, 'email', rep.email) as reporter,
      json_build_object('id', res.id, 'name', res.name) as resolved_by
    FROM listing_reports r
    JOIN car_listings cl ON r.listing_id = cl.id
    JOIN users s ON cl.user_id = s.id
    LEFT JOIN users rep ON r.reporter_id = rep.id
    LEFT JOIN users res ON r.resolved_by = res.id
    WHERE 1=1
  `;
  
  const params = [];
  let paramIndex = 1;
  
  if (filters.status) {
    query += ` AND r.status = $${paramIndex}`;
    params.push(filters.status);
    paramIndex++;
  }
  
  if (filters.reason) {
    query += ` AND r.reason = $${paramIndex}`;
    params.push(filters.reason);
    paramIndex++;
  }
  
  query += ` ORDER BY r.created_at DESC LIMIT $${paramIndex} OFFSET $${paramIndex + 1}`;
  params.push(limit, offset);
  
  const reports = await db.query(query, params);
  
  // Get total count
  let countQuery = 'SELECT COUNT(*) FROM listing_reports WHERE 1=1';
  // ... add same filters
  const total = await db.query(countQuery, params.slice(0, -2));
  
  return {
    reports: reports.rows,
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

## Validation Rules

```javascript
const reportValidation = {
  listingId: {
    required: true,
    format: 'uuid'
  },
  reason: {
    required: true,
    enum: ['misleading', 'duplicate', 'sold', 'spam', 'inappropriate', 'other']
  },
  details: {
    maxLength: 2000
  }
};

const resolveValidation = {
  status: {
    required: true,
    enum: ['pending', 'reviewing', 'resolved', 'dismissed']
  },
  adminNotes: {
    maxLength: 2000
  },
  actionTaken: {
    enum: ['listing_removed', 'listing_edited', 'user_warned', 'user_suspended', 'no_violation']
  }
};
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

// 400 - Self Report
{
  "success": false,
  "error": {
    "code": "BAD_REQUEST",
    "message": "You cannot report your own listing"
  }
}

// 409 - Duplicate Report
{
  "success": false,
  "error": {
    "code": "CONFLICT",
    "message": "You have already reported this listing recently"
  }
}

// 403 - Admin Required
{
  "success": false,
  "error": {
    "code": "FORBIDDEN",
    "message": "Admin access required"
  }
}
```

---

## Related Services

- **Listings Service** - Updates listing status based on report actions
- **User Service** - Suspends users if needed
- **Notification Service** - Sends notifications to admins and affected users

