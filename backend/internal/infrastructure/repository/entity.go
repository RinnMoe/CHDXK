package repository

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
	"gorm.io/datatypes"
)

type TeacherEntity struct {
	ID         int    `gorm:"column:id"`
	Code       string `gorm:"column:code;uniqueIndex"`
	Name       string `gorm:"column:name"`
	Department string `gorm:"column:department;index"`
	Title      string `gorm:"column:title"`

	SearchVector string `gorm:"column:search_vector;type:tsvector;index:,type:gin"`

	LastSemester string `gorm:"column:last_semester"`

	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (TeacherEntity) TableName() string {
	return "teachers"
}

type OfferedCourseEntity struct {
	ID          int            `gorm:"column:id"`
	CourseID    int            `gorm:"column:course_id;index;uniqueIndex:uniq_offered_courses_course_semester"`
	Semester    string         `gorm:"column:semester;uniqueIndex:uniq_offered_courses_course_semester"`
	Language    string         `gorm:"column:language"`
	TargetYears pq.StringArray `gorm:"column:target_years;type:text[]"`
	Categories  pq.StringArray `gorm:"column:categories;type:text[]"`
	TeacherIDs  pq.Int64Array  `gorm:"column:teacher_ids;type:integer[]"`
	CreatedAt   time.Time      `gorm:"column:created_at"`
}

func (OfferedCourseEntity) TableName() string {
	return "offered_courses"
}

type CourseSectionEntity struct {
	ID              int           `gorm:"column:id"`
	OfferedCourseID int           `gorm:"column:offered_course_id;index"`
	SourceLessonID  string        `gorm:"column:source_lesson_id"`
	LessonNo        string        `gorm:"column:lesson_no"`
	CourseName      string        `gorm:"column:course_name"`
	CourseType      string        `gorm:"column:course_type"`
	Credit          float32       `gorm:"column:credit"`
	Department      string        `gorm:"column:department"`
	Language        string        `gorm:"column:language"`
	Campus          string        `gorm:"column:campus"`
	Assessment      string        `gorm:"column:assessment"`
	TeachClass      string        `gorm:"column:teach_class"`
	ActualCount     *int          `gorm:"column:actual_count"`
	LimitCount      *int          `gorm:"column:limit_count"`
	CoursePeriod    string        `gorm:"column:course_period"`
	TotalHours      *float32      `gorm:"column:total_hours"`
	WeeklyHours     *float32      `gorm:"column:weekly_hours"`
	WeekRange       string        `gorm:"column:week_range"`
	StartWeek       *int          `gorm:"column:start_week"`
	EndWeek         *int          `gorm:"column:end_week"`
	WeekCount       *int          `gorm:"column:week_count"`
	Remark          string        `gorm:"column:remark"`
	SyllabusURL     string        `gorm:"column:syllabus_url"`
	TeacherIDs      pq.Int64Array `gorm:"column:teacher_ids;type:integer[]"`
	CreatedAt       time.Time     `gorm:"column:created_at"`
	UpdatedAt       time.Time     `gorm:"column:updated_at"`
}

func (CourseSectionEntity) TableName() string {
	return "course_sections"
}

type CourseEnrollmentEntity struct {
	ID        int           `gorm:"column:id"`
	UserID    int           `gorm:"column:user_id;index;uniqueIndex:uniq_course_enrollments_user_course_semester"`
	CourseID  int           `gorm:"column:course_id;index;uniqueIndex:uniq_course_enrollments_user_course_semester"`
	Semester  string        `gorm:"column:semester;uniqueIndex:uniq_course_enrollments_user_course_semester"`
	CreatedAt time.Time     `gorm:"column:created_at"`
	Course    *CourseEntity `gorm:"foreignKey:course_id;references:id"`
}

func (CourseEnrollmentEntity) TableName() string {
	return "course_enrollments"
}

type CourseEntity struct {
	ID              int            `gorm:"column:id"`
	Code            string         `gorm:"column:code;uniqueIndex:idx_courses_code_teacher"`
	Name            string         `gorm:"column:name"`
	Credit          float32        `gorm:"column:credit"`
	Department      string         `gorm:"column:department;index"`
	MainTeacherID   int            `gorm:"column:main_teacher_id;uniqueIndex:idx_courses_code_teacher"`
	TargetYears     pq.StringArray `gorm:"column:target_years;type:text[]"`
	Language        string         `gorm:"column:language"`
	Categories      pq.StringArray `gorm:"column:categories;type:text[]"`
	TeacherIDs      pq.Int64Array  `gorm:"column:teacher_ids;type:integer[]"`
	SearchVector    string         `gorm:"column:search_vector;type:tsvector;index:,type:gin;<-:false"`
	LastSemester    string         `gorm:"column:last_semester"`
	ModeratorRemark string         `gorm:"column:moderator_remark"`
	RatingCount     int            `gorm:"column:rating_count"`
	RatingAvg       float64        `gorm:"column:rating_avg"`
	RatingScore     float64        `gorm:"column:rating_score"`
	CreatedAt       time.Time      `gorm:"column:created_at"`
	MainTeacher     *TeacherEntity `gorm:"foreignKey:main_teacher_id;references:id"`
}

func (CourseEntity) TableName() string {
	return "courses"
}

type ReviewEntity struct {
	ID              int           `gorm:"column:id"`
	CourseID        int           `gorm:"column:course_id;index;uniqueIndex:uniq_reviews_user_course"`
	Semester        string        `gorm:"column:semester"`
	UserID          int           `gorm:"column:user_id;index;uniqueIndex:uniq_reviews_user_course"`
	Rating          int           `gorm:"column:rating"`
	Content         string        `gorm:"column:content"`
	Score           string        `gorm:"column:score"`
	ModeratorRemark string        `gorm:"column:moderator_remark"`
	LikeCount       int           `gorm:"column:like_count"`
	DislikeCount    int           `gorm:"column:dislike_count"`
	SearchVector    string        `gorm:"column:search_vector;type:tsvector;index:,type:gin;<-:false"`
	CreatedAt       time.Time     `gorm:"column:created_at"`
	UpdatedAt       time.Time     `gorm:"column:updated_at"`
	Course          *CourseEntity `gorm:"foreignKey:course_id;references:id"`
}

func (ReviewEntity) TableName() string {
	return "reviews"
}

type ReviewRevisionEntity struct {
	ID        int       `gorm:"column:id"`
	ReviewID  int       `gorm:"column:review_id;index"`
	CourseID  int       `gorm:"column:course_id;index"`
	Semester  string    `gorm:"column:semester"`
	CreatedBy int       `gorm:"column:created_by"`
	Rating    int       `gorm:"column:rating"`
	Content   string    `gorm:"column:content"`
	Score     string    `gorm:"column:score"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (ReviewRevisionEntity) TableName() string {
	return "review_revisions"
}

type UserEntity struct {
	ID           int            `gorm:"column:id"`
	Username     string         `gorm:"column:username;uniqueIndex"`
	Email        sql.NullString `gorm:"column:email;uniqueIndex"`
	Role         string         `gorm:"column:role"`
	PasswordHash string         `gorm:"column:password_hash"`

	CreatedAt  time.Time `gorm:"column:created_at"`
	LastSeenAt time.Time `gorm:"column:last_seen_at"`

	SuspendedAt *time.Time `gorm:"column:suspended_at"`
	SuspendTill *time.Time `gorm:"column:suspend_till"`
}

func (UserEntity) TableName() string {
	return "users"
}

// UserIdentityEntity stores the res-hub subject separately from the local
// projection used by course and review foreign keys.
type UserIdentityEntity struct {
	UserID       int       `gorm:"column:user_id;primaryKey"`
	Issuer       string    `gorm:"column:issuer;uniqueIndex:uniq_user_identities_issuer_subject"`
	Subject      string    `gorm:"column:subject;uniqueIndex:uniq_user_identities_issuer_subject"`
	Username     string    `gorm:"column:username"`
	Email        string    `gorm:"column:email"`
	IdentityRole string    `gorm:"column:identity_role"`
	IdentityType string    `gorm:"column:identity_type"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (UserIdentityEntity) TableName() string {
	return "user_identities"
}

type SystemSettingsEntity struct {
	Key       string    `gorm:"column:key;primaryKey"`
	Value     string    `gorm:"column:value"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (SystemSettingsEntity) TableName() string {
	return "system_settings"
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

type PointRewardEntity struct {
	ID                  int        `gorm:"column:id"`
	UserID              int        `gorm:"column:user_id;index"`
	Reason              string     `gorm:"column:reason;uniqueIndex:uniq_point_rewards_reason_source"`
	Amount              int        `gorm:"column:amount"`
	SourceType          string     `gorm:"column:source_type;uniqueIndex:uniq_point_rewards_reason_source"`
	SourceKey           string     `gorm:"column:source_key;uniqueIndex:uniq_point_rewards_reason_source"`
	Description         string     `gorm:"column:description"`
	Status              string     `gorm:"column:status;index"`
	CreatedAt           time.Time  `gorm:"column:created_at"`
	GrantedAt           *time.Time `gorm:"column:granted_at"`
	OperationKey        string     `gorm:"column:operation_key"`
	ExternalOperationID string     `gorm:"column:external_operation_id"`
	ExternalReversalID  string     `gorm:"column:external_reversal_id"`
	Attempts            int        `gorm:"column:attempts"`
	LastError           string     `gorm:"column:last_error"`
	NextRetryAt         *time.Time `gorm:"column:next_retry_at;index"`
}

func (PointRewardEntity) TableName() string {
	return "point_rewards"
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

type CourseHotScoreEntity struct {
	Period    string    `gorm:"column:period;primaryKey"`
	PeriodKey string    `gorm:"column:period_key;primaryKey"`
	CourseID  int       `gorm:"column:course_id;primaryKey"`
	Score     int64     `gorm:"column:score"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (CourseHotScoreEntity) TableName() string {
	return "course_hot_scores"
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

type ApiKeyEntity struct {
	ID         int64      `gorm:"column:id;primaryKey;autoIncrement:false"`
	Name       string     `gorm:"column:name"`
	SecretHash string     `gorm:"column:secret_hash"`
	Role       string     `gorm:"column:role"`
	LastUsedAt *time.Time `gorm:"column:last_used_at"`
	CreatedAt  time.Time  `gorm:"column:created_at"`
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
	LinkURL   string    `gorm:"column:link_url"`
	LinkTitle string    `gorm:"column:link_title"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (AnnouncementEntity) TableName() string {
	return "announcements"
}

type AuditLogEntity struct {
	ID          int64             `gorm:"column:id"`
	OccurredAt  time.Time         `gorm:"column:occurred_at"`
	ActorUserID int               `gorm:"column:actor_user_id;index"`
	Action      string            `gorm:"column:action;index"`
	TargetType  string            `gorm:"column:target_type"`
	TargetID    string            `gorm:"column:target_id"`
	Details     datatypes.JSONMap `gorm:"column:details;type:jsonb"`
	CreatedAt   time.Time         `gorm:"column:created_at"`
}

func (AuditLogEntity) TableName() string {
	return "audit_logs"
}
