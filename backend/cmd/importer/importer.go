package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	teacherdomain "jcourse/internal/domain/teacher"
	"jcourse/internal/infrastructure/repository"
	"jcourse/pkg/logx"
)

const batchSize = 100

type Importer struct {
	db       *gorm.DB
	semester string
}

func NewImporter(db *gorm.DB, semester string) *Importer {
	return &Importer{db: db, semester: semester}
}

func (imp *Importer) Run(ctx context.Context, rows []CSVRow) error {
	logx.Info(ctx, "collecting unique entities from csv data")
	teachers, courses := collectUnique(rows)

	logx.Info(ctx, "upserting teachers")
	imp.upsertTeachers(ctx, teachers)

	logx.Info(ctx, "resolving teacher ids")
	teacherIDMap := imp.resolveTeacherIDs(ctx)

	logx.Info(ctx, "upserting courses")
	imp.upsertCourses(ctx, courses, teacherIDMap)

	logx.Info(ctx, "resolving course ids")
	courseMap := imp.resolveCourseIDs(ctx)

	logx.Info(ctx, "creating offered courses")
	imp.upsertOfferedCourses(ctx, rows, teacherIDMap, courseMap)

	logx.Info(ctx, "syncing course languages from last offered courses")
	if err := imp.syncCourseLanguagesFromLastOfferings(ctx); err != nil {
		return err
	}

	logx.Info(ctx, "refreshing course search vectors")
	if err := repository.RefreshCourseSearchVectors(imp.db); err != nil {
		return err
	}

	logx.Info(ctx, "import complete")
	return nil
}

// courseKey returns the composite key for a course: code|teacherCode
func courseKey(code, teacherCode string) string {
	return code + "|" + teacherCode
}

func collectUnique(rows []CSVRow) (
	teachers map[string]TeacherInfo,
	courses map[string]CSVRow,
) {
	teachers = make(map[string]TeacherInfo)
	courses = make(map[string]CSVRow)

	for _, r := range rows {
		for _, t := range r.AllTeachers {
			if t.Code != "" {
				teachers[t.Code] = t
			}
		}
		if r.CourseCode != "" && r.MainTeacher.Code != "" {
			courses[courseKey(r.CourseCode, r.MainTeacher.Code)] = r
		}
	}
	return
}

func (imp *Importer) upsertTeachers(ctx context.Context, teachers map[string]TeacherInfo) {
	config := repository.SearchConfig(imp.db)
	onConflict := clause.OnConflict{
		Columns: []clause.Column{{Name: "code"}},
		DoUpdates: clause.Assignments(map[string]any{
			"name":          clause.Column{Table: "excluded", Name: "name"},
			"department":    clause.Column{Table: "excluded", Name: "department"},
			"title":         clause.Column{Table: "excluded", Name: "title"},
			"search_vector": clause.Column{Table: "excluded", Name: "search_vector"},
			"last_semester": clause.Column{Table: "excluded", Name: "last_semester"},
			"updated_at":    clause.Column{Table: "excluded", Name: "updated_at"},
		}),
		Where: clause.Where{
			Exprs: []clause.Expression{
				clause.Expr{SQL: "teachers.last_semester <= ?", Vars: []any{imp.semester}},
			},
		},
	}

	var batch []map[string]any
	count := 0
	for code, info := range teachers {
		now := time.Now()
		searchName := teacherdomain.NewSearchName(info.Name)
		batch = append(batch, map[string]any{
			"code":          code,
			"name":          info.Name,
			"department":    info.Department,
			"title":         info.Title,
			"search_vector": repository.TeacherSearchVectorExpr(config, code, searchName),
			"last_semester": imp.semester,
			"created_at":    now,
			"updated_at":    now,
		})
		if len(batch) >= batchSize {
			imp.db.Model(&repository.TeacherEntity{}).Clauses(onConflict).Create(&batch)
			batch = batch[:0]
		}
		count++
		if count%500 == 0 {
			logx.Info(ctx, "processed teachers", "processed", count, "total", len(teachers))
		}
	}
	if len(batch) > 0 {
		imp.db.Model(&repository.TeacherEntity{}).Clauses(onConflict).Create(&batch)
	}
	logx.Info(ctx, "teachers processed", "count", count)
}

func (imp *Importer) resolveTeacherIDs(ctx context.Context) map[string]int {
	var teachers []repository.TeacherEntity
	imp.db.Select("id, code").Find(&teachers)
	m := make(map[string]int, len(teachers))
	for _, t := range teachers {
		m[t.Code] = t.ID
	}
	logx.Info(ctx, "resolved teacher ids", "count", len(m))
	return m
}

func (imp *Importer) upsertCourses(ctx context.Context, courses map[string]CSVRow, teacherIDMap map[string]int) {
	onConflict := clause.OnConflict{
		Columns: []clause.Column{{Name: "code"}, {Name: "main_teacher_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"department":    clause.Column{Table: "excluded", Name: "department"},
			"language":      clause.Column{Table: "excluded", Name: "language"},
			"last_semester": clause.Column{Table: "excluded", Name: "last_semester"},
		}),
		Where: clause.Where{
			Exprs: []clause.Expression{
				clause.Expr{SQL: "courses.last_semester <= ?", Vars: []any{imp.semester}},
			},
		},
	}

	var batch []repository.CourseEntity
	skipped := 0
	for _, row := range courses {
		mainTeacherID := teacherIDMap[row.MainTeacher.Code]
		if mainTeacherID == 0 {
			logx.Warn(ctx, "skipping course without main teacher", "course_code", row.CourseCode, "course_name", row.CourseName, "teacher_code", row.MainTeacher.Code)
			skipped++
			continue
		}
		batch = append(batch, repository.CourseEntity{
			Code:          row.CourseCode,
			Name:          row.CourseName,
			Credit:        row.Credit,
			Department:    row.Department,
			MainTeacherID: mainTeacherID,
			Language:      row.Language,
			LastSemester:  imp.semester,
			CreatedAt:     time.Now(),
		})
		if len(batch) >= batchSize {
			imp.db.Clauses(onConflict).Create(&batch)
			batch = batch[:0]
		}
	}
	if len(batch) > 0 {
		imp.db.Clauses(onConflict).Create(&batch)
	}
	logx.Info(ctx, "courses imported", "imported", len(courses)-skipped, "skipped", skipped)
}

func (imp *Importer) resolveCourseIDs(ctx context.Context) map[string]repository.CourseEntity {
	var courses []repository.CourseEntity
	imp.db.Model(&repository.CourseEntity{}).
		Joins("MainTeacher").
		Find(&courses)
	m := make(map[string]repository.CourseEntity, len(courses))
	for _, c := range courses {
		m[courseKey(c.Code, c.MainTeacher.Code)] = c
	}
	logx.Info(ctx, "resolved course ids", "count", len(m))
	return m
}

func (imp *Importer) syncCourseLanguagesFromLastOfferings(ctx context.Context) error {
	result := imp.db.Exec(`
		UPDATE courses AS c
		SET language = oc.language
		FROM offered_courses AS oc
		WHERE oc.course_id = c.id
		  AND oc.semester = c.last_semester
		  AND c.language IS DISTINCT FROM oc.language
	`)
	if result.Error != nil {
		return result.Error
	}
	logx.Info(ctx, "course languages synced", "updated", result.RowsAffected)
	return nil
}

type courseAgg struct {
	mainTeacherCode string
	yearSet         map[string]bool
	catSet          map[string]bool
	tidSet          map[int64]bool
	language        string
}

type courseOfferingUpdate struct {
	CourseID      int            `json:"id"`
	MainTeacherID int            `json:"main_teacher_id"`
	TargetYears   pq.StringArray `json:"target_years"`
	Categories    pq.StringArray `json:"categories"`
	TeacherIDs    pq.Int64Array  `json:"teacher_ids"`
}

func (imp *Importer) updateCoursesFromOfferingsBatch(ctx context.Context, updates []courseOfferingUpdate) {
	if len(updates) == 0 {
		return
	}

	payload, err := json.Marshal(updates)
	if err != nil {
		logx.Error(ctx, "failed to marshal course offering updates", "error", err)
		return
	}

	result := imp.db.WithContext(ctx).Exec(`
		WITH updates AS (
			SELECT id, main_teacher_id, target_years, categories, teacher_ids
			FROM jsonb_to_recordset(?::jsonb) AS v(
				id integer,
				main_teacher_id integer,
				target_years text[],
				categories text[],
				teacher_ids integer[]
			)
		)
		UPDATE courses AS c
		SET main_teacher_id = v.main_teacher_id,
		    target_years = v.target_years,
		    categories = v.categories,
		    teacher_ids = v.teacher_ids
		FROM updates AS v
		WHERE c.id = v.id
		  AND c.last_semester = ?
	`, string(payload), imp.semester)
	if result.Error != nil {
		logx.Error(ctx, "failed to batch update courses from offerings", "error", result.Error)
	}
}

func (imp *Importer) upsertOfferedCourses(ctx context.Context, rows []CSVRow, teacherIDMap map[string]int, courseMap map[string]repository.CourseEntity) {
	onConflict := clause.OnConflict{
		Columns: []clause.Column{{Name: "course_id"}, {Name: "semester"}},
		DoUpdates: clause.Assignments(map[string]any{
			"language":     clause.Column{Table: "excluded", Name: "language"},
			"target_years": clause.Column{Table: "excluded", Name: "target_years"},
			"categories":   clause.Column{Table: "excluded", Name: "categories"},
			"teacher_ids":  clause.Column{Table: "excluded", Name: "teacher_ids"},
			"created_at":   clause.Column{Table: "excluded", Name: "created_at"},
		}),
	}

	aggMap := make(map[string]*courseAgg)
	for _, r := range rows {
		if r.CourseCode == "" || r.MainTeacher.Code == "" {
			continue
		}
		key := courseKey(r.CourseCode, r.MainTeacher.Code)
		agg, ok := aggMap[key]
		if !ok {
			agg = &courseAgg{
				mainTeacherCode: r.MainTeacher.Code,
				yearSet:         make(map[string]bool),
				catSet:          make(map[string]bool),
				tidSet:          make(map[int64]bool),
			}
			aggMap[key] = agg
		}

		for _, t := range r.AllTeachers {
			if id, ok := teacherIDMap[t.Code]; ok {
				agg.tidSet[int64(id)] = true
			}
		}
		for _, y := range r.TargetYears {
			agg.yearSet[y] = true
		}
		for _, c := range r.Categories {
			agg.catSet[c] = true
		}
		if agg.language == "" && r.Language != "" {
			agg.language = r.Language
		}
	}

	var batch []repository.OfferedCourseEntity
	var courseUpdateBatch []courseOfferingUpdate
	processed := 0
	skipped := 0
	for key, agg := range aggMap {
		processed++
		course, ok := courseMap[key]
		if !ok {
			skipped++
			if processed%500 == 0 {
				logx.Info(ctx, "processed offered courses", "processed", processed, "total", len(aggMap))
			}
			continue
		}

		var targetYears pq.StringArray
		for y := range agg.yearSet {
			targetYears = append(targetYears, y)
		}
		var courseCats pq.StringArray
		for c := range agg.catSet {
			courseCats = append(courseCats, c)
		}
		var allTIDs pq.Int64Array
		for tid := range agg.tidSet {
			allTIDs = append(allTIDs, tid)
		}
		mainTeacherID := teacherIDMap[agg.mainTeacherCode]

		if course.LastSemester == imp.semester {
			courseUpdateBatch = append(courseUpdateBatch, courseOfferingUpdate{
				CourseID:      course.ID,
				MainTeacherID: mainTeacherID,
				TargetYears:   targetYears,
				Categories:    courseCats,
				TeacherIDs:    allTIDs,
			})
			if len(courseUpdateBatch) >= batchSize {
				imp.updateCoursesFromOfferingsBatch(ctx, courseUpdateBatch)
				courseUpdateBatch = courseUpdateBatch[:0]
			}
		}

		batch = append(batch, repository.OfferedCourseEntity{
			CourseID:    course.ID,
			Semester:    imp.semester,
			Language:    agg.language,
			TargetYears: targetYears,
			Categories:  courseCats,
			TeacherIDs:  allTIDs,
			CreatedAt:   time.Now(),
		})
		if len(batch) >= batchSize {
			imp.db.Clauses(onConflict).Create(&batch)
			batch = batch[:0]
		}
		if processed%500 == 0 {
			logx.Info(ctx, "processed offered courses", "processed", processed, "total", len(aggMap))
		}
	}
	if len(batch) > 0 {
		imp.db.Clauses(onConflict).Create(&batch)
	}
	if len(courseUpdateBatch) > 0 {
		imp.updateCoursesFromOfferingsBatch(ctx, courseUpdateBatch)
	}
	logx.Info(ctx, "offered courses imported", "imported", processed-skipped, "skipped", skipped)
}
