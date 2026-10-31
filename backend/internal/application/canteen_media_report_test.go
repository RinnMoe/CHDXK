package application

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/canteen"
)

func newMediaCommandFixture(t *testing.T) (*CanteenMediaCommandService, *canteen.MockMediaRepository, *canteen.MockMediaStorage) {
	t.Helper()
	repo := canteen.NewMockMediaRepository()
	storage := canteen.NewMockMediaStorage()
	service := NewCanteenMediaCommandService(
		repo,
		repo,
		storage,
		canteen.MockImageProcessor{Result: canteen.ProcessedImage{
			Data: []byte("processed"), MIME: "image/jpeg", Width: 10, Height: 10, SHA256: "hash",
		}},
		StaticCanteenSettingsProvider{Runtime: CanteenRuntimeConfig{
			Enabled: true, ReviewsEnabled: true, MediaEnabled: true,
		}},
		CanteenMediaLimits{MaxBytes: 1024, MaxPerReview: 3, MaxUploadsPerDay: 5, MaxPixels: 1000},
	)
	return service, repo, storage
}

func TestCanteenMediaUploadFlow(t *testing.T) {
	service, repo, storage := newMediaCommandFixture(t)
	owner := &auth.User{ID: 2, Role: auth.RoleUser}
	media, uploadURL, err := service.CreateIntent(context.Background(), owner, "review", "image/jpeg", 512)
	if err != nil {
		t.Fatalf("CreateIntent() = %v", err)
	}
	if media.ID == 0 || uploadURL == "" || media.Status != canteen.MediaStatusPendingUpload {
		t.Fatalf("media = %+v url=%q", media, uploadURL)
	}
	uploaded, err := service.PutContent(context.Background(), owner, media.ID, bytes.NewReader([]byte("raw-image")))
	if err != nil {
		t.Fatalf("PutContent() = %v", err)
	}
	if uploaded.Status != canteen.MediaStatusPending || uploaded.ByteSize != len("processed") {
		t.Fatalf("uploaded = %+v", uploaded)
	}
	if _, ok := storage.Files[media.StorageKey]; !ok {
		t.Fatalf("storage missing key %q", media.StorageKey)
	}
	completed, err := service.Complete(context.Background(), owner, media.ID)
	if err != nil || completed.Status != canteen.MediaStatusPending {
		t.Fatalf("Complete() = %+v err=%v", completed, err)
	}

	other := &auth.User{ID: 3, Role: auth.RoleUser}
	if _, _, err := service.Open(context.Background(), other, media.ID); !errors.Is(err, canteen.ErrMediaNotFound) {
		t.Fatalf("Open(pending as other) = %v", err)
	}
	admin := &auth.User{ID: 1, Role: auth.RoleAdmin}
	if err := service.Approve(context.Background(), admin, media.ID, canteen.MediaStatusApproved, ""); err != nil {
		t.Fatalf("Approve() = %v", err)
	}
	repo.Public[media.ID] = true
	reader, approved, err := service.Open(context.Background(), other, media.ID)
	if err != nil {
		t.Fatalf("Open(approved) = %v", err)
	}
	_ = reader.Close()
	if approved.Status != canteen.MediaStatusApproved {
		t.Fatalf("approved = %+v", approved)
	}
	if err := service.Approve(context.Background(), admin, media.ID, canteen.MediaStatusApproved, ""); !errors.Is(err, canteen.ErrMediaState) {
		t.Fatalf("second Approve() = %v, want ErrMediaState", err)
	}
	if len(repo.Items) != 1 {
		t.Fatalf("media rows = %d", len(repo.Items))
	}
}

func TestCanteenMediaApprovedWithoutLiveParentIsPrivate(t *testing.T) {
	service, repo, _ := newMediaCommandFixture(t)
	owner := &auth.User{ID: 2, Role: auth.RoleUser}
	media, _, err := service.CreateIntent(context.Background(), owner, "review", "image/png", 100)
	if err != nil {
		t.Fatalf("CreateIntent() = %v", err)
	}
	if _, err := service.PutContent(context.Background(), owner, media.ID, bytes.NewReader([]byte("raw"))); err != nil {
		t.Fatalf("PutContent() = %v", err)
	}
	admin := &auth.User{ID: 1, Role: auth.RoleAdmin}
	if err := service.Approve(context.Background(), admin, media.ID, canteen.MediaStatusApproved, ""); err != nil {
		t.Fatalf("Approve() = %v", err)
	}
	// The media itself is approved, but no active review/dish exposes it.
	repo.Public[media.ID] = false
	other := &auth.User{ID: 3, Role: auth.RoleUser}
	if _, _, err := service.Open(context.Background(), other, media.ID); !errors.Is(err, canteen.ErrMediaNotFound) {
		t.Fatalf("Open(orphan approved as other) = %v, want ErrMediaNotFound", err)
	}
	reader, _, err := service.Open(context.Background(), owner, media.ID)
	if err != nil {
		t.Fatalf("Open(orphan approved as owner) = %v", err)
	}
	_ = reader.Close()
}

func TestCanteenMediaOwnership(t *testing.T) {
	service, _, _ := newMediaCommandFixture(t)
	owner := &auth.User{ID: 2, Role: auth.RoleUser}
	media, _, err := service.CreateIntent(context.Background(), owner, "review", "image/png", 100)
	if err != nil {
		t.Fatalf("CreateIntent() = %v", err)
	}
	other := &auth.User{ID: 3, Role: auth.RoleUser}
	if _, err := service.PutContent(context.Background(), other, media.ID, bytes.NewReader([]byte("x"))); !errors.Is(err, canteen.ErrMediaNotOwned) {
		t.Fatalf("PutContent(other) = %v, want ErrMediaNotOwned", err)
	}
}

func TestCanteenReportCreateAndResolve(t *testing.T) {
	repo := canteen.NewMockReportRepository()
	service := NewCanteenReportCommandService(repo, StaticCanteenSettingsProvider{
		Runtime: CanteenRuntimeConfig{Enabled: true, ReviewsEnabled: true},
	})
	user := &auth.User{ID: 2, Role: auth.RoleUser}
	report, err := service.Create(context.Background(), user, CreateCanteenReportCommand{
		SubjectType: canteen.ReportSubjectReview,
		SubjectID:   7,
		Reason:      canteen.ReportReasonSpam,
		Detail:      "重复广告",
	})
	if err != nil {
		t.Fatalf("Create() = %v", err)
	}
	if report.ID == 0 || report.Status != canteen.ReportStatusOpen {
		t.Fatalf("report = %+v", report)
	}
	if _, _, _, _, err := service.List(context.Background(), user, canteen.ReportFilter{}); !errors.Is(err, canteen.ErrCannotModerate) {
		t.Fatalf("List(user) = %v", err)
	}
	admin := &auth.User{ID: 1, Role: auth.RoleAdmin}
	items, total, _, _, err := service.List(context.Background(), admin, canteen.ReportFilter{})
	if err != nil || total != 1 || len(items) != 1 {
		t.Fatalf("List(admin) = %+v total=%d err=%v", items, total, err)
	}
	if err := service.Resolve(context.Background(), admin, report.ID, canteen.ReportStatusResolved, "已处理"); err != nil {
		t.Fatalf("Resolve() = %v", err)
	}
	if err := service.Resolve(context.Background(), admin, report.ID, canteen.ReportStatusOpen, ""); err == nil {
		t.Fatal("Resolve(open) should be rejected")
	}
}
