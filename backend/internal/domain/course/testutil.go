//go:build test

package course

import "context"

type MockCourseRepository struct {
	Courses          map[int]*CourseView
	OfferedCourses   map[int]map[string]bool
	OfferedSemesters map[string]bool
	Filters          *CourseFilters
	RatingConfig     RatingScoreConfig

	OnGet                   func(context.Context, int) (*CourseView, error)
	OnFindBy                func(context.Context, CourseFilter) ([]CourseView, int64, error)
	OnGetDetail             func(context.Context, int) (*CourseDetailView, error)
	OnFindOfferedCourses    func(context.Context, int) ([]OfferedCourseView, error)
	OnGetFilters            func(context.Context) (*CourseFilters, error)
	OnRefreshRatingScores   func(context.Context, RatingScoreConfig) error
	OnOfferedCourseExists   func(context.Context, int, string) (bool, error)
	OnOfferedSemesterExists func(context.Context, string) (bool, error)
}

func NewMockCourseRepository() *MockCourseRepository {
	return &MockCourseRepository{
		Courses:          map[int]*CourseView{},
		OfferedCourses:   map[int]map[string]bool{},
		OfferedSemesters: map[string]bool{},
	}
}

func (r *MockCourseRepository) Get(ctx context.Context, courseID int) (*CourseView, error) {
	if r.OnGet != nil {
		return r.OnGet(ctx, courseID)
	}
	r.ensureMaps()
	c, ok := r.Courses[courseID]
	if !ok {
		return nil, nil
	}
	copy := *c
	return &copy, nil
}

func (r *MockCourseRepository) FindBy(ctx context.Context, filter CourseFilter) ([]CourseView, int64, error) {
	if r.OnFindBy != nil {
		return r.OnFindBy(ctx, filter)
	}
	r.ensureMaps()
	views := make([]CourseView, 0, len(r.Courses))
	for id, c := range r.Courses {
		if len(filter.CourseIDs) > 0 && !containsInt(filter.CourseIDs, id) {
			continue
		}
		if filter.TeacherID > 0 && c.MainTeacherID != filter.TeacherID {
			continue
		}
		if filter.ExcludeID > 0 && id == filter.ExcludeID {
			continue
		}
		if filter.Code != "" && c.Code != filter.Code {
			continue
		}
		if filter.Name != "" && c.Name != filter.Name {
			continue
		}
		if filter.Department != "" && c.Department != filter.Department {
			continue
		}
		if filter.Credit != nil && c.Credit != *filter.Credit {
			continue
		}
		if filter.Language != "" && c.Language != filter.Language {
			continue
		}
		copy := *c
		views = append(views, copy)
	}
	return views, int64(len(views)), nil
}

func (r *MockCourseRepository) GetDetail(ctx context.Context, courseID int) (*CourseDetailView, error) {
	if r.OnGetDetail != nil {
		return r.OnGetDetail(ctx, courseID)
	}
	c, err := r.Get(ctx, courseID)
	if err != nil || c == nil {
		return nil, err
	}
	return &CourseDetailView{
		ID:            c.ID,
		Code:          c.Code,
		Name:          c.Name,
		Credit:        c.Credit,
		Department:    c.Department,
		MainTeacherID: c.MainTeacherID,
		MainTeacher:   c.MainTeacher,
		LastSemester:  c.LastSemester,
		Categories:    c.Categories,
		Language:      c.Language,
		TargetYears:   c.TargetYears,
		Rating:        c.Rating,
	}, nil
}

func (r *MockCourseRepository) FindOfferedCourses(ctx context.Context, courseID int) ([]OfferedCourseView, error) {
	if r.OnFindOfferedCourses != nil {
		return r.OnFindOfferedCourses(ctx, courseID)
	}
	return []OfferedCourseView{}, nil
}

func (r *MockCourseRepository) GetFilters(ctx context.Context) (*CourseFilters, error) {
	if r.OnGetFilters != nil {
		return r.OnGetFilters(ctx)
	}
	return r.Filters, nil
}

func (r *MockCourseRepository) RefreshRatingScores(ctx context.Context, config RatingScoreConfig) error {
	if r.OnRefreshRatingScores != nil {
		return r.OnRefreshRatingScores(ctx, config)
	}
	r.RatingConfig = config
	return nil
}

func (r *MockCourseRepository) OfferedCourseExists(ctx context.Context, courseID int, semester string) (bool, error) {
	if r.OnOfferedCourseExists != nil {
		return r.OnOfferedCourseExists(ctx, courseID, semester)
	}
	r.ensureMaps()
	if semesters, ok := r.OfferedCourses[courseID]; ok {
		return semesters[semester], nil
	}
	return false, nil
}

func (r *MockCourseRepository) OfferedSemesterExists(ctx context.Context, semester string) (bool, error) {
	if r.OnOfferedSemesterExists != nil {
		return r.OnOfferedSemesterExists(ctx, semester)
	}
	r.ensureMaps()
	return r.OfferedSemesters[semester], nil
}

func (r *MockCourseRepository) ensureMaps() {
	if r.Courses == nil {
		r.Courses = map[int]*CourseView{}
	}
	if r.OfferedCourses == nil {
		r.OfferedCourses = map[int]map[string]bool{}
	}
	if r.OfferedSemesters == nil {
		r.OfferedSemesters = map[string]bool{}
	}
}

func containsInt(values []int, target int) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

type MockCourseEnrollmentRepository struct {
	Created             *CourseEnrollment
	DeletedEnrollmentID int
	DeletedUserID       int
	SyncedUserID        int
	SyncedSemester      string
	SyncedPairs         []CourseCodeTeacher
	SyncCount           int64

	OnCreate              func(context.Context, *CourseEnrollment) error
	OnSyncFromCoursePairs func(context.Context, int, string, []CourseCodeTeacher) (int64, error)
	OnDelete              func(context.Context, int, int) error
}

func (r *MockCourseEnrollmentRepository) Create(ctx context.Context, enrollment *CourseEnrollment) error {
	if r.OnCreate != nil {
		return r.OnCreate(ctx, enrollment)
	}
	copy := *enrollment
	r.Created = &copy
	return nil
}

func (r *MockCourseEnrollmentRepository) SyncFromCoursePairs(ctx context.Context, userID int, semester string, pairs []CourseCodeTeacher) (int64, error) {
	if r.OnSyncFromCoursePairs != nil {
		return r.OnSyncFromCoursePairs(ctx, userID, semester, pairs)
	}
	r.SyncedUserID = userID
	r.SyncedSemester = semester
	r.SyncedPairs = append([]CourseCodeTeacher(nil), pairs...)
	return r.SyncCount, nil
}

func (r *MockCourseEnrollmentRepository) Delete(ctx context.Context, enrollmentID, userID int) error {
	if r.OnDelete != nil {
		return r.OnDelete(ctx, enrollmentID, userID)
	}
	r.DeletedEnrollmentID = enrollmentID
	r.DeletedUserID = userID
	return nil
}

type MockCourseNotificationRepository struct {
	Levels map[CourseNotificationKey]NotificationLevel

	OnGetLevel          func(context.Context, int, int) (NotificationLevel, error)
	OnSetLevel          func(context.Context, int, int, NotificationLevel) error
	OnGetCoursesByLevel func(context.Context, int, NotificationLevel) ([]int, error)
}

type CourseNotificationKey struct {
	UserID   int
	CourseID int
}

func NewMockCourseNotificationRepository() *MockCourseNotificationRepository {
	return &MockCourseNotificationRepository{Levels: map[CourseNotificationKey]NotificationLevel{}}
}

func (r *MockCourseNotificationRepository) GetLevel(ctx context.Context, userID, courseID int) (NotificationLevel, error) {
	if r.OnGetLevel != nil {
		return r.OnGetLevel(ctx, userID, courseID)
	}
	r.ensureLevels()
	if level, ok := r.Levels[CourseNotificationKey{UserID: userID, CourseID: courseID}]; ok {
		return level, nil
	}
	return NotificationLevelNormal, nil
}

func (r *MockCourseNotificationRepository) SetLevel(ctx context.Context, userID, courseID int, level NotificationLevel) error {
	if r.OnSetLevel != nil {
		return r.OnSetLevel(ctx, userID, courseID, level)
	}
	r.ensureLevels()
	r.Levels[CourseNotificationKey{UserID: userID, CourseID: courseID}] = level
	return nil
}

func (r *MockCourseNotificationRepository) GetCoursesByLevel(ctx context.Context, userID int, level NotificationLevel) ([]int, error) {
	if r.OnGetCoursesByLevel != nil {
		return r.OnGetCoursesByLevel(ctx, userID, level)
	}
	r.ensureLevels()
	var ids []int
	for k, v := range r.Levels {
		if k.UserID == userID && v == level {
			ids = append(ids, k.CourseID)
		}
	}
	return ids, nil
}

func (r *MockCourseNotificationRepository) ensureLevels() {
	if r.Levels == nil {
		r.Levels = map[CourseNotificationKey]NotificationLevel{}
	}
}

type MockHotCourseRepository struct {
	Ranks []HotCourseRank
	Calls []HotCourseRank

	OnAddScore func(context.Context, int, int64, ...HotCoursePeriod) error
	OnTop      func(context.Context, HotCoursePeriod, int64) ([]HotCourseRank, error)
}

func (r *MockHotCourseRepository) AddScore(ctx context.Context, courseID int, score int64, periods ...HotCoursePeriod) error {
	if r.OnAddScore != nil {
		return r.OnAddScore(ctx, courseID, score, periods...)
	}
	r.Calls = append(r.Calls, HotCourseRank{CourseID: courseID, Score: score})
	return nil
}

func (r *MockHotCourseRepository) Top(ctx context.Context, period HotCoursePeriod, limit int64) ([]HotCourseRank, error) {
	if r.OnTop != nil {
		return r.OnTop(ctx, period, limit)
	}
	if limit < int64(len(r.Ranks)) {
		return r.Ranks[:limit], nil
	}
	return r.Ranks, nil
}
