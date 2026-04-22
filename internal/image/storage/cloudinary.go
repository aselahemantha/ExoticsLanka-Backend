package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryClient struct {
	cld *cloudinary.Cloudinary
}

func NewCloudinaryClient(cloudinaryURL string) (*CloudinaryClient, error) {
	cld, err := cloudinary.NewFromURL(cloudinaryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Cloudinary: %w", err)
	}
	return &CloudinaryClient{cld: cld}, nil
}

// UploadFile uploads a file to Cloudinary and returns the secure URL and the Public ID
func (c *CloudinaryClient) UploadFile(ctx context.Context, file io.Reader, folder string) (string, string, error) {
	resp, err := c.cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder: folder,
	})
	if err != nil {
		return "", "", fmt.Errorf("cloudinary upload failed: %w", err)
	}

	return resp.SecureURL, resp.PublicID, nil
}

// DeleteFile removes a file from Cloudinary using its Public ID
func (c *CloudinaryClient) DeleteFile(ctx context.Context, publicID string) error {
	_, err := c.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})
	if err != nil {
		return fmt.Errorf("cloudinary delete failed: %w", err)
	}
	return nil
}
