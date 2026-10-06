package repository

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"jcourse/internal/domain/course"
	teacherdomain "jcourse/internal/domain/teacher"
	"jcourse/pkg/apperr"
)

type managedOfferingRow struct {
	ID          int            `gorm:"column:id"`
	CourseID    int            `gorm:"column:course_id"`
	Semester    string         `gorm:"column:semester"`
	Language    string         `gorm:"column:language"`
	TargetYears pq.StringArray `gorm:"column:target_years;type:text[]"`
	Categories  pq.StringArray `gorm:"column:categories;type:text[]"`
	Code        string         `gorm:"column:code"`
}

func (r *CourseRepository) ListCourseSemesters(ctx context.Context) ([]string, error) {
	var semesters []string
	err := r.db.WithContext(ctx).Model(&OfferedCourseEntity{}).
		Distinct("semester").Where("semester <> ''").Order("semester DESC").Pluck("semester", &semesters).Error
	return semesters, err
}

func (r *CourseRepository) ListManagedOfferings(ctx context.Context, filter course.ManagedOfferingFilter) ([]course.ManagedOffering, int64, error) {
	db := r.db.WithContext(ctx).Table("offered_courses AS oc").
		Joins("JOIN courses AS c ON c.id = oc.course_id").
		Where("oc.semester = ?", filter.Semester)
	// Treat wildcard characters as literal text; column filters combine with AND.
	like := func(value string) string {
		return "%" + strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`).Replace(strings.TrimSpace(value)) + "%"
	}
	if strings.TrimSpace(filter.Code) != "" {
		db = db.Where("c.code ILIKE ?", like(filter.Code))
	}
	sections := r.db.Table("course_sections AS cs").Select("1").Where("cs.offered_course_id = oc.id")
	filtered := false
	for _, field := range []struct{ column, value string }{
		{"cs.course_name", filter.Name}, {"cs.department", filter.Department}, {"cs.lesson_no", filter.LessonNo},
	} {
		if strings.TrimSpace(field.value) != "" {
			sections = sections.Where(field.column+" ILIKE ?", like(field.value))
			filtered = true
		}
	}
	if strings.TrimSpace(filter.Teacher) != "" {
		sections = sections.Where("EXISTS (SELECT 1 FROM teachers t WHERE t.id = cs.teacher_ids[1] AND t.name ILIKE ?)", like(filter.Teacher))
		filtered = true
	}
	if filtered {
		db = db.Where("EXISTS (?)", sections)
	}
	sectionCount := "(SELECT COUNT(*) FROM course_sections cs WHERE cs.offered_course_id = oc.id)"
	if filter.MinSections != nil {
		db = db.Where(sectionCount+" >= ?", *filter.MinSections)
	}
	if filter.MaxSections != nil {
		db = db.Where(sectionCount+" <= ?", *filter.MaxSections)
	}
	sortColumn := map[string]string{
		"code":       "c.code",
		"name":       "(SELECT cs.course_name FROM course_sections cs WHERE cs.offered_course_id = oc.id ORDER BY cs.id LIMIT 1)",
		"department": "(SELECT cs.department FROM course_sections cs WHERE cs.offered_course_id = oc.id ORDER BY cs.id LIMIT 1)",
		"teacher":    "(SELECT t.name FROM course_sections cs JOIN teachers t ON t.id = cs.teacher_ids[1] WHERE cs.offered_course_id = oc.id ORDER BY cs.id LIMIT 1)",
		"sections":   sectionCount,
	}[filter.Sort]
	if sortColumn == "" {
		sortColumn = "c.code"
	}
	direction := " ASC"
	if filter.Desc {
		direction = " DESC"
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 50
	}
	var rows []managedOfferingRow
	err := db.Select("oc.id, oc.course_id, oc.semester, oc.language, oc.target_years, oc.categories, c.code").
		Order(sortColumn + direction + " NULLS LAST, oc.id ASC").
		Offset((filter.Page - 1) * filter.PageSize).Limit(filter.PageSize).Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	offeringIDs := make([]int, len(rows))
	for i := range rows {
		offeringIDs[i] = rows[i].ID
	}
	sectionsByOffering, err := r.loadManagedSections(ctx, offeringIDs)
	if err != nil {
		return nil, 0, err
	}
	offerings := make([]course.ManagedOffering, len(rows))
	for i, row := range rows {
		offerings[i] = course.ManagedOffering{
			ID:       row.ID,
			CourseID: row.CourseID,
			Code:     row.Code,
			Semester: row.Semester,
			Sections: sectionsByOffering[row.ID],
		}
	}
	return offerings, total, nil
}

func (r *CourseRepository) loadManagedSections(ctx context.Context, offeringIDs []int) (map[int][]course.CourseSection, error) {
	result := make(map[int][]course.CourseSection, len(offeringIDs))
	for _, id := range offeringIDs {
		result[id] = []course.CourseSection{}
	}
	if len(offeringIDs) == 0 {
		return result, nil
	}
	sections, err := gorm.G[CourseSectionEntity](r.db).
		Where("offered_course_id IN ?", offeringIDs).Order("id ASC").Find(ctx)
	if err != nil {
		return nil, err
	}
	allTeacherIDs := make([]int64, 0)
	for _, section := range sections {
		allTeacherIDs = append(allTeacherIDs, section.TeacherIDs...)
	}
	teacherNames := map[int]string{}
	if len(allTeacherIDs) > 0 {
		allTeacherIDs = uniqueInt64s(allTeacherIDs)
		var teachers []TeacherEntity
		if err := r.db.WithContext(ctx).Select("id, name").Where("id IN ?", allTeacherIDs).Find(&teachers).Error; err != nil {
			return nil, err
		}
		for _, teacher := range teachers {
			teacherNames[teacher.ID] = teacher.Name
		}
	}
	for _, section := range sections {
		teachers := make([]string, 0, len(section.TeacherIDs))
		profiles := make([]course.CourseSectionTeacher, 0, len(section.TeacherIDs))
		for position, id := range section.TeacherIDs {
			if name, ok := teacherNames[int(id)]; ok {
				teachers = append(teachers, name)
				role := "co-teacher"
				if position == 0 {
					role = "primary"
				}
				profiles = append(profiles, course.CourseSectionTeacher{
					ID: int(id), Name: name, Role: role, Position: position,
				})
			}
		}
		result[section.OfferedCourseID] = append(result[section.OfferedCourseID], course.CourseSection{
			ID:              section.ID,
			SourceLessonID:  section.SourceLessonID,
			LessonNo:        section.LessonNo,
			CourseName:      section.CourseName,
			CourseType:      section.CourseType,
			Credit:          section.Credit,
			Department:      section.Department,
			Language:        section.Language,
			Campus:          section.Campus,
			Assessment:      section.Assessment,
			TeachClass:      section.TeachClass,
			ActualCount:     section.ActualCount,
			LimitCount:      section.LimitCount,
			CoursePeriod:    section.CoursePeriod,
			TotalHours:      section.TotalHours,
			WeeklyHours:     section.WeeklyHours,
			WeekRange:       section.WeekRange,
			StartWeek:       section.StartWeek,
			EndWeek:         section.EndWeek,
			WeekCount:       section.WeekCount,
			Remark:          section.Remark,
			SyllabusURL:     section.SyllabusURL,
			Teachers:        teachers,
			TeacherProfiles: profiles,
		})
	}
	return result, nil
}

func (r *CourseRepository) ReplaceSemester(ctx context.Context, semester string, offerings []course.ManagedOffering) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("semester = ?", semester).Delete(&OfferedCourseEntity{}).Error; err != nil {
			return err
		}
		return r.saveOfferingsTx(ctx, tx, semester, offerings)
	})
	if err == nil {
		r.invalidateCourseData(ctx)
	}
	return err
}

func (r *CourseRepository) SaveManagedOffering(ctx context.Context, id int, offering course.ManagedOffering) (*course.ManagedOffering, error) {
	var result course.ManagedOffering
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if id > 0 {
			var existing OfferedCourseEntity
			if err := tx.Where("id = ?", id).Take(&existing).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return apperr.NotFound("课程开课记录不存在")
				}
				return err
			}
			if existing.Semester != offering.Semester {
				return apperr.BadRequest("编辑课程时不能更改所属学期")
			}
			offering.ID = existing.ID
		}
		offerings := []course.ManagedOffering{offering}
		if err := r.saveOfferingsTx(ctx, tx, offering.Semester, offerings); err != nil {
			return err
		}
		result = offerings[0]
		return nil
	})
	if err == nil {
		r.invalidateCourseData(ctx)
		return &result, nil
	}
	return nil, err
}

func (r *CourseRepository) DeleteManagedOffering(ctx context.Context, id int) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&OfferedCourseEntity{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperr.NotFound("课程开课记录不存在")
	}
	r.invalidateCourseData(ctx)
	return nil
}

func (r *CourseRepository) saveOfferingsTx(ctx context.Context, tx *gorm.DB, semester string, offerings []course.ManagedOffering) error {
	teacherRecords := map[string]TeacherEntity{}
	for _, offering := range offerings {
		for _, section := range offering.Sections {
			for _, name := range section.Teachers {
				code := course.InternalTeacherCode(section.Department, name)
				teacherRecords[code] = TeacherEntity{Code: code, Name: course.NormalizeTeacherName(name), Department: section.Department}
			}
		}
	}
	teacherIDs, err := upsertCourseTeachers(ctx, tx, r.searchConfig, semester, teacherRecords)
	if err != nil {
		return err
	}
	courseIDs, err := upsertOfferingCourses(ctx, tx, semester, offerings, teacherIDs)
	if err != nil {
		return err
	}
	for i := range offerings {
		offering := &offerings[i]
		courseID := courseIDs[offeringKey(offering.Code, *offering, teacherIDs)]
		var existing OfferedCourseEntity
		var err error
		if offering.ID > 0 {
			err = tx.Where("id = ? AND semester = ?", offering.ID, semester).Take(&existing).Error
		} else {
			err = tx.Where("course_id = ? AND semester = ?", courseID, semester).Take(&existing).Error
		}
		var offeringID int
		if err == nil {
			offeringID = existing.ID
			var duplicate int64
			if err := tx.Model(&OfferedCourseEntity{}).
				Where("course_id = ? AND semester = ? AND id <> ?", courseID, semester, offeringID).
				Count(&duplicate).Error; err != nil {
				return err
			}
			if duplicate > 0 {
				return apperr.BadRequest("该课程本学期已存在开课记录")
			}
			if err := tx.Model(&OfferedCourseEntity{}).Where("id = ?", offeringID).Updates(map[string]any{
				"course_id":    courseID,
				"language":     firstSectionLanguage(*offering),
				"target_years": pq.StringArray{},
				"categories":   pq.StringArray(offeringCategories(*offering)),
				"teacher_ids":  pq.Int64Array(offeringTeacherIDs(*offering, teacherIDs)),
			}).Error; err != nil {
				return err
			}
			if err := tx.Where("offered_course_id = ?", offeringID).Delete(&CourseSectionEntity{}).Error; err != nil {
				return err
			}
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			entity := OfferedCourseEntity{
				CourseID:    courseID,
				Semester:    semester,
				Language:    firstSectionLanguage(*offering),
				TargetYears: pq.StringArray{},
				Categories:  pq.StringArray(offeringCategories(*offering)),
				TeacherIDs:  pq.Int64Array(offeringTeacherIDs(*offering, teacherIDs)),
				CreatedAt:   time.Now(),
			}
			if err := tx.Create(&entity).Error; err != nil {
				return err
			}
			offeringID = entity.ID
		} else if offering.ID > 0 {
			return apperr.NotFound("课程开课记录不存在")
		} else {
			return err
		}
		sectionRows := make([]CourseSectionEntity, 0, len(offering.Sections))
		for _, section := range offering.Sections {
			sectionRows = append(sectionRows, newCourseSectionEntity(offeringID, section, teacherIDs))
		}
		if len(sectionRows) > 0 {
			if err := tx.Create(&sectionRows).Error; err != nil {
				return err
			}
		}
		offering.ID = offeringID
		offering.CourseID = courseID
	}
	if err := RefreshCourseSearchVectors(tx, semester); err != nil {
		return err
	}
	return nil
}

func upsertCourseTeachers(ctx context.Context, tx *gorm.DB, searchConfig, semester string, records map[string]TeacherEntity) (map[string]int, error) {
	if len(records) == 0 {
		return map[string]int{}, nil
	}
	codes := make([]string, 0, len(records))
	entities := make([]map[string]any, 0, len(records))
	for code, teacher := range records {
		codes = append(codes, code)
		entities = append(entities, map[string]any{
			"code":          code,
			"name":          teacher.Name,
			"department":    teacher.Department,
			"title":         "",
			"search_vector": TeacherSearchVectorExpr(searchConfig, code, teacherdomain.NewSearchName(teacher.Name)),
			"last_semester": semester,
			"created_at":    time.Now(),
			"updated_at":    time.Now(),
		})
	}
	conflict := clause.OnConflict{
		Columns: []clause.Column{{Name: "code"}},
		DoUpdates: clause.Assignments(map[string]any{
			"name":          gorm.Expr("CASE WHEN teachers.profile_override = '{}'::jsonb THEN excluded.name ELSE teachers.name END"),
			"department":    gorm.Expr("CASE WHEN teachers.profile_override = '{}'::jsonb THEN excluded.department ELSE teachers.department END"),
			"search_vector": gorm.Expr("CASE WHEN teachers.profile_override = '{}'::jsonb THEN excluded.search_vector ELSE teachers.search_vector END"),
			"last_semester": clause.Column{Table: "excluded", Name: "last_semester"},
			"updated_at":    clause.Column{Table: "excluded", Name: "updated_at"},
		}),
		Where: clause.Where{Exprs: []clause.Expression{
			clause.Expr{SQL: "teachers.last_semester <= ?", Vars: []any{semester}},
		}},
	}
	if err := tx.WithContext(ctx).Model(&TeacherEntity{}).Clauses(conflict).Create(&entities).Error; err != nil {
		return nil, err
	}
	var teachers []TeacherEntity
	if err := tx.WithContext(ctx).Select("id, code").Where("code IN ?", codes).Find(&teachers).Error; err != nil {
		return nil, err
	}
	result := make(map[string]int, len(teachers))
	for _, teacher := range teachers {
		result[teacher.Code] = teacher.ID
	}
	return result, nil
}

func upsertOfferingCourses(ctx context.Context, tx *gorm.DB, semester string, offerings []course.ManagedOffering, teacherIDs map[string]int) (map[string]int, error) {
	type target struct {
		code       string
		mainID     int
		hasMain    bool
		section    course.CourseSection
		teacherID  []int64
		categories []string
	}
	targets := map[string]target{}
	for _, offering := range offerings {
		if len(offering.Sections) == 0 {
			continue
		}
		mainCode := ""
		mainID := 0
		if len(offering.Sections[0].Teachers) > 0 {
			mainCode = course.InternalTeacherCode(offering.Sections[0].Department, offering.Sections[0].Teachers[0])
			mainID = teacherIDs[mainCode]
		}
		ids := offeringTeacherIDs(offering, teacherIDs)
		key := offeringKey(offering.Code, offering, teacherIDs)
		targets[key] = target{
			code: offering.Code, mainID: mainID, hasMain: mainCode != "", section: offering.Sections[0],
			teacherID: ids, categories: categoriesFromSections(offering.Sections),
		}
	}
	teacherTargets := make([]map[string]any, 0, len(targets))
	teacherlessTargets := make([]target, 0)
	for _, item := range targets {
		values := map[string]any{
			"code":            item.code,
			"main_teacher_id": nil,
			"name":            item.section.CourseName,
			"credit":          item.section.Credit,
			"department":      item.section.Department,
			"target_years":    pq.StringArray{},
			"language":        item.section.Language,
			"categories":      pq.StringArray(item.categories),
			"teacher_ids":     pq.Int64Array(item.teacherID),
			"last_semester":   semester,
			"created_at":      time.Now(),
		}
		if item.hasMain {
			values["main_teacher_id"] = item.mainID
			teacherTargets = append(teacherTargets, values)
		} else {
			teacherlessTargets = append(teacherlessTargets, item)
		}
	}
	if len(teacherTargets) > 0 {
		conflict := clause.OnConflict{
			Columns: []clause.Column{{Name: "code"}, {Name: "main_teacher_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"name":          clause.Column{Table: "excluded", Name: "name"},
				"credit":        clause.Column{Table: "excluded", Name: "credit"},
				"department":    clause.Column{Table: "excluded", Name: "department"},
				"language":      clause.Column{Table: "excluded", Name: "language"},
				"categories":    clause.Column{Table: "excluded", Name: "categories"},
				"teacher_ids":   clause.Column{Table: "excluded", Name: "teacher_ids"},
				"last_semester": clause.Column{Table: "excluded", Name: "last_semester"},
			}),
			Where: clause.Where{Exprs: []clause.Expression{
				clause.Expr{SQL: "courses.last_semester <= ?", Vars: []any{semester}},
			}},
		}
		if err := tx.WithContext(ctx).Model(&CourseEntity{}).Clauses(conflict).Create(&teacherTargets).Error; err != nil {
			return nil, err
		}
	}
	for _, item := range teacherlessTargets {
		var existing CourseEntity
		err := tx.WithContext(ctx).Where("code = ? AND main_teacher_id IS NULL", item.code).Order("id ASC").Take(&existing).Error
		values := map[string]any{
			"name":          item.section.CourseName,
			"credit":        item.section.Credit,
			"department":    item.section.Department,
			"language":      item.section.Language,
			"categories":    pq.StringArray(item.categories),
			"target_years":  pq.StringArray{},
			"teacher_ids":   pq.Int64Array{},
			"last_semester": semester,
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			values["code"] = item.code
			values["main_teacher_id"] = nil
			values["created_at"] = time.Now()
			if err := tx.WithContext(ctx).Model(&CourseEntity{}).Create(values).Error; err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		} else if existing.LastSemester <= semester {
			if err := tx.WithContext(ctx).Model(&CourseEntity{}).Where("id = ?", existing.ID).Updates(values).Error; err != nil {
				return nil, err
			}
		}
	}

	codes := make([]string, 0, len(targets))
	for _, item := range targets {
		codes = append(codes, item.code)
	}
	codes = uniqueStrings(codes)
	var stored []CourseEntity
	if len(codes) > 0 {
		if err := tx.WithContext(ctx).Select("id, code, main_teacher_id").Where("code IN ?", codes).Find(&stored).Error; err != nil {
			return nil, err
		}
	}
	byKey := make(map[string]int, len(stored))
	for _, item := range stored {
		byKey[fmt.Sprintf("%s\x00%d", item.Code, item.MainTeacherID)] = item.ID
	}
	result := make(map[string]int, len(targets))
	for key, item := range targets {
		result[key] = byKey[fmt.Sprintf("%s\x00%d", item.code, item.mainID)]
		if result[key] == 0 {
			return nil, fmt.Errorf("resolve course id for %s", item.code)
		}
	}
	return result, nil
}

func newCourseSectionEntity(offeringID int, section course.CourseSection, teacherIDs map[string]int) CourseSectionEntity {
	ids := make([]int64, 0, len(section.Teachers))
	for _, name := range section.Teachers {
		if id := teacherIDs[course.InternalTeacherCode(section.Department, name)]; id > 0 {
			ids = append(ids, int64(id))
		}
	}
	return CourseSectionEntity{
		OfferedCourseID: offeringID,
		SourceLessonID:  section.SourceLessonID,
		LessonNo:        section.LessonNo,
		CourseName:      section.CourseName,
		CourseType:      section.CourseType,
		Credit:          section.Credit,
		Department:      section.Department,
		Language:        section.Language,
		Campus:          section.Campus,
		Assessment:      section.Assessment,
		TeachClass:      section.TeachClass,
		ActualCount:     section.ActualCount,
		LimitCount:      section.LimitCount,
		CoursePeriod:    section.CoursePeriod,
		TotalHours:      section.TotalHours,
		WeeklyHours:     section.WeeklyHours,
		WeekRange:       section.WeekRange,
		StartWeek:       section.StartWeek,
		EndWeek:         section.EndWeek,
		WeekCount:       section.WeekCount,
		Remark:          section.Remark,
		SyllabusURL:     section.SyllabusURL,
		TeacherIDs:      pq.Int64Array(ids),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

func offeringTeacherIDs(offering course.ManagedOffering, teachers map[string]int) []int64 {
	ids := make([]int64, 0)
	for _, section := range offering.Sections {
		for _, name := range section.Teachers {
			if id := teachers[course.InternalTeacherCode(section.Department, name)]; id > 0 {
				ids = append(ids, int64(id))
			}
		}
	}
	return uniqueInt64s(ids)
}

func offeringKey(code string, offering course.ManagedOffering, teachers map[string]int) string {
	mainID := 0
	if len(offering.Sections) > 0 && len(offering.Sections[0].Teachers) > 0 {
		mainID = teachers[course.InternalTeacherCode(offering.Sections[0].Department, offering.Sections[0].Teachers[0])]
	}
	return fmt.Sprintf("%s\x00%d", code, mainID)
}

func firstSectionLanguage(offering course.ManagedOffering) string {
	if len(offering.Sections) == 0 {
		return ""
	}
	return offering.Sections[0].Language
}

func categoriesFromSections(sections []course.CourseSection) []string {
	values := make([]string, 0, len(sections))
	for _, section := range sections {
		if section.CourseType != "" {
			values = append(values, section.CourseType)
		}
	}
	return uniqueStrings(values)
}

func offeringCategories(offering course.ManagedOffering) []string {
	return categoriesFromSections(offering.Sections)
}

func uniqueStrings(values []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func uniqueInt64s(values []int64) []int64 {
	seen := map[int64]struct{}{}
	result := make([]int64, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func (r *CourseRepository) invalidateCourseData(ctx context.Context) {
	cacheDeletePattern(ctx, r.cache, "jcourse:course:*")
	cacheDeletePattern(ctx, r.cache, "jcourse:teacher:*")
	cacheDeletePattern(ctx, r.cache, "jcourse:review:*")
}

var _ course.CourseDataRepository = (*CourseRepository)(nil)
