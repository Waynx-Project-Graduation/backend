package services

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryService struct {
	client *cloudinary.Cloudinary
}

func NewCloudinaryService(url string) *CloudinaryService {
	cld, err := cloudinary.NewFromURL(url)
	if err != nil {
		log.Printf("⚠️ Warning: Failed to initialize Cloudinary client: %v\n", err)
		return &CloudinaryService{}
	}
	log.Println("✅ Cloudinary client initialized successfully")
	return &CloudinaryService{
		client: cld,
	}
}

// UploadImage uploads an image file (e.g. multipart.File or file path) to Cloudinary
func (s *CloudinaryService) UploadImage(ctx context.Context, file interface{}, folder string) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("Cloudinary client is not initialized")
	}

	resp, err := s.client.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder: folder,
	})

	if err != nil {
		return "", err
	}

	return resp.SecureURL, nil
}
