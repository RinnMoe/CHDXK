package repository

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/datatypes"
)

type DepartmentEntity struct {
	ID        int       `gorm:"column:id"`
	Name      string    `gorm:"column:name"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (DepartmentEntity) TableName() string {
	return "departments"
}

type SemesterEntity struct {
	ID        int       `gorm:"column:id"`
	Name      string    `gorm:"column:name"`
	CanReview bool      `gorm:"column:can_review"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (SemesterEntity) TableName() string {
	return "semesters"
}

type TeacherEntity struct {
	ID         int    `gorm:"column:id"`
	Code       string `gorm:"column:code;uniqueIndex"`
	Name       string `gorm:"column:name"`
	Department string `gorm:"column:department;index"`
	Title      string `gorm:"column:title"`

	Pinyin       string `gorm:"column:pinyin;index"`
	PinyinAbbr   string `gorm:"column:pinyin_abbr;index"`
	SearchVector string `gorm:"column:search_vector;type:tsvector;index:,type:gin;<-:false"`

	LastSemester string `gorm:"column:last_semester"`

	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (TeacherEntity) TableName() string {
	return "teachers"
}

type OfferedCourseEntity struct {
	ID          int            `gorm:"column:id"`
	CourseID    int            `gorm:"column:course_id;index"`
	Semester    string         `gorm:"column:semester"`
	Language    string         `gorm:"column:language"`
	TargetYears pq.StringArray `gorm:"column:target_years;type:text[]"`
	Categories  pq.StringArray `gorm:"column:categories;type:text[]"`
	TeacherIDs  pq.Int64Array  `gorm:"column:teacher_ids;type:integer[]"`
	CreatedAt   time.Time      `gorm:"column:created_at"`
}

func (OfferedCourseEntity) TableName() string {
	return "offered_courses"
}

type CourseEntity struct {
	ID            int            `gorm:"column:id"`
	Code          string         `gorm:"column:code;uniqueIndex:idx_courses_code_teacher"`
	Name          string         `gorm:"column:name"`
	Credit        float32        `gorm:"column:credit"`
	Department    string         `gorm:"column:department;index"`
	MainTeacherID int            `gorm:"column:main_teacher_id;uniqueIndex:idx_courses_code_teacher"`
	TargetYears   pq.StringArray `gorm:"column:target_years;type:text[]"`
	Language      string         `gorm:"column:language"`
	Categories    pq.StringArray `gorm:"column:categories;type:text[]"`
	TeacherIDs    pq.Int64Array  `gorm:"column:teacher_ids;type:integer[]"`
	SearchVector  string         `gorm:"column:search_vector;type:tsvector;index:,type:gin;<-:false"`
	LastSemester  string         `gorm:"column:last_semester"`
	RatingCount   int            `gorm:"column:rating_count"`
	RatingAvg     float64        `gorm:"column:rating_avg"`
	CreatedAt     time.Time      `gorm:"column:created_at"`
	MainTeacher   *TeacherEntity `gorm:"foreignKey:MainTeacherID;references:ID"`
}

func (CourseEntity) TableName() string {
	return "courses"
}

type ReviewEntity struct {
	ID           int           `gorm:"column:id"`
	CourseID     int           `gorm:"column:course_id;index"`
	Semester     string        `gorm:"column:semester"`
	UserID       int           `gorm:"column:user_id;index"`
	Rating       int           `gorm:"column:rating"`
	Content      string        `gorm:"column:content"`
	Score        string        `gorm:"column:score"`
	LikeCount    int           `gorm:"column:like_count"`
	DislikeCount int           `gorm:"column:dislike_count"`
	CreatedAt    time.Time     `gorm:"column:created_at"`
	UpdatedAt    time.Time     `gorm:"column:updated_at"`
	Course       *CourseEntity `gorm:"foreignKey:CourseID;references:ID"`
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

type UserPointRecordEntity struct {
	ID          int       `gorm:"column:id"`
	UserID      int       `gorm:"column:user_id;index"`
	Reason      string    `gorm:"column:reason"`
	Amount      int       `gorm:"column:amount"`
	Description string    `gorm:"column:description"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

func (UserPointRecordEntity) TableName() string {
	return "user_point_records"
}

type PointTransferEntity struct {
	ID              int       `gorm:"column:id"`
	SenderUserID    int       `gorm:"column:sender_user_id;index"`
	RecipientUserID int       `gorm:"column:recipient_user_id;index"`
	Amount          int       `gorm:"column:amount"`
	Fee             int       `gorm:"column:fee"`
	FeePayer        string    `gorm:"column:fee_payer"`
	SenderDelta     int       `gorm:"column:sender_delta"`
	RecipientDelta  int       `gorm:"column:recipient_delta"`
	CreatedAt       time.Time `gorm:"column:created_at"`
}

func (PointTransferEntity) TableName() string {
	return "point_transfers"
}

type ReviewVoteEntity struct {
	ReviewID  int       `gorm:"column:review_id;primaryKey"`
	UserID    int       `gorm:"column:user_id;primaryKey"`
	VoteType  int       `gorm:"column:vote_type"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (ReviewVoteEntity) TableName() string {
	return "review_votes"
}

type CourseNotificationEntity struct {
	UserID    int       `gorm:"column:user_id;primaryKey"`
	CourseID  int       `gorm:"column:course_id;primaryKey"`
	Level     int       `gorm:"column:level"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (CourseNotificationEntity) TableName() string {
	return "course_notifications"
}

type SiteDailyStatEntity struct {
	StatDate    time.Time         `gorm:"column:stat_date;primaryKey;type:date"`
	Metrics     datatypes.JSONMap `gorm:"column:metrics;type:jsonb"`
	GeneratedAt time.Time         `gorm:"column:generated_at"`
	CreatedAt   time.Time         `gorm:"column:created_at"`
	UpdatedAt   time.Time         `gorm:"column:updated_at"`
}

func (SiteDailyStatEntity) TableName() string {
	return "site_daily_stats"
}

type CategoryEntity struct {
	ID        int       `gorm:"column:id"`
	Name      string    `gorm:"column:name;uniqueIndex"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (CategoryEntity) TableName() string {
	return "categories"
}

type ApiKeyEntity struct {
	ID        int       `gorm:"column:id"`
	Name      string    `gorm:"column:name"`
	Key       string    `gorm:"column:key;uniqueIndex"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (ApiKeyEntity) TableName() string {
	return "api_keys"
}

type AnnouncementEntity struct {
	ID        int       `gorm:"column:id"`
	Title     string    `gorm:"column:title"`
	Body      string    `gorm:"column:body"`
	Priority  int       `gorm:"column:priority"`
	ShowStart time.Time `gorm:"column:show_start"`
	ShowEnd   time.Time `gorm:"column:show_end"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (AnnouncementEntity) TableName() string {
	return "announcements"
}
