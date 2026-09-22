package app

import (
	"jcourse/config"
	"jcourse/internal/application"
	"jcourse/internal/domain/account"
	"jcourse/internal/domain/account/credential"
	"jcourse/internal/domain/account/identity"
	"jcourse/internal/domain/announcement"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/email"
	"jcourse/internal/domain/point"
	"jcourse/internal/infrastructure/persistence"
	"jcourse/internal/infrastructure/repository"
	"jcourse/internal/infrastructure/reshub"
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
	AnnouncementService   *announcement.Service
	ApiKeySvc             *auth.ApiKeyService
	ApiKeyQuery           *application.ApiKeyQueryService
	ApiKeyCommand         *application.ApiKeyCommandService
	SystemSettingsQuery   *application.SystemSettingsQueryService
	SystemSettingsCommand *application.SystemSettingsCommandService
	AuditLogQuery         *application.AuditLogQueryService
	AuditLogCommand       *application.AuditLogCommandService
	EmailSender           email.Sender
	ResHubOIDC            *application.ResHubOIDCService
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
	identityProjection := repository.NewUserIdentityRepository(db, conf.ResHub.Issuer)
	accountRepo := repository.NewAccountRepository(db, redisClient)
	userRepo := repository.NewUserRepository(db, redisClient)
	systemSettingsRepo := repository.NewSystemSettingsRepository(db, redisClient)
	systemSettingsService := newSystemSettingsService(systemSettingsRepo, courseRepo)
	siteSettings := application.NewSystemSiteSettingsProvider(systemSettingsService)
	apiKeyRepo := repository.NewApiKeyRepository(db)
	auditLogRepo := repository.NewAuditLogRepository(db)
	accessTracker := repository.NewAccessTrackerRepository(db, redisClient, conf.Auth.Access.FlushBatchSize)
	statRepo := repository.NewSiteDailyStatRepositoryWithIntegration(db, redisClient, true)
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
	)
	reviewCommand.SetGlobalSuspensionEnabled(false)
	courseHotService := course.NewCourseHotService(courseHotRepo)
	courseService := course.NewService(courseRepo)
	courseRatingCommand := course.NewCourseRatingCommandService(courseRepo, conf.Course.RatingScore)
	courseQuery := application.NewCourseQueryService(courseRepo, teacherRepo, reviewRepo, notificationRepo, courseEnrollmentRepo, courseHotRepo)
	courseEnrollmentService := course.NewEnrollmentService(courseEnrollmentRepo)
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
	announcementService := announcement.NewService(announcementRepo)
	var pointQueryRepository point.Query = pointRepo
	var rewardRepository point.RewardRepository = pointRepo
	resHubClient := reshub.NewClient(conf.ResHub)
	pointQueryRepository = reshub.NewPointQuery(resHubClient, identityProjection)
	rewardRepository = reshub.NewRewardRepository(db, resHubClient, identityProjection)
	resHubOIDC := application.NewResHubOIDCService(resHubClient, identityProjection, conf.ResHub.RedirectURI, conf.ResHub.LogoutReturnURI)
	resHubDependency := []application.ResHubAuthDependency{{
		Client:     resHubClient,
		Projection: identityProjection,
	}}
	pointQuery := application.NewPointQueryService(pointQueryRepository, accountRepo, usernameDeriver)
	pointRewardCommand := application.NewPointRewardCommandService(rewardRepository)
	statsConfig := conf.Stats
	siteStatsQuery := application.NewSiteStatsQueryService(statRepo, statsConfig)
	siteStatsCommand := application.NewSiteStatsCommandService(statRepo, statRepo, statsConfig)
	currentUserService := auth.NewCurrentUserService(userRepo)
	accountQuery := application.NewAccountQueryService(accountRepo)
	adminUserQuery := application.NewAdminUserQueryService(accountRepo, userRepo, reviewRepo, usernameDeriver)
	hasher := credential.NewDjangoPBKDF2SHA256PasswordHasher(conf.Auth.PasswordHash)
	adminUserCommand := application.NewAdminUserCommandService(userRepo, accountRepo, hasher)
	smtpSender := smtp.NewSMTPSender(conf.SMTP)
	registrationService := account.NewRegistrationService(
		accountRepo,
		verificationRepo,
		hasher,
		usernameDeriver,
	)
	loginService := account.NewLoginService(
		accountRepo,
		hasher,
		loginAttemptRepo,
		usernameDeriver,
	)
	passwordResetService := account.NewPasswordResetService(
		accountRepo,
		resetCodeRepo,
		hasher,
		usernameDeriver,
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
	authResolution := application.NewAuthResolutionService(currentUserService, apiKeySvc, accessTracker, sessionAuthService, resHubDependency...)
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
		AnnouncementService:   announcementService,
		ApiKeySvc:             apiKeySvc,
		ApiKeyQuery:           apiKeyQuery,
		ApiKeyCommand:         apiKeyCommand,
		SystemSettingsQuery:   systemSettingsQuery,
		SystemSettingsCommand: systemSettingsCommand,
		AuditLogQuery:         auditLogQuery,
		AuditLogCommand:       auditLogCommand,
		EmailSender:           smtpSender,
		ResHubOIDC:            resHubOIDC,
	}
}
