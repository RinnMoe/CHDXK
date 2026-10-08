//go:build test

package application

import (
	"context"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/setting"
	"testing"
)

func TestImportSettingsAuthorizationRuntimeAndPersistence(t *testing.T) {
	ctx := context.Background()
	repo := setting.NewMockSystemRepository()
	settings := setting.NewSystemSettingsService(repo, setting.NewRegistry(ImportSettingDefinitions()))
	q := NewCollectorQueue(128, 64)
	q.UseSettings(settings)
	command := NewSystemSettingsCommandService(settings)
	command.SetCollectorQueue(q)
	admin := &auth.User{ID: 2, Role: auth.RoleAdmin, Permissions: []string{auth.PermissionSettings}}
	super := &auth.User{ID: 1, Role: auth.RoleSuperAdmin}
	update := func(actor *auth.User, key, value string) error {
		_, err := command.Update(ctx, actor, key, UpdateSystemSettingCommand{Value: value})
		return err
	}
	if update(admin, ImportConcurrency, "2") == nil {
		t.Fatal("delegated admin changed import settings")
	}
	for _, value := range []string{"0", "65", "1.5"} {
		if update(super, ImportConcurrency, value) == nil {
			t.Fatal("invalid concurrency accepted", value)
		}
	}
	if err := update(super, ImportConcurrency, "1"); err != nil {
		t.Fatal(err)
	}
	first, _ := q.enqueue()
	second, _ := q.enqueue()
	defer first.release()
	defer second.release()
	if second.started {
		t.Fatal("queue bypassed configured concurrency")
	}
	if err := update(super, ImportConcurrency, "2"); err != nil {
		t.Fatal(err)
	}
	if !second.started {
		t.Fatal("increase did not wake waiting job")
	}
	if err := update(super, ImportConcurrency, "1"); err != nil {
		t.Fatal(err)
	}
	third, _ := q.enqueue()
	defer third.release()
	if third.started || !first.started || !second.started {
		t.Fatal("decrease interrupted running tasks or admitted new task")
	}
	for _, key := range []string{ImportSchoolEnabled, ImportAdminEnabled, ImportManualEnabled} {
		if err := update(super, key, "false"); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := q.enqueue(); err == nil {
		t.Fatal("disabled personal import accepted")
	}
	if _, err := q.enqueue(true); err == nil {
		t.Fatal("disabled administrative import accepted")
	}
	// A direct application call cannot bypass the manual-add switch.
	manual := &CourseCommandService{importSettings: settings}
	if err := manual.SaveEnrollment(ctx, 2, 0, 1); err == nil {
		t.Fatal("disabled manual import accepted")
	}
	restarted := NewCollectorQueue(128, 64)
	restarted.UseSettings(settings)
	_, err := restarted.enqueue()
	if err == nil || restarted.concurrency != 1 {
		t.Fatal("saved concurrency or switch not restored", err)
	}
	query := NewSystemSettingsQueryService(settings)
	items, err := query.ListAdmin(auth.WithUser(ctx, admin))
	if err != nil || len(items) != 0 {
		t.Fatal("import settings exposed as editable to delegated admin", err)
	}
	items, err = query.ListAdmin(auth.WithUser(ctx, super))
	if err != nil || len(items) != 5 {
		t.Fatal("superadmin settings missing", err)
	}
	public, err := query.List(ctx)
	if err != nil || len(public) != 4 {
		t.Fatal("public availability switches missing", err)
	}
	first.release()
	second.release()
	if !third.started {
		t.Fatal("accepted waiter lost after disabling")
	}
}
