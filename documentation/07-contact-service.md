# Contact Service

> Handles contact form submissions and support inquiries.

---

## Overview

The Contact Service manages all contact form submissions from the website. It provides a way for users (and non-users) to reach out to the platform for support, inquiries, or feedback.

**Responsibilities:**
- Receive contact form submissions
- Store inquiries in database
- Provide admin interface for managing inquiries
- Track response status
- Send confirmation emails

---

## Database Tables

### Contact Inquiries Table

```sql
CREATE TABLE contact_inquiries (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Contact Info
    name            VARCHAR(255) NOT NULL,
    email           VARCHAR(255) NOT NULL,
    phone           VARCHAR(20),
    
    -- Inquiry Details
    subject         VARCHAR(100),
    message         TEXT NOT NULL,
    
    -- Status
    status          VARCHAR(20) DEFAULT 'pending' 
                    CHECK (status IN ('pending', 'in_progress', 'responded', 'closed')),
    priority        VARCHAR(20) DEFAULT 'normal'
                    CHECK (priority IN ('low', 'normal', 'high', 'urgent')),
    
    -- Response
    admin_response  TEXT,
    responded_by    UUID REFERENCES users(id),
    responded_at    TIMESTAMP,
    
    -- Metadata
    user_id         UUID REFERENCES users(id) ON DELETE SET NULL,
    ip_address      VARCHAR(45),
    user_agent      TEXT,
    
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_contact_inquiries_status ON contact_inquiries(status);
CREATE INDEX idx_contact_inquiries_email ON contact_inquiries(email);
CREATE INDEX idx_contact_inquiries_created_at ON contact_inquiries(created_at DESC);
CREATE INDEX idx_contact_inquiries_subject ON contact_inquiries(subject);
CREATE INDEX idx_contact_inquiries_priority ON contact_inquiries(priority);
```

---

## Field Descriptions

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Inquiry identifier |
| `name` | VARCHAR(255) | Sender's full name |
| `email` | VARCHAR(255) | Sender's email address |
| `phone` | VARCHAR(20) | Optional phone number |
| `subject` | VARCHAR(100) | Inquiry category/subject |
| `message` | TEXT | Inquiry content |
| `status` | VARCHAR(20) | pending, in_progress, responded, closed |
| `priority` | VARCHAR(20) | low, normal, high, urgent |
| `admin_response` | TEXT | Response from support |
| `responded_by` | UUID | Admin who responded |
| `responded_at` | TIMESTAMP | When response was sent |
| `user_id` | UUID | Linked user account (if logged in) |
| `ip_address` | VARCHAR(45) | Sender's IP address |
| `user_agent` | TEXT | Browser user agent |

### Subject Options

| Value | Label |
|-------|-------|
| `general` | General Inquiry |
| `support` | Technical Support |
| `listing` | Listing Question |
| `account` | Account Issue |
| `partnership` | Partnership |
| `feedback` | Feedback |
| `complaint` | Complaint |
| `other` | Other |

---

## API Endpoints

### Public Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| `POST` | `/api/contact` | Submit contact inquiry | No |

### Admin Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| `GET` | `/api/contact` | Get all inquiries | Yes (Admin) |
| `GET` | `/api/contact/:id` | Get inquiry details | Yes (Admin) |
| `PUT` | `/api/contact/:id` | Update inquiry/respond | Yes (Admin) |
| `DELETE` | `/api/contact/:id` | Delete inquiry | Yes (Admin) |
| `GET` | `/api/contact/stats` | Get inquiry statistics | Yes (Admin) |

---

## Request/Response Examples

### POST /api/contact

Submit a contact form inquiry.

**Request Body:**
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "phone": "+94771234567",
  "subject": "listing",
  "message": "I'm having trouble uploading images to my listing. The images keep failing to upload even though they are under 10MB. Can you help?"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Message sent! We'll get back to you within 24 hours.",
  "data": {
    "id": "inquiry-uuid",
    "referenceNumber": "INQ-20240116-001",
    "createdAt": "2024-01-16T10:00:00Z"
  }
}
```

### GET /api/contact (Admin)

Get all contact inquiries.

**Query Parameters:**
```
?status=pending
&subject=support
&priority=high
&search=john@example.com
&page=1
&limit=20
&sortBy=date_desc
```

**Response:**
```json
{
  "success": true,
  "data": {
    "inquiries": [
      {
        "id": "inquiry-uuid-1",
        "referenceNumber": "INQ-20240116-001",
        "name": "John Doe",
        "email": "john@example.com",
        "phone": "+94771234567",
        "subject": "support",
        "subjectLabel": "Technical Support",
        "message": "I'm having trouble uploading images...",
        "status": "pending",
        "priority": "normal",
        "user": {
          "id": "user-uuid",
          "name": "John Doe",
          "role": "seller"
        },
        "createdAt": "2024-01-16T10:00:00Z"
      },
      {
        "id": "inquiry-uuid-2",
        "referenceNumber": "INQ-20240115-003",
        "name": "Jane Smith",
        "email": "jane@example.com",
        "phone": null,
        "subject": "general",
        "subjectLabel": "General Inquiry",
        "message": "What are the fees for listing a vehicle?",
        "status": "responded",
        "priority": "low",
        "user": null,
        "respondedAt": "2024-01-15T14:30:00Z",
        "createdAt": "2024-01-15T09:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 156,
      "totalPages": 8
    }
  }
}
```

### GET /api/contact/:id (Admin)

Get detailed inquiry information.

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "inquiry-uuid-1",
    "referenceNumber": "INQ-20240116-001",
    "name": "John Doe",
    "email": "john@example.com",
    "phone": "+94771234567",
    "subject": "support",
    "subjectLabel": "Technical Support",
    "message": "I'm having trouble uploading images to my listing. The images keep failing to upload even though they are under 10MB. Can you help?",
    "status": "pending",
    "priority": "normal",
    "adminResponse": null,
    "respondedBy": null,
    "respondedAt": null,
    "user": {
      "id": "user-uuid",
      "name": "John Doe",
      "email": "john@example.com",
      "role": "seller",
      "createdAt": "2023-06-15T00:00:00Z",
      "listingsCount": 3
    },
    "metadata": {
      "ipAddress": "192.168.1.1",
      "userAgent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)...",
      "referrer": "https://exotics.lk/sell"
    },
    "history": [
      {
        "action": "created",
        "timestamp": "2024-01-16T10:00:00Z"
      }
    ],
    "createdAt": "2024-01-16T10:00:00Z",
    "updatedAt": "2024-01-16T10:00:00Z"
  }
}
```

### PUT /api/contact/:id (Admin)

Update inquiry status or respond.

**Request Body (Update Status):**
```json
{
  "status": "in_progress",
  "priority": "high"
}
```

**Request Body (Respond):**
```json
{
  "status": "responded",
  "adminResponse": "Hello John,\n\nThank you for reaching out. We've identified an issue with our image upload service that was affecting some users.\n\nThe issue has been resolved. Please try uploading your images again. If you continue to experience problems, please clear your browser cache and try again.\n\nBest regards,\nExotics Lanka Support"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Response sent to john@example.com",
  "data": {
    "id": "inquiry-uuid",
    "status": "responded",
    "respondedAt": "2024-01-16T14:00:00Z",
    "respondedBy": {
      "id": "admin-uuid",
      "name": "Support Admin"
    }
  }
}
```

### GET /api/contact/stats (Admin)

Get inquiry statistics.

**Query Parameters:**
```
?period=30d
```

**Response:**
```json
{
  "success": true,
  "data": {
    "period": "30d",
    "total": 156,
    "byStatus": {
      "pending": 12,
      "in_progress": 5,
      "responded": 125,
      "closed": 14
    },
    "bySubject": {
      "general": 45,
      "support": 38,
      "listing": 32,
      "account": 18,
      "partnership": 8,
      "feedback": 10,
      "complaint": 3,
      "other": 2
    },
    "byPriority": {
      "low": 40,
      "normal": 98,
      "high": 15,
      "urgent": 3
    },
    "avgResponseTimeHours": 8.5,
    "responseRate": 94.2,
    "dailyTrend": [
      { "date": "2024-01-16", "count": 8 },
      { "date": "2024-01-15", "count": 12 },
      { "date": "2024-01-14", "count": 5 }
    ]
  }
}
```

---

## Business Logic

### Submitting an Inquiry

```javascript
async function submitInquiry(data, metadata) {
  const { name, email, phone, subject, message } = data;
  const { userId, ipAddress, userAgent } = metadata;
  
  // Rate limiting check
  const recentInquiries = await db.query(`
    SELECT COUNT(*) FROM contact_inquiries 
    WHERE (email = $1 OR ip_address = $2)
    AND created_at > NOW() - INTERVAL '1 hour'
  `, [email, ipAddress]);
  
  if (parseInt(recentInquiries.rows[0].count) >= 5) {
    throw new TooManyRequestsError('Too many inquiries. Please try again later.');
  }
  
  // Generate reference number
  const today = new Date().toISOString().slice(0, 10).replace(/-/g, '');
  const countToday = await db.query(`
    SELECT COUNT(*) FROM contact_inquiries 
    WHERE DATE(created_at) = CURRENT_DATE
  `);
  const refNumber = `INQ-${today}-${String(parseInt(countToday.rows[0].count) + 1).padStart(3, '0')}`;
  
  // Determine priority based on subject
  let priority = 'normal';
  if (subject === 'complaint') priority = 'high';
  if (subject === 'support' && message.toLowerCase().includes('urgent')) priority = 'high';
  
  // Create inquiry
  const inquiry = await db.query(`
    INSERT INTO contact_inquiries (
      name, email, phone, subject, message,
      priority, user_id, ip_address, user_agent
    )
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    RETURNING *
  `, [name, email, phone, subject, message, priority, userId, ipAddress, userAgent]);
  
  // Send confirmation email to user
  await emailService.sendContactConfirmation({
    to: email,
    name,
    referenceNumber: refNumber,
    subject
  });
  
  // Notify support team
  await notificationService.notifyNewInquiry(inquiry.rows[0]);
  
  return {
    ...inquiry.rows[0],
    referenceNumber: refNumber
  };
}
```

### Responding to an Inquiry

```javascript
async function respondToInquiry(inquiryId, adminId, response) {
  // Get inquiry
  const inquiry = await db.query(
    'SELECT * FROM contact_inquiries WHERE id = $1',
    [inquiryId]
  );
  
  if (!inquiry.rows[0]) {
    throw new NotFoundError('Inquiry not found');
  }
  
  // Update inquiry
  const updated = await db.query(`
    UPDATE contact_inquiries 
    SET 
      status = 'responded',
      admin_response = $1,
      responded_by = $2,
      responded_at = NOW(),
      updated_at = NOW()
    WHERE id = $3
    RETURNING *
  `, [response, adminId, inquiryId]);
  
  // Send response email
  await emailService.sendContactResponse({
    to: inquiry.rows[0].email,
    name: inquiry.rows[0].name,
    subject: inquiry.rows[0].subject,
    originalMessage: inquiry.rows[0].message,
    response
  });
  
  return updated.rows[0];
}
```

### Getting Inquiries with Filters

```javascript
async function getInquiries(filters, page = 1, limit = 20) {
  const offset = (page - 1) * limit;
  let query = `
    SELECT 
      ci.*,
      json_build_object('id', u.id, 'name', u.name, 'role', u.role) as user,
      json_build_object('id', admin.id, 'name', admin.name) as responded_by_user
    FROM contact_inquiries ci
    LEFT JOIN users u ON ci.user_id = u.id
    LEFT JOIN users admin ON ci.responded_by = admin.id
    WHERE 1=1
  `;
  
  const params = [];
  let paramIndex = 1;
  
  if (filters.status) {
    query += ` AND ci.status = $${paramIndex}`;
    params.push(filters.status);
    paramIndex++;
  }
  
  if (filters.subject) {
    query += ` AND ci.subject = $${paramIndex}`;
    params.push(filters.subject);
    paramIndex++;
  }
  
  if (filters.priority) {
    query += ` AND ci.priority = $${paramIndex}`;
    params.push(filters.priority);
    paramIndex++;
  }
  
  if (filters.search) {
    query += ` AND (ci.email ILIKE $${paramIndex} OR ci.name ILIKE $${paramIndex} OR ci.message ILIKE $${paramIndex})`;
    params.push(`%${filters.search}%`);
    paramIndex++;
  }
  
  query += ` ORDER BY ci.created_at DESC LIMIT $${paramIndex} OFFSET $${paramIndex + 1}`;
  params.push(limit, offset);
  
  const inquiries = await db.query(query, params);
  
  // Get total
  let countQuery = 'SELECT COUNT(*) FROM contact_inquiries WHERE 1=1';
  // Add same filters...
  const total = await db.query(countQuery, params.slice(0, -2));
  
  return {
    inquiries: inquiries.rows,
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

## Email Templates

### Contact Confirmation Email

```html
Subject: We received your message - Reference #INQ-20240116-001

Hello John,

Thank you for contacting Exotics Lanka. We have received your inquiry and will get back to you within 24 hours.

Reference Number: INQ-20240116-001
Subject: Technical Support

Your Message:
"I'm having trouble uploading images to my listing..."

If you need immediate assistance, please call us at +94 77 123 4567.

Best regards,
Exotics Lanka Support Team
```

### Contact Response Email

```html
Subject: Re: Your inquiry - Reference #INQ-20240116-001

Hello John,

Thank you for reaching out to Exotics Lanka.

Your original inquiry:
"I'm having trouble uploading images to my listing..."

Our response:
"We've identified an issue with our image upload service..."

If you have any further questions, please reply to this email.

Best regards,
Exotics Lanka Support Team
```

---

## Validation Rules

```javascript
const contactValidation = {
  name: {
    required: true,
    minLength: 2,
    maxLength: 255
  },
  email: {
    required: true,
    format: 'email',
    maxLength: 255
  },
  phone: {
    format: 'phone',
    maxLength: 20
  },
  subject: {
    enum: ['general', 'support', 'listing', 'account', 'partnership', 'feedback', 'complaint', 'other']
  },
  message: {
    required: true,
    minLength: 10,
    maxLength: 5000
  }
};
```

---

## Error Responses

```json
// 429 - Rate Limited
{
  "success": false,
  "error": {
    "code": "TOO_MANY_REQUESTS",
    "message": "Too many inquiries. Please try again later."
  }
}

// 404 - Inquiry Not Found
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Inquiry not found"
  }
}

// 400 - Validation Error
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": [
      { "field": "email", "message": "Invalid email format" }
    ]
  }
}
```

---

## Related Services

- **Notification Service** - Sends emails and alerts
- **User Service** - Links inquiries to user accounts

