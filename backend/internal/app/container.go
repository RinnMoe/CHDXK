package app

import (
	"jcourse/config"
	"jcourse/internal/application"
	"jcourse/internal/domain/account"
	"jcourse/internal/domain/account/credential"
	"jcourse/internal/domain/account/identity"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/email"
	"jcourse/internal/domain/point"
	"jcourse/internal/domain/review"
	"jcourse/internal/domain/review/policy"
	"jcourse/internal/infrastructure/jaccount"
	"jcourse/internal/infrastructure/moderation"
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
	PointCommand          *application.PointCommandService
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
	UserSettingsQuery     *application.UserSettingsQueryService
	UserSettingsCommand   *application.UserSettingsCommandService
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
	userSettingsRepo := repository.NewUserSettingsRepository(db, redisClient)
	apiKeyRepo := repository.NewApiKeyRepository(db)
	auditLogRepo := repository.NewAuditLogRepository(db)
	accessTracker := repository.NewAccessTrackerRepository(db, redisClient, conf.Auth.Access.FlushBatchSize)
	statRepo := repository.NewSiteDailyStatRepository(db, redisClient)
	courseHotRepo := repository.NewGormCourseHotRepository(db, redisClient)
	verificationRepo := repository.NewVerificationCodeRepository(redisClient)
	resetCodeRepo := repository.NewVerificationCodeRepositoryWithPrefix(redisClient, "reset")
	loginAttemptRepo := repository.NewLoginAttemptRepository(redisClient, conf.Auth.Login.Lockout)
	usernameDeriver := identity.NewBLAKE2bUsernameDeriver(conf.Auth.UsernameDeriver)

	reviewQuery := application.NewReviewQueryService(reviewRepo, voteRepo, notificationRepo)

	freqPolicy := policy.NewFrequencyPolicy(reviewRepo, conf.Review.FrequencyPolicy)
	moderator, _ := moderation.NewAliyunGreenModerator(conf.Review.SafetyPolicy)
	safetyPolicy := policy.NewSafetyPolicy(moderator)

	reviewCommand := application.NewReviewCommandService(
		courseRepo,
		reviewRepo,
		voteRepo,
		conf.Review.Command,
		[]review.CreatePolicy{freqPolicy, safetyPolicy},
	)
	courseHotService := course.NewCourseHotService(courseHotRepo, conf.Review.Command.HotScores)
	courseRatingCommand := course.NewCourseRatingCommandService(courseRepo, conf.Course.RatingScore)
	courseQuery := application.NewCourseQueryService(courseRepo, teacherRepo, reviewRepo, notificationRepo, courseEnrollmentRepo, courseHotRepo)
	jaccountClient := jaccount.NewOAuthClient(conf.JAccount)
	courseEnrollmentService := course.NewEnrollmentService(courseRepo, courseEnrollmentRepo, jaccountClient)
	courseCommand := application.NewCourseCommandService(
		course.NewNotificationService(courseRepo, notificationRepo),
		courseEnrollmentService,
		courseHotService,
		courseRatingCommand,
	)
	courseEnrollmentQuery := application.NewCourseEnrollmentQueryService(courseEnrollmentRepo)
	teacherQuery := application.NewTeacherQueryService(teacherRepo)
	announcementQuery := application.NewAnnouncementQueryService(announcementRepo)
	transferService := point.NewTransferService(conf.Point)
	pointQuery := application.NewPointQueryService(pointRepo, accountRepo, transferService, usernameDeriver)
	pointCommand := application.NewPointCommandService(accountRepo, pointRepo, transferService, usernameDeriver)
	statsConfig := conf.Stats
	siteStatsQuery := application.NewSiteStatsQueryService(statRepo, statsConfig)
	siteStatsCommand := application.NewSiteStatsCommandService(statRepo, statRepo, statsConfig)
	currentUserService := auth.NewCurrentUserService(userRepo)
	accountQuery := application.NewAccountQueryService(accountRepo)
	adminUserQuery := application.NewAdminUserQueryService(accountRepo, userRepo, usernameDeriver)
	adminUserCommand := application.NewAdminUserCommandService(userRepo, conf.Admin)
	hasher := credential.NewDjangoPBKDF2SHA256PasswordHasher(conf.Auth.PasswordHash)
	smtpSender := smtp.NewSMTPSender(conf.SMTP)
	registrationService := account.NewRegistrationService(
		accountRepo,
		verificationRepo,
		hasher,
		usernameDeriver,
		conf.Auth.Registration,
		conf.Auth.Verification,
	)
	loginService := account.NewLoginService(
		accountRepo,
		hasher,
		loginAttemptRepo,
		usernameDeriver,
		conf.Auth.Login,
	)
	passwordResetService := account.NewPasswordResetService(
		accountRepo,
		resetCodeRepo,
		hasher,
		usernameDeriver,
		conf.Auth.Verification,
	)
	accountCommand := application.NewAccountCommandService(
		registrationService,
		loginService,
		passwordResetService,
		currentUserService,
	)
	apiKeySvc := auth.NewApiKeyService(apiKeyRepo, apiKeyRepo, conf.APIKey)
	authResolution := application.NewAuthResolutionService(currentUserService, apiKeySvc, accessTracker)
	apiKeyQuery := application.NewApiKeyQueryService(apiKeySvc)
	apiKeyCommand := application.NewApiKeyCommandService(apiKeySvc)
	userSettingsQuery := application.NewUserSettingsQueryService(userSettingsRepo, courseRepo)
	userSettingsCommand := application.NewUserSettingsCommandService(userSettingsRepo, courseRepo)
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
		PointCommand:          pointCommand,
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
		UserSettingsQuery:     userSettingsQuery,
		UserSettingsCommand:   userSettingsCommand,
		AuditLogQuery:         auditLogQuery,
		AuditLogCommand:       auditLogCommand,
		EmailSender:           smtpSender,
	}
}
