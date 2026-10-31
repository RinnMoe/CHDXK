package app

import (
	"jcourse/config"
	"jcourse/internal/application"
	"jcourse/internal/domain/announcement"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/email"
	"jcourse/internal/domain/point"
	"jcourse/internal/infrastructure/bkjw"
	"jcourse/internal/infrastructure/chdteachers"
	inframedia "jcourse/internal/infrastructure/media"
	"jcourse/internal/infrastructure/persistence"
	"jcourse/internal/infrastructure/repository"
	"jcourse/internal/infrastructure/reshub"
	"jcourse/internal/infrastructure/smtp"
	"path/filepath"
	"time"
)

type ServiceContainer struct {
	ResHubRefresh         auth.ResHubSessionRefresher
	ResHubOIDCClient      auth.ResHubOIDCClient
	Accounts              *application.LocalAccountService
	CourseSync            *application.CourseSyncService
	EnrollmentSync        *application.EnrollmentSyncService
	TeacherHotQuery       *application.TeacherHotQueryService
	ReviewQuery           *application.ReviewQueryService
	ReviewCommand         *application.ReviewCommandService
	CourseQuery           *application.CourseQueryService
	CourseCommand         *application.CourseCommandService
	CourseManagement      *application.CourseManagementService
	CourseEnrollmentQuery *application.CourseEnrollmentQueryService
	TeacherQuery          *application.TeacherQueryService
	TeacherProfiles       *application.TeacherProfileService
	PointQuery            *application.PointQueryService
	PointRewardCommand    *application.PointRewardCommandService
	SiteStatsQuery        *application.SiteStatsQueryService
	SiteStatsCommand      *application.SiteStatsCommandService
	AccountQuery          *application.AccountQueryService
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
	CanteenSettings       application.CanteenSettingsProvider
	CanteenHalls          *repository.CanteenHallRepository
	CanteenStores         *repository.CanteenStoreRepository
	CanteenDishes         *repository.CanteenDishRepository
	CanteenReviews        *repository.CanteenReviewRepository
	CanteenCatalogQuery   *application.CanteenCatalogQueryService
	CanteenAdmin          *application.CanteenAdminService
	CanteenReviewCommand  *application.CanteenReviewCommandService
	CanteenReviewQuery    *application.CanteenReviewQueryService
	CanteenDishCommand    *application.CanteenDishCommandService
	CanteenDishQuery      *application.CanteenDishQueryService
	CanteenSearchQuery    *application.CanteenSearchQueryService
	CanteenMediaCommand   *application.CanteenMediaCommandService
	CanteenMediaQuery     *application.CanteenMediaQueryService
	CanteenReportCommand  *application.CanteenReportCommandService
	EmailSender           email.Sender
	ResHubOIDC            *application.ResHubOIDCService
}

func NewServiceContainer(conf config.AppConfig) (*ServiceContainer, error) {
	resHubClient, err := reshub.NewClient(conf.ResHub)
	if err != nil {
		return nil, err
	}

	db := persistence.NewPostgres(conf.Postgres)
	redisClient := persistence.NewRedisClient(conf.Redis)
	resHubRefresh := reshub.NewRefreshCoordinator(redisClient, resHubClient, time.Duration(conf.Session.MaxAge)*time.Second)

	reviewRepo := repository.NewReviewRepositoryWithRatingScore(db, redisClient, conf.Course.RatingScore)
	voteRepo := repository.NewReviewVoteRepository(db, redisClient)
	courseRepo := repository.NewCourseRepository(db, redisClient)
	courseEnrollmentRepo := repository.NewCourseEnrollmentRepository(db, redisClient)
	teacherRepo := repository.NewTeacherRepository(db, redisClient)
	announcementRepo := repository.NewAnnouncementRepository(db)
	notificationRepo := repository.NewCourseNotificationRepository(db, redisClient)
	pointRepo := repository.NewPointRepository(db, redisClient)
	identityProjection := repository.NewUserIdentityRepository(db, conf.ResHub.Issuer)

	userRepo := repository.NewUserRepository(db, redisClient)
	systemSettingsRepo := repository.NewSystemSettingsRepository(db, redisClient)
	systemSettingsService := newSystemSettingsService(systemSettingsRepo, courseRepo)
	canteenHalls := repository.NewCanteenHallRepository(db, redisClient)
	canteenStores := repository.NewCanteenStoreRepository(db, redisClient)
	canteenStores.SetPriorCount(conf.Canteen.Rating.Normalized().PriorCount)
	canteenDishes := repository.NewCanteenDishRepository(db, redisClient)
	canteenReviews := repository.NewCanteenReviewRepository(db, redisClient)
	canteenReviews.SetPriorCount(conf.Canteen.Rating.Normalized().PriorCount)
	canteenVotes := repository.NewCanteenVoteRepository(db)
	canteenSettings := newSystemCanteenSettingsProvider(systemSettingsService)
	canteenStoreQuery := repository.NewCanteenStoreQueryRepository(db, redisClient)
	canteenCatalogQuery := application.NewCanteenCatalogQueryService(canteenStoreQuery, canteenStoreQuery, canteenSettings)
	canteenAdmin := application.NewCanteenAdminService(canteenHalls, canteenStores)
	canteenMediaRepo := repository.NewCanteenMediaRepository(db)
	canteenMediaStorage, err := inframedia.NewLocalStorage(conf.Canteen.Media.Normalized().Dir)
	if err != nil {
		return nil, err
	}
	canteenMediaCommand := application.NewCanteenMediaCommandService(
		canteenMediaRepo,
		canteenMediaRepo,
		canteenMediaStorage,
		inframedia.Processor{},
		canteenSettings,
		application.CanteenMediaLimits{
			MaxBytes:         conf.Canteen.Media.Normalized().MaxBytes,
			MaxPerReview:     conf.Canteen.Media.Normalized().MaxPerReview,
			MaxUploadsPerDay: conf.Canteen.Media.Normalized().MaxUploadsPerDay,
			MaxPixels:        conf.Canteen.Media.Normalized().MaxPixels,
		},
	)
	canteenMediaQuery := application.NewCanteenMediaQueryService(canteenMediaRepo, canteenSettings)
	canteenReportCommand := application.NewCanteenReportCommandService(repository.NewCanteenReportRepository(db), canteenSettings)

	canteenReviewCommand := application.NewCanteenReviewCommandService(
		canteenStores,
		canteenReviews,
		canteenVotes,
		canteenStoreQuery,
		canteenSettings,
		conf.Canteen.Review.Frequency,
		conf.Canteen.Review.MaxDailyReviews,
		canteenMediaCommand,
	)
	canteenReviewQuery := application.NewCanteenReviewQueryService(canteenStoreQuery, canteenSettings)
	canteenDishQueryRepo := repository.NewCanteenDishQueryRepository(db)
	canteenDishReviews := repository.NewCanteenDishReviewRepository(db, redisClient)
	canteenDishReviews.SetPriorCount(conf.Canteen.Rating.Normalized().PriorCount)
	canteenDishQueryService := application.NewCanteenDishQueryService(canteenDishQueryRepo, canteenSettings)
	canteenSearchQuery := application.NewCanteenSearchQueryService(canteenStoreQuery, canteenDishQueryRepo, canteenSettings)
	canteenDishCommand := application.NewCanteenDishCommandService(
		canteenStores,
		canteenDishes,
		canteenDishQueryRepo,
		canteenDishReviews,
		canteenStoreQuery,
		canteenSettings,
		canteenMediaCommand,
	)
	siteSettings := application.NewSystemSiteSettingsProvider(systemSettingsService)
	apiKeyRepo := repository.NewApiKeyRepository(db)
	auditLogRepo := repository.NewAuditLogRepository(db)
	accessTracker := repository.NewAccessTrackerRepository(db, redisClient, conf.Auth.Access.FlushBatchSize)
	statRepo := repository.NewSiteDailyStatRepositoryWithIntegration(db, redisClient, true)
	courseHotRepo := repository.NewGormCourseHotRepository(db, redisClient)
	reviewQuery := application.NewReviewQueryService(reviewRepo, voteRepo, notificationRepo, courseRepo)

	reviewCommand := application.NewReviewCommandService(
		courseRepo,
		reviewRepo,
		voteRepo,
		siteSettings,
	)
	reviewCommand.SetGlobalSuspensionEnabled(true)
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
	courseManagement := application.NewCourseManagementService(courseRepo)
	teacherQuery := application.NewTeacherQueryService(teacherRepo)
	announcementQuery := application.NewAnnouncementQueryService(announcementRepo)
	announcementService := announcement.NewService(announcementRepo)
	var pointQueryRepository point.Query = pointRepo
	var rewardRepository point.RewardRepository = pointRepo
	pointQueryRepository = reshub.NewPointQuery(resHubClient, identityProjection)
	rewardRepository = reshub.NewRewardRepository(db, resHubClient, identityProjection)
	resHubOIDC := application.NewResHubOIDCService(resHubClient, identityProjection, conf.ResHub.RedirectURI, conf.ResHub.LogoutReturnURI)
	pointQuery := application.NewPointQueryService(pointQueryRepository)
	pointRewardCommand := application.NewPointRewardCommandService(rewardRepository)
	statsConfig := conf.Stats
	siteStatsQuery := application.NewSiteStatsQueryService(statRepo, statsConfig)
	siteStatsCommand := application.NewSiteStatsCommandService(statRepo, statRepo, statsConfig)
	currentUserService := auth.NewCurrentUserService(userRepo)
	accountQuery := application.NewAccountQueryService()
	smtpSender := smtp.NewSMTPSender(conf.SMTP)
	apiKeySvc := auth.NewApiKeyService(apiKeyRepo, apiKeyRepo, conf.APIKey)
	authResolution := application.NewAuthResolutionService(apiKeySvc, accessTracker, resHubClient, identityProjection)
	authResolution.SetAccounts(identityProjection)
	apiKeyQuery := application.NewApiKeyQueryService(apiKeySvc)
	apiKeyCommand := application.NewApiKeyCommandService(apiKeySvc)
	systemSettingsQuery := application.NewSystemSettingsQueryService(systemSettingsService)
	systemSettingsCommand := application.NewSystemSettingsCommandService(systemSettingsService)
	auditLogQuery := application.NewAuditLogQueryService(auditLogRepo)
	auditLogCommand := application.NewAuditLogCommandService(auditLogRepo)
	collectorQueue := application.NewCollectorQueue(128, 8)
	requestBudget := bkjw.NewRequestBudget()
	jobStore := bkjw.NewJobStore(conf.CourseSource.CacheDir)
	collectorQueue.UseSettings(systemSettingsService)
	systemSettingsCommand.SetCollectorQueue(collectorQueue)
	courseCommand.SetImportSettings(systemSettingsService)

	return &ServiceContainer{
		ResHubRefresh:         resHubRefresh,
		Accounts:              application.NewLocalAccountService(identityProjection, systemSettingsService, resHubClient, auditLogRepo),
		ResHubOIDCClient:      resHubClient,
		ReviewQuery:           reviewQuery,
		ReviewCommand:         reviewCommand,
		CourseQuery:           courseQuery,
		CourseCommand:         courseCommand,
		CourseManagement:      courseManagement,
		CourseSync:            application.NewCourseSyncService(&bkjw.Source{Config: conf.CourseSource, Budget: requestBudget}, courseManagement, collectorQueue).WithJobStore(jobStore),
		EnrollmentSync:        application.NewEnrollmentSyncService(&bkjw.Source{Config: conf.CourseSource, Personal: true, Budget: requestBudget}, courseEnrollmentRepo, collectorQueue).WithJobStore(jobStore),
		CourseEnrollmentQuery: courseEnrollmentQuery,
		TeacherQuery:          teacherQuery,
		TeacherProfiles:       application.NewTeacherProfileService(teacherRepo, chdteachers.NewClient(filepath.Join(conf.CourseSource.CacheDir, "teacher-profiles"))),
		TeacherHotQuery:       application.NewTeacherHotQueryService(courseHotRepo),
		PointQuery:            pointQuery,
		PointRewardCommand:    pointRewardCommand,
		SiteStatsQuery:        siteStatsQuery,
		SiteStatsCommand:      siteStatsCommand,
		AccountQuery:          accountQuery,
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
		CanteenSettings:       canteenSettings,
		CanteenHalls:          canteenHalls,
		CanteenStores:         canteenStores,
		CanteenDishes:         canteenDishes,
		CanteenReviews:        canteenReviews,
		CanteenCatalogQuery:   canteenCatalogQuery,
		CanteenAdmin:          canteenAdmin,
		CanteenReviewCommand:  canteenReviewCommand,
		CanteenReviewQuery:    canteenReviewQuery,
		CanteenDishCommand:    canteenDishCommand,
		CanteenDishQuery:      canteenDishQueryService,
		CanteenSearchQuery:    canteenSearchQuery,
		CanteenMediaCommand:   canteenMediaCommand,
		CanteenMediaQuery:     canteenMediaQuery,
		CanteenReportCommand:  canteenReportCommand,
		EmailSender:           smtpSender,
		ResHubOIDC:            resHubOIDC,
	}, nil
}
