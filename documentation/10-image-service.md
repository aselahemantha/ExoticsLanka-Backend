# Image Service

> Handles image uploads, processing, and storage.

---

## Overview

The Image Service manages all image uploads for the platform, including listing photos, user avatars, and review photos. It handles image processing, resizing, and cloud storage.

**Responsibilities:**
- Handle multipart file uploads
- Validate file types and sizes
- Resize images to multiple sizes
- Generate thumbnails
- Upload to cloud storage (S3/Cloudinary)
- Manage image ordering for listings
- Delete images

---

## Configuration

### Supported Formats

```javascript
const ALLOWED_MIME_TYPES = [
  'image/jpeg',
  'image/jpg',
  'image/png',
  'image/webp'
];

const MAX_FILE_SIZE = 10 * 1024 * 1024; // 10MB
const MAX_LISTING_IMAGES = 15;
const MAX_REVIEW_PHOTOS = 5;
```

### Image Sizes

```javascript
const IMAGE_SIZES = {
  listing: {
    full: { width: 1200, height: 900, quality: 85 },
    medium: { width: 800, height: 600, quality: 80 },
    thumbnail: { width: 400, height: 300, quality: 75 },
    card: { width: 600, height: 400, quality: 80 }
  },
  avatar: {
    full: { width: 400, height: 400, quality: 85 },
    thumbnail: { width: 100, height: 100, quality: 80 }
  },
  review: {
    full: { width: 800, height: 600, quality: 80 },
    thumbnail: { width: 200, height: 150, quality: 75 }
  }
};
```

---

## API Endpoints

### Listing Images

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| `POST` | `/api/listings/:id/images` | Upload images | Yes (Owner) |
| `PUT` | `/api/listings/:id/images/reorder` | Reorder images | Yes (Owner) |
| `PUT` | `/api/listings/:id/images/:imageId/cover` | Set cover image | Yes (Owner) |
| `DELETE` | `/api/listings/:id/images/:imageId` | Delete image | Yes (Owner) |

### User Avatar

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| `PUT` | `/api/users/me/avatar` | Upload avatar | Yes |
| `DELETE` | `/api/users/me/avatar` | Delete avatar | Yes |

### Review Photos

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| `POST` | `/api/reviews/:id/photos` | Upload photos | Yes (Author) |
| `DELETE` | `/api/reviews/:id/photos/:photoId` | Delete photo | Yes (Author) |

---

## Request/Response Examples

### POST /api/listings/:id/images

Upload images to a listing.

**Request:** Multipart form data
- Field: `images` (multiple files allowed)

**Response:**
```json
{
  "success": true,
  "message": "3 images uploaded successfully",
  "data": {
    "images": [
      {
        "id": "img-uuid-1",
        "imageUrl": "https://storage.exotics.lk/listings/full/abc123.jpg",
        "thumbnailUrl": "https://storage.exotics.lk/listings/thumb/abc123.jpg",
        "isCover": false,
        "sortOrder": 3
      },
      {
        "id": "img-uuid-2",
        "imageUrl": "https://storage.exotics.lk/listings/full/def456.jpg",
        "thumbnailUrl": "https://storage.exotics.lk/listings/thumb/def456.jpg",
        "isCover": false,
        "sortOrder": 4
      },
      {
        "id": "img-uuid-3",
        "imageUrl": "https://storage.exotics.lk/listings/full/ghi789.jpg",
        "thumbnailUrl": "https://storage.exotics.lk/listings/thumb/ghi789.jpg",
        "isCover": false,
        "sortOrder": 5
      }
    ],
    "totalImages": 5,
    "maxImages": 15
  }
}
```

### PUT /api/listings/:id/images/reorder

Reorder images for a listing.

**Request Body:**
```json
{
  "order": [
    { "id": "img-uuid-2", "sortOrder": 0 },
    { "id": "img-uuid-1", "sortOrder": 1 },
    { "id": "img-uuid-3", "sortOrder": 2 }
  ]
}
```

**Response:**
```json
{
  "success": true,
  "message": "Images reordered"
}
```

### PUT /api/listings/:id/images/:imageId/cover

Set an image as the cover image.

**Response:**
```json
{
  "success": true,
  "message": "Cover image updated",
  "data": {
    "coverId": "img-uuid-2"
  }
}
```

### PUT /api/users/me/avatar

Upload user avatar.

**Request:** Multipart form data
- Field: `avatar` (single file)

**Response:**
```json
{
  "success": true,
  "message": "Avatar updated",
  "data": {
    "avatarUrl": "https://storage.exotics.lk/avatars/full/user123.jpg",
    "thumbnailUrl": "https://storage.exotics.lk/avatars/thumb/user123.jpg"
  }
}
```

---

## Implementation

### Upload Middleware (Multer + S3)

```javascript
const multer = require('multer');
const multerS3 = require('multer-s3');
const { S3Client } = require('@aws-sdk/client-s3');
const sharp = require('sharp');

const s3 = new S3Client({
  region: process.env.AWS_REGION,
  credentials: {
    accessKeyId: process.env.AWS_ACCESS_KEY,
    secretAccessKey: process.env.AWS_SECRET_KEY
  }
});

// Memory storage for processing before upload
const upload = multer({
  storage: multer.memoryStorage(),
  limits: {
    fileSize: MAX_FILE_SIZE,
    files: MAX_LISTING_IMAGES
  },
  fileFilter: (req, file, cb) => {
    if (ALLOWED_MIME_TYPES.includes(file.mimetype)) {
      cb(null, true);
    } else {
      cb(new ValidationError('Invalid file type. Allowed: JPEG, PNG, WebP'));
    }
  }
});
```

### Image Processing

```javascript
const sharp = require('sharp');
const { PutObjectCommand, DeleteObjectCommand } = require('@aws-sdk/client-s3');
const { v4: uuidv4 } = require('uuid');

async function processAndUploadImage(buffer, type = 'listing') {
  const id = uuidv4();
  const sizes = IMAGE_SIZES[type];
  const urls = {};
  
  for (const [sizeName, config] of Object.entries(sizes)) {
    // Process image
    const processed = await sharp(buffer)
      .resize(config.width, config.height, {
        fit: 'cover',
        position: 'center'
      })
      .jpeg({ quality: config.quality })
      .toBuffer();
    
    // Upload to S3
    const key = `${type}s/${sizeName}/${id}.jpg`;
    
    await s3.send(new PutObjectCommand({
      Bucket: process.env.S3_BUCKET,
      Key: key,
      Body: processed,
      ContentType: 'image/jpeg',
      ACL: 'public-read'
    }));
    
    urls[sizeName] = `https://${process.env.S3_BUCKET}.s3.${process.env.AWS_REGION}.amazonaws.com/${key}`;
  }
  
  return {
    id,
    imageUrl: urls.full,
    thumbnailUrl: urls.thumbnail,
    ...urls
  };
}

async function deleteImage(imageUrl) {
  // Extract key from URL
  const url = new URL(imageUrl);
  const key = url.pathname.substring(1); // Remove leading slash
  
  // Delete all sizes
  const sizes = ['full', 'medium', 'thumbnail', 'card'];
  const baseKey = key.replace(/\/(full|medium|thumbnail|card)\//, '/');
  
  for (const size of sizes) {
    const sizeKey = baseKey.replace('/', `/${size}/`);
    try {
      await s3.send(new DeleteObjectCommand({
        Bucket: process.env.S3_BUCKET,
        Key: sizeKey
      }));
    } catch (error) {
      // Ignore if file doesn't exist
    }
  }
}
```

### Upload Handler

```javascript
async function uploadListingImages(listingId, userId, files) {
  // Verify ownership
  const listing = await db.query(
    'SELECT user_id FROM car_listings WHERE id = $1',
    [listingId]
  );
  
  if (!listing.rows[0]) {
    throw new NotFoundError('Listing not found');
  }
  
  if (listing.rows[0].user_id !== userId) {
    throw new ForbiddenError('You can only upload images to your own listings');
  }
  
  // Check current image count
  const currentCount = await db.query(
    'SELECT COUNT(*) FROM listing_images WHERE listing_id = $1',
    [listingId]
  );
  
  const count = parseInt(currentCount.rows[0].count);
  if (count + files.length > MAX_LISTING_IMAGES) {
    throw new LimitExceededError(`Maximum ${MAX_LISTING_IMAGES} images allowed. You have ${count}.`);
  }
  
  // Process and upload each image
  const uploadedImages = [];
  let sortOrder = count;
  
  for (const file of files) {
    const processed = await processAndUploadImage(file.buffer, 'listing');
    
    // Save to database
    const image = await db.query(`
      INSERT INTO listing_images (listing_id, image_url, thumbnail_url, is_cover, sort_order)
      VALUES ($1, $2, $3, $4, $5)
      RETURNING *
    `, [listingId, processed.imageUrl, processed.thumbnailUrl, count === 0 && sortOrder === 0, sortOrder]);
    
    uploadedImages.push(image.rows[0]);
    sortOrder++;
  }
  
  // Recalculate health score (images affect it)
  await recalculateHealthScore(listingId);
  
  return {
    images: uploadedImages,
    totalImages: count + files.length,
    maxImages: MAX_LISTING_IMAGES
  };
}
```

### Reorder Images

```javascript
async function reorderImages(listingId, userId, order) {
  // Verify ownership
  const listing = await db.query(
    'SELECT user_id FROM car_listings WHERE id = $1',
    [listingId]
  );
  
  if (listing.rows[0]?.user_id !== userId) {
    throw new ForbiddenError('Not authorized');
  }
  
  // Update sort orders in transaction
  await db.query('BEGIN');
  
  try {
    for (const item of order) {
      await db.query(
        'UPDATE listing_images SET sort_order = $1 WHERE id = $2 AND listing_id = $3',
        [item.sortOrder, item.id, listingId]
      );
    }
    
    // Set first image as cover
    await db.query(
      'UPDATE listing_images SET is_cover = FALSE WHERE listing_id = $1',
      [listingId]
    );
    
    const firstImage = order.find(o => o.sortOrder === 0);
    if (firstImage) {
      await db.query(
        'UPDATE listing_images SET is_cover = TRUE WHERE id = $1',
        [firstImage.id]
      );
    }
    
    await db.query('COMMIT');
  } catch (error) {
    await db.query('ROLLBACK');
    throw error;
  }
}
```

### Delete Image

```javascript
async function deleteListingImage(listingId, imageId, userId) {
  // Verify ownership
  const image = await db.query(`
    SELECT li.*, cl.user_id 
    FROM listing_images li
    JOIN car_listings cl ON li.listing_id = cl.id
    WHERE li.id = $1 AND li.listing_id = $2
  `, [imageId, listingId]);
  
  if (!image.rows[0]) {
    throw new NotFoundError('Image not found');
  }
  
  if (image.rows[0].user_id !== userId) {
    throw new ForbiddenError('Not authorized');
  }
  
  // Delete from S3
  await deleteImage(image.rows[0].image_url);
  
  // Delete from database
  await db.query('DELETE FROM listing_images WHERE id = $1', [imageId]);
  
  // If it was cover, set new cover
  if (image.rows[0].is_cover) {
    await db.query(`
      UPDATE listing_images 
      SET is_cover = TRUE 
      WHERE listing_id = $1 
      ORDER BY sort_order 
      LIMIT 1
    `, [listingId]);
  }
  
  // Recalculate health score
  await recalculateHealthScore(listingId);
}
```

---

## Cloudinary Alternative

If using Cloudinary instead of S3:

```javascript
const cloudinary = require('cloudinary').v2;

cloudinary.config({
  cloud_name: process.env.CLOUDINARY_CLOUD_NAME,
  api_key: process.env.CLOUDINARY_API_KEY,
  api_secret: process.env.CLOUDINARY_API_SECRET
});

async function uploadToCloudinary(buffer, folder) {
  return new Promise((resolve, reject) => {
    const uploadStream = cloudinary.uploader.upload_stream(
      {
        folder: `exotics-lanka/${folder}`,
        transformation: [
          { width: 1200, height: 900, crop: 'fill' }
        ],
        eager: [
          { width: 400, height: 300, crop: 'fill' }, // thumbnail
          { width: 600, height: 400, crop: 'fill' }  // card
        ]
      },
      (error, result) => {
        if (error) reject(error);
        else resolve(result);
      }
    );
    
    uploadStream.end(buffer);
  });
}
```

---

## Error Responses

```json
// 400 - Invalid File Type
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid file type. Allowed: JPEG, PNG, WebP"
  }
}

// 400 - File Too Large
{
  "success": false,
  "error": {
    "code": "FILE_TOO_LARGE",
    "message": "File size exceeds 10MB limit"
  }
}

// 400 - Too Many Files
{
  "success": false,
  "error": {
    "code": "LIMIT_EXCEEDED",
    "message": "Maximum 15 images allowed. You have 12."
  }
}

// 403 - Not Authorized
{
  "success": false,
  "error": {
    "code": "FORBIDDEN",
    "message": "You can only upload images to your own listings"
  }
}
```

---

## Related Services

- **Listings Service** - Manages listing images
- **User Service** - Manages avatars
- **Reviews Service** - Manages review photos

