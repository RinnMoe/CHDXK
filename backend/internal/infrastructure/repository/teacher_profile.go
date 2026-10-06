package repository

import (
	"context"
	"encoding/json"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"jcourse/internal/domain/teacher"
	"jcourse/pkg/apperr"
)

type TeacherProfileSourceEntity struct {
	URL  string         `gorm:"primaryKey"`
	Data datatypes.JSON `gorm:"type:jsonb;not null"`
}

func (TeacherProfileSourceEntity) TableName() string { return "teacher_profile_sources" }
func (r *TeacherRepository) Teachers(ctx context.Context) ([]teacher.TeacherView, error) {
	rows, _, err := r.FindBy(ctx, teacher.TeacherFilter{})
	return rows, err
}
func (r *TeacherRepository) Save(ctx context.Context, teachers []teacher.TeacherView, profiles []teacher.Profile) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range profiles {
			p := profiles[i]
			if p.UpdatedAt.IsZero() {
				var old TeacherProfileSourceEntity
				if err := tx.Where("url = ?", p.URL).Take(&old).Error; err == nil {
					var previous teacher.Profile
					if json.Unmarshal(old.Data, &previous) == nil {
						p.Fields = previous.Fields
						p.UpdatedAt = previous.UpdatedAt
						profiles[i] = p
					}
				} else if err != gorm.ErrRecordNotFound {
					return err
				}
			}
			data, _ := json.Marshal(p)
			row := TeacherProfileSourceEntity{URL: p.URL, Data: data}
			if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "url"}}, DoUpdates: clause.AssignmentColumns([]string{"data"})}).Create(&row).Error; err != nil {
				return err
			}
		}
		var entities []TeacherEntity
		if err := tx.Find(&entities).Error; err != nil {
			return err
		}
		// Re-read local identities inside the transaction: a concurrent course import
		// may have introduced another teacher with the same name and college.
		current := make([]teacher.TeacherView, len(entities))
		counts := map[string]int{}
		byIdentity := map[string][]teacher.Profile{}
		for _, p := range profiles {
			key := teacher.ProfileIdentity(p.Name, p.Department)
			byIdentity[key] = append(byIdentity[key], p)
		}
		for i := range entities {
			current[i] = newTeacherViewFromEntity(&entities[i])
			counts[teacher.ProfileIdentity(current[i].Name, current[i].Department)]++
		}
		for _, e := range entities {
			t := newTeacherViewFromEntity(&e)
			key := teacher.ProfileIdentity(t.Name, t.Department)
			candidates := byIdentity[key]
			if candidates == nil {
				candidates = []teacher.Profile{}
			}
			unique := counts[key] == 1 && len(candidates) == 1
			status := "unmatched"
			selected := ""
			profile := []byte(`{}`)
			if len(candidates) > 0 || counts[key] > 1 {
				status = "ambiguous"
			}
			for _, p := range candidates {
				if e.ProfileStatus == "manual" && e.ProfileSelectedURL == p.URL {
					status = "manual"
					selected = p.URL
					profile, _ = json.Marshal(p)
					break
				}
			}
			if status != "manual" && unique {
				status = "matched"
				selected = candidates[0].URL
				profile, _ = json.Marshal(candidates[0])
			}
			if len(e.ProfileOverride) > 2 {
				profile = e.ProfileOverride
			}
			data, _ := json.Marshal(candidates)
			if err := tx.Model(&TeacherEntity{}).Where("id = ?", e.ID).Updates(map[string]any{"profile": datatypes.JSON(profile), "profile_candidates": datatypes.JSON(data), "profile_status": status, "profile_selected_url": selected}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err == nil {
		r.invalidateProfileCaches(ctx)
	}
	return err
}
func (r *TeacherRepository) Matches(ctx context.Context) ([]teacher.ProfileMatch, error) {
	var entities []TeacherEntity
	if err := r.db.WithContext(ctx).Where("profile_status IN ?", []string{"ambiguous", "manual"}).Order("department, name, id").Find(&entities).Error; err != nil {
		return nil, err
	}
	result := make([]teacher.ProfileMatch, 0, len(entities))
	for _, e := range entities {
		candidates := []teacher.Profile{}
		if err := json.Unmarshal(e.ProfileCandidates, &candidates); err != nil {
			return nil, err
		}
		result = append(result, teacher.ProfileMatch{TeacherID: e.ID, Name: e.Name, Department: e.Department, Status: e.ProfileStatus, SelectedURL: e.ProfileSelectedURL, Candidates: candidates})
	}
	return result, nil
}
func (r *TeacherRepository) Bind(ctx context.Context, id int, address string) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var entity TeacherEntity
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&entity, id).Error; err != nil {
			return err
		}
		var candidates []teacher.Profile
		if err := json.Unmarshal(entity.ProfileCandidates, &candidates); err != nil {
			return err
		}
		for _, p := range candidates {
			if p.URL == address {
				data, _ := json.Marshal(p)
				if len(entity.ProfileOverride) > 2 {
					data = entity.ProfileOverride
				}
				return tx.Model(&entity).Updates(map[string]any{"profile": datatypes.JSON(data), "profile_selected_url": address, "profile_status": "manual"}).Error
			}
		}
		return apperr.BadRequest("请选择该教师的候选主页")
	})
	if err == nil {
		r.invalidateProfileCaches(ctx)
	}
	return err
}
func (r *TeacherRepository) invalidateProfileCaches(ctx context.Context) {
	cacheDeletePattern(ctx, r.cache, cacheKey("teacher", "*"))
	cacheDeletePattern(ctx, r.cache, cacheKey("course", "*"))
}

var _ teacher.ProfileStore = (*TeacherRepository)(nil)

func (r *TeacherRepository) Edit(ctx context.Context, id int, p teacher.Profile) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var e TeacherEntity
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&e, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return apperr.BadRequest("教师不存在")
			}
			return err
		}
		changes := map[string]any{"name": p.Name, "department": p.Department, "title": p.Title, "profile": datatypes.JSON(data), "profile_override": datatypes.JSON(data), "search_vector": TeacherSearchVectorExpr(r.searchConfig, e.Code, teacher.NewSearchName(p.Name))}
		if teacher.ProfileIdentity(e.Name, e.Department) != teacher.ProfileIdentity(p.Name, p.Department) {
			changes["profile_status"] = "unmatched"
			changes["profile_selected_url"] = ""
			changes["profile_candidates"] = datatypes.JSON([]byte(`[]`))
		}
		return tx.Model(&e).Updates(changes).Error
	})
	if err == nil {
		r.invalidateProfileCaches(ctx)
	}
	return err
}

func (r *TeacherRepository) List(ctx context.Context, f teacher.TeacherFilter) ([]teacher.TeacherView, int64, error) {
	return r.FindBy(ctx, f)
}
