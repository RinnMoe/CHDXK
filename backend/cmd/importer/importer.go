package main

import (
	"log"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	teacherdomain "jcourse/internal/domain/teacher"
	"jcourse/internal/infrastructure/repository"
)

const batchSize = 100

type Importer struct {
	db       *gorm.DB
	semester string
}

func NewImporter(db *gorm.DB, semester string) *Importer {
	return &Importer{db: db, semester: semester}
}

func (imp *Importer) Run(rows []CSVRow) error {
	log.Println("Collecting unique entities from CSV data...")
	teachers, courses := collectUnique(rows)

	log.Println("Upserting teachers...")
	imp.upsertTeachers(teachers)

	log.Println("Resolving teacher IDs...")
	teacherIDMap := imp.resolveTeacherIDs()

	log.Println("Upserting courses...")
	imp.upsertCourses(courses, teacherIDMap)

	log.Println("Resolving course IDs...")
	courseIDMap := imp.resolveCourseIDs()

	log.Println("Creating offered courses...")
	imp.upsertOfferedCourses(rows, teacherIDMap, courseIDMap)

	log.Println("Refreshing course search vectors...")
	if err := repository.RefreshCourseSearchVectors(imp.db); err != nil {
		return err
	}

	log.Println("Import complete!")
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

func (imp *Importer) upsertTeachers(teachers map[string]TeacherInfo) {
	config := repository.SearchConfig(imp.db)
	onConflict := clause.OnConflict{
		Columns: []clause.Column{{Name: "code"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"name":          clause.Column{Table: "excluded", Name: "name"},
			"department":    clause.Column{Table: "excluded", Name: "department"},
			"title":         clause.Column{Table: "excluded", Name: "title"},
			"search_vector": clause.Column{Table: "excluded", Name: "search_vector"},
			"last_semester": clause.Column{Table: "excluded", Name: "last_semester"},
			"updated_at":    clause.Column{Table: "excluded", Name: "updated_at"},
		}),
		Where: clause.Where{
			Exprs: []clause.Expression{
				clause.Expr{SQL: "teachers.last_semester < ?", Vars: []interface{}{imp.semester}},
			},
		},
	}

	var batch []map[string]interface{}
	count := 0
	for code, info := range teachers {
		now := time.Now()
		searchName := teacherdomain.NewSearchName(info.Name)
		batch = append(batch, map[string]interface{}{
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
			log.Printf("  Processed %d/%d teachers...", count, len(teachers))
		}
	}
	if len(batch) > 0 {
		imp.db.Model(&repository.TeacherEntity{}).Clauses(onConflict).Create(&batch)
	}
	log.Printf("  Teachers: %d processed", count)
}

func (imp *Importer) resolveTeacherIDs() map[string]int {
	var teachers []repository.TeacherEntity
	imp.db.Select("id, code").Find(&teachers)
	m := make(map[string]int, len(teachers))
	for _, t := range teachers {
		m[t.Code] = t.ID
	}
	log.Printf("  Resolved %d teacher IDs", len(m))
	return m
}

func (imp *Importer) upsertCourses(courses map[string]CSVRow, teacherIDMap map[string]int) {
	onConflict := clause.OnConflict{
		Columns: []clause.Column{{Name: "code"}, {Name: "main_teacher_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"department":    clause.Column{Table: "excluded", Name: "department"},
			"language":      clause.Column{Table: "excluded", Name: "language"},
			"last_semester": clause.Column{Table: "excluded", Name: "last_semester"},
		}),
		Where: clause.Where{
			Exprs: []clause.Expression{
				clause.Expr{SQL: "courses.last_semester < ?", Vars: []interface{}{imp.semester}},
			},
		},
	}

	var batch []repository.CourseEntity
	skipped := 0
	for _, row := range courses {
		mainTeacherID := teacherIDMap[row.MainTeacher.Code]
		if mainTeacherID == 0 {
			log.Printf("  Skipping course %s (%s): no main teacher found (code=%q)", row.CourseCode, row.CourseName, row.MainTeacher.Code)
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
	log.Printf("  Courses: %d imported, %d skipped", len(courses)-skipped, skipped)
}

func (imp *Importer) resolveCourseIDs() map[string]int {
	var courses []repository.CourseEntity
	imp.db.Model(&repository.CourseEntity{}).
		Joins("MainTeacher").
		Find(&courses)
	m := make(map[string]int, len(courses))
	for _, c := range courses {
		m[courseKey(c.Code, c.MainTeacher.Code)] = c.ID
	}
	log.Printf("  Resolved %d course IDs", len(m))
	return m
}

type courseAgg struct {
	mainTeacherCode string
	yearSet         map[string]bool
	catSet          map[string]bool
	tidSet          map[int64]bool
	language        string
}

func (imp *Importer) upsertOfferedCourses(rows []CSVRow, teacherIDMap map[string]int, courseIDMap map[string]int) {
	onConflict := clause.OnConflict{
		Columns: []clause.Column{{Name: "course_id"}, {Name: "semester"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
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

		var tids []int64
		for _, t := range r.AllTeachers {
			if id, ok := teacherIDMap[t.Code]; ok {
				tid := int64(id)
				agg.tidSet[tid] = true
				tids = append(tids, tid)
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

	for key, agg := range aggMap {
		courseID, ok := courseIDMap[key]
		if !ok {
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

		imp.db.Model(&repository.CourseEntity{}).Where("id = ? AND last_semester = ?", courseID, imp.semester).Updates(map[string]interface{}{
			"main_teacher_id": mainTeacherID,
			"target_years":    targetYears,
			"categories":      courseCats,
			"teacher_ids":     allTIDs,
		})

		entity := repository.OfferedCourseEntity{
			CourseID:    courseID,
			Semester:    imp.semester,
			Language:    agg.language,
			TargetYears: targetYears,
			Categories:  courseCats,
			TeacherIDs:  allTIDs,
			CreatedAt:   time.Now(),
		}
		imp.db.Clauses(onConflict).Create(&entity)
	}
}
