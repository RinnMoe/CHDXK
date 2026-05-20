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
	ID          int      `gorm:"column:id"`
	CourseID    int      `gorm:"column:course_id;index"`
	Semester    string   `gorm:"column:semester"`
	Language    string   `gorm:"column:language"`
	TargetYears []string `gorm:"column:target_years;type:text[];serializer:json"`
	Categories  []string `gorm:"column:categories;type:text[];serializer:json"`
}

func (OfferedCourseEntity) TableName() string {
	return "offered_courses"
}

type CourseTeacherGroupEntity struct {
	OfferedCourseID int `gorm:"column:offered_course_id;index:idx_offered_teacher"`
	TeacherID       int `gorm:"column:teacher_id;index:idx_offered_teacher"`
}

func (CourseTeacherGroupEntity) TableName() string {
	return "course_teacher_groups"
}

type CourseEntity struct {
	ID            int       `gorm:"column:id"`
	Code          string    `gorm:"column:code;uniqueIndex"`
	Name          string    `gorm:"column:name"`
	Credit        float32   `gorm:"column:credit"`
	Department    string    `gorm:"column:department;index"`
	MainTeacherID int       `gorm:"column:main_teacher_id;index"`
	TargetYears   []string  `gorm:"column:target_years;type:text[];serializer:json"`
	Language      string    `gorm:"column:language"`
	Categories    []string  `gorm:"column:categories;type:text[];serializer:json"`
	ReviewCount   int       `gorm:"column:review_count"`
	AvgRating     float64   `gorm:"column:avg_rating"`
	CreatedAt     time.Time `gorm:"column:created_at"`
}

func (CourseEntity) TableName() string {
	return "courses"
}

type ReviewEntity struct {
	ID        int       `gorm:"column:id"`
	CourseID  int       `gorm:"column:course_id;index"`
	Semester  string    `gorm:"column:semester"`
	UserID    int       `gorm:"column:user_id;index"`
	Rating    int       `gorm:"column:rating"`
	Content   string    `gorm:"column:content"`
	Score     string    `gorm:"column:score"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (ReviewEntity) TableName() string {
	return "reviews"
}

type ReviewRevisionEntity struct {
	ReviewID  int       `gorm:"column:review_id;index"`
	CourseID  int       `gorm:"column:course_id;index"`
	Semester  string    `gorm:"column:semester"`
	UserID    int       `gorm:"column:user_id"`
	Rating    int       `gorm:"column:rating"`
	Content   string    `gorm:"column:content"`
	Score     string    `gorm:"column:score"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (ReviewRevisionEntity) TableName() string {
	return "review_revisions"
}

type UserEntity struct {
	ID       int    `gorm:"column:id"`
	Username string `gorm:"column:username;uniqueIndex"`
	Email    string `gorm:"column:email;uniqueIndex"`
	Role     string `gorm:"column:role"`
	Password string `gorm:"column:password"`

	CreatedAt  time.Time `gorm:"column:created_at"`
	LastSeenAt time.Time `gorm:"column:last_seen_at"`

	SuspendedAt *time.Time `gorm:"column:suspended_at"`
	SuspendTill *time.Time `gorm:"column:suspend_till"`
}

func (UserEntity) TableName() string {
	return "users"
}
