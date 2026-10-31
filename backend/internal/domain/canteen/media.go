package canteen

import (
	"context"
	"io"
	"time"
)

type MediaStatus string

const (
	MediaStatusPendingUpload MediaStatus = "pending_upload"
	MediaStatusPending       MediaStatus = "pending"
	MediaStatusApproved      MediaStatus = "approved"
	MediaStatusRejected      MediaStatus = "rejected"
	MediaStatusArchived      MediaStatus = "archived"
)

func (s MediaStatus) Valid() bool {
	switch s {
	case MediaStatusPendingUpload, MediaStatusPending, MediaStatusApproved, MediaStatusRejected, MediaStatusArchived:
		return true
	default:
		return false
	}
}

// PubliclyVisible gates every unauthenticated or third-party read. Pending and
// rejected images stay visible to their owner and to administrators only.
func (s MediaStatus) PubliclyVisible() bool {
	return s == MediaStatusApproved
}

type MediaPurpose string

const (
	MediaPurposeReview     MediaPurpose = "review"
	MediaPurposeDish       MediaPurpose = "dish"
	MediaPurposeStoreCover MediaPurpose = "store_cover"
	MediaPurposeReport     MediaPurpose = "report"
)

func (p MediaPurpose) Valid() bool {
	switch p {
	case MediaPurposeReview, MediaPurposeDish, MediaPurposeStoreCover, MediaPurposeReport:
		return true
	default:
		return false
	}
}

// Media is stored metadata only; bytes live behind MediaStorage.
type Media struct {
	ID            int
	OwnerUserID   int
	Purpose       MediaPurpose
	StorageDriver string
	StorageKey    string
	MIME          string
	ByteSize      int
	Width         int
	Height        int
	SHA256        string
	Status        MediaStatus
	ReviewedBy    *int
	ReviewedAt    *time.Time
	RejectReason  string
	ExpiresAt     *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// MediaStorage is the byte-store port. The first release ships a local
// filesystem driver; an S3-compatible driver can implement the same port later.
type MediaStorage interface {
	Put(ctx context.Context, key string, reader io.Reader) error
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
}

// ProcessedImage is the normalized, metadata-free image written to storage.
type ProcessedImage struct {
	Data   []byte
	MIME   string
	Width  int
	Height int
	SHA256 string
}

// ImageProcessor decodes, validates and re-encodes uploaded images. The
// re-encode strips EXIF, including GPS data.
type ImageProcessor interface {
	Process(raw []byte, maxPixels int) (ProcessedImage, error)
}
