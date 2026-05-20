package repository

import "time"

type DepartmentEntity struct {
	ID        int
	Name      string
	CreatedAt time.Time
}

func (DepartmentEntity) TableName() string {
	return "departments"
}

type SemesterEntity struct {
	ID        int
	Name      string
	CanReview bool
	CreatedAt time.Time
}

func (SemesterEntity) TableName() string {
	return "semesters"
}

type TeacherEntity struct {
	ID         int
	Code       string
	Name       string
	Department string
	Title      string

	Pinyin     string
	PinyinAbbr string

	LateSemester string

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (TeacherEntity) TableName() string {
	return "teachers"
}

type OfferedCourseEntity struct {
	ID         int
	CourseID   int
	Semester   string
	Language   string
	Grades     []string `gorm:"type:text[];serializer:json"`
	Categories []string `gorm:"type:text[];serializer:json"`
}

func (OfferedCourseEntity) TableName() string {
	return "offered_courses"
}

type CourseTeacherGroupEntity struct {
	OfferedCourseID int
	TeacherID       int
}

func (CourseTeacherGroupEntity) TableName() string {
	return "course_teacher_groups"
}

type CourseEntity struct {
	ID            int
	Code          string
	Name          string
	Credit        float32
	Department    string
	MainTeacherID int
	Categories    []string `gorm:"type:text[];serializer:json"`
	Language      string
	Grades        []string `gorm:"type:text[];serializer:json"`
	ReviewCount   int
	AvgRating     float64
	CreatedAt     time.Time
}

func (CourseEntity) TableName() string {
	return "courses"
}

type ReviewEntity struct {
	ID        int
	CourseID  int
	Semester  string
	UserID    int
	Rating    int
	Content   string
	Grade     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (ReviewEntity) TableName() string {
	return "reviews"
}

type ReviewRevisionEntity struct {
	ReviewID  int
	CourseID  int
	Semester  string
	UserID    int
	Rating    int
	Content   string
	Grade     string
	CreatedAt time.Time
}

func (ReviewRevisionEntity) TableName() string {
	return "review_revisions"
}

type UserEntity struct {
	ID       int
	Username string
	Email    string
	Role     string
	Password string

	CreatedAt  time.Time
	LastSeenAt time.Time

	SuspendedAt *time.Time
	SuspendTill *time.Time
}

func (UserEntity) TableName() string {
	return "users"
}
