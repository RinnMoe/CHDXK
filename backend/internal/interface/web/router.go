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
	"jcourse/internal/domain/auth"
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
	g.Use(middleware.SiteCopy(container.SystemSettingsQuery))
	g.Use(middleware.ResolveCurrentUser(container.AuthResolution))
	g.Use(middleware.UserIDRateLimit())
	g.Use(middleware.CSRF())

	reviewController := controller.NewReviewController(container.ReviewQuery, container.ReviewCommand)
	courseController := controller.NewCourseController(container.CourseQuery, container.CourseCommand)
	enrollmentSyncController := controller.NewCourseSyncController(container.EnrollmentSync)
	courseSyncController := controller.NewCourseSyncController(container.CourseSync)
	courseAdminController := controller.NewCourseAdminController(container.CourseManagement)
	courseEnrollmentController := controller.NewCourseEnrollmentController(container.CourseEnrollmentQuery, container.CourseCommand)
	teacherController := controller.NewTeacherController(container.TeacherQuery, container.CourseQuery)
	pointController := controller.NewPointController(container.PointQuery)
	accountController := controller.NewAccountController(container.AccountQuery, container.ResHubOIDC, container.Accounts, container.AuthResolution)
	apiKeyController := controller.NewApiKeyController(container.ApiKeyQuery, container.ApiKeyCommand)
	systemSettingsController := controller.NewSystemSettingsController(container.SystemSettingsQuery, container.SystemSettingsCommand)
	siteStatsController := controller.NewSiteStatsController(container.SiteStatsQuery)
	announcementController := controller.NewAnnouncementController(container.AnnouncementQuery, container.AnnouncementService)
	auditLogController := controller.NewAuditLogController(container.AuditLogQuery)

	apiGroup := g.Group("/api", middleware.NoStore())
	apiGroup.GET("/site-copy", systemSettingsController.Copy)
	apiGroup.GET("/site-pages", systemSettingsController.Pages)
	apiGroup.GET("/site-theme", systemSettingsController.Theme)
	apiGroup.GET("/site.webmanifest", systemSettingsController.Manifest)
	apiGroup.GET("/system-settings", systemSettingsController.List)
	publicAuthGroup := apiGroup.Group("/auth")
	{
		publicAuthGroup.GET("/csrf", func(c *gin.Context) { c.Status(http.StatusNoContent) })
		publicAuthGroup.POST("/login", accountController.PasswordLogin)
		publicAuthGroup.POST("/register", accountController.Register)
		publicAuthGroup.GET("/options", accountController.Options)
		publicAuthGroup.GET("/accounts", accountController.Accounts)
		publicAuthGroup.POST("/accounts/select", accountController.SelectAccount)
		publicAuthGroup.POST("/native/exchange", accountController.NativeExchange)
		publicAuthGroup.POST("/logout", accountController.Logout)
		publicAuthGroup.GET("/oidc/start", accountController.OIDCStart)
		publicAuthGroup.GET("/oidc/callback", accountController.OIDCCallback)
	}

	apiGroup.Use(middleware.RequireAuth(), middleware.RequirePasswordChanged())

	authGroup := apiGroup.Group("/auth")
	{
		authGroup.GET("/me", accountController.Me)
		authGroup.POST("/password", accountController.ChangePassword)
		authGroup.PUT("/profile", accountController.Profile)
		authGroup.DELETE("/binding", accountController.Unbind)
	}
	courseGroup := apiGroup.Group("/course")
	{
		courseGroup.GET("/filter", courseController.GetCourseFilters)
		courseGroup.GET("/discover", courseController.DiscoverCourses)
		courseGroup.GET("/review-prompts", courseController.ReviewPrompts)
		courseGroup.GET("/hot", courseController.ListHotCourses)
		courseGroup.GET("/", courseController.ListCourses)
		courseGroup.POST("/enrollment-sync", enrollmentSyncController.Start)
		courseGroup.GET("/enrollment-sync/:jobID", enrollmentSyncController.Status)
		courseGroup.POST("/enrollment-sync/:jobID/run", enrollmentSyncController.Sync)
		courseGroup.POST("/enrollment-sync/:jobID/cancel", enrollmentSyncController.Cancel)
		courseGroup.GET("/enrolled", courseEnrollmentController.ListMyEnrollments)
		courseGroup.POST("/enrollment", courseEnrollmentController.SaveEnrollment)
		courseGroup.PUT("/enrollment/:enrollmentID", courseEnrollmentController.SaveEnrollment)
		courseGroup.DELETE("/enrollment/:enrollmentID", courseEnrollmentController.DeleteEnrollment)
		courseGroup.GET("/followed", courseController.ListFollowedCourses)
		courseGroup.GET("/ignored", courseController.ListIgnoredCourses)
		courseGroup.GET("/:courseID", courseController.GetCourse)
		courseGroup.GET("/:courseID/review/filter", reviewController.GetCourseReviewFilters)
		courseGroup.GET("/:courseID/review/trend", reviewController.GetCourseReviewTrend)
		courseGroup.GET("/:courseID/review", reviewController.ListCourseReviews)
		courseGroup.POST("/:courseID/notification", courseController.SetNotificationLevel)
		courseGroup.PUT("/:courseID/moderator-remark", middleware.RequireAdmin(), courseController.UpdateModeratorRemark)
	}
	teacherGroup := apiGroup.Group("/teacher")
	{
		teacherGroup.GET("/filter", teacherController.GetTeacherFilters)
		teacherGroup.GET("/hot", controller.TeacherHotHandler(container.TeacherHotQuery))
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
		userGroup.GET("/:userID/point", middleware.RequireSelfOrAdmin("userID"), pointController.GetUserPoints)
		userGroup.GET("/:userID/review", middleware.RequireSelfOrAdmin("userID"), reviewController.ListUserReviews)
	}
	siteStatsGroup := apiGroup.Group("/site-stat", middleware.RequireAdmin())
	{
		siteStatsGroup.GET("/daily/:date", siteStatsController.GetByDate)
		siteStatsGroup.GET("/daily", siteStatsController.ListDaily)
	}
	adminApiKeyGroup := apiGroup.Group("/admin/api-key", middleware.RequirePermission(auth.PermissionMaintenance))
	{
		adminApiKeyGroup.GET("/system", apiKeyController.ListSystemApiKeys)
		adminApiKeyGroup.POST("/system", apiKeyController.CreateSystemApiKey)
		adminApiKeyGroup.DELETE("/system/:apiKeyID", apiKeyController.DeleteSystemApiKey)
	}
	adminSystemSettingsGroup := apiGroup.Group("/admin/system-settings", middleware.RequirePermission(auth.PermissionSettings))
	{
		adminSystemSettingsGroup.GET("", systemSettingsController.ListAdmin)
		adminSystemSettingsGroup.PUT("/:key", systemSettingsController.Update)
	}
	adminAuditGroup := apiGroup.Group("/admin/audit-log", middleware.RequirePermission(auth.PermissionMaintenance))
	{
		adminAuditGroup.GET("", auditLogController.ListAuditLogs)
	}
	adminAnnouncementGroup := apiGroup.Group("/admin/announcement", middleware.RequireAdmin())
	{
		adminAnnouncementGroup.GET("", announcementController.ListAdminAnnouncements)
		adminAnnouncementGroup.POST("", announcementController.CreateAnnouncement)
		adminAnnouncementGroup.PUT("/:announcementID", announcementController.UpdateAnnouncement)
		adminAnnouncementGroup.DELETE("/:announcementID", announcementController.DeleteAnnouncement)
	}
	syncGroup := apiGroup.Group("/admin/course-sync", middleware.RequirePermission(auth.PermissionCourseData))
	syncGroup.POST("", courseSyncController.Start)
	syncGroup.GET("/:jobID", courseSyncController.Status)
	syncGroup.POST("/:jobID/run", courseSyncController.Sync)
	syncGroup.POST("/:jobID/cancel", courseSyncController.Cancel)
	adminCourseGroup := apiGroup.Group("/admin/course", middleware.RequirePermission(auth.PermissionCourseData))
	profileController := controller.NewTeacherProfileController(container.TeacherProfiles)
	profileGroup := apiGroup.Group("/admin/teacher-profiles", middleware.RequirePermission(auth.PermissionCourseData))
	profileGroup.GET("", profileController.Status)
	profileGroup.POST("/sync", profileController.Start)
	profileGroup.POST("/cancel", profileController.Cancel)
	profileGroup.GET("/matches", profileController.Matches)
	profileGroup.GET("/teachers", profileController.Teachers)
	profileGroup.PUT("/:teacherID/details", profileController.Edit)
	profileGroup.POST("/:teacherID/preview", profileController.Preview)
	profileGroup.PUT("/:teacherID", profileController.Bind)
	{
		adminCourseGroup.GET("/semester", courseAdminController.ListSemesters)
		adminCourseGroup.GET("", courseAdminController.List)
		adminCourseGroup.POST("", courseAdminController.Create)
		adminCourseGroup.PUT("/:offeringID", courseAdminController.Update)
		adminCourseGroup.DELETE("/:offeringID", courseAdminController.Delete)
	}
	announcementGroup := apiGroup.Group("/announcement")
	{
		announcementGroup.GET("/", announcementController.ListAnnouncements)
	}

	userAdmin := controller.NewUserManagementController(container.Accounts)
	users := apiGroup.Group("/admin/users", middleware.RequireAdmin())
	users.GET("", userAdmin.List)
	users.POST("", middleware.RequirePermission(auth.PermissionUsers), userAdmin.Create)
	users.PUT("/:userID", middleware.RequirePermission(auth.PermissionUsers), userAdmin.Update)
	users.DELETE("/:userID", middleware.RequirePermission(auth.PermissionUsers), userAdmin.Delete)
	users.POST("/:userID/password", middleware.RequirePermission(auth.PermissionUsers), userAdmin.ResetPassword)
	users.PUT("/:userID/role", middleware.RequirePermission(auth.PermissionRoles), userAdmin.Role)
	users.POST("/:userID/moderation", userAdmin.Moderate)
	users.POST("/:userID/points", middleware.RequirePermission(auth.PermissionPoints), userAdmin.Points)
	return g
}
