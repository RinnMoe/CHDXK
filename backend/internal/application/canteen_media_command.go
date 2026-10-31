package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"jcourse/internal/domain/audit"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/canteen"
	"jcourse/pkg/apperr"
)

// CanteenMediaLimits is the media budget enforced by the service.
type CanteenMediaLimits struct {
	MaxBytes         int64
	MaxPerReview     int
	MaxUploadsPerDay int
	MaxPixels        int
}

// CanteenMediaCommandService handles upload intents, proxy uploads and
// moderation. Bytes are written through the MediaStorage port.
type CanteenMediaCommandService struct {
	media     canteen.MediaRepository
	query     canteen.MediaQuery
	storage   canteen.MediaStorage
	processor canteen.ImageProcessor
	settings  CanteenSettingsProvider
	limits    CanteenMediaLimits
	now       func() time.Time
}

func NewCanteenMediaCommandService(
	media canteen.MediaRepository,
	query canteen.MediaQuery,
	storage canteen.MediaStorage,
	processor canteen.ImageProcessor,
	settings CanteenSettingsProvider,
	limits CanteenMediaLimits,
) *CanteenMediaCommandService {
	if limits.MaxBytes <= 0 {
		limits.MaxBytes = 5 * 1024 * 1024
	}
	if limits.MaxPerReview <= 0 {
		limits.MaxPerReview = 3
	}
	if limits.MaxUploadsPerDay <= 0 {
		limits.MaxUploadsPerDay = 20
	}
	if limits.MaxPixels <= 0 {
		limits.MaxPixels = 8000000
	}
	return &CanteenMediaCommandService{
		media:     media,
		query:     query,
		storage:   storage,
		processor: processor,
		settings:  settings,
		limits:    limits,
		now:       time.Now,
	}
}

func (s *CanteenMediaCommandService) requireEnabled(ctx context.Context) error {
	if s.settings == nil {
		return nil
	}
	config, err := s.settings.CanteenRuntimeConfig(ctx)
	if err != nil {
		return err
	}
	if !config.Enabled {
		return apperr.NotFound("食堂评价暂未开放")
	}
	if !config.MediaEnabled {
		return apperr.Forbidden("实拍图片暂未开放")
	}
	return nil
}

func canteenMediaExtension(mime string) string {
	switch mime {
	case "image/png":
		return "png"
	default:
		return "jpg"
	}
}

func canteenRandomToken() string {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buffer)
}

// CreateIntent reserves a media row and returns the proxy upload URL.
func (s *CanteenMediaCommandService) CreateIntent(ctx context.Context, user *auth.User, purpose string, mime string, size int64) (*canteen.Media, string, error) {
	if err := s.requireEnabled(ctx); err != nil {
		return nil, "", err
	}
	mediaPurpose := canteen.MediaPurpose(purpose)
	if mediaPurpose != canteen.MediaPurposeReview && mediaPurpose != canteen.MediaPurposeDish {
		if !(user != nil && user.IsAdmin() && mediaPurpose == canteen.MediaPurposeStoreCover) {
			return nil, "", apperr.BadRequest("图片用途无效")
		}
	}
	mime = strings.TrimSpace(mime)
	if mime != "image/jpeg" && mime != "image/png" {
		return nil, "", canteen.ErrMediaType
	}
	if size <= 0 || size > s.limits.MaxBytes {
		return nil, "", canteen.ErrMediaTooLarge
	}
	now := s.now()
	count, err := s.media.CountByOwnerSince(ctx, user.ID, now.Add(-24*time.Hour))
	if err != nil {
		return nil, "", err
	}
	if count >= s.limits.MaxUploadsPerDay {
		return nil, "", canteen.ErrMediaTooManyDaily
	}
	expires := now.Add(30 * time.Minute)
	key := fmt.Sprintf("%s/%d/%s/%s.%s",
		mediaPurpose, user.ID, now.Format("2006/01"), canteenRandomToken(), canteenMediaExtension(mime))
	media := &canteen.Media{
		OwnerUserID:   user.ID,
		Purpose:       mediaPurpose,
		StorageDriver: "local",
		StorageKey:    key,
		MIME:          mime,
		ByteSize:      int(size),
		Status:        canteen.MediaStatusPendingUpload,
		ExpiresAt:     &expires,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.media.Create(ctx, media); err != nil {
		return nil, "", err
	}
	return media, fmt.Sprintf("/api/canteen/media/%d/content", media.ID), nil
}

// PutContent reads the uploaded bytes, strips metadata and marks the image
// pending moderation.
func (s *CanteenMediaCommandService) PutContent(ctx context.Context, user *auth.User, mediaID int, reader io.Reader) (*canteen.Media, error) {
	if err := s.requireEnabled(ctx); err != nil {
		return nil, err
	}
	media, err := s.query.GetMedia(ctx, mediaID)
	if err != nil {
		return nil, err
	}
	if media == nil {
		return nil, canteen.ErrMediaNotFound
	}
	if media.OwnerUserID != user.ID {
		return nil, canteen.ErrMediaNotOwned
	}
	if media.Status != canteen.MediaStatusPendingUpload {
		return nil, canteen.ErrMediaState
	}
	now := s.now()
	if media.ExpiresAt != nil && now.After(*media.ExpiresAt) {
		return nil, canteen.ErrMediaState
	}
	raw, err := io.ReadAll(io.LimitReader(reader, s.limits.MaxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(raw)) > s.limits.MaxBytes {
		return nil, canteen.ErrMediaTooLarge
	}
	processed, err := s.processor.Process(raw, s.limits.MaxPixels)
	if err != nil {
		return nil, err
	}
	if err := s.storage.Put(ctx, media.StorageKey, strings.NewReader(string(processed.Data))); err != nil {
		return nil, err
	}
	media.MIME = processed.MIME
	media.ByteSize = len(processed.Data)
	media.Width = processed.Width
	media.Height = processed.Height
	media.SHA256 = processed.SHA256
	media.Status = canteen.MediaStatusPending
	media.ExpiresAt = nil
	media.UpdatedAt = now
	if err := s.media.Update(ctx, media); err != nil {
		return nil, err
	}
	return media, nil
}

// Complete is idempotent confirmation that the upload finished.
func (s *CanteenMediaCommandService) Complete(ctx context.Context, user *auth.User, mediaID int) (*canteen.Media, error) {
	media, err := s.query.GetMedia(ctx, mediaID)
	if err != nil {
		return nil, err
	}
	if media == nil {
		return nil, canteen.ErrMediaNotFound
	}
	if media.OwnerUserID != user.ID {
		return nil, canteen.ErrMediaNotOwned
	}
	if media.Status == canteen.MediaStatusPendingUpload {
		return nil, canteen.ErrMediaState
	}
	return media, nil
}

func (s *CanteenMediaCommandService) Approve(ctx context.Context, reviewer *auth.User, mediaID int, status canteen.MediaStatus, reason string) error {
	if reviewer == nil || !reviewer.IsAdmin() {
		return canteen.ErrCannotModerate
	}
	if status != canteen.MediaStatusApproved && status != canteen.MediaStatusRejected {
		return apperr.BadRequest("图片审核状态无效")
	}
	media, err := s.query.GetMedia(ctx, mediaID)
	if err != nil {
		return err
	}
	if media == nil {
		return canteen.ErrMediaNotFound
	}
	if media.Status != canteen.MediaStatusPending {
		return canteen.ErrMediaState
	}
	now := s.now()
	media.Status = status
	media.ReviewedBy = &reviewer.ID
	media.ReviewedAt = &now
	media.RejectReason = reason
	media.UpdatedAt = now
	if err := s.media.Update(ctx, media); err != nil {
		return err
	}
	audit.EnqueueLog(ctx, audit.Log{
		OccurredAt:  now,
		ActorUserID: reviewer.ID,
		Action:      "canteen.media.review",
		TargetType:  "canteen_media",
		TargetID:    fmt.Sprintf("%d", media.ID),
		Details:     audit.Details{"status": string(status)},
	})
	return nil
}

// Open returns bytes only when the caller may see the image status.
func (s *CanteenMediaCommandService) Open(ctx context.Context, viewer *auth.User, mediaID int) (io.ReadCloser, *canteen.Media, error) {
	media, err := s.query.GetMedia(ctx, mediaID)
	if err != nil {
		return nil, nil, err
	}
	if media == nil {
		return nil, nil, canteen.ErrMediaNotFound
	}
	if !media.Status.PubliclyVisible() {
		if viewer == nil {
			return nil, nil, canteen.ErrMediaNotFound
		}
		if viewer.ID != media.OwnerUserID && !viewer.IsAdmin() {
			return nil, nil, canteen.ErrMediaNotFound
		}
	}
	reader, err := s.storage.Open(ctx, media.StorageKey)
	if err != nil {
		return nil, nil, err
	}
	return reader, media, nil
}

// ValidateAttachments checks ownership, purpose and status before a review is
// created. Duplicate ids are collapsed.
func (s *CanteenMediaCommandService) ValidateAttachments(ctx context.Context, user *auth.User, mediaIDs []int) ([]canteen.Media, error) {
	if len(mediaIDs) == 0 {
		return nil, nil
	}
	if len(mediaIDs) > s.limits.MaxPerReview {
		return nil, canteen.ErrMediaLimit
	}
	seen := make(map[int]struct{}, len(mediaIDs))
	result := make([]canteen.Media, 0, len(mediaIDs))
	for _, mediaID := range mediaIDs {
		if _, ok := seen[mediaID]; ok {
			continue
		}
		seen[mediaID] = struct{}{}
		media, err := s.query.GetMedia(ctx, mediaID)
		if err != nil {
			return nil, err
		}
		if media == nil {
			return nil, canteen.ErrMediaNotFound
		}
		if media.OwnerUserID != user.ID {
			return nil, canteen.ErrMediaNotOwned
		}
		if media.Purpose != canteen.MediaPurposeReview && media.Purpose != canteen.MediaPurposeDish {
			return nil, canteen.ErrMediaState
		}
		if media.Status != canteen.MediaStatusPending && media.Status != canteen.MediaStatusApproved {
			return nil, canteen.ErrMediaState
		}
		result = append(result, *media)
	}
	return result, nil
}

// Attach validates and links media to a review after creation.
func (s *CanteenMediaCommandService) Attach(ctx context.Context, user *auth.User, reviewID int, mediaIDs []int) error {
	media, err := s.ValidateAttachments(ctx, user, mediaIDs)
	if err != nil {
		return err
	}
	for index, item := range media {
		if err := s.media.AttachToReview(ctx, reviewID, item.ID, index); err != nil {
			return err
		}
	}
	return nil
}

// AttachToDish validates and links media to a dish gallery.
func (s *CanteenMediaCommandService) AttachToDish(ctx context.Context, user *auth.User, dishID int, mediaIDs []int) error {
	media, err := s.ValidateAttachments(ctx, user, mediaIDs)
	if err != nil {
		return err
	}
	now := s.now()
	for index, item := range media {
		if err := s.media.AttachToDish(ctx, dishID, item.ID, user.ID, index, now); err != nil {
			return err
		}
	}
	return nil
}

// IsMediaLimitError reports whether an attachment validation error is the
// per-review cap, used to keep error mapping stable.
func IsMediaLimitError(err error) bool {
	return errors.Is(err, canteen.ErrMediaLimit)
}
