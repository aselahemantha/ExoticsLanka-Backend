# Notification Service

> Handles email, SMS, and push notifications.

---

## Overview

The Notification Service manages all outbound communications including emails, SMS messages, and push notifications. It provides templates, scheduling, and delivery tracking.

**Responsibilities:**
- Send transactional emails
- Send SMS notifications (optional)
- Push notifications (optional)
- Manage notification preferences
- Track delivery status
- Queue and rate limit notifications

---

## Configuration

### Email Provider (SendGrid)

```javascript
const sgMail = require('@sendgrid/mail');
sgMail.setApiKey(process.env.SENDGRID_API_KEY);

const EMAIL_FROM = {
  email: 'noreply@exotics.lk',
  name: 'Exotics Lanka'
};
```

### SMS Provider (Twilio)

```javascript
const twilio = require('twilio');
const twilioClient = twilio(
  process.env.TWILIO_ACCOUNT_SID,
  process.env.TWILIO_AUTH_TOKEN
);
const SMS_FROM = process.env.TWILIO_PHONE_NUMBER;
```

---

## Email Templates

### Template List

| Template | Trigger | Description |
|----------|---------|-------------|
| `welcome` | User registration | Welcome new user |
| `verify-email` | Registration | Email verification link |
| `password-reset` | Forgot password | Password reset link |
| `new-message` | New message | Notify of new message |
| `new-lead` | New conversation | Notify seller of inquiry |
| `search-alert` | New matches | Saved search matches |
| `listing-approved` | Admin approval | Listing is now live |
| `listing-rejected` | Admin rejection | Listing was rejected |
| `new-review` | Review submitted | Notify seller of review |
| `review-response` | Seller responds | Notify buyer of response |
| `contact-confirmation` | Contact form | Confirm inquiry received |
| `contact-response` | Admin responds | Support response |
| `weekly-report` | Scheduled | Dealer weekly summary |

---

## API Endpoints

### User Preferences

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| `GET` | `/api/notifications/preferences` | Get preferences | Yes |
| `PUT` | `/api/notifications/preferences` | Update preferences | Yes |

---

## Notification Preferences Schema

```javascript
const defaultPreferences = {
  email: {
    messages: true,
    leads: true,
    reviews: true,
    searchAlerts: true,
    marketing: false,
    weeklyReport: true
  },
  sms: {
    messages: false,
    leads: true,
    urgent: true
  },
  push: {
    messages: true,
    leads: true,
    reviews: true
  }
};
```

---

## Email Templates

### Welcome Email

```javascript
async function sendWelcomeEmail(user) {
  const template = {
    to: user.email,
    from: EMAIL_FROM,
    subject: 'Welcome to Exotics Lanka!',
    templateId: 'd-welcome-template-id',
    dynamicTemplateData: {
      name: user.name,
      loginUrl: `${process.env.FRONTEND_URL}/login`,
      exploreUrl: `${process.env.FRONTEND_URL}/collection`
    }
  };
  
  await sgMail.send(template);
}
```

**Template HTML:**
```html
<!DOCTYPE html>
<html>
<head>
  <style>
    .container { max-width: 600px; margin: 0 auto; font-family: Arial, sans-serif; }
    .header { background: linear-gradient(135deg, #D4AF37, #B8860B); padding: 30px; text-align: center; }
    .header h1 { color: white; margin: 0; }
    .content { padding: 30px; background: #f9f9f9; }
    .button { display: inline-block; background: #D4AF37; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; }
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h1>Exotics Lanka</h1>
    </div>
    <div class="content">
      <h2>Welcome, {{name}}!</h2>
      <p>Thank you for joining Exotics Lanka, Sri Lanka's premier luxury car marketplace.</p>
      <p>Start exploring our collection of premium vehicles:</p>
      <p style="text-align: center;">
        <a href="{{exploreUrl}}" class="button">Browse Collection</a>
      </p>
      <p>Need help? Contact us at support@exotics.lk</p>
    </div>
  </div>
</body>
</html>
```

### New Message Notification

```javascript
async function sendNewMessageNotification(recipientId, message, conversation) {
  const recipient = await getUser(recipientId);
  
  // Check preferences
  if (!recipient.notificationPreferences?.email?.messages) {
    return;
  }
  
  const template = {
    to: recipient.email,
    from: EMAIL_FROM,
    subject: `New message about ${conversation.listingTitle}`,
    templateId: 'd-new-message-template-id',
    dynamicTemplateData: {
      recipientName: recipient.name,
      senderName: message.senderName,
      listingTitle: conversation.listingTitle,
      listingImage: conversation.listingImage,
      messagePreview: message.content.substring(0, 100),
      conversationUrl: `${process.env.FRONTEND_URL}/inbox/${conversation.id}`
    }
  };
  
  await sgMail.send(template);
}
```

### Search Alert Email

```javascript
async function sendSearchAlert(data) {
  const { email, userName, searchName, newCount, listings } = data;
  
  const template = {
    to: email,
    from: EMAIL_FROM,
    subject: `${newCount} new matches for "${searchName}"`,
    templateId: 'd-search-alert-template-id',
    dynamicTemplateData: {
      userName,
      searchName,
      newCount,
      listings: listings.slice(0, 5).map(l => ({
        title: l.title,
        price: formatPrice(l.price),
        image: l.coverImage,
        url: `${process.env.FRONTEND_URL}/car/${l.id}`
      })),
      viewAllUrl: `${process.env.FRONTEND_URL}/saved-searches`
    }
  };
  
  await sgMail.send(template);
}
```

### New Review Notification

```javascript
async function sendNewReviewNotification(sellerId, review) {
  const seller = await getUser(sellerId);
  
  if (!seller.notificationPreferences?.email?.reviews) {
    return;
  }
  
  const template = {
    to: seller.email,
    from: EMAIL_FROM,
    subject: `New ${review.rating}-star review from ${review.buyerName}`,
    templateId: 'd-new-review-template-id',
    dynamicTemplateData: {
      sellerName: seller.name,
      buyerName: review.buyerName,
      rating: review.rating,
      stars: '★'.repeat(review.rating) + '☆'.repeat(5 - review.rating),
      title: review.title,
      comment: review.comment,
      respondUrl: `${process.env.FRONTEND_URL}/dashboard/reviews/${review.id}`
    }
  };
  
  await sgMail.send(template);
}
```

### Weekly Report Email

```javascript
async function sendWeeklyReport(dealerId) {
  const dealer = await getUser(dealerId);
  const analytics = await getWeeklyAnalytics(dealerId);
  
  if (!dealer.notificationPreferences?.email?.weeklyReport) {
    return;
  }
  
  const template = {
    to: dealer.email,
    from: EMAIL_FROM,
    subject: `Your Weekly Performance Report - ${formatDate(new Date())}`,
    templateId: 'd-weekly-report-template-id',
    dynamicTemplateData: {
      dealerName: dealer.name,
      weekEnding: formatDate(new Date()),
      summary: {
        views: analytics.totalViews,
        viewsChange: analytics.viewsChange,
        leads: analytics.totalLeads,
        leadsChange: analytics.leadsChange,
        sales: analytics.totalSales,
        revenue: formatPrice(analytics.totalRevenue)
      },
      topListings: analytics.topListings.slice(0, 3),
      insights: analytics.insights,
      dashboardUrl: `${process.env.FRONTEND_URL}/dealer/analytics`
    }
  };
  
  await sgMail.send(template);
}
```

---

## SMS Notifications

### Send SMS

```javascript
async function sendSMS(to, message) {
  try {
    const result = await twilioClient.messages.create({
      body: message,
      from: SMS_FROM,
      to: to
    });
    
    return { success: true, messageId: result.sid };
  } catch (error) {
    console.error('SMS send failed:', error);
    return { success: false, error: error.message };
  }
}
```

### New Lead SMS (for dealers)

```javascript
async function sendNewLeadSMS(dealerId, lead) {
  const dealer = await getUser(dealerId);
  
  if (!dealer.notificationPreferences?.sms?.leads || !dealer.phone) {
    return;
  }
  
  const message = `[Exotics Lanka] New inquiry for ${lead.listingTitle} from ${lead.buyerName}. Check your inbox: ${process.env.FRONTEND_URL}/inbox`;
  
  await sendSMS(dealer.phone, message);
}
```

---

## Push Notifications (Firebase)

### Setup

```javascript
const admin = require('firebase-admin');

admin.initializeApp({
  credential: admin.credential.cert({
    projectId: process.env.FIREBASE_PROJECT_ID,
    clientEmail: process.env.FIREBASE_CLIENT_EMAIL,
    privateKey: process.env.FIREBASE_PRIVATE_KEY
  })
});
```

### Send Push Notification

```javascript
async function sendPushNotification(userId, notification) {
  // Get user's device tokens
  const tokens = await db.query(
    'SELECT token FROM user_devices WHERE user_id = $1 AND active = true',
    [userId]
  );
  
  if (tokens.rows.length === 0) return;
  
  const message = {
    notification: {
      title: notification.title,
      body: notification.body,
      icon: '/icons/notification-icon.png'
    },
    data: notification.data || {},
    tokens: tokens.rows.map(t => t.token)
  };
  
  try {
    const response = await admin.messaging().sendMulticast(message);
    
    // Handle failed tokens
    if (response.failureCount > 0) {
      const failedTokens = [];
      response.responses.forEach((resp, idx) => {
        if (!resp.success) {
          failedTokens.push(tokens.rows[idx].token);
        }
      });
      
      // Remove invalid tokens
      await db.query(
        'UPDATE user_devices SET active = false WHERE token = ANY($1)',
        [failedTokens]
      );
    }
  } catch (error) {
    console.error('Push notification failed:', error);
  }
}
```

---

## Notification Queue (Bull)

### Setup Queue

```javascript
const Queue = require('bull');
const notificationQueue = new Queue('notifications', process.env.REDIS_URL);

// Process jobs
notificationQueue.process('email', async (job) => {
  const { type, data } = job.data;
  
  switch (type) {
    case 'welcome':
      await sendWelcomeEmail(data);
      break;
    case 'new-message':
      await sendNewMessageNotification(data.recipientId, data.message, data.conversation);
      break;
    case 'search-alert':
      await sendSearchAlert(data);
      break;
    // ... other types
  }
});

notificationQueue.process('sms', async (job) => {
  const { to, message } = job.data;
  await sendSMS(to, message);
});
```

### Queue Notification

```javascript
async function queueNotification(channel, type, data, options = {}) {
  await notificationQueue.add(channel, { type, data }, {
    attempts: 3,
    backoff: {
      type: 'exponential',
      delay: 2000
    },
    delay: options.delay || 0,
    priority: options.priority || 3
  });
}

// Usage
await queueNotification('email', 'welcome', { user });
await queueNotification('sms', 'new-lead', { to: phone, message }, { priority: 1 });
```

---

## Request/Response Examples

### GET /api/notifications/preferences

**Response:**
```json
{
  "success": true,
  "data": {
    "email": {
      "messages": true,
      "leads": true,
      "reviews": true,
      "searchAlerts": true,
      "marketing": false,
      "weeklyReport": true
    },
    "sms": {
      "messages": false,
      "leads": true,
      "urgent": true
    },
    "push": {
      "messages": true,
      "leads": true,
      "reviews": true
    }
  }
}
```

### PUT /api/notifications/preferences

**Request Body:**
```json
{
  "email": {
    "messages": true,
    "marketing": true
  },
  "sms": {
    "leads": false
  }
}
```

**Response:**
```json
{
  "success": true,
  "message": "Preferences updated"
}
```

---

## Background Jobs

| Job | Schedule | Description |
|-----|----------|-------------|
| `sendWeeklyReports` | Sun 08:00 | Send weekly reports to dealers |
| `sendDailyDigests` | Daily 08:00 | Send daily search alert digests |
| `cleanupOldNotifications` | Daily 03:00 | Archive old notification logs |

---

## Error Handling

```javascript
async function sendWithRetry(fn, maxRetries = 3) {
  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    try {
      return await fn();
    } catch (error) {
      if (attempt === maxRetries) throw error;
      
      // Exponential backoff
      await new Promise(resolve => 
        setTimeout(resolve, Math.pow(2, attempt) * 1000)
      );
    }
  }
}
```

---

## Related Services

- **User Service** - Provides user contact info and preferences
- **Messaging Service** - Triggers message notifications
- **Reviews Service** - Triggers review notifications
- **Saved Searches Service** - Triggers search alert notifications
- **Analytics Service** - Provides data for weekly reports

