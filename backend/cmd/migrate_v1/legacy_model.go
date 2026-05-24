package main

import (
	"database/sql"
	"time"
)

type legacyTeacher struct {
	ID             int
	TID            sql.NullString `gorm:"column:tid"`
	Name           string
	DepartmentID   *int
	Department     *legacyDepartment `gorm:"foreignKey:DepartmentID"`
	Title          sql.NullString
	LastSemesterID *int
	LastSemester   *legacySemester `gorm:"foreignKey:LastSemesterID"`
}

func (legacyTeacher) TableName() string {
	return "jcourse_api_teacher"
}

type legacyDepartment struct {
	ID   int
	Name string
}

func (legacyDepartment) TableName() string {
	return "jcourse_api_department"
}

type legacySemester struct {
	ID   int
	Name string
}

func (legacySemester) TableName() string {
	return "jcourse_api_semester"
}

type legacyCategory struct {
	ID   int
	Name string
}

func (legacyCategory) TableName() string {
	return "jcourse_api_category"
}

type legacyCourse struct {
	ID              int
	Code            string
	Name            string
	Credit          float32
	DepartmentID    *int
	Department      *legacyDepartment `gorm:"foreignKey:DepartmentID"`
	MainTeacherID   int
	ModeratorRemark sql.NullString
	ReviewCount     sql.NullInt64
	ReviewAvg       sql.NullFloat64
	LastSemesterID  *int
	LastSemester    *legacySemester  `gorm:"foreignKey:LastSemesterID"`
	Categories      []legacyCategory `gorm:"many2many:jcourse_api_course_categories;joinForeignKey:CourseID;joinReferences:CategoryID"`
	TeacherGroup    []legacyTeacher  `gorm:"many2many:jcourse_api_course_teacher_group;joinForeignKey:CourseID;joinReferences:TeacherID"`
}

func (legacyCourse) TableName() string {
	return "jcourse_api_course"
}

type legacyUser struct {
	ID           int
	PasswordHash string
	Username     string
	IsStaff      bool
	IsSuperuser  bool
	DateJoined   time.Time
	LastLogin    sql.NullTime
	LastSeenAt   sql.NullTime
}

type legacyReview struct {
	ID              int
	UserID          int
	CourseID        int
	SemesterName    sql.NullString
	Rating          int
	Comment         string
	CreatedAt       time.Time
	ModifiedAt      sql.NullTime
	Score           sql.NullString
	ModeratorRemark sql.NullString
	ApproveCount    sql.NullInt64
	DisapproveCount sql.NullInt64
}

type legacyReviewRevision struct {
	ID           int
	ReviewID     sql.NullInt64
	CourseID     sql.NullInt64
	SemesterName sql.NullString
	UserID       sql.NullInt64
	Rating       int
	Comment      string
	CreatedAt    time.Time
	Score        sql.NullString
}

type legacyReaction struct {
	ID         int
	UserID     int
	ReviewID   int
	Reaction   int
	ModifiedAt sql.NullTime
}

type legacyUserPoint struct {
	ID          int
	UserID      int
	Value       int
	Description sql.NullString
	Time        time.Time
}
