package controller

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/canteen"
	"jcourse/internal/infrastructure/media"
)

func TestCanteenMediaGetUsesPrivateNoStoreForAllResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := canteen.NewMockMediaRepository()
	repo.Items[1] = &canteen.Media{
		ID: 1, OwnerUserID: 2, Purpose: canteen.MediaPurposeReview, StorageDriver: "local",
		StorageKey: "review/hidden.png", MIME: "image/png", ByteSize: 4, Width: 8, Height: 8,
		SHA256: "hidden", Status: canteen.MediaStatusApproved,
	}
	storage, err := media.NewLocalStorage(t.TempDir())
	if err != nil {
		t.Fatalf("storage: %v", err)
	}
	if err := storage.Put(context.Background(), "review/hidden.png", bytes.NewReader([]byte("png!"))); err != nil {
		t.Fatalf("put: %v", err)
	}
	service := application.NewCanteenMediaCommandService(
		repo, repo, storage, nil,
		application.StaticCanteenSettingsProvider{Runtime: application.CanteenRuntimeConfig{Enabled: true, MediaEnabled: true}},
		application.CanteenMediaLimits{},
	)
	ctrl := NewCanteenMediaController(service, nil)

	request := func(userID int) *httptest.ResponseRecorder {
		t.Helper()
		router := gin.New()
		router.GET("/media/:mediaID", func(ctx *gin.Context) {
			ctx.Request = ctx.Request.WithContext(auth.WithUser(ctx.Request.Context(), &auth.User{ID: userID, Role: auth.RoleUser}))
			ctrl.Get(ctx)
		})
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/media/1", nil)
		router.ServeHTTP(recorder, req)
		return recorder
	}

	// Hidden parent: owner can still read, but the response must never be
	// stored by a shared cache.
	owner := request(2)
	if owner.Code != http.StatusOK {
		t.Fatalf("owner status = %d, body=%s", owner.Code, owner.Body.String())
	}
	if got := owner.Header().Get("Cache-Control"); got != "private, no-store" {
		t.Fatalf("owner Cache-Control = %q", got)
	}
	if other := request(3); other.Code != http.StatusNotFound {
		t.Fatalf("other status = %d, want 404", other.Code)
	}

	// Even when the parent is currently public, keep the conservative policy
	// until an explicit public-cache invalidation design exists.
	repo.Public[1] = true
	public := request(2)
	if public.Code != http.StatusOK {
		t.Fatalf("public status = %d", public.Code)
	}
	if got := public.Header().Get("Cache-Control"); got != "private, no-store" {
		t.Fatalf("public Cache-Control = %q", got)
	}
}
