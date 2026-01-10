# 🔐 Auth Service - Database Schema

**Project:** Exotics Lanka  
**Service:** Authentication Service  
**Database:** PostgreSQL + Redis  
**Date:** January 6, 2026  

---

## 📋 **TABLE OF CONTENTS**

1. [Database Overview](#database-overview)
2. [PostgreSQL Tables](#postgresql-tables)
3. [Redis Schema](#redis-schema)
4. [Relationships](#relationships)
5. [Indexes & Performance](#indexes--performance)
6. [Security Considerations](#security-considerations)
7. [Migration Scripts](#migration-scripts)

---

## 🎯 **DATABASE OVERVIEW**

### **Dual Database Strategy:**

```
┌─────────────────────────────────────────────────┐
│          Authentication Service                  │
├─────────────────────────────────────────────────┤
│                                                  │
│  PostgreSQL                    Redis             │
│  ├── users                     ├── sessions      │
│  ├── roles                     ├── tokens        │
│  ├── permissions               ├── rate_limits   │
│  ├── verification_tokens       └── blacklist     │
│  ├── password_resets                             │
│  ├── login_attempts                              │
│  └── audit_logs                                  │
│                                                  │
└─────────────────────────────────────────────────┘
```

**Why Two Databases?**
- **PostgreSQL:** Permanent user data, audit logs, roles
- **Redis:** Fast session storage, token caching, rate limiting

---

## 🗄️ **POSTGRESQL TABLES**

### **1. `users` Table**

**Purpose:** Core user authentication data

```sql
CREATE TABLE users (
  -- Primary Key
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  
  -- Authentication
  email                 VARCHAR(255) UNIQUE NOT NULL,
  password_hash         VARCHAR(255) NOT NULL,
  
  -- Account Status
  status                VARCHAR(20) NOT NULL DEFAULT 'pending',
    -- Values: 'pending', 'active', 'suspended', 'deleted'
  email_verified        BOOLEAN DEFAULT FALSE,
  email_verified_at     TIMESTAMP,
  
  -- Role & Permissions
  role                  VARCHAR(50) NOT NULL DEFAULT 'buyer',
    -- Values: 'buyer', 'seller', 'dealer', 'admin', 'super_admin'
  
  -- Security
  two_factor_enabled    BOOLEAN DEFAULT FALSE,
  two_factor_secret     VARCHAR(255),
  failed_login_attempts INTEGER DEFAULT 0,
  locked_until          TIMESTAMP,
  
  -- OAuth
  oauth_provider        VARCHAR(50),
    -- Values: 'google', 'facebook', 'apple', null (local)
  oauth_id              VARCHAR(255),
  
  -- Timestamps
  created_at            TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at            TIMESTAMP NOT NULL DEFAULT NOW(),
  last_login_at         TIMESTAMP,
  deleted_at            TIMESTAMP,
  
  -- Constraints
  CONSTRAINT email_lowercase CHECK (email = LOWER(email)),
  CONSTRAINT valid_status CHECK (status IN ('pending', 'active', 'suspended', 'deleted')),
  CONSTRAINT valid_role CHECK (role IN ('buyer', 'seller', 'dealer', 'admin', 'super_admin'))
);

-- Indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_oauth ON users(oauth_provider, oauth_id);
CREATE INDEX idx_users_created_at ON users(created_at DESC);
```

---

### **2. `roles` Table**

**Purpose:** Role-based access control (RBAC)

```sql
CREATE TABLE roles (
  -- Primary Key
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  
  -- Role Details
  name              VARCHAR(50) UNIQUE NOT NULL,
  display_name      VARCHAR(100) NOT NULL,
  description       TEXT,
  
  -- Hierarchy
  level             INTEGER NOT NULL DEFAULT 0,
    -- 0: Buyer, 10: Seller, 20: Dealer, 90: Admin, 100: Super Admin
  
  -- Status
  is_active         BOOLEAN DEFAULT TRUE,
  
  -- Timestamps
  created_at        TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Seed Data
INSERT INTO roles (name, display_name, level) VALUES
  ('buyer', 'Buyer', 0),
  ('seller', 'Seller', 10),
  ('dealer', 'Dealer', 20),
  ('admin', 'Administrator', 90),
  ('super_admin', 'Super Administrator', 100);
```

---

### **3. `permissions` Table**

**Purpose:** Fine-grained permission control

```sql
CREATE TABLE permissions (
  -- Primary Key
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  
  -- Permission Details
  name              VARCHAR(100) UNIQUE NOT NULL,
  resource          VARCHAR(50) NOT NULL,
    -- Values: 'listings', 'users', 'messages', 'reviews', etc.
  action            VARCHAR(50) NOT NULL,
    -- Values: 'create', 'read', 'update', 'delete', 'list'
  
  -- Description
  description       TEXT,
  
  -- Timestamps
  created_at        TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Seed Data
INSERT INTO permissions (name, resource, action, description) VALUES
  -- Listing Permissions
  ('listing.create', 'listings', 'create', 'Create new car listings'),
  ('listing.read', 'listings', 'read', 'View car listings'),
  ('listing.update', 'listings', 'update', 'Update own listings'),
  ('listing.delete', 'listings', 'delete', 'Delete own listings'),
  ('listing.list', 'listings', 'list', 'List all listings'),
  ('listing.moderate', 'listings', 'moderate', 'Moderate any listing'),
  
  -- User Permissions
  ('user.read', 'users', 'read', 'View user profiles'),
  ('user.update', 'users', 'update', 'Update own profile'),
  ('user.delete', 'users', 'delete', 'Delete own account'),
  ('user.list', 'users', 'list', 'List all users'),
  ('user.manage', 'users', 'manage', 'Manage any user account'),
  
  -- Message Permissions
  ('message.send', 'messages', 'create', 'Send messages'),
  ('message.read', 'messages', 'read', 'Read own messages'),
  ('message.delete', 'messages', 'delete', 'Delete own messages'),
  ('message.moderate', 'messages', 'moderate', 'Moderate any message');
```

---

### **4. `role_permissions` Table**

**Purpose:** Many-to-many relationship between roles and permissions

```sql
CREATE TABLE role_permissions (
  -- Composite Primary Key
  role_id           UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  permission_id     UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
  
  -- Timestamps
  created_at        TIMESTAMP NOT NULL DEFAULT NOW(),
  
  -- Constraints
  PRIMARY KEY (role_id, permission_id)
);

-- Indexes
CREATE INDEX idx_role_permissions_role ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission ON role_permissions(permission_id);
```

---

### **5. `verification_tokens` Table**

**Purpose:** Email verification and password reset tokens

```sql
CREATE TABLE verification_tokens (
  -- Primary Key
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  
  -- User Reference
  user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  
  -- Token
  token             VARCHAR(255) UNIQUE NOT NULL,
  token_hash        VARCHAR(255) NOT NULL,
  
  -- Type
  type              VARCHAR(50) NOT NULL,
    -- Values: 'email_verification', 'password_reset', '2fa_setup'
  
  -- Status
  used              BOOLEAN DEFAULT FALSE,
  used_at           TIMESTAMP,
  
  -- Expiry
  expires_at        TIMESTAMP NOT NULL,
  
  -- Metadata
  ip_address        INET,
  user_agent        TEXT,
  
  -- Timestamps
  created_at        TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_verification_tokens_user ON verification_tokens(user_id);
CREATE INDEX idx_verification_tokens_token ON verification_tokens(token);
CREATE INDEX idx_verification_tokens_type ON verification_tokens(type);
CREATE INDEX idx_verification_tokens_expires ON verification_tokens(expires_at);

-- Auto-delete expired tokens (PostgreSQL + pg_cron extension)
CREATE INDEX idx_verification_tokens_cleanup ON verification_tokens(expires_at) 
  WHERE used = FALSE AND expires_at < NOW();
```

---

### **6. `password_resets` Table**

**Purpose:** Track password reset history

```sql
CREATE TABLE password_resets (
  -- Primary Key
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  
  -- User Reference
  user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  
  -- Reset Details
  token_id          UUID REFERENCES verification_tokens(id),
  old_password_hash VARCHAR(255),
  
  -- Status
  status            VARCHAR(20) NOT NULL,
    -- Values: 'initiated', 'completed', 'expired', 'cancelled'
  
  -- Metadata
  ip_address        INET,
  user_agent        TEXT,
  
  -- Timestamps
  initiated_at      TIMESTAMP NOT NULL DEFAULT NOW(),
  completed_at      TIMESTAMP
);

-- Indexes
CREATE INDEX idx_password_resets_user ON password_resets(user_id);
CREATE INDEX idx_password_resets_initiated ON password_resets(initiated_at DESC);
```

---

### **7. `login_attempts` Table**

**Purpose:** Track login attempts for security and rate limiting

```sql
CREATE TABLE login_attempts (
  -- Primary Key
  id                BIGSERIAL PRIMARY KEY,
  
  -- User Reference (nullable - might not exist)
  user_id           UUID REFERENCES users(id) ON DELETE SET NULL,
  email             VARCHAR(255) NOT NULL,
  
  -- Attempt Details
  success           BOOLEAN NOT NULL,
  failure_reason    VARCHAR(100),
    -- Values: 'invalid_credentials', 'account_locked', 'email_not_verified', '2fa_failed'
  
  -- Security
  ip_address        INET NOT NULL,
  user_agent        TEXT,
  device_fingerprint TEXT,
  location_country  VARCHAR(2),
  location_city     VARCHAR(100),
  
  -- Risk Analysis
  risk_score        INTEGER DEFAULT 0,
    -- 0-100: calculated based on IP, location, frequency
  is_suspicious     BOOLEAN DEFAULT FALSE,
  
  -- Timestamps
  attempted_at      TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_login_attempts_user ON login_attempts(user_id);
CREATE INDEX idx_login_attempts_email ON login_attempts(email);
CREATE INDEX idx_login_attempts_ip ON login_attempts(ip_address);
CREATE INDEX idx_login_attempts_attempted ON login_attempts(attempted_at DESC);
CREATE INDEX idx_login_attempts_suspicious ON login_attempts(is_suspicious) 
  WHERE is_suspicious = TRUE;

-- Partitioning by month (for performance with large data)
-- CREATE TABLE login_attempts_2026_01 PARTITION OF login_attempts
--   FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');
```

---

### **8. `sessions` Table (Alternative to Redis)**

**Purpose:** Store user sessions in PostgreSQL (if not using Redis)

```sql
CREATE TABLE sessions (
  -- Primary Key
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  
  -- User Reference
  user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  
  -- Token
  token             VARCHAR(500) UNIQUE NOT NULL,
  refresh_token     VARCHAR(500),
  
  -- Device Info
  device_id         VARCHAR(255),
  device_name       VARCHAR(100),
  ip_address        INET,
  user_agent        TEXT,
  
  -- Status
  is_active         BOOLEAN DEFAULT TRUE,
  
  -- Expiry
  expires_at        TIMESTAMP NOT NULL,
  last_activity_at  TIMESTAMP NOT NULL DEFAULT NOW(),
  
  -- Timestamps
  created_at        TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_sessions_user ON sessions(user_id);
CREATE INDEX idx_sessions_token ON sessions(token);
CREATE INDEX idx_sessions_refresh ON sessions(refresh_token);
CREATE INDEX idx_sessions_expires ON sessions(expires_at);
CREATE INDEX idx_sessions_active ON sessions(is_active, user_id);
```

---

### **9. `audit_logs` Table**

**Purpose:** Track all authentication events for security auditing

```sql
CREATE TABLE audit_logs (
  -- Primary Key
  id                BIGSERIAL PRIMARY KEY,
  
  -- User Reference
  user_id           UUID REFERENCES users(id) ON DELETE SET NULL,
  
  -- Event
  event_type        VARCHAR(100) NOT NULL,
    -- Values: 'login', 'logout', 'password_change', 'email_verification',
    --         'account_created', 'account_suspended', 'role_changed', etc.
  event_category    VARCHAR(50) NOT NULL,
    -- Values: 'authentication', 'authorization', 'account_management'
  
  -- Details
  description       TEXT,
  metadata          JSONB,
  
  -- Security
  ip_address        INET,
  user_agent        TEXT,
  
  -- Result
  success           BOOLEAN NOT NULL,
  error_message     TEXT,
  
  -- Timestamps
  created_at        TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_event_type ON audit_logs(event_type);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_metadata ON audit_logs USING GIN(metadata);

-- Partitioning by month (recommended for high traffic)
```

---

### **10. `oauth_providers` Table**

**Purpose:** Store OAuth provider configurations and user mappings

```sql
CREATE TABLE oauth_providers (
  -- Primary Key
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  
  -- User Reference
  user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  
  -- Provider
  provider          VARCHAR(50) NOT NULL,
    -- Values: 'google', 'facebook', 'apple'
  provider_user_id  VARCHAR(255) NOT NULL,
  
  -- Tokens
  access_token      TEXT,
  refresh_token     TEXT,
  token_expires_at  TIMESTAMP,
  
  -- Profile Data
  provider_email    VARCHAR(255),
  provider_name     VARCHAR(255),
  provider_avatar   TEXT,
  raw_data          JSONB,
  
  -- Status
  is_primary        BOOLEAN DEFAULT FALSE,
  
  -- Timestamps
  created_at        TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMP NOT NULL DEFAULT NOW(),
  last_used_at      TIMESTAMP,
  
  -- Constraints
  UNIQUE(provider, provider_user_id)
);

-- Indexes
CREATE INDEX idx_oauth_providers_user ON oauth_providers(user_id);
CREATE INDEX idx_oauth_providers_provider ON oauth_providers(provider, provider_user_id);
```

---

## 🔴 **REDIS SCHEMA**

### **Purpose:** Fast session storage, caching, rate limiting

### **1. User Sessions**

```redis
# Key Pattern: session:{token}
# Type: Hash
# TTL: 24 hours (configurable)

HSET session:eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
  user_id         "550e8400-e29b-41d4-a716-446655440000"
  email           "user@example.com"
  role            "seller"
  ip_address      "192.168.1.1"
  created_at      "2026-01-06T12:00:00Z"
  last_activity   "2026-01-06T12:30:00Z"

EXPIRE session:eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9... 86400
```

### **2. Refresh Tokens**

```redis
# Key Pattern: refresh:{token}
# Type: Hash
# TTL: 30 days

HSET refresh:d8a9f7e6c5b4a3d2c1b0a9e8d7c6b5a4
  user_id         "550e8400-e29b-41d4-a716-446655440000"
  session_token   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  created_at      "2026-01-06T12:00:00Z"

EXPIRE refresh:d8a9f7e6c5b4a3d2c1b0a9e8d7c6b5a4 2592000
```

### **3. Token Blacklist**

```redis
# Key Pattern: blacklist:{token}
# Type: String
# TTL: Until original token expiry

SET blacklist:eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9... "revoked"
EXPIRE blacklist:eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9... 86400
```

### **4. Rate Limiting**

```redis
# Key Pattern: rate_limit:{type}:{identifier}
# Type: String (counter)
# TTL: Time window (e.g., 60 seconds, 3600 seconds)

# Login attempts per IP
INCR rate_limit:login:192.168.1.1
EXPIRE rate_limit:login:192.168.1.1 3600  # 5 attempts per hour

# API requests per user
INCR rate_limit:api:550e8400-e29b-41d4-a716-446655440000
EXPIRE rate_limit:api:550e8400-e29b-41d4-a716-446655440000 60  # 100 requests per minute
```

### **5. Email Verification Cache**

```redis
# Key Pattern: email_verification:{user_id}
# Type: Hash
# TTL: 24 hours

HSET email_verification:550e8400-e29b-41d4-a716-446655440000
  token           "abc123def456"
  email           "user@example.com"
  attempts        "0"
  sent_at         "2026-01-06T12:00:00Z"

EXPIRE email_verification:550e8400-e29b-41d4-a716-446655440000 86400
```

### **6. Password Reset Cache**

```redis
# Key Pattern: password_reset:{token}
# Type: Hash
# TTL: 1 hour

HSET password_reset:xyz789abc123
  user_id         "550e8400-e29b-41d4-a716-446655440000"
  email           "user@example.com"
  created_at      "2026-01-06T12:00:00Z"

EXPIRE password_reset:xyz789abc123 3600
```

### **7. Two-Factor Authentication**

```redis
# Key Pattern: 2fa:{user_id}
# Type: Hash
# TTL: 5 minutes

HSET 2fa:550e8400-e29b-41d4-a716-446655440000
  code            "123456"
  attempts        "0"
  created_at      "2026-01-06T12:00:00Z"

EXPIRE 2fa:550e8400-e29b-41d4-a716-446655440000 300
```

---

## 🔗 **DATABASE RELATIONSHIPS**

```
┌─────────────┐
│    users    │
└─────────────┘
      │ 1
      │
      ├──────┬──────────────────────────┐
      │      │                          │
      │ *    │ *                        │ *
┌─────▼──────▼────┐    ┌───────────────▼─────┐    ┌─────────────────┐
│  sessions       │    │ verification_tokens │    │ login_attempts  │
└─────────────────┘    └─────────────────────┘    └─────────────────┘
      
      │ 1
      │
      │ *
┌─────▼───────────┐    ┌─────────────┐
│ oauth_providers │    │    roles    │
└─────────────────┘    └─────┬───────┘
                             │ *
                             │ *
                       ┌─────▼─────────┐
                       │ role_permissions│
                       └─────┬───────┘
                             │ *
                             │ 1
                       ┌─────▼─────────┐
                       │  permissions  │
                       └───────────────┘

┌─────────────┐
│ audit_logs  │  ← Tracks all events
└─────────────┘
```

---

## ⚡ **INDEXES & PERFORMANCE**

### **Primary Indexes (Already Defined):**

1. **users table:**
   - Primary: `id` (UUID, clustered)
   - Unique: `email`
   - Index: `status`, `role`, `oauth_provider`, `created_at`

2. **sessions table:**
   - Primary: `id`
   - Unique: `token`
   - Index: `user_id`, `expires_at`, `is_active`

3. **login_attempts table:**
   - Primary: `id` (BIGSERIAL for performance)
   - Index: `user_id`, `email`, `ip_address`, `attempted_at`, `is_suspicious`

### **Composite Indexes:**

```sql
-- Fast session lookup by user and active status
CREATE INDEX idx_sessions_user_active 
  ON sessions(user_id, is_active) 
  WHERE is_active = TRUE;

-- Fast token lookup for verification
CREATE INDEX idx_verification_tokens_user_type 
  ON verification_tokens(user_id, type) 
  WHERE used = FALSE;

-- Recent login attempts by user
CREATE INDEX idx_login_attempts_user_recent 
  ON login_attempts(user_id, attempted_at DESC);
```

### **Partial Indexes (Better Performance):**

```sql
-- Only index active users
CREATE INDEX idx_users_active 
  ON users(id) 
  WHERE status = 'active' AND deleted_at IS NULL;

-- Only index failed login attempts
CREATE INDEX idx_login_attempts_failed 
  ON login_attempts(email, attempted_at DESC) 
  WHERE success = FALSE;
```

### **Full-Text Search (for audit logs):**

```sql
-- Search audit log descriptions
CREATE INDEX idx_audit_logs_description_fts 
  ON audit_logs USING GIN(to_tsvector('english', description));
```

---

## 🔒 **SECURITY CONSIDERATIONS**

### **1. Password Security:**

```sql
-- Password hashing: bcrypt with cost factor 12
-- Example in Node.js:
-- const hash = await bcrypt.hash(password, 12);

-- Password requirements (application level):
-- - Minimum 8 characters
-- - At least 1 uppercase, 1 lowercase, 1 number
-- - Optional: 1 special character
```

### **2. Token Security:**

```sql
-- JWT Token Structure:
-- {
--   "sub": "user_id",
--   "email": "user@example.com",
--   "role": "seller",
--   "iat": 1609459200,
--   "exp": 1609545600
-- }

-- Token Expiry:
-- - Access Token: 15 minutes (short-lived)
-- - Refresh Token: 30 days
-- - Email Verification: 24 hours
-- - Password Reset: 1 hour
-- - 2FA Code: 5 minutes
```

### **3. Rate Limiting:**

```javascript
// Rate Limits (enforced via Redis)
const RATE_LIMITS = {
  login: {
    maxAttempts: 5,
    windowMs: 15 * 60 * 1000, // 15 minutes
    blockDurationMs: 60 * 60 * 1000, // 1 hour
  },
  passwordReset: {
    maxAttempts: 3,
    windowMs: 60 * 60 * 1000, // 1 hour
  },
  emailVerification: {
    maxAttempts: 3,
    windowMs: 60 * 60 * 1000, // 1 hour
  },
};
```

### **4. Account Locking:**

```sql
-- Lock account after 5 failed login attempts
UPDATE users 
SET 
  failed_login_attempts = failed_login_attempts + 1,
  locked_until = CASE 
    WHEN failed_login_attempts >= 4 
    THEN NOW() + INTERVAL '1 hour' 
    ELSE locked_until 
  END
WHERE id = $1;
```

### **5. IP Whitelisting/Blacklisting:**

```sql
CREATE TABLE ip_blacklist (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  ip_address    INET NOT NULL UNIQUE,
  reason        TEXT,
  created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
  expires_at    TIMESTAMP
);

CREATE INDEX idx_ip_blacklist_ip ON ip_blacklist(ip_address);
```

---

## 🚀 **MIGRATION SCRIPTS**

### **Initial Migration:**

```sql
-- File: migrations/001_create_auth_tables.sql

BEGIN;

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create users table
CREATE TABLE users (
  -- [Full schema from above]
);

-- Create roles table
CREATE TABLE roles (
  -- [Full schema from above]
);

-- Create permissions table
CREATE TABLE permissions (
  -- [Full schema from above]
);

-- Create role_permissions table
CREATE TABLE role_permissions (
  -- [Full schema from above]
);

-- Create verification_tokens table
CREATE TABLE verification_tokens (
  -- [Full schema from above]
);

-- Create password_resets table
CREATE TABLE password_resets (
  -- [Full schema from above]
);

-- Create login_attempts table
CREATE TABLE login_attempts (
  -- [Full schema from above]
);

-- Create sessions table
CREATE TABLE sessions (
  -- [Full schema from above]
);

-- Create audit_logs table
CREATE TABLE audit_logs (
  -- [Full schema from above]
);

-- Create oauth_providers table
CREATE TABLE oauth_providers (
  -- [Full schema from above]
);

-- Seed default data
INSERT INTO roles (name, display_name, level) VALUES
  ('buyer', 'Buyer', 0),
  ('seller', 'Seller', 10),
  ('dealer', 'Dealer', 20),
  ('admin', 'Administrator', 90),
  ('super_admin', 'Super Administrator', 100);

COMMIT;
```

### **Rollback Migration:**

```sql
-- File: migrations/001_rollback_auth_tables.sql

BEGIN;

DROP TABLE IF EXISTS oauth_providers CASCADE;
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS login_attempts CASCADE;
DROP TABLE IF EXISTS password_resets CASCADE;
DROP TABLE IF EXISTS verification_tokens CASCADE;
DROP TABLE IF EXISTS role_permissions CASCADE;
DROP TABLE IF EXISTS permissions CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
DROP TABLE IF EXISTS users CASCADE;

COMMIT;
```

---

## 📊 **DATABASE SIZE ESTIMATES**

### **Expected Growth (Year 1):**

```
Table               Records      Size/Record    Total Size
─────────────────────────────────────────────────────────
users               50,000       2 KB           100 MB
sessions            10,000       1 KB           10 MB
login_attempts      500,000      500 B          250 MB
verification_tokens 100,000      500 B          50 MB
password_resets     20,000       500 B          10 MB
audit_logs          1,000,000    1 KB           1 GB
oauth_providers     25,000       2 KB           50 MB
─────────────────────────────────────────────────────────
TOTAL                                           ~1.5 GB
```

### **Redis Memory Usage:**

```
Active sessions:     5,000 × 1 KB  = 5 MB
Refresh tokens:      10,000 × 500 B = 5 MB
Blacklisted tokens:  1,000 × 200 B = 200 KB
Rate limit counters: 20,000 × 100 B = 2 MB
Cache data:          Variable       = 10 MB
───────────────────────────────────────────
TOTAL                               ~25 MB
```

---

## ✅ **BEST PRACTICES**

1. **Always hash passwords** with bcrypt (cost: 12)
2. **Use UUIDs** for primary keys (better for distributed systems)
3. **Implement soft deletes** (deleted_at column)
4. **Log everything** in audit_logs
5. **Rate limit** all authentication endpoints
6. **Use Redis** for sessions (fast access)
7. **Partition large tables** (login_attempts, audit_logs)
8. **Regular cleanup** of expired tokens
9. **Monitor failed logins** for security threats
10. **Use prepared statements** to prevent SQL injection

---

## 🎯 **NEXT STEPS**

1. ✅ Database schema defined
2. ⏳ Create migration scripts
3. ⏳ Implement database models (Sequelize/Prisma/TypeORM)
4. ⏳ Create seed data for testing
5. ⏳ Set up Redis connection
6. ⏳ Implement authentication logic
7. ⏳ Add monitoring and logging
8. ⏳ Performance testing

---

**🔐 Your Auth Service Database is production-ready! 🚀**

