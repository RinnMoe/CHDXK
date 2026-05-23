package app

import (
	"time"

	"jcourse/config"
	"jcourse/internal/application"
	"jcourse/internal/domain/account"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/point"
	"jcourse/internal/domain/review"
	"jcourse/internal/domain/review/policy"
	"jcourse/internal/infrastructure/email"
	"jcourse/internal/infrastructure/persistence"
	"jcourse/internal/infrastructure/repository"
)

type ServiceContainer struct {
	ReviewQuery       *application.ReviewQueryService
	ReviewCommand     *application.ReviewCommandService
	CourseQuery       *application.CourseQueryService
	CourseCommand     *application.CourseCommandService
	TeacherQuery      *application.TeacherQueryService
	PointQuery        *application.PointQueryService
	PointCommand      *application.PointCommandService
	SiteStatsQuery    *application.SiteStatsQueryService
	SiteStatsCommand  *application.SiteStatsCommandService
	AccountQuery      *application.AccountQueryService
	AccountCommand    *application.AccountCommandService
	AuthUserService   *auth.AuthUserService
	AnnouncementQuery *application.AnnouncementQueryService
	ApiKeySvc         *auth.ApiKeyService
	ApiKeyQuery       *application.ApiKeyQueryService
	ApiKeyCommand     *application.ApiKeyCommandService
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
	accountRepo := repository.NewAccountRepository(db)
	userRepo := repository.NewUserRepository(db)
	apiKeyRepo := repository.NewApiKeyRepository(db)
	statRepo := repository.NewSiteDailyStatRepository(db)
	courseHotRepo := repository.NewGormCourseHotRepository(db)
	verificationRepo := repository.NewVerificationCodeRepository(redisClient)
	resetCodeRepo := repository.NewVerificationCodeRepositoryWithPrefix(redisClient, "reset")
	loginAttemptRepo := repository.NewLoginAttemptRepository(redisClient, time.Duration(conf.Auth.LoginLockoutDuration)*time.Second)
	usernameDeriver := account.NewBLAKE2bUsernameDeriver(conf.Auth.UsernameSalt)

	reviewQuery := application.NewReviewQueryService(reviewRepo, voteRepo, notificationRepo)

	freqPolicy := policy.NewFrequencyPolicy(reviewRepo, policy.DefaultFrequencyPolicyConfig())
	safetyPolicy := policy.NewSafetyPolicy(nil)

	reviewCommand := application.NewReviewCommandService(
		courseRepo,
		reviewRepo,
		voteRepo,
		courseHotRepo,
		application.CourseHotScoreConfig{
			ReviewCreateScore: conf.CourseHot.ReviewCreateScore,
			ReviewUpdateScore: conf.CourseHot.ReviewUpdateScore,
			ReviewVoteScore:   conf.CourseHot.ReviewVoteScore,
		},
		[]review.CreatePolicy{freqPolicy, safetyPolicy},
	)
	courseQuery := application.NewCourseQueryService(courseRepo, teacherRepo, reviewRepo, notificationRepo, courseHotRepo)
	courseCommand := application.NewCourseCommandService(courseRepo, notificationRepo)
	teacherQuery := application.NewTeacherQueryService(teacherRepo)
	announcementQuery := application.NewAnnouncementQueryService(announcementRepo)
	pointQuery := application.NewPointQueryService(pointRepo, accountRepo, point.NewTransferService(point.TransferFeeConfig{
		RateBps: conf.Point.TransferFeeRateBps,
		MinFee:  conf.Point.TransferMinFee,
	}), usernameDeriver)
	pointCommand := application.NewPointCommandService(accountRepo, pointRepo, point.NewTransferService(point.TransferFeeConfig{
		RateBps: conf.Point.TransferFeeRateBps,
		MinFee:  conf.Point.TransferMinFee,
	}))
	siteStatsQuery := application.NewSiteStatsQueryService(statRepo)
	siteStatsCommand := application.NewSiteStatsCommandService(statRepo, statRepo)
	currentUserService := auth.NewCurrentUserService(userRepo)
	accountQuery := application.NewAccountQueryService(accountRepo)
	accountCommand := application.NewAccountCommandService(
		accountRepo,
		currentUserService,
		verificationRepo,
		resetCodeRepo,
		email.NewSMTPVerificationCodeSender(conf.SMTP),
		account.NewDjangoPBKDF2SHA256PasswordHasher(0),
		usernameDeriver,
		application.AccountCommandConfig{
			EmailWhitelist: conf.Auth.EmailWhitelist,
			CodeInterval:   time.Duration(conf.Auth.VerificationCodeInterval) * time.Second,
			CodeTTL:        time.Duration(conf.Auth.VerificationCodeTTL) * time.Second,
		},
		account.PasswordResetConfig{
			CodeInterval: time.Duration(conf.Auth.VerificationCodeInterval) * time.Second,
			CodeTTL:      time.Duration(conf.Auth.VerificationCodeTTL) * time.Second,
		},
		loginAttemptRepo,
		conf.Auth.MaxLoginAttempts,
		time.Duration(conf.Auth.LoginLockoutDuration)*time.Second,
	)
	apiKeySvc := auth.NewApiKeyService(apiKeyRepo)
	apiKeyQuery := application.NewApiKeyQueryService(apiKeySvc)
	apiKeyCommand := application.NewApiKeyCommandService(apiKeySvc)

	return &ServiceContainer{
		ReviewQuery:       reviewQuery,
		ReviewCommand:     reviewCommand,
		CourseQuery:       courseQuery,
		CourseCommand:     courseCommand,
		TeacherQuery:      teacherQuery,
		PointQuery:        pointQuery,
		PointCommand:      pointCommand,
		SiteStatsQuery:    siteStatsQuery,
		SiteStatsCommand:  siteStatsCommand,
		AccountQuery:      accountQuery,
		AccountCommand:    accountCommand,
		AuthUserService:   currentUserService,
		AnnouncementQuery: announcementQuery,
		ApiKeySvc:         apiKeySvc,
		ApiKeyQuery:       apiKeyQuery,
		ApiKeyCommand:     apiKeyCommand,
	}
}
