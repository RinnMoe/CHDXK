package web

import (
	stdslog "log/slog"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	ginslog "github.com/gin-contrib/slog"
	"github.com/gin-gonic/gin"

	"jcourse/config"
	"jcourse/internal/app"
	"jcourse/internal/interface/web/controller"
	"jcourse/internal/interface/web/middleware"
	"jcourse/pkg/logx"
	"jcourse/pkg/requestid"
)

func NewRouter(container *app.ServiceContainer, conf config.AppConfig) *gin.Engine {
	if conf.Server.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	g := gin.New()
	g.Use(middleware.RequestID())
	g.Use(ginslog.SetLogger(
		ginslog.WithLogger(func(_ *gin.Context, _ *stdslog.Logger) *stdslog.Logger {
			return logx.Logger()
		}),
		ginslog.WithContext(func(c *gin.Context, rec *stdslog.Record) *stdslog.Record {
			if id, ok := requestid.FromContext(c.Request.Context()); ok {
				rec.Add(requestid.GinKey, id)
			}
			return rec
		}),
	))
	g.Use(gin.Recovery())
	g.Use(cors.New(cors.Config{
		AllowOrigins: conf.Server.Cors.AllowedOrigins,
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
		},
		ExposeHeaders: []string{
			"X-CSRF-Token",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	store, err := middleware.NewSessionStore(conf.Redis, conf.Session)
	if err != nil {
		panic(err)
	}
	g.Use(middleware.GlobalRateLimit())
	g.Use(sessions.Sessions("jcourse_session", store))
	g.Use(middleware.ResolveCurrentUser(container.AuthResolution))
	g.Use(middleware.UserIDRateLimit())
	g.Use(middleware.CSRF())

	reviewController := controller.NewReviewController(container.ReviewQuery, container.ReviewCommand)
	courseController := controller.NewCourseController(container.CourseQuery, container.CourseCommand)
	courseEnrollmentController := controller.NewCourseEnrollmentController(container.CourseEnrollmentQuery, container.CourseCommand)
	courseEnrollmentSyncController := controller.NewCourseEnrollmentSyncController(container.CourseCommand, conf.JAccount)
	teacherController := controller.NewTeacherController(container.TeacherQuery, container.CourseQuery)
	pointController := controller.NewPointController(container.PointQuery, container.PointCommand)
	accountController := controller.NewAccountController(container.AccountCommand, container.AccountQuery)
	apiKeyController := controller.NewApiKeyController(container.ApiKeyQuery, container.ApiKeyCommand)
	userSettingsController := controller.NewUserSettingsController(container.UserSettingsQuery, container.UserSettingsCommand)
	siteStatsController := controller.NewSiteStatsController(container.SiteStatsQuery)
	announcementController := controller.NewAnnouncementController(container.AnnouncementQuery)
	adminUserController := controller.NewAdminUserController(container.AdminUserQuery, container.AdminUserCommand)
	auditLogController := controller.NewAuditLogController(container.AuditLogQuery)

	apiGroup := g.Group("/api", middleware.NoStore())
	publicAuthGroup := apiGroup.Group("/auth")
	{
		publicAuthGroup.GET("/csrf", func(c *gin.Context) { c.Status(http.StatusNoContent) })
		publicAuthGroup.POST("/register/code", accountController.SendRegisterCode)
		publicAuthGroup.POST("/register", accountController.Register)
		publicAuthGroup.POST("/login", accountController.Login)
		publicAuthGroup.POST("/password-reset/code", accountController.SendResetCode)
		publicAuthGroup.POST("/password-reset", accountController.ResetPassword)
	}

	extGroup := apiGroup.Group("/ext", middleware.SystemAPIKeyAuth())
	{
		extGroup.GET("/point", pointController.GetPointsByEmail)
	}

	enrollmentSyncGroup := apiGroup.Group("/course/enrollment-sync")
	{
		enrollmentSyncGroup.GET("/callback", courseEnrollmentSyncController.Callback)
		enrollmentSyncGroup.GET("/start", middleware.RequireAuth(), courseEnrollmentSyncController.Start)
	}

	apiGroup.Use(middleware.RequireAuth())

	authGroup := apiGroup.Group("/auth")
	{
		authGroup.POST("/logout", accountController.Logout)
		authGroup.GET("/me", accountController.Me)
	}
	courseGroup := apiGroup.Group("/course")
	{
		courseGroup.GET("/filter", courseController.GetCourseFilters)
		courseGroup.GET("/hot", courseController.ListHotCourses)
		courseGroup.GET("/", courseController.ListCourses)
		courseGroup.GET("/enrolled", courseEnrollmentController.ListMyEnrollments)
		courseGroup.DELETE("/enrollment/:enrollmentID", courseEnrollmentController.DeleteEnrollment)
		courseGroup.GET("/followed", courseController.ListFollowedCourses)
		courseGroup.GET("/ignored", courseController.ListIgnoredCourses)
		courseGroup.GET("/:courseID", courseController.GetCourse)
		courseGroup.GET("/:courseID/review/filter", reviewController.GetCourseReviewFilters)
		courseGroup.GET("/:courseID/review/trend", reviewController.GetCourseReviewTrend)
		courseGroup.GET("/:courseID/review", reviewController.ListCourseReviews)
		courseGroup.POST("/:courseID/enrollment", courseEnrollmentController.CreateEnrollment)
		courseGroup.POST("/:courseID/notification", courseController.SetNotificationLevel)
		courseGroup.PUT("/:courseID/moderator-remark", middleware.RequireAdmin(), courseController.UpdateModeratorRemark)
	}
	teacherGroup := apiGroup.Group("/teacher")
	{
		teacherGroup.GET("/filter", teacherController.GetTeacherFilters)
		teacherGroup.GET("/", teacherController.ListTeachers)
		teacherGroup.GET("/:teacherID", teacherController.GetTeacher)
		teacherGroup.GET("/:teacherID/course", teacherController.ListTeacherCourses)
	}
	reviewGroup := apiGroup.Group("/review")
	{
		reviewGroup.GET("", reviewController.ListReviews)
		reviewGroup.GET("/followed", reviewController.ListFollowedReviews)
		reviewGroup.GET("/:reviewID/revision", middleware.RequireAdmin(), reviewController.ListReviewRevisions)
		reviewGroup.GET("/:reviewID", reviewController.GetReview)
		reviewGroup.POST("/", reviewController.CreateReview)
		reviewGroup.POST("/:reviewID/vote", reviewController.VoteReview)
		reviewGroup.PUT("/:reviewID/moderator-remark", middleware.RequireAdmin(), reviewController.UpdateModeratorRemark)
		reviewGroup.PUT("/:reviewID", reviewController.UpdateReview)
		reviewGroup.DELETE("/:reviewID", reviewController.DeleteReview)
	}
	userGroup := apiGroup.Group("/user")
	{
		userGroup.GET("/settings", userSettingsController.GetMySettings)
		userGroup.PUT("/settings", userSettingsController.UpdateMySettings)
		userGroup.GET("/:userID/point", middleware.RequireSelfOrAdmin("userID"), pointController.GetUserPoints)
		userGroup.GET("/:userID/review", middleware.RequireSelfOrAdmin("userID"), reviewController.ListUserReviews)
	}
	apiKeyGroup := apiGroup.Group("/api-key")
	{
		apiKeyGroup.GET("/", apiKeyController.ListMyApiKeys)
		apiKeyGroup.POST("/", apiKeyController.CreateMyApiKey)
		apiKeyGroup.DELETE("/:apiKeyID", apiKeyController.DeleteMyApiKey)
	}
	pointGroup := apiGroup.Group("/point")
	{
		pointGroup.POST("/transfer/preview", pointController.PreviewTransfer)
		pointGroup.POST("/transfer", pointController.CreateTransfer)
	}
	siteStatsGroup := apiGroup.Group("/site-stat", middleware.RequireAdmin())
	{
		siteStatsGroup.GET("/daily/:date", siteStatsController.GetByDate)
		siteStatsGroup.GET("/daily", siteStatsController.ListDaily)
	}
	adminUserGroup := apiGroup.Group("/admin/user", middleware.RequireAdmin())
	{
		adminUserGroup.GET("/admin", adminUserController.ListAdmins)
		adminUserGroup.GET("/by-email", adminUserController.GetUserByEmail)
		adminUserGroup.PUT("/:userID/suspension", adminUserController.SuspendUser)
		adminUserGroup.DELETE("/:userID/suspension", adminUserController.ClearSuspension)
		adminUserGroup.PUT("/:userID/password", middleware.RequireSuperAdmin(), adminUserController.ResetPassword)
		adminUserGroup.PUT("/:userID/admin", middleware.RequireSuperAdmin(), adminUserController.GrantAdmin)
		adminUserGroup.DELETE("/:userID/admin", middleware.RequireSuperAdmin(), adminUserController.RevokeAdmin)
	}
	adminApiKeyGroup := apiGroup.Group("/admin/api-key", middleware.RequireAdmin())
	{
		adminApiKeyGroup.GET("/system", apiKeyController.ListSystemApiKeys)
		adminApiKeyGroup.POST("/system", apiKeyController.CreateSystemApiKey)
		adminApiKeyGroup.DELETE("/system/:apiKeyID", apiKeyController.DeleteSystemApiKey)
	}
	adminAuditGroup := apiGroup.Group("/admin/audit-log", middleware.RequireAdmin())
	{
		adminAuditGroup.GET("", auditLogController.ListAuditLogs)
	}
	announcementGroup := apiGroup.Group("/announcement")
	{
		announcementGroup.GET("/", announcementController.ListAnnouncements)
	}

	return g
}
