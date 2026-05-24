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
	"jcourse/internal/domain/stat"
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
	AdminUserQuery    *application.AdminUserQueryService
	AdminUserCommand  *application.AdminUserCommandService
	AuthUserService   *auth.AuthUserService
	AnnouncementQuery *application.AnnouncementQueryService
	ApiKeySvc         *auth.ApiKeyService
	ApiKeyQuery       *application.ApiKeyQueryService
	ApiKeyCommand     *application.ApiKeyCommandService
}

func NewServiceContainer(conf config.AppConfig) *ServiceContainer {

	db := persistence.NewPostgres(conf.Postgres)
	redisClient := persistence.NewRedisClient(conf.Redis)

	reviewRepo := repository.NewReviewRepository(db, redisClient)
	voteRepo := repository.NewReviewVoteRepository(db, redisClient)
	courseRepo := repository.NewCourseRepository(db, redisClient)
	teacherRepo := repository.NewTeacherRepository(db, redisClient)
	announcementRepo := repository.NewAnnouncementRepository(db)
	notificationRepo := repository.NewCourseNotificationRepository(db, redisClient)
	pointRepo := repository.NewPointRepository(db, redisClient)
	accountRepo := repository.NewAccountRepository(db, redisClient)
	userRepo := repository.NewUserRepository(db, redisClient)
	apiKeyRepo := repository.NewApiKeyRepository(db)
	statRepo := repository.NewSiteDailyStatRepository(db, redisClient)
	courseHotRepo := repository.NewGormCourseHotRepository(db, redisClient)
	verificationRepo := repository.NewVerificationCodeRepository(redisClient)
	resetCodeRepo := repository.NewVerificationCodeRepositoryWithPrefix(redisClient, "reset")
	loginAttemptRepo := repository.NewLoginAttemptRepository(redisClient, time.Duration(conf.Auth.LoginLockoutDuration)*time.Second)
	usernameDeriver := account.NewBLAKE2bUsernameDeriver(conf.Auth.UsernameSalt)

	reviewQuery := application.NewReviewQueryService(reviewRepo, voteRepo, notificationRepo)

	freqPolicy := policy.NewFrequencyPolicy(reviewRepo, policy.FrequencyPolicyConfig{
		Window:          time.Duration(conf.Review.FrequencyWindowSeconds) * time.Second,
		MaxReviews:      conf.Review.FrequencyMaxReviews,
		SimilarityRatio: conf.Review.SimilarityRatio,
	})
	safetyPolicy := policy.NewSafetyPolicy(nil)

	reviewCommand := application.NewReviewCommandService(
		courseRepo,
		reviewRepo,
		voteRepo,
		courseHotRepo,
		application.ReviewCommandConfig{
			HotScores: application.CourseHotScoreConfig{
				ReviewCreateScore: conf.CourseHot.ReviewCreateScore,
				ReviewUpdateScore: conf.CourseHot.ReviewUpdateScore,
				ReviewVoteScore:   conf.CourseHot.ReviewVoteScore,
			},
			Vote: review.VoteConfig{MaxDailyVotes: conf.Review.MaxDailyVotes},
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
	statsConfig := stat.Config{Timezone: conf.Stats.Timezone}
	siteStatsQuery := application.NewSiteStatsQueryService(statRepo, statsConfig)
	siteStatsCommand := application.NewSiteStatsCommandService(statRepo, statRepo, statsConfig)
	currentUserService := auth.NewCurrentUserService(userRepo)
	accountQuery := application.NewAccountQueryService(accountRepo)
	adminUserQuery := application.NewAdminUserQueryService(accountRepo, userRepo, usernameDeriver)
	adminUserCommand := application.NewAdminUserCommandService(userRepo, application.AdminUserCommandConfig{DefaultSuspendDays: conf.Admin.DefaultSuspendDays})
	accountCommand := application.NewAccountCommandService(
		accountRepo,
		currentUserService,
		verificationRepo,
		resetCodeRepo,
		email.NewSMTPVerificationCodeSender(conf.SMTP),
		account.NewDjangoPBKDF2SHA256PasswordHasher(account.PasswordHashConfig{
			Iterations: conf.Auth.PasswordHashIterations,
			SaltLength: conf.Auth.PasswordSaltLength,
		}),
		usernameDeriver,
		application.AccountCommandConfig{
			Registration: account.RegistrationConfig{
				EmailWhitelist: conf.Auth.EmailWhitelist,
				CodeInterval:   time.Duration(conf.Auth.VerificationCodeInterval) * time.Second,
				CodeTTL:        time.Duration(conf.Auth.VerificationCodeTTL) * time.Second,
				CodeLength:     conf.Auth.VerificationCodeLength,
			},
			PasswordReset: account.PasswordResetConfig{
				CodeInterval: time.Duration(conf.Auth.VerificationCodeInterval) * time.Second,
				CodeTTL:      time.Duration(conf.Auth.VerificationCodeTTL) * time.Second,
				CodeLength:   conf.Auth.VerificationCodeLength,
			},
			Login: account.LoginConfig{
				MaxAttempts: conf.Auth.MaxLoginAttempts,
				Lockout:     time.Duration(conf.Auth.LoginLockoutDuration) * time.Second,
			},
		},
		loginAttemptRepo,
	)
	apiKeySvc := auth.NewApiKeyService(apiKeyRepo, auth.ApiKeyConfig{MaxUserKeys: conf.APIKey.MaxUserKeys})
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
		AdminUserQuery:    adminUserQuery,
		AdminUserCommand:  adminUserCommand,
		AuthUserService:   currentUserService,
		AnnouncementQuery: announcementQuery,
		ApiKeySvc:         apiKeySvc,
		ApiKeyQuery:       apiKeyQuery,
		ApiKeyCommand:     apiKeyCommand,
	}
}
