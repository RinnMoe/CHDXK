package app

import (
	"time"

	"jcourse/config"
	"jcourse/internal/application"
	domainauth "jcourse/internal/domain/auth"
	"jcourse/internal/domain/review"
	"jcourse/internal/domain/review/policy"
	"jcourse/internal/infrastructure/email"
	"jcourse/internal/infrastructure/persistence"
	"jcourse/internal/infrastructure/repository"
)

type ServiceContainer struct {
	ReviewQuery      *application.ReviewQueryService
	ReviewCommand    *application.ReviewCommandService
	CourseQuery      *application.CourseQueryService
	CourseCommand    *application.CourseCommandService
	TeacherQuery     *application.TeacherQueryService
	PointQuery       *application.PointQueryService
	PointCommand     *application.PointCommandService
	SiteStatsQuery   *application.SiteStatsQueryService
	SiteStatsCommand *application.SiteStatsCommandService
	AuthCommand      *application.AuthCommandService
	AuthService      *domainauth.AuthService
	AnnouncementQuery *application.AnnouncementQueryService
}

func NewServiceContainer(conf config.AppConfig) *ServiceContainer {

	db := persistence.NewPostgres(conf.Postgres)
	redisClient := persistence.NewRedisClient(conf.Redis)

	reviewRepo := repository.NewReviewRepository(db)
	voteRepo := repository.NewReviewVoteRepository(db)
	courseRepo := repository.NewCourseRepository(db)
	teacherRepo := repository.NewTeacherRepository(db)
	announcementRepo := repository.NewAnnouncementRepository(db)
	notificationRepo := repository.NewCourseNotificationRepository(db)
	pointRepo := repository.NewPointRepository(db)
	userRepo := repository.NewUserRepository(db)
	statRepo := repository.NewSiteDailyStatRepository(db)
	verificationRepo := repository.NewVerificationCodeRepository(redisClient)

	reviewQuery := application.NewReviewQueryService(reviewRepo, voteRepo, notificationRepo)

	freqPolicy := policy.NewFrequencyPolicy(reviewRepo, policy.DefaultFrequencyPolicyConfig())
	safetyPolicy := policy.NewSafetyPolicy(nil)

	reviewCommand := application.NewReviewCommandService(
		courseRepo, reviewRepo, voteRepo, []review.CreatePolicy{freqPolicy, safetyPolicy},
	)
	courseQuery := application.NewCourseQueryService(courseRepo, notificationRepo)
	courseCommand := application.NewCourseCommandService(courseRepo, notificationRepo)
	teacherQuery := application.NewTeacherQueryService(teacherRepo)
	announcementQuery := application.NewAnnouncementQueryService(announcementRepo)
	pointQuery := application.NewPointQueryService(pointRepo)
	pointCommand := application.NewPointCommandService(userRepo, pointRepo, application.PointTransferFeeConfig{
		RateBps: conf.Point.TransferFeeRateBps,
		MinFee:  conf.Point.TransferMinFee,
	})
	siteStatsQuery := application.NewSiteStatsQueryService(statRepo)
	siteStatsCommand := application.NewSiteStatsCommandService(statRepo, statRepo)
	authCommand := application.NewAuthCommandService(
		userRepo,
		verificationRepo,
		email.NewSMTPVerificationCodeSender(conf.SMTP),
		domainauth.NewDjangoPBKDF2SHA256PasswordHasher(0),
		application.AuthCommandConfig{
			EmailWhitelist: conf.Auth.EmailWhitelist,
			CodeInterval:   time.Duration(conf.Auth.VerificationCodeInterval) * time.Second,
			CodeTTL:        time.Duration(conf.Auth.VerificationCodeTTL) * time.Second,
		},
	)
	authService := domainauth.NewAuthService(userRepo)

	return &ServiceContainer{
		ReviewQuery:      reviewQuery,
		ReviewCommand:    reviewCommand,
		CourseQuery:      courseQuery,
		CourseCommand:    courseCommand,
		TeacherQuery:     teacherQuery,
		PointQuery:       pointQuery,
		PointCommand:     pointCommand,
		SiteStatsQuery:   siteStatsQuery,
		SiteStatsCommand: siteStatsCommand,
		AuthCommand:      authCommand,
		AuthService:      authService,
		AnnouncementQuery: announcementQuery,
	}
}
