package app

import (
	"jcourse/config"
	"jcourse/internal/application"
	"jcourse/internal/domain/account"
	"jcourse/internal/domain/account/credential"
	"jcourse/internal/domain/account/identity"
	"jcourse/internal/domain/account/verification"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/email"
	"jcourse/internal/infrastructure/jaccount"
	"jcourse/internal/infrastructure/persistence"
	"jcourse/internal/infrastructure/repository"
	"jcourse/internal/infrastructure/smtp"
)

type ServiceContainer struct {
	ReviewQuery           *application.ReviewQueryService
	ReviewCommand         *application.ReviewCommandService
	CourseQuery           *application.CourseQueryService
	CourseCommand         *application.CourseCommandService
	CourseEnrollmentQuery *application.CourseEnrollmentQueryService
	TeacherQuery          *application.TeacherQueryService
	PointQuery            *application.PointQueryService
	PointRewardCommand    *application.PointRewardCommandService
	SiteStatsQuery        *application.SiteStatsQueryService
	SiteStatsCommand      *application.SiteStatsCommandService
	AccountQuery          *application.AccountQueryService
	AccountCommand        *application.AccountCommandService
	AdminUserQuery        *application.AdminUserQueryService
	AdminUserCommand      *application.AdminUserCommandService
	AuthUserService       *auth.AuthUserService
	AuthResolution        *application.AuthResolutionService
	AccessTracker         auth.AccessTracker
	AnnouncementQuery     *application.AnnouncementQueryService
	ApiKeySvc             *auth.ApiKeyService
	ApiKeyQuery           *application.ApiKeyQueryService
	ApiKeyCommand         *application.ApiKeyCommandService
	SystemSettingsQuery   *application.SystemSettingsQueryService
	SystemSettingsCommand *application.SystemSettingsCommandService
	AuditLogQuery         *application.AuditLogQueryService
	AuditLogCommand       *application.AuditLogCommandService
	EmailSender           email.Sender
}

func NewServiceContainer(conf config.AppConfig) *ServiceContainer {

	db := persistence.NewPostgres(conf.Postgres)
	redisClient := persistence.NewRedisClient(conf.Redis)

	reviewRepo := repository.NewReviewRepositoryWithRatingScore(db, redisClient, conf.Course.RatingScore)
	voteRepo := repository.NewReviewVoteRepository(db, redisClient)
	courseRepo := repository.NewCourseRepository(db, redisClient)
	courseEnrollmentRepo := repository.NewCourseEnrollmentRepository(db, redisClient)
	teacherRepo := repository.NewTeacherRepository(db, redisClient)
	announcementRepo := repository.NewAnnouncementRepository(db)
	notificationRepo := repository.NewCourseNotificationRepository(db, redisClient)
	pointRepo := repository.NewPointRepository(db, redisClient)
	accountRepo := repository.NewAccountRepository(db, redisClient)
	userRepo := repository.NewUserRepository(db, redisClient)
	systemSettingsRepo := repository.NewSystemSettingsRepository(db, redisClient)
	systemSettingsService := newSystemSettingsService(systemSettingsRepo, courseRepo)
	siteSettings := application.NewSystemSiteSettingsProvider(systemSettingsService)
	apiKeyRepo := repository.NewApiKeyRepository(db)
	auditLogRepo := repository.NewAuditLogRepository(db)
	accessTracker := repository.NewAccessTrackerRepository(db, redisClient, conf.Auth.Access.FlushBatchSize)
	statRepo := repository.NewSiteDailyStatRepository(db, redisClient)
	courseHotRepo := repository.NewGormCourseHotRepository(db, redisClient)
	verificationRepo := repository.NewVerificationCodeRepository(redisClient)
	resetCodeRepo := repository.NewVerificationCodeRepositoryWithPrefix(redisClient, "reset")
	loginAttemptRepo := repository.NewLoginAttemptRepository(redisClient, account.DefaultLoginConfig.Lockout)
	usernameDeriver := identity.NewBLAKE2bUsernameDeriver(conf.Auth.UsernameDeriver)

	reviewQuery := application.NewReviewQueryService(reviewRepo, voteRepo, notificationRepo)

	reviewCommand := application.NewReviewCommandService(
		courseRepo,
		reviewRepo,
		voteRepo,
		siteSettings,
		conf.Review.Command,
		nil,
	)
	courseHotService := course.NewCourseHotService(courseHotRepo, course.DefaultHotScoreConfig)
	courseService := course.NewService(courseRepo)
	courseRatingCommand := course.NewCourseRatingCommandService(courseRepo, conf.Course.RatingScore)
	courseQuery := application.NewCourseQueryService(courseRepo, teacherRepo, reviewRepo, notificationRepo, courseEnrollmentRepo, courseHotRepo)
	jaccountClient := jaccount.NewOAuthClient(conf.JAccount)
	courseEnrollmentService := course.NewEnrollmentService(courseRepo, courseEnrollmentRepo, jaccountClient)
	courseCommand := application.NewCourseCommandService(
		courseService,
		course.NewNotificationService(courseRepo, notificationRepo),
		courseEnrollmentService,
		courseHotService,
		courseRatingCommand,
		siteSettings,
	)
	courseEnrollmentQuery := application.NewCourseEnrollmentQueryService(courseEnrollmentRepo)
	teacherQuery := application.NewTeacherQueryService(teacherRepo)
	announcementQuery := application.NewAnnouncementQueryService(announcementRepo)
	pointQuery := application.NewPointQueryService(pointRepo, accountRepo, usernameDeriver)
	pointRewardCommand := application.NewPointRewardCommandService(pointRepo)
	statsConfig := conf.Stats
	siteStatsQuery := application.NewSiteStatsQueryService(statRepo, statsConfig)
	siteStatsCommand := application.NewSiteStatsCommandService(statRepo, statRepo, statsConfig)
	currentUserService := auth.NewCurrentUserService(userRepo)
	accountQuery := application.NewAccountQueryService(accountRepo)
	adminUserQuery := application.NewAdminUserQueryService(accountRepo, userRepo, usernameDeriver)
	hasher := credential.NewDjangoPBKDF2SHA256PasswordHasher(conf.Auth.PasswordHash)
	adminUserCommand := application.NewAdminUserCommandService(userRepo, accountRepo, hasher, siteSettings)
	smtpSender := smtp.NewSMTPSender(conf.SMTP)
	registrationService := account.NewRegistrationService(
		accountRepo,
		verificationRepo,
		hasher,
		usernameDeriver,
		account.DefaultRegistrationConfig,
		verification.DefaultConfig,
	)
	loginService := account.NewLoginService(
		accountRepo,
		hasher,
		loginAttemptRepo,
		usernameDeriver,
		account.DefaultLoginConfig,
	)
	passwordResetService := account.NewPasswordResetService(
		accountRepo,
		resetCodeRepo,
		hasher,
		usernameDeriver,
		verification.DefaultConfig,
	)
	sessionAuthService := auth.NewSessionAuthService(accountRepo, conf.Session.Secret)
	accountCommand := application.NewAccountCommandService(
		registrationService,
		loginService,
		passwordResetService,
		currentUserService,
		sessionAuthService,
		siteSettings,
	)
	apiKeySvc := auth.NewApiKeyService(apiKeyRepo, apiKeyRepo, conf.APIKey)
	authResolution := application.NewAuthResolutionService(currentUserService, apiKeySvc, accessTracker, sessionAuthService)
	apiKeyQuery := application.NewApiKeyQueryService(apiKeySvc)
	apiKeyCommand := application.NewApiKeyCommandService(apiKeySvc, siteSettings)
	systemSettingsQuery := application.NewSystemSettingsQueryService(systemSettingsService)
	systemSettingsCommand := application.NewSystemSettingsCommandService(systemSettingsService)
	auditLogQuery := application.NewAuditLogQueryService(auditLogRepo)
	auditLogCommand := application.NewAuditLogCommandService(auditLogRepo)

	return &ServiceContainer{
		ReviewQuery:           reviewQuery,
		ReviewCommand:         reviewCommand,
		CourseQuery:           courseQuery,
		CourseCommand:         courseCommand,
		CourseEnrollmentQuery: courseEnrollmentQuery,
		TeacherQuery:          teacherQuery,
		PointQuery:            pointQuery,
		PointRewardCommand:    pointRewardCommand,
		SiteStatsQuery:        siteStatsQuery,
		SiteStatsCommand:      siteStatsCommand,
		AccountQuery:          accountQuery,
		AccountCommand:        accountCommand,
		AdminUserQuery:        adminUserQuery,
		AdminUserCommand:      adminUserCommand,
		AuthUserService:       currentUserService,
		AuthResolution:        authResolution,
		AccessTracker:         accessTracker,
		AnnouncementQuery:     announcementQuery,
		ApiKeySvc:             apiKeySvc,
		ApiKeyQuery:           apiKeyQuery,
		ApiKeyCommand:         apiKeyCommand,
		SystemSettingsQuery:   systemSettingsQuery,
		SystemSettingsCommand: systemSettingsCommand,
		AuditLogQuery:         auditLogQuery,
		AuditLogCommand:       auditLogCommand,
		EmailSender:           smtpSender,
	}
}
