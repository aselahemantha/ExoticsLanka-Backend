# 🔐 Auth Service - Quick Reference

**Quick lookup for Auth Service database structure and APIs**

---

## 📊 **DATABASE TABLES SUMMARY**

```
┌─────────────────────────────────────────────────────────────┐
│                   POSTGRESQL TABLES (10)                     │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  1. users                    - Core user authentication      │
│  2. roles                    - RBAC roles                    │
│  3. permissions              - Fine-grained permissions      │
│  4. role_permissions         - Role ↔ Permission mapping     │
│  5. verification_tokens      - Email & reset tokens          │
│  6. password_resets          - Password reset history        │
│  7. login_attempts           - Security audit trail          │
│  8. sessions                 - Active user sessions          │
│  9. audit_logs               - Complete event log            │
│  10. oauth_providers         - OAuth integrations            │
│                                                              │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                   REDIS CACHE KEYS (7)                       │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  session:{token}             - Active user sessions          │
│  refresh:{token}             - Refresh tokens                │
│  blacklist:{token}           - Revoked tokens                │
│  rate_limit:{type}:{id}      - Rate limiting counters        │
│  email_verification:{user}   - Email verification cache      │
│  password_reset:{token}      - Password reset cache          │
│  2fa:{user_id}               - Two-factor auth codes         │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## 🗄️ **KEY TABLE STRUCTURES**

### **`users` Table (Primary)**

| Column | Type | Purpose |
|--------|------|---------|
| `id` | UUID | Primary key |
| `email` | VARCHAR(255) | Login identifier (unique) |
| `password_hash` | VARCHAR(255) | Bcrypt hashed password |
| `status` | VARCHAR(20) | pending, active, suspended, deleted |
| `role` | VARCHAR(50) | buyer, seller, dealer, admin |
| `email_verified` | BOOLEAN | Email confirmation status |
| `two_factor_enabled` | BOOLEAN | 2FA status |
| `failed_login_attempts` | INTEGER | Security counter |
| `locked_until` | TIMESTAMP | Account lock expiry |
| `oauth_provider` | VARCHAR(50) | google, facebook, apple |
| `last_login_at` | TIMESTAMP | Last successful login |

### **`sessions` Table**

| Column | Type | Purpose |
|--------|------|---------|
| `id` | UUID | Primary key |
| `user_id` | UUID | Foreign key → users |
| `token` | VARCHAR(500) | JWT access token |
| `refresh_token` | VARCHAR(500) | JWT refresh token |
| `device_id` | VARCHAR(255) | Device identifier |
| `ip_address` | INET | Login IP |
| `is_active` | BOOLEAN | Session status |
| `expires_at` | TIMESTAMP | Token expiry |

### **`login_attempts` Table**

| Column | Type | Purpose |
|--------|------|---------|
| `id` | BIGSERIAL | Primary key |
| `user_id` | UUID | Foreign key → users (nullable) |
| `email` | VARCHAR(255) | Login email |
| `success` | BOOLEAN | Attempt result |
| `failure_reason` | VARCHAR(100) | Error type |
| `ip_address` | INET | Request IP |
| `risk_score` | INTEGER | 0-100 security score |
| `is_suspicious` | BOOLEAN | Security flag |
| `attempted_at` | TIMESTAMP | Event time |

---

## 🔗 **RELATIONSHIPS**

```
users
  ├── sessions (1:N)
  ├── verification_tokens (1:N)
  ├── login_attempts (1:N)
  ├── password_resets (1:N)
  ├── audit_logs (1:N)
  └── oauth_providers (1:N)

roles
  └── role_permissions (N:M via role_permissions)
        └── permissions
```

---

## 🚀 **API ENDPOINTS**

### **Authentication:**

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/auth/register` | Create new account |
| POST | `/api/auth/login` | Login with credentials |
| POST | `/api/auth/logout` | Logout current session |
| POST | `/api/auth/refresh` | Refresh access token |
| GET | `/api/auth/me` | Get current user |
| POST | `/api/auth/verify-email` | Verify email address |

### **Password Management:**

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/auth/forgot-password` | Request password reset |
| POST | `/api/auth/reset-password` | Complete password reset |
| POST | `/api/auth/change-password` | Change password (authenticated) |

### **OAuth:**

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/auth/google` | Google OAuth login |
| GET | `/api/auth/google/callback` | Google OAuth callback |
| GET | `/api/auth/facebook` | Facebook OAuth login |
| GET | `/api/auth/facebook/callback` | Facebook OAuth callback |

### **Security:**

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/auth/2fa/enable` | Enable two-factor auth |
| POST | `/api/auth/2fa/verify` | Verify 2FA code |
| GET | `/api/auth/sessions` | List active sessions |
| DELETE | `/api/auth/sessions/:id` | Revoke session |

---

## 🔒 **SECURITY FEATURES**

### **Password Security:**
- **Algorithm:** bcrypt
- **Cost Factor:** 12
- **Min Length:** 8 characters
- **Requirements:** Uppercase, lowercase, number

### **Token Management:**
- **Access Token:** 15 minutes (short-lived)
- **Refresh Token:** 30 days
- **Algorithm:** JWT with HS256
- **Storage:** Redis + PostgreSQL

### **Rate Limiting:**
```javascript
Login attempts:     5 per hour per IP
Password reset:     3 per hour per IP
Email verification: 3 per hour per user
API requests:       100 per minute per user
```

### **Account Locking:**
- **Trigger:** 5 failed login attempts
- **Duration:** 1 hour
- **Reset:** Successful login or manual unlock

### **Session Management:**
- **Max Active Sessions:** 5 per user
- **Auto Logout:** 24 hours of inactivity
- **Forced Logout:** On password change

---

## 📊 **DATABASE INDEXES**

### **Critical Indexes:**
```sql
users:
  - idx_users_email (UNIQUE)
  - idx_users_status
  - idx_users_role

sessions:
  - idx_sessions_token (UNIQUE)
  - idx_sessions_user_active (COMPOSITE)

login_attempts:
  - idx_login_attempts_ip
  - idx_login_attempts_email
  - idx_login_attempts_suspicious

verification_tokens:
  - idx_verification_tokens_token (UNIQUE)
  - idx_verification_tokens_user_type (COMPOSITE)
```

---

## 🔧 **IMPLEMENTATION CHECKLIST**

### **Phase 1: Core Setup** ✅
- [ ] Create PostgreSQL database
- [ ] Setup Redis instance
- [ ] Run initial migrations
- [ ] Seed roles and permissions
- [ ] Configure JWT secrets

### **Phase 2: Basic Auth** ✅
- [ ] Implement registration
- [ ] Implement login
- [ ] Implement logout
- [ ] JWT token generation
- [ ] Token validation middleware

### **Phase 3: Email Verification** ✅
- [ ] Generate verification tokens
- [ ] Send verification emails
- [ ] Verify email endpoint
- [ ] Resend verification

### **Phase 4: Password Reset** ✅
- [ ] Forgot password endpoint
- [ ] Generate reset tokens
- [ ] Send reset emails
- [ ] Reset password endpoint
- [ ] Change password endpoint

### **Phase 5: Security** ✅
- [ ] Rate limiting (Redis)
- [ ] Account locking
- [ ] Login attempt tracking
- [ ] IP blacklisting
- [ ] Session management

### **Phase 6: OAuth** ✅
- [ ] Google OAuth integration
- [ ] Facebook OAuth integration
- [ ] Apple OAuth integration
- [ ] Link/unlink providers

### **Phase 7: Advanced** ✅
- [ ] Two-factor authentication
- [ ] Device management
- [ ] Session analytics
- [ ] Risk scoring
- [ ] Audit logging

---

## 💻 **QUICK START COMMANDS**

### **Setup:**
```bash
# Install dependencies
npm install @prisma/client bcrypt jsonwebtoken redis

# Initialize Prisma
npx prisma init

# Run migrations
npx prisma migrate dev

# Generate Prisma client
npx prisma generate

# Seed database
npx prisma db seed
```

### **Development:**
```bash
# Start development server
npm run dev

# Run Redis locally
docker run -d -p 6379:6379 redis:alpine

# Run PostgreSQL locally
docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=password postgres:14
```

### **Testing:**
```bash
# Run tests
npm test

# Run with coverage
npm run test:coverage

# Test specific service
npm test -- auth.test.ts
```

---

## 📈 **PERFORMANCE METRICS**

### **Expected Response Times:**
```
Login:              < 200ms
Registration:       < 300ms
Token validation:   < 50ms  (Redis cache)
Password reset:     < 400ms
OAuth callback:     < 500ms
```

### **Database Queries:**
```
Average per login:     3-4 queries
Average per register:  2-3 queries
Cache hit rate:        > 95% (sessions)
```

### **Scalability:**
```
Concurrent logins:     10,000+/second
Active sessions:       1,000,000+
Database size (1 year): ~1.5 GB
Redis memory:          ~25 MB
```

---

## 🔑 **ENVIRONMENT VARIABLES**

```env
# Database
DATABASE_URL=postgresql://user:pass@localhost:5432/exotics_lanka
REDIS_URL=redis://localhost:6379

# JWT Secrets
JWT_SECRET=your-super-secret-jwt-key-change-this
JWT_REFRESH_SECRET=your-refresh-token-secret-change-this

# OAuth
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
FACEBOOK_APP_ID=your-facebook-app-id
FACEBOOK_APP_SECRET=your-facebook-app-secret

# Email
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password

# App
APP_URL=http://localhost:3000
API_URL=http://localhost:8000
NODE_ENV=development
```

---

## 🎯 **COMMON QUERIES**

### **Check user status:**
```sql
SELECT id, email, status, email_verified, last_login_at 
FROM users 
WHERE email = 'user@example.com';
```

### **Recent failed logins:**
```sql
SELECT email, ip_address, failure_reason, attempted_at 
FROM login_attempts 
WHERE success = FALSE 
ORDER BY attempted_at DESC 
LIMIT 10;
```

### **Active sessions for user:**
```sql
SELECT token, device_name, ip_address, last_activity_at, expires_at 
FROM sessions 
WHERE user_id = 'user-uuid' AND is_active = TRUE;
```

### **Suspicious login attempts:**
```sql
SELECT email, ip_address, COUNT(*) as attempts 
FROM login_attempts 
WHERE success = FALSE 
  AND attempted_at > NOW() - INTERVAL '1 hour'
GROUP BY email, ip_address 
HAVING COUNT(*) > 3;
```

---

## 📚 **RELATED DOCUMENTATION**

- Full Schema: `AUTH_SERVICE_DATABASE_SCHEMA.md`
- Implementation: `AUTH_SERVICE_IMPLEMENTATION_EXAMPLE.md`
- API Specs: `API_SPECIFICATIONS.md`
- Architecture: `MICROSERVICES_ARCHITECTURE.md`

---

**🔐 Auth Service Quick Reference - Ready to Build! 🚀**

