# 🔐 Auth Service - Implementation Examples

**Project:** Exotics Lanka  
**Language:** Node.js + TypeScript  
**ORM:** Prisma (recommended)  
**Date:** January 6, 2026  

---

## 📋 **QUICK START**

### **1. Prisma Schema Setup**

```prisma
// prisma/schema.prisma

generator client {
  provider = "prisma-client-js"
}

datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
}

model User {
  id                    String    @id @default(uuid())
  email                 String    @unique
  passwordHash          String    @map("password_hash")
  
  // Status
  status                String    @default("pending")
  emailVerified         Boolean   @default(false) @map("email_verified")
  emailVerifiedAt       DateTime? @map("email_verified_at")
  
  // Role
  role                  String    @default("buyer")
  
  // Security
  twoFactorEnabled      Boolean   @default(false) @map("two_factor_enabled")
  twoFactorSecret       String?   @map("two_factor_secret")
  failedLoginAttempts   Int       @default(0) @map("failed_login_attempts")
  lockedUntil           DateTime? @map("locked_until")
  
  // OAuth
  oauthProvider         String?   @map("oauth_provider")
  oauthId               String?   @map("oauth_id")
  
  // Timestamps
  createdAt             DateTime  @default(now()) @map("created_at")
  updatedAt             DateTime  @updatedAt @map("updated_at")
  lastLoginAt           DateTime? @map("last_login_at")
  deletedAt             DateTime? @map("deleted_at")
  
  // Relations
  sessions              Session[]
  verificationTokens    VerificationToken[]
  loginAttempts         LoginAttempt[]
  passwordResets        PasswordReset[]
  auditLogs             AuditLog[]
  oauthProviders        OAuthProvider[]
  
  @@index([email])
  @@index([status])
  @@index([role])
  @@map("users")
}

model Session {
  id                String    @id @default(uuid())
  userId            String    @map("user_id")
  token             String    @unique
  refreshToken      String?   @map("refresh_token")
  
  deviceId          String?   @map("device_id")
  deviceName        String?   @map("device_name")
  ipAddress         String?   @map("ip_address")
  userAgent         String?   @map("user_agent")
  
  isActive          Boolean   @default(true) @map("is_active")
  expiresAt         DateTime  @map("expires_at")
  lastActivityAt    DateTime  @default(now()) @map("last_activity_at")
  createdAt         DateTime  @default(now()) @map("created_at")
  
  user              User      @relation(fields: [userId], references: [id], onDelete: Cascade)
  
  @@index([userId])
  @@index([token])
  @@index([expiresAt])
  @@map("sessions")
}

model VerificationToken {
  id                String    @id @default(uuid())
  userId            String    @map("user_id")
  token             String    @unique
  tokenHash         String    @map("token_hash")
  type              String
  
  used              Boolean   @default(false)
  usedAt            DateTime? @map("used_at")
  expiresAt         DateTime  @map("expires_at")
  
  ipAddress         String?   @map("ip_address")
  userAgent         String?   @map("user_agent")
  createdAt         DateTime  @default(now()) @map("created_at")
  
  user              User      @relation(fields: [userId], references: [id], onDelete: Cascade)
  
  @@index([userId])
  @@index([token])
  @@index([type])
  @@map("verification_tokens")
}

model LoginAttempt {
  id                BigInt    @id @default(autoincrement())
  userId            String?   @map("user_id")
  email             String
  
  success           Boolean
  failureReason     String?   @map("failure_reason")
  
  ipAddress         String    @map("ip_address")
  userAgent         String?   @map("user_agent")
  deviceFingerprint String?   @map("device_fingerprint")
  locationCountry   String?   @map("location_country")
  locationCity      String?   @map("location_city")
  
  riskScore         Int       @default(0) @map("risk_score")
  isSuspicious      Boolean   @default(false) @map("is_suspicious")
  attemptedAt       DateTime  @default(now()) @map("attempted_at")
  
  user              User?     @relation(fields: [userId], references: [id], onDelete: SetNull)
  
  @@index([userId])
  @@index([email])
  @@index([ipAddress])
  @@index([attemptedAt])
  @@map("login_attempts")
}

model AuditLog {
  id                BigInt    @id @default(autoincrement())
  userId            String?   @map("user_id")
  
  eventType         String    @map("event_type")
  eventCategory     String    @map("event_category")
  description       String?
  metadata          Json?
  
  ipAddress         String?   @map("ip_address")
  userAgent         String?   @map("user_agent")
  
  success           Boolean
  errorMessage      String?   @map("error_message")
  createdAt         DateTime  @default(now()) @map("created_at")
  
  user              User?     @relation(fields: [userId], references: [id], onDelete: SetNull)
  
  @@index([userId])
  @@index([eventType])
  @@index([createdAt])
  @@map("audit_logs")
}

model PasswordReset {
  id                String    @id @default(uuid())
  userId            String    @map("user_id")
  tokenId           String?   @map("token_id")
  oldPasswordHash   String?   @map("old_password_hash")
  
  status            String
  ipAddress         String?   @map("ip_address")
  userAgent         String?   @map("user_agent")
  
  initiatedAt       DateTime  @default(now()) @map("initiated_at")
  completedAt       DateTime? @map("completed_at")
  
  user              User      @relation(fields: [userId], references: [id], onDelete: Cascade)
  
  @@index([userId])
  @@map("password_resets")
}

model OAuthProvider {
  id                String    @id @default(uuid())
  userId            String    @map("user_id")
  
  provider          String
  providerUserId    String    @map("provider_user_id")
  
  accessToken       String?   @map("access_token")
  refreshToken      String?   @map("refresh_token")
  tokenExpiresAt    DateTime? @map("token_expires_at")
  
  providerEmail     String?   @map("provider_email")
  providerName      String?   @map("provider_name")
  providerAvatar    String?   @map("provider_avatar")
  rawData           Json?     @map("raw_data")
  
  isPrimary         Boolean   @default(false) @map("is_primary")
  createdAt         DateTime  @default(now()) @map("created_at")
  updatedAt         DateTime  @updatedAt @map("updated_at")
  lastUsedAt        DateTime? @map("last_used_at")
  
  user              User      @relation(fields: [userId], references: [id], onDelete: Cascade)
  
  @@unique([provider, providerUserId])
  @@index([userId])
  @@map("oauth_providers")
}
```

---

## 🔧 **IMPLEMENTATION EXAMPLES**

### **1. User Registration**

```typescript
// src/services/auth/register.service.ts

import { PrismaClient } from '@prisma/client';
import bcrypt from 'bcrypt';
import { v4 as uuidv4 } from 'uuid';
import { sendVerificationEmail } from '../email/email.service';
import { logAudit } from '../audit/audit.service';

const prisma = new PrismaClient();

interface RegisterInput {
  email: string;
  password: string;
  role?: 'buyer' | 'seller';
  ipAddress?: string;
  userAgent?: string;
}

export async function registerUser(input: RegisterInput) {
  const { email, password, role = 'buyer', ipAddress, userAgent } = input;
  
  // 1. Check if user exists
  const existingUser = await prisma.user.findUnique({
    where: { email: email.toLowerCase() },
  });
  
  if (existingUser) {
    throw new Error('Email already registered');
  }
  
  // 2. Hash password
  const passwordHash = await bcrypt.hash(password, 12);
  
  // 3. Create user
  const user = await prisma.user.create({
    data: {
      email: email.toLowerCase(),
      passwordHash,
      role,
      status: 'pending',
    },
  });
  
  // 4. Generate verification token
  const token = uuidv4();
  const tokenHash = await bcrypt.hash(token, 10);
  
  await prisma.verificationToken.create({
    data: {
      userId: user.id,
      token,
      tokenHash,
      type: 'email_verification',
      expiresAt: new Date(Date.now() + 24 * 60 * 60 * 1000), // 24 hours
      ipAddress,
      userAgent,
    },
  });
  
  // 5. Send verification email
  await sendVerificationEmail(user.email, token);
  
  // 6. Log audit
  await logAudit({
    userId: user.id,
    eventType: 'account_created',
    eventCategory: 'account_management',
    description: `New ${role} account created`,
    ipAddress,
    userAgent,
    success: true,
  });
  
  return {
    id: user.id,
    email: user.email,
    role: user.role,
    status: user.status,
  };
}
```

---

### **2. User Login**

```typescript
// src/services/auth/login.service.ts

import { PrismaClient } from '@prisma/client';
import bcrypt from 'bcrypt';
import jwt from 'jsonwebtoken';
import { redisClient } from '../redis/redis.client';
import { logAudit } from '../audit/audit.service';

const prisma = new PrismaClient();

interface LoginInput {
  email: string;
  password: string;
  ipAddress?: string;
  userAgent?: string;
  deviceId?: string;
}

export async function loginUser(input: LoginInput) {
  const { email, password, ipAddress, userAgent, deviceId } = input;
  
  // 1. Rate limiting check (Redis)
  const rateLimitKey = `rate_limit:login:${ipAddress}`;
  const attempts = await redisClient.incr(rateLimitKey);
  
  if (attempts === 1) {
    await redisClient.expire(rateLimitKey, 3600); // 1 hour window
  }
  
  if (attempts > 5) {
    await logLoginAttempt({
      email,
      success: false,
      failureReason: 'rate_limit_exceeded',
      ipAddress,
      userAgent,
    });
    
    throw new Error('Too many login attempts. Please try again later.');
  }
  
  // 2. Find user
  const user = await prisma.user.findUnique({
    where: { email: email.toLowerCase() },
  });
  
  if (!user) {
    await logLoginAttempt({
      email,
      success: false,
      failureReason: 'user_not_found',
      ipAddress,
      userAgent,
    });
    
    throw new Error('Invalid credentials');
  }
  
  // 3. Check if account is locked
  if (user.lockedUntil && user.lockedUntil > new Date()) {
    await logLoginAttempt({
      userId: user.id,
      email,
      success: false,
      failureReason: 'account_locked',
      ipAddress,
      userAgent,
    });
    
    throw new Error('Account is locked. Please try again later.');
  }
  
  // 4. Check if email is verified
  if (!user.emailVerified) {
    await logLoginAttempt({
      userId: user.id,
      email,
      success: false,
      failureReason: 'email_not_verified',
      ipAddress,
      userAgent,
    });
    
    throw new Error('Please verify your email first');
  }
  
  // 5. Verify password
  const passwordValid = await bcrypt.compare(password, user.passwordHash);
  
  if (!passwordValid) {
    // Increment failed attempts
    await prisma.user.update({
      where: { id: user.id },
      data: {
        failedLoginAttempts: { increment: 1 },
        lockedUntil:
          user.failedLoginAttempts >= 4
            ? new Date(Date.now() + 60 * 60 * 1000) // Lock for 1 hour
            : undefined,
      },
    });
    
    await logLoginAttempt({
      userId: user.id,
      email,
      success: false,
      failureReason: 'invalid_credentials',
      ipAddress,
      userAgent,
    });
    
    throw new Error('Invalid credentials');
  }
  
  // 6. Reset failed attempts
  await prisma.user.update({
    where: { id: user.id },
    data: {
      failedLoginAttempts: 0,
      lockedUntil: null,
      lastLoginAt: new Date(),
    },
  });
  
  // 7. Generate tokens
  const accessToken = jwt.sign(
    {
      sub: user.id,
      email: user.email,
      role: user.role,
    },
    process.env.JWT_SECRET!,
    { expiresIn: '15m' }
  );
  
  const refreshToken = jwt.sign(
    {
      sub: user.id,
      type: 'refresh',
    },
    process.env.JWT_REFRESH_SECRET!,
    { expiresIn: '30d' }
  );
  
  // 8. Store session in database
  const session = await prisma.session.create({
    data: {
      userId: user.id,
      token: accessToken,
      refreshToken,
      deviceId,
      ipAddress,
      userAgent,
      expiresAt: new Date(Date.now() + 15 * 60 * 1000), // 15 minutes
    },
  });
  
  // 9. Store session in Redis
  await redisClient.hset(`session:${accessToken}`, {
    user_id: user.id,
    email: user.email,
    role: user.role,
    ip_address: ipAddress || '',
    created_at: new Date().toISOString(),
  });
  await redisClient.expire(`session:${accessToken}`, 900); // 15 minutes
  
  // 10. Log successful login
  await logLoginAttempt({
    userId: user.id,
    email,
    success: true,
    ipAddress,
    userAgent,
  });
  
  await logAudit({
    userId: user.id,
    eventType: 'login',
    eventCategory: 'authentication',
    description: 'User logged in',
    ipAddress,
    userAgent,
    success: true,
  });
  
  return {
    user: {
      id: user.id,
      email: user.email,
      role: user.role,
    },
    tokens: {
      accessToken,
      refreshToken,
      expiresIn: 900, // 15 minutes
    },
  };
}

async function logLoginAttempt(data: {
  userId?: string;
  email: string;
  success: boolean;
  failureReason?: string;
  ipAddress?: string;
  userAgent?: string;
}) {
  await prisma.loginAttempt.create({
    data: {
      userId: data.userId,
      email: data.email,
      success: data.success,
      failureReason: data.failureReason,
      ipAddress: data.ipAddress || '',
      userAgent: data.userAgent,
    },
  });
}
```

---

### **3. Token Validation Middleware**

```typescript
// src/middleware/auth.middleware.ts

import { Request, Response, NextFunction } from 'express';
import jwt from 'jsonwebtoken';
import { redisClient } from '../redis/redis.client';

export async function authenticateToken(
  req: Request,
  res: Response,
  next: NextFunction
) {
  const authHeader = req.headers['authorization'];
  const token = authHeader && authHeader.split(' ')[1]; // Bearer TOKEN
  
  if (!token) {
    return res.status(401).json({ error: 'Access token required' });
  }
  
  try {
    // 1. Check if token is blacklisted
    const isBlacklisted = await redisClient.exists(`blacklist:${token}`);
    if (isBlacklisted) {
      return res.status(401).json({ error: 'Token has been revoked' });
    }
    
    // 2. Verify JWT
    const payload = jwt.verify(token, process.env.JWT_SECRET!) as {
      sub: string;
      email: string;
      role: string;
    };
    
    // 3. Check session in Redis
    const session = await redisClient.hgetall(`session:${token}`);
    if (!session || !session.user_id) {
      return res.status(401).json({ error: 'Invalid session' });
    }
    
    // 4. Update last activity
    await redisClient.hset(`session:${token}`, 'last_activity', new Date().toISOString());
    
    // 5. Attach user to request
    req.user = {
      id: payload.sub,
      email: payload.email,
      role: payload.role,
    };
    
    next();
  } catch (error) {
    if (error instanceof jwt.TokenExpiredError) {
      return res.status(401).json({ error: 'Token expired' });
    }
    
    return res.status(403).json({ error: 'Invalid token' });
  }
}

// Role-based authorization
export function authorizeRole(...allowedRoles: string[]) {
  return (req: Request, res: Response, next: NextFunction) => {
    if (!req.user) {
      return res.status(401).json({ error: 'Authentication required' });
    }
    
    if (!allowedRoles.includes(req.user.role)) {
      return res.status(403).json({ error: 'Insufficient permissions' });
    }
    
    next();
  };
}
```

---

### **4. Password Reset**

```typescript
// src/services/auth/password-reset.service.ts

import { PrismaClient } from '@prisma/client';
import bcrypt from 'bcrypt';
import { v4 as uuidv4 } from 'uuid';
import { sendPasswordResetEmail } from '../email/email.service';
import { redisClient } from '../redis/redis.client';

const prisma = new PrismaClient();

// Initiate password reset
export async function initiatePasswordReset(email: string, ipAddress?: string) {
  // 1. Rate limiting
  const rateLimitKey = `rate_limit:password_reset:${ipAddress}`;
  const attempts = await redisClient.incr(rateLimitKey);
  
  if (attempts === 1) {
    await redisClient.expire(rateLimitKey, 3600); // 1 hour
  }
  
  if (attempts > 3) {
    throw new Error('Too many password reset attempts');
  }
  
  // 2. Find user
  const user = await prisma.user.findUnique({
    where: { email: email.toLowerCase() },
  });
  
  if (!user) {
    // Don't reveal if user exists or not
    return { message: 'If email exists, reset link has been sent' };
  }
  
  // 3. Generate reset token
  const token = uuidv4();
  const tokenHash = await bcrypt.hash(token, 10);
  
  const verificationToken = await prisma.verificationToken.create({
    data: {
      userId: user.id,
      token,
      tokenHash,
      type: 'password_reset',
      expiresAt: new Date(Date.now() + 60 * 60 * 1000), // 1 hour
      ipAddress,
    },
  });
  
  // 4. Create password reset record
  await prisma.passwordReset.create({
    data: {
      userId: user.id,
      tokenId: verificationToken.id,
      status: 'initiated',
      ipAddress,
    },
  });
  
  // 5. Store in Redis for fast access
  await redisClient.hset(`password_reset:${token}`, {
    user_id: user.id,
    email: user.email,
    created_at: new Date().toISOString(),
  });
  await redisClient.expire(`password_reset:${token}`, 3600); // 1 hour
  
  // 6. Send email
  await sendPasswordResetEmail(user.email, token);
  
  return { message: 'If email exists, reset link has been sent' };
}

// Complete password reset
export async function completePasswordReset(
  token: string,
  newPassword: string
) {
  // 1. Check Redis cache first
  const cachedReset = await redisClient.hgetall(`password_reset:${token}`);
  
  if (!cachedReset || !cachedReset.user_id) {
    throw new Error('Invalid or expired reset token');
  }
  
  // 2. Verify token in database
  const verificationToken = await prisma.verificationToken.findUnique({
    where: { token },
    include: { user: true },
  });
  
  if (
    !verificationToken ||
    verificationToken.used ||
    verificationToken.expiresAt < new Date()
  ) {
    throw new Error('Invalid or expired reset token');
  }
  
  // 3. Hash new password
  const newPasswordHash = await bcrypt.hash(newPassword, 12);
  
  // 4. Update user password
  await prisma.user.update({
    where: { id: verificationToken.userId },
    data: {
      passwordHash: newPasswordHash,
      failedLoginAttempts: 0,
      lockedUntil: null,
    },
  });
  
  // 5. Mark token as used
  await prisma.verificationToken.update({
    where: { id: verificationToken.id },
    data: {
      used: true,
      usedAt: new Date(),
    },
  });
  
  // 6. Update password reset record
  await prisma.passwordReset.updateMany({
    where: {
      userId: verificationToken.userId,
      tokenId: verificationToken.id,
    },
    data: {
      status: 'completed',
      completedAt: new Date(),
    },
  });
  
  // 7. Remove from Redis
  await redisClient.del(`password_reset:${token}`);
  
  // 8. Invalidate all user sessions
  await invalidateAllUserSessions(verificationToken.userId);
  
  return { message: 'Password reset successful' };
}

async function invalidateAllUserSessions(userId: string) {
  // Mark all sessions as inactive in database
  await prisma.session.updateMany({
    where: { userId },
    data: { isActive: false },
  });
  
  // Note: Redis sessions will expire naturally
}
```

---

## 🚀 **REDIS HELPER FUNCTIONS**

```typescript
// src/services/redis/redis.client.ts

import { createClient } from 'redis';

export const redisClient = createClient({
  url: process.env.REDIS_URL || 'redis://localhost:6379',
});

redisClient.on('error', (err) => console.error('Redis Client Error', err));

await redisClient.connect();

// Session management
export async function createSession(userId: string, token: string, data: any) {
  await redisClient.hset(`session:${token}`, {
    user_id: userId,
    ...data,
    created_at: new Date().toISOString(),
  });
  await redisClient.expire(`session:${token}`, 900); // 15 minutes
}

export async function getSession(token: string) {
  return await redisClient.hgetall(`session:${token}`);
}

export async function deleteSession(token: string) {
  await redisClient.del(`session:${token}`);
}

// Token blacklist
export async function blacklistToken(token: string, expiresIn: number) {
  await redisClient.set(`blacklist:${token}`, 'revoked');
  await redisClient.expire(`blacklist:${token}`, expiresIn);
}

export async function isTokenBlacklisted(token: string): Promise<boolean> {
  return (await redisClient.exists(`blacklist:${token}`)) === 1;
}

// Rate limiting
export async function checkRateLimit(
  key: string,
  maxAttempts: number,
  windowSeconds: number
): Promise<boolean> {
  const attempts = await redisClient.incr(`rate_limit:${key}`);
  
  if (attempts === 1) {
    await redisClient.expire(`rate_limit:${key}`, windowSeconds);
  }
  
  return attempts <= maxAttempts;
}
```

---

## ✅ **TESTING EXAMPLES**

```typescript
// tests/auth.test.ts

import { registerUser, loginUser } from '../services/auth';

describe('Authentication', () => {
  describe('User Registration', () => {
    it('should register a new user', async () => {
      const result = await registerUser({
        email: 'test@example.com',
        password: 'SecurePassword123!',
        role: 'buyer',
      });
      
      expect(result).toHaveProperty('id');
      expect(result.email).toBe('test@example.com');
      expect(result.status).toBe('pending');
    });
    
    it('should not allow duplicate emails', async () => {
      await expect(
        registerUser({
          email: 'test@example.com',
          password: 'SecurePassword123!',
        })
      ).rejects.toThrow('Email already registered');
    });
  });
  
  describe('User Login', () => {
    it('should login with valid credentials', async () => {
      const result = await loginUser({
        email: 'test@example.com',
        password: 'SecurePassword123!',
      });
      
      expect(result).toHaveProperty('tokens');
      expect(result.tokens).toHaveProperty('accessToken');
      expect(result.tokens).toHaveProperty('refreshToken');
    });
    
    it('should not login with invalid password', async () => {
      await expect(
        loginUser({
          email: 'test@example.com',
          password: 'WrongPassword',
        })
      ).rejects.toThrow('Invalid credentials');
    });
  });
});
```

---

**🔐 Your Auth Service implementation is ready to code! 🚀**

