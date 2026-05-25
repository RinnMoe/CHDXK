package app

import (
	"jcourse/config"
	"jcourse/internal/application"
	"jcourse/internal/domain/account"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/point"
	"jcourse/internal/domain/review"
	"jcourse/internal/domain/review/policy"
	"jcourse/internal/infrastructure/email"
	"jcourse/internal/infrastructure/jaccount"
	"jcourse/internal/infrastructure/persistence"
	"jcourse/internal/infrastructure/repository"
)

type ServiceContainer struct {
	ReviewQuery             *application.ReviewQueryService
	ReviewCommand           *application.ReviewCommandService
	CourseHotCommand        *application.CourseHotCommandService
	CourseQuery             *application.CourseQueryService
	CourseCommand           *application.CourseCommandService
	CourseEnrollmentQuery   *application.CourseEnrollmentQueryService
	CourseEnrollmentCommand *application.CourseEnrollmentCommandService
	CourseEnrollmentSync    *application.CourseEnrollmentSyncService
	TeacherQuery            *application.TeacherQueryService
	PointQuery              *application.PointQueryService
	PointCommand            *application.PointCommandService
	SiteStatsQuery          *application.SiteStatsQueryService
	SiteStatsCommand        *application.SiteStatsCommandService
	AccountQuery            *application.AccountQueryService
	AccountCommand          *application.AccountCommandService
	AdminUserQuery          *application.AdminUserQueryService
	AdminUserCommand        *application.AdminUserCommandService
	AuthUserService         *auth.AuthUserService
	AnnouncementQuery       *application.AnnouncementQueryService
	ApiKeySvc               *auth.ApiKeyService
	ApiKeyQuery             *application.ApiKeyQueryService
	ApiKeyCommand           *application.ApiKeyCommandService
	UserSettingsQuery       *application.UserSettingsQueryService
	UserSettingsCommand     *application.UserSettingsCommandService
}

func NewServiceContainer(conf config.AppConfig) *ServiceContainer {

	db := persistence.NewPostgres(conf.Postgres)
	redisClient := persistence.NewRedisClient(conf.Redis)

	reviewRepo := repository.NewReviewRepository(db, redisClient)
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
	statRepo := repository.NewSiteDailyStatRepository(db, redisClient)
	courseHotRepo := repository.NewGormCourseHotRepository(db, redisClient)
	verificationRepo := repository.NewVerificationCodeRepository(redisClient)
	resetCodeRepo := repository.NewVerificationCodeRepositoryWithPrefix(redisClient, "reset")
	loginAttemptRepo := repository.NewLoginAttemptRepository(redisClient, conf.Auth.Login.Lockout)
	usernameDeriver := account.NewBLAKE2bUsernameDeriver(conf.Auth.UsernameDeriver)

	reviewQuery := application.NewReviewQueryService(reviewRepo, voteRepo, notificationRepo)

	freqPolicy := policy.NewFrequencyPolicy(reviewRepo, conf.Review.FrequencyPolicy)
	safetyPolicy := policy.NewSafetyPolicy(nil)

	reviewCommand := application.NewReviewCommandService(
		courseRepo,
		reviewRepo,
		voteRepo,
		conf.Review.Command,
		[]review.CreatePolicy{freqPolicy, safetyPolicy},
	)
	courseHotCommand := application.NewCourseHotCommandService(courseHotRepo, conf.Review.Command.HotScores)
	courseQuery := application.NewCourseQueryService(courseRepo, teacherRepo, reviewRepo, notificationRepo, courseEnrollmentRepo, courseHotRepo)
	courseCommand := application.NewCourseCommandService(courseRepo, notificationRepo)
	jaccountClient := jaccount.NewOAuthClient(conf.JAccount)
	courseEnrollmentQuery := application.NewCourseEnrollmentQueryService(courseEnrollmentRepo)
	courseEnrollmentCommand := application.NewCourseEnrollmentCommandService(courseRepo, courseEnrollmentRepo)
	courseEnrollmentSync := application.NewCourseEnrollmentSyncService(courseEnrollmentRepo, courseRepo, jaccountClient)
	teacherQuery := application.NewTeacherQueryService(teacherRepo)
	announcementQuery := application.NewAnnouncementQueryService(announcementRepo)
	transferService := point.NewTransferService(conf.Point)
	pointQuery := application.NewPointQueryService(pointRepo, accountRepo, transferService, usernameDeriver)
	pointCommand := application.NewPointCommandService(accountRepo, pointRepo, transferService)
	statsConfig := conf.Stats
	siteStatsQuery := application.NewSiteStatsQueryService(statRepo, statsConfig)
	siteStatsCommand := application.NewSiteStatsCommandService(statRepo, statRepo, statsConfig)
	currentUserService := auth.NewCurrentUserService(userRepo)
	accountQuery := application.NewAccountQueryService(accountRepo)
	adminUserQuery := application.NewAdminUserQueryService(accountRepo, userRepo, usernameDeriver)
	adminUserCommand := application.NewAdminUserCommandService(userRepo, conf.Admin)
	hasher := account.NewDjangoPBKDF2SHA256PasswordHasher(conf.Auth.PasswordHash)
	verificationSender := email.NewSMTPSender(conf.SMTP)
	registrationService := account.NewRegistrationService(
		accountRepo,
		verificationRepo,
		verificationSender,
		hasher,
		usernameDeriver,
		conf.Auth.Registration,
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
		verificationSender,
		hasher,
		usernameDeriver,
		conf.Auth.PasswordReset,
	)
	accountCommand := application.NewAccountCommandService(
		registrationService,
		loginService,
		passwordResetService,
		currentUserService,
	)
	apiKeySvc := auth.NewApiKeyService(apiKeyRepo, conf.APIKey)
	apiKeyQuery := application.NewApiKeyQueryService(apiKeySvc)
	apiKeyCommand := application.NewApiKeyCommandService(apiKeySvc)
	userSettingsQuery := application.NewUserSettingsQueryService(userSettingsRepo, courseRepo, courseRepo)
	userSettingsCommand := application.NewUserSettingsCommandService(userSettingsRepo, courseRepo)

	return &ServiceContainer{
		ReviewQuery:             reviewQuery,
		ReviewCommand:           reviewCommand,
		CourseHotCommand:        courseHotCommand,
		CourseQuery:             courseQuery,
		CourseCommand:           courseCommand,
		CourseEnrollmentQuery:   courseEnrollmentQuery,
		CourseEnrollmentCommand: courseEnrollmentCommand,
		CourseEnrollmentSync:    courseEnrollmentSync,
		TeacherQuery:            teacherQuery,
		PointQuery:              pointQuery,
		PointCommand:            pointCommand,
		SiteStatsQuery:          siteStatsQuery,
		SiteStatsCommand:        siteStatsCommand,
		AccountQuery:            accountQuery,
		AccountCommand:          accountCommand,
		AdminUserQuery:          adminUserQuery,
		AdminUserCommand:        adminUserCommand,
		AuthUserService:         currentUserService,
		AnnouncementQuery:       announcementQuery,
		ApiKeySvc:               apiKeySvc,
		ApiKeyQuery:             apiKeyQuery,
		ApiKeyCommand:           apiKeyCommand,
		UserSettingsQuery:       userSettingsQuery,
		UserSettingsCommand:     userSettingsCommand,
	}
}
