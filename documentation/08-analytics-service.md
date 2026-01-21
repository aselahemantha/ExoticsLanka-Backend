# Analytics Service

> Provides dealer dashboard metrics and business intelligence.

---

## Overview

The Analytics Service provides comprehensive analytics and insights for dealers to understand their performance, track inventory, and make data-driven decisions. It powers the dealer dashboard with real-time and aggregated metrics.

**Responsibilities:**
- Track listing views and engagement
- Calculate conversion funnels
- Aggregate daily/weekly/monthly metrics
- Generate AI-powered insights
- Market trend analysis
- Inventory performance tracking

---

## Database Tables

### 1. Listing Views Table

```sql
CREATE TABLE listing_views (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    listing_id      UUID NOT NULL REFERENCES car_listings(id) ON DELETE CASCADE,
    user_id         UUID REFERENCES users(id) ON DELETE SET NULL,
    
    -- Tracking Data
    session_id      VARCHAR(100),
    ip_address      VARCHAR(45),
    user_agent      TEXT,
    referrer        TEXT,
    
    -- Event Type
    event_type      VARCHAR(20) DEFAULT 'view'
                    CHECK (event_type IN ('view', 'click', 'contact_view', 'phone_click', 'share')),
    
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_listing_views_listing_id ON listing_views(listing_id);
CREATE INDEX idx_listing_views_user_id ON listing_views(user_id);
CREATE INDEX idx_listing_views_created_at ON listing_views(created_at);
CREATE INDEX idx_listing_views_date ON listing_views(DATE(created_at));
CREATE INDEX idx_listing_views_event ON listing_views(event_type);
```

### 2. Dealer Analytics Table (Daily Aggregates)

```sql
CREATE TABLE dealer_analytics (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dealer_id               UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date                    DATE NOT NULL,
    
    -- Engagement Metrics
    total_views             INT DEFAULT 0,
    unique_viewers          INT DEFAULT 0,
    total_clicks            INT DEFAULT 0,
    total_favorites         INT DEFAULT 0,
    total_shares            INT DEFAULT 0,
    
    -- Conversion Metrics
    total_leads             INT DEFAULT 0,
    total_messages          INT DEFAULT 0,
    phone_reveals           INT DEFAULT 0,
    
    -- Sales Metrics
    total_sales             INT DEFAULT 0,
    total_revenue           DECIMAL(20, 2) DEFAULT 0,
    
    -- Performance Metrics
    avg_response_time_mins  INT,
    response_rate           DECIMAL(5, 2),
    
    -- Inventory Metrics
    inventory_count         INT DEFAULT 0,
    inventory_value         DECIMAL(20, 2) DEFAULT 0,
    avg_health_score        INT DEFAULT 0,
    avg_days_listed         INT DEFAULT 0,
    
    created_at              TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(dealer_id, date)
);

-- Indexes
CREATE INDEX idx_dealer_analytics_dealer_id ON dealer_analytics(dealer_id);
CREATE INDEX idx_dealer_analytics_date ON dealer_analytics(date DESC);
```

### 3. Market Trends Table

```sql
CREATE TABLE market_trends (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date            DATE NOT NULL,
    
    -- Model Trends
    make            VARCHAR(100) NOT NULL,
    model           VARCHAR(100),
    
    -- Metrics
    search_count    INT DEFAULT 0,
    view_count      INT DEFAULT 0,
    listing_count   INT DEFAULT 0,
    avg_price       DECIMAL(15, 2),
    price_change    DECIMAL(5, 2),
    
    UNIQUE(date, make, model)
);

CREATE INDEX idx_market_trends_date ON market_trends(date DESC);
CREATE INDEX idx_market_trends_make ON market_trends(make);
```

---

## API Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| `GET` | `/api/analytics/overview` | Get dashboard overview | Yes (Dealer) |
| `GET` | `/api/analytics/funnel` | Get conversion funnel | Yes (Dealer) |
| `GET` | `/api/analytics/trends` | Get trending models | Yes (Dealer) |
| `GET` | `/api/analytics/heatmap` | Get location heatmap | Yes (Dealer) |
| `GET` | `/api/analytics/depreciation` | Get depreciation chart | Yes (Dealer) |
| `GET` | `/api/analytics/sentiment` | Get review sentiment | Yes (Dealer) |
| `GET` | `/api/analytics/inventory` | Get inventory breakdown | Yes (Dealer) |
| `GET` | `/api/analytics/insights` | Get AI insights | Yes (Dealer) |
| `GET` | `/api/analytics/listing/:id` | Get listing analytics | Yes (Owner) |
| `POST` | `/api/analytics/track` | Track event | No |

---

## Request/Response Examples

### GET /api/analytics/overview

Get dashboard overview statistics.

**Query Parameters:**
```
?period=30d  (7d, 30d, 90d)
```

**Response:**
```json
{
  "success": true,
  "data": {
    "period": "30d",
    "summary": {
      "totalInventory": 24,
      "totalValue": 1240000000,
      "avgDaysToSell": 18,
      "responseTimeAvg": 45,
      "reviewScore": 4.8,
      "inventoryDepreciation": -2.4
    },
    "engagement": {
      "totalViews": 12500,
      "uniqueViewers": 8400,
      "totalFavorites": 840,
      "totalShares": 156,
      "viewsChange": 12.5,
      "favoritesChange": 8.2
    },
    "conversions": {
      "totalLeads": 124,
      "totalMessages": 89,
      "phoneReveals": 67,
      "totalSales": 8,
      "conversionRate": 0.64,
      "leadsChange": 15.3
    },
    "performance": {
      "avgHealthScore": 82,
      "listingsAbove80": 18,
      "listingsBelow60": 2,
      "priceCompetitiveness": 94
    }
  }
}
```

### GET /api/analytics/funnel

Get conversion funnel data.

**Response:**
```json
{
  "success": true,
  "data": {
    "stages": [
      { "name": "Views", "value": 12500, "color": "#3B82F6" },
      { "name": "Clicks", "value": 4500, "color": "#3B82F6" },
      { "name": "Favorites", "value": 840, "color": "#F59E0B" },
      { "name": "Leads", "value": 124, "color": "#10B981" },
      { "name": "Sales", "value": 8, "color": "#10B981" }
    ],
    "conversionRates": {
      "viewToClick": 36.0,
      "clickToFavorite": 18.7,
      "favoriteToLead": 14.8,
      "leadToSale": 6.5,
      "overall": 0.064
    },
    "benchmarks": {
      "viewToClick": { "avg": 32, "top10": 45 },
      "clickToFavorite": { "avg": 15, "top10": 25 },
      "favoriteToLead": { "avg": 12, "top10": 20 },
      "leadToSale": { "avg": 5, "top10": 10 }
    }
  }
}
```

### GET /api/analytics/trends

Get trending models in the market.

**Response:**
```json
{
  "success": true,
  "data": {
    "trendingModels": [
      {
        "model": "Toyota Aqua 2017",
        "growth": 22,
        "searches": 4500,
        "avgPrice": 8500000,
        "priceChange": -2.3,
        "recommendation": "High demand - consider sourcing"
      },
      {
        "model": "Honda Civic Turbo",
        "growth": 18,
        "searches": 3200,
        "avgPrice": 16800000,
        "priceChange": 1.5,
        "recommendation": "Strong interest - good inventory"
      },
      {
        "model": "Toyota Allion 260",
        "growth": 15,
        "searches": 2800,
        "avgPrice": 12500000,
        "priceChange": -1.2,
        "recommendation": "Steady demand"
      },
      {
        "model": "Land Cruiser Sahara",
        "growth": 12,
        "searches": 1900,
        "avgPrice": 95000000,
        "priceChange": 0.8,
        "recommendation": "Premium segment growing"
      },
      {
        "model": "Defender P400",
        "growth": 8,
        "searches": 1200,
        "avgPrice": 68000000,
        "priceChange": -0.5,
        "recommendation": "Niche but stable"
      }
    ],
    "searchKeywords": [
      "Vitz Safety",
      "Allion 260",
      "Land Cruiser Sahara",
      "Defender P400",
      "Civic Turbo",
      "Aqua G"
    ]
  }
}
```

### GET /api/analytics/heatmap

Get buyer location distribution.

**Response:**
```json
{
  "success": true,
  "data": {
    "locations": [
      { "location": "Colombo", "percentage": 45, "buyers": 892, "leads": 56 },
      { "location": "Kandy", "percentage": 20, "buyers": 398, "leads": 25 },
      { "location": "Galle", "percentage": 15, "buyers": 298, "leads": 18 },
      { "location": "Negombo", "percentage": 12, "buyers": 239, "leads": 15 },
      { "location": "Kurunegala", "percentage": 8, "buyers": 159, "leads": 10 }
    ],
    "insights": [
      "45% of your buyers are from Colombo",
      "Consider targeting Kandy with more listings",
      "Galle shows growing interest (+12% this month)"
    ]
  }
}
```

### GET /api/analytics/depreciation

Get inventory depreciation trend.

**Response:**
```json
{
  "success": true,
  "data": {
    "chart": [
      { "date": "Week 1", "value": 1280000000 },
      { "date": "Week 2", "value": 1265000000 },
      { "date": "Week 3", "value": 1252000000 },
      { "date": "Week 4", "value": 1240000000 }
    ],
    "summary": {
      "startValue": 1280000000,
      "currentValue": 1240000000,
      "change": -40000000,
      "changePercent": -3.1,
      "projectedNextMonth": 1210000000
    },
    "byVehicle": [
      {
        "id": "listing-uuid-1",
        "title": "Honda Vezel RS",
        "depreciation": -485000,
        "daysListed": 35,
        "recommendation": "Price reduction recommended"
      },
      {
        "id": "listing-uuid-2",
        "title": "Toyota Land Cruiser",
        "depreciation": -250000,
        "daysListed": 22,
        "recommendation": "Within normal range"
      }
    ]
  }
}
```

### GET /api/analytics/sentiment

Get review sentiment analysis.

**Response:**
```json
{
  "success": true,
  "data": {
    "overall": {
      "positive": 78,
      "neutral": 15,
      "negative": 7
    },
    "breakdown": {
      "communication": { "score": 4.9, "sentiment": "positive" },
      "accuracy": { "score": 4.7, "sentiment": "positive" },
      "professionalism": { "score": 4.8, "sentiment": "positive" },
      "pricing": { "score": 4.2, "sentiment": "neutral" }
    },
    "recentReviews": [
      {
        "rating": 5,
        "sentiment": "positive",
        "highlight": "Very professional and honest dealer",
        "date": "2024-01-15"
      },
      {
        "rating": 4,
        "sentiment": "positive",
        "highlight": "Good experience overall",
        "date": "2024-01-14"
      }
    ],
    "insights": [
      "Your communication is rated higher than 95% of dealers",
      "Consider improving pricing transparency based on feedback"
    ]
  }
}
```

### GET /api/analytics/inventory

Get inventory breakdown and performance.

**Response:**
```json
{
  "success": true,
  "data": {
    "summary": {
      "total": 24,
      "totalValue": 1240000000,
      "avgPrice": 51666667,
      "avgDaysListed": 18
    },
    "byStatus": {
      "active": 20,
      "pending": 2,
      "sold": 8,
      "expired": 2
    },
    "byBrand": [
      { "brand": "Mercedes-Benz", "count": 6, "value": 480000000 },
      { "brand": "BMW", "count": 5, "value": 350000000 },
      { "brand": "Toyota", "count": 8, "value": 280000000 },
      { "brand": "Honda", "count": 5, "value": 130000000 }
    ],
    "byHealthScore": {
      "excellent": { "range": "80-100", "count": 18, "percentage": 75 },
      "good": { "range": "60-79", "count": 4, "percentage": 17 },
      "needsAttention": { "range": "0-59", "count": 2, "percentage": 8 }
    },
    "slowMoving": [
      {
        "id": "listing-uuid-1",
        "title": "Honda Vezel RS",
        "daysListed": 35,
        "views": 2400,
        "healthScore": 58,
        "issue": "Overpriced by 485K"
      }
    ],
    "topPerformers": [
      {
        "id": "listing-uuid-2",
        "title": "Porsche 911 Carrera S",
        "daysListed": 3,
        "views": 2100,
        "healthScore": 98,
        "leads": 12
      }
    ]
  }
}
```

### GET /api/analytics/insights

Get AI-powered insights and recommendations.

**Response:**
```json
{
  "success": true,
  "data": {
    "insights": [
      {
        "type": "prediction",
        "icon": "trending_up",
        "title": "High Demand Alert",
        "message": "Toyota Aqua 2017 demand is up 22% this week. Sourcing recommended.",
        "priority": "high",
        "action": {
          "label": "View Market Data",
          "link": "/dealer/analytics/trends"
        }
      },
      {
        "type": "warning",
        "icon": "warning",
        "title": "Price Alert",
        "message": "Your Honda Vezel is overpriced by LKR 485,000 vs Market Avg.",
        "priority": "warning",
        "action": {
          "label": "Adjust Price",
          "link": "/dealer/inventory/listing-uuid"
        }
      },
      {
        "type": "opportunity",
        "icon": "lightbulb",
        "title": "Buyer Interest",
        "message": "3 buyers searching for Land Cruiser Sahara in your area.",
        "priority": "medium",
        "action": {
          "label": "View Leads",
          "link": "/dealer/inbox"
        }
      },
      {
        "type": "alert",
        "icon": "notification",
        "title": "Response Time",
        "message": "2 unanswered leads are over 4 hours old.",
        "priority": "urgent",
        "action": {
          "label": "Respond Now",
          "link": "/dealer/inbox?filter=unanswered"
        }
      },
      {
        "type": "tip",
        "icon": "photo",
        "title": "Listing Quality",
        "message": "3 listings are missing interior photos. Adding photos can increase views by 40%.",
        "priority": "low",
        "action": {
          "label": "Update Listings",
          "link": "/dealer/inventory?filter=low_quality"
        }
      }
    ],
    "qualityAlerts": [
      { "type": "warning", "message": "3 Listings missing interior photos" },
      { "type": "urgent", "message": "2 Unanswered leads > 4 hours old" },
      { "type": "info", "message": "5 Listings need price review" }
    ]
  }
}
```

### POST /api/analytics/track

Track a view or engagement event.

**Request Body:**
```json
{
  "listingId": "listing-uuid",
  "eventType": "view",
  "sessionId": "session-uuid",
  "referrer": "https://google.com"
}
```

**Response:**
```json
{
  "success": true
}
```

---

## Business Logic

### Aggregating Daily Analytics

```javascript
// Run at 1:00 AM daily
async function aggregateDailyAnalytics() {
  const yesterday = new Date();
  yesterday.setDate(yesterday.getDate() - 1);
  const dateStr = yesterday.toISOString().slice(0, 10);
  
  // Get all dealers
  const dealers = await db.query(
    "SELECT id FROM users WHERE role = 'dealer'"
  );
  
  for (const dealer of dealers.rows) {
    await aggregateDealerDay(dealer.id, dateStr);
  }
}

async function aggregateDealerDay(dealerId, date) {
  // Get engagement metrics
  const engagement = await db.query(`
    SELECT 
      COUNT(*) as total_views,
      COUNT(DISTINCT user_id) as unique_viewers,
      COUNT(*) FILTER (WHERE event_type = 'click') as total_clicks,
      COUNT(*) FILTER (WHERE event_type = 'share') as total_shares,
      COUNT(*) FILTER (WHERE event_type = 'phone_click') as phone_reveals
    FROM listing_views lv
    JOIN car_listings cl ON lv.listing_id = cl.id
    WHERE cl.user_id = $1 AND DATE(lv.created_at) = $2
  `, [dealerId, date]);
  
  // Get favorites count
  const favorites = await db.query(`
    SELECT COUNT(*) as total
    FROM favorites f
    JOIN car_listings cl ON f.listing_id = cl.id
    WHERE cl.user_id = $1 AND DATE(f.created_at) = $2
  `, [dealerId, date]);
  
  // Get leads (new conversations)
  const leads = await db.query(`
    SELECT COUNT(*) as total
    FROM conversations
    WHERE seller_id = $1 AND DATE(created_at) = $2
  `, [dealerId, date]);
  
  // Get messages
  const messages = await db.query(`
    SELECT COUNT(*) as total
    FROM messages m
    JOIN conversations c ON m.conversation_id = c.id
    WHERE c.seller_id = $1 AND DATE(m.created_at) = $2
  `, [dealerId, date]);
  
  // Get inventory metrics
  const inventory = await db.query(`
    SELECT 
      COUNT(*) as count,
      COALESCE(SUM(price), 0) as total_value,
      COALESCE(AVG(health_score), 0) as avg_health,
      COALESCE(AVG(days_listed), 0) as avg_days
    FROM car_listings
    WHERE user_id = $1 AND status = 'active'
  `, [dealerId]);
  
  // Calculate response time
  const responseTime = await db.query(`
    SELECT AVG(EXTRACT(EPOCH FROM (m2.created_at - m1.created_at))/60) as avg_mins
    FROM messages m1
    JOIN messages m2 ON m1.conversation_id = m2.conversation_id
    JOIN conversations c ON m1.conversation_id = c.id
    WHERE c.seller_id = $1
      AND m1.sender_id != $1
      AND m2.sender_id = $1
      AND m2.created_at > m1.created_at
      AND DATE(m1.created_at) = $2
  `, [dealerId, date]);
  
  // Upsert analytics record
  await db.query(`
    INSERT INTO dealer_analytics (
      dealer_id, date,
      total_views, unique_viewers, total_clicks, total_favorites, total_shares,
      total_leads, total_messages, phone_reveals,
      avg_response_time_mins,
      inventory_count, inventory_value, avg_health_score, avg_days_listed
    )
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
    ON CONFLICT (dealer_id, date) DO UPDATE SET
      total_views = EXCLUDED.total_views,
      unique_viewers = EXCLUDED.unique_viewers,
      total_clicks = EXCLUDED.total_clicks,
      total_favorites = EXCLUDED.total_favorites,
      total_shares = EXCLUDED.total_shares,
      total_leads = EXCLUDED.total_leads,
      total_messages = EXCLUDED.total_messages,
      phone_reveals = EXCLUDED.phone_reveals,
      avg_response_time_mins = EXCLUDED.avg_response_time_mins,
      inventory_count = EXCLUDED.inventory_count,
      inventory_value = EXCLUDED.inventory_value,
      avg_health_score = EXCLUDED.avg_health_score,
      avg_days_listed = EXCLUDED.avg_days_listed
  `, [
    dealerId, date,
    engagement.rows[0].total_views,
    engagement.rows[0].unique_viewers,
    engagement.rows[0].total_clicks,
    favorites.rows[0].total,
    engagement.rows[0].total_shares,
    leads.rows[0].total,
    messages.rows[0].total,
    engagement.rows[0].phone_reveals,
    responseTime.rows[0]?.avg_mins || null,
    inventory.rows[0].count,
    inventory.rows[0].total_value,
    Math.round(inventory.rows[0].avg_health),
    Math.round(inventory.rows[0].avg_days)
  ]);
}
```

### Generating AI Insights

```javascript
async function generateInsights(dealerId) {
  const insights = [];
  
  // Check for trending models dealer doesn't have
  const trending = await getTrendingModels();
  const dealerInventory = await getDealerMakes(dealerId);
  
  for (const model of trending.slice(0, 3)) {
    if (!dealerInventory.includes(model.make)) {
      insights.push({
        type: 'prediction',
        icon: 'trending_up',
        title: 'High Demand Alert',
        message: `${model.model} demand is up ${model.growth}% this week. Sourcing recommended.`,
        priority: 'high'
      });
    }
  }
  
  // Check for overpriced listings
  const overpriced = await db.query(`
    SELECT id, title, price, market_avg_price
    FROM car_listings
    WHERE user_id = $1 
      AND status = 'active'
      AND market_avg_price IS NOT NULL
      AND price > market_avg_price * 1.1
  `, [dealerId]);
  
  for (const listing of overpriced.rows) {
    const diff = listing.price - listing.market_avg_price;
    insights.push({
      type: 'warning',
      icon: 'warning',
      title: 'Price Alert',
      message: `Your ${listing.title} is overpriced by LKR ${formatNumber(diff)} vs Market Avg.`,
      priority: 'warning',
      action: { label: 'Adjust Price', link: `/dealer/inventory/${listing.id}` }
    });
  }
  
  // Check for unanswered leads
  const unanswered = await db.query(`
    SELECT COUNT(*) as count
    FROM conversations c
    WHERE c.seller_id = $1
      AND c.seller_unread_count > 0
      AND c.last_message_at < NOW() - INTERVAL '4 hours'
  `, [dealerId]);
  
  if (parseInt(unanswered.rows[0].count) > 0) {
    insights.push({
      type: 'alert',
      icon: 'notification',
      title: 'Response Time',
      message: `${unanswered.rows[0].count} unanswered leads are over 4 hours old.`,
      priority: 'urgent',
      action: { label: 'Respond Now', link: '/dealer/inbox?filter=unanswered' }
    });
  }
  
  // Check for listings missing photos
  const lowQuality = await db.query(`
    SELECT COUNT(*) as count
    FROM car_listings cl
    WHERE cl.user_id = $1
      AND cl.status = 'active'
      AND (SELECT COUNT(*) FROM listing_images WHERE listing_id = cl.id) < 4
  `, [dealerId]);
  
  if (parseInt(lowQuality.rows[0].count) > 0) {
    insights.push({
      type: 'tip',
      icon: 'photo',
      title: 'Listing Quality',
      message: `${lowQuality.rows[0].count} listings are missing interior photos. Adding photos can increase views by 40%.`,
      priority: 'low',
      action: { label: 'Update Listings', link: '/dealer/inventory?filter=low_quality' }
    });
  }
  
  return insights;
}
```

---

## Background Jobs

| Job | Schedule | Description |
|-----|----------|-------------|
| `aggregateDailyAnalytics` | Daily 01:00 | Aggregate previous day's metrics |
| `updateMarketTrends` | Daily 02:00 | Calculate market trends |
| `generateWeeklyReport` | Weekly Sun 08:00 | Email weekly performance report |
| `cleanupOldViews` | Daily 03:00 | Remove views older than 90 days |

---

## Related Services

- **Listings Service** - Provides listing data
- **Messaging Service** - Provides lead and response data
- **Reviews Service** - Provides sentiment data
- **Notification Service** - Sends analytics reports

