package main

import (
	"backend/internal/handler"
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/service"
	"backend/internal/storage"
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	//load env variables
	godotenv.Load()
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		panic("DB_DSN is not set")
	}

	// db connection
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // Mengatasi error prepared statement pada Supabase Pooler
	}), &gorm.Config{})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	//initialize repo, service, handler

	schoolRepo := repository.NewSchoolRepository(db)
	schoolService := service.NewSchoolService(schoolRepo)
	schoolHandler := handler.NewSchoolHandler(schoolService)

	academicYearRepo := repository.NewAcademicYearRepository(db)
	academicYearService := service.NewAcademicYearService(academicYearRepo, schoolService)
	academicYearHandler := handler.NewAcademicYearHandler(academicYearService, schoolService)

	termRepo := repository.NewTermRepository(db)
	termService := service.NewTermService(termRepo)
	termHandler := handler.NewTermHandler(termService)

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	schoolUserRepo := repository.NewSchoolUserRepository(db)
	schoolUserService := service.NewSchoolUserService(schoolUserRepo, schoolService)
	schoolUserHandler := handler.NewSchoolUserHandler(schoolUserService, schoolService)

	authService := service.NewAuthService(userRepo, schoolUserRepo)
	authHandler := handler.NewAuthHandler(authService)

	subjectRepo := repository.NewSubjectRepository(db)
	subjectService := service.NewSubjectService(subjectRepo, schoolService)
	subjectHandler := handler.NewSubjectHandler(subjectService, schoolService)

	rbacRepo := repository.NewRBACRepository(db)
	rbacService := service.NewRBACService(rbacRepo, userService, schoolRepo)
	rbacHandler := handler.NewRBACHandler(rbacService)

	classRepo := repository.NewClassRepository(db)
	classService := service.NewClassService(classRepo, schoolService)
	classHandler := handler.NewClassHandler(classService, schoolService)

	subjectClassRepo := repository.NewSubjectClassRepository(db)
	subjectClassService := service.NewSubjectClassService(subjectClassRepo)
	subjectClassHandler := handler.NewSubjectClassHandler(subjectClassService, classService)

	enrollmentRepo := repository.NewEnrollmentRepository(db)
	enrollmentService := service.NewEnrollmentService(enrollmentRepo, classRepo, schoolUserRepo)
	enrollmentHandler := handler.NewEnrollmentHandler(enrollmentService, classService)

	mediaRepo := repository.NewMediaRepository(db)
	storageProvider, err := buildStorageProvider()
	if err != nil {
		panic("failed to initialize storage provider: " + err.Error())
	}
	mediaService := service.NewMediaService(mediaRepo, storageProvider)
	mediaHandler := handler.NewMediaHandler(mediaService)

	attachmentRepo := repository.NewAttachmentRepository(db)
	attachmentService := service.NewAttachmentService(attachmentRepo)

	notificationRepo := repository.NewNotificationRepository(db)
	notificationService := service.NewNotificationService(notificationRepo)
	notificationHandler := handler.NewNotificationHandler(notificationService)

	materialRepo := repository.NewMaterialRepository(db)
	materialService := service.NewMaterialService(materialRepo, attachmentService, mediaRepo, storageProvider, notificationService, subjectClassRepo, enrollmentRepo)
	materialHandler := handler.NewMaterialHandler(materialService, subjectClassService)

	feedRepo := repository.NewFeedRepository(db)
	feedService := service.NewFeedService(feedRepo, attachmentService, notificationService, enrollmentRepo, classRepo, subjectClassRepo)
	commentRepo := repository.NewCommentRepository(db)
	contentOwnerRepo := repository.NewContentOwnerRepository(db)
	commentService := service.NewCommentService(commentRepo, contentOwnerRepo, notificationService)
	feedHandler := handler.NewFeedHandler(feedService, commentService, classService)
	commentHandler := handler.NewCommentHandler(commentService)

	assignmentRepo := repository.NewAssignmentRepository(db)
	assignmentService := service.NewAssignmentService(assignmentRepo, attachmentService, notificationService, enrollmentRepo)
	assignmentHandler := handler.NewAssignmentHandler(assignmentService, schoolService, subjectClassService)

	gradeHandler := handler.NewGradeHandler(service.NewGradeService(
		repository.NewAssessmentWeightRepository(db),
		repository.NewGradeRepository(db),
		subjectRepo,
		classRepo,
		userRepo,
	))

	logRepo := repository.NewLogRepository(db)
	logService := service.NewLogService(logRepo)
	logHandler := handler.NewLogHandler(logService)

	dashboardRepo := repository.NewDashboardRepository(db)
	dashboardService := service.NewDashboardService(dashboardRepo)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)

	// Initialize RBAC middleware
	middleware.InitRBAC(rbacRepo)

	//router setup
	r := gin.Default()
	r.Use(corsMiddleware())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	api := r.Group("/api")
	{
		//public routes
		api.POST("/login", authHandler.Login)
		api.POST("/register", authHandler.Register)

		//protected routes
		api.Use(middleware.AuthRequired())

		schoolAPI := api.Group("/schools")
		{
			schoolAPI.POST("", middleware.RequireRole(schoolService, "super_admin"), schoolHandler.CreateSchool)
			schoolAPI.GET("", middleware.RequireRole(schoolService, "super_admin"), schoolHandler.GetSchools)
			schoolAPI.GET("/summary", middleware.RequireRole(schoolService, "super_admin"), schoolHandler.GetSchoolSummary)
			schoolAPI.GET("/check-code/:schoolCode", schoolHandler.CheckCodeAvailability)
			schoolAPI.GET("/:schoolCode", middleware.RequireSchoolMember(schoolService), schoolHandler.GetSchoolByCode)
			schoolAPI.PATCH("/:schoolCode", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin", "super_admin"), schoolHandler.UpdateSchool)
			schoolAPI.PATCH("/restore/:schoolCode", middleware.RequireRole(schoolService, "super_admin"), schoolHandler.RestoreDeletedSchool)
			schoolAPI.DELETE("/:schoolCode", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin", "super_admin"), schoolHandler.DeleteSchool)
			schoolAPI.DELETE("/permanent/:schoolCode", middleware.RequireRole(schoolService, "super_admin"), schoolHandler.HardDeleteSchool)
		}

		academicYearAPI := api.Group("/academic-years")
		{
			academicYearAPI.POST("", middleware.RequireRole(schoolService, "admin", "super_admin"), academicYearHandler.Create)
			academicYearAPI.GET("", academicYearHandler.FindAll)
			academicYearAPI.GET("/:id", academicYearHandler.GetByID)
			academicYearAPI.GET("/school/:schoolCode", middleware.RequireSchoolMember(schoolService), academicYearHandler.GetBySchool)
			academicYearAPI.PATCH("/:id", middleware.RequireRole(schoolService, "admin", "super_admin"), academicYearHandler.Update)
			academicYearAPI.PATCH("/activate/:id", middleware.RequireRole(schoolService, "admin", "super_admin"), academicYearHandler.Activate)
			academicYearAPI.PATCH("/deactivate/:id", middleware.RequireRole(schoolService, "admin", "super_admin"), academicYearHandler.Deactivate)
			academicYearAPI.DELETE("/:id", middleware.RequireRole(schoolService, "admin", "super_admin"), academicYearHandler.Delete)
		}

		termAPI := api.Group("/terms")
		{
			termAPI.POST("", middleware.RequireRole(schoolService, "admin", "super_admin"), termHandler.Create)
			termAPI.GET("", termHandler.FindAll)
			termAPI.GET("/:id", termHandler.GetByID)
			termAPI.GET("/academic-year/:academicYearId", termHandler.GetByAcademicYear)
			termAPI.PATCH("/:id", middleware.RequireRole(schoolService, "admin", "super_admin"), termHandler.Update)
			termAPI.PATCH("/activate/:id", middleware.RequireRole(schoolService, "admin", "super_admin"), termHandler.Activate)
			termAPI.PATCH("/deactivate/:id", middleware.RequireRole(schoolService, "admin", "super_admin"), termHandler.Deactivate)
			termAPI.DELETE("/:id", middleware.RequireRole(schoolService, "admin", "super_admin"), termHandler.Delete)
		}

		userAPI := api.Group("/users")
		{
			userAPI.POST("", middleware.RequireSystemSuperAdmin(schoolService), userHandler.Create)
			userAPI.GET("", middleware.RequireRole(schoolService, "admin", "super_admin"), userHandler.FindAll)
			userAPI.GET("/:id", middleware.RequireSystemSuperAdmin(schoolService), userHandler.GetByID)
			userAPI.PATCH("/:id", middleware.RequireSystemSuperAdmin(schoolService), userHandler.Update)
			userAPI.PATCH("/change-password/:id", middleware.RequireSystemSuperAdmin(schoolService), userHandler.ChangePassword)
			userAPI.DELETE("/:id", middleware.RequireSystemSuperAdmin(schoolService), userHandler.Delete)
		}

		schoolUserAPI := api.Group("/school-users")
		{
			schoolUserAPI.POST("/enroll", middleware.RequireRole(schoolService, "admin", "super_admin"), schoolUserHandler.Enroll)
			schoolUserAPI.GET("/school/:schoolCode", middleware.RequireSchoolMember(schoolService), schoolUserHandler.GetMembersBySchool)
			schoolUserAPI.GET("/user/:userId", schoolUserHandler.GetSchoolsByUser)
			schoolUserAPI.DELETE("/:userId", middleware.RequireRole(schoolService, "admin", "super_admin"), schoolUserHandler.Unenroll)
		}

		subjectAPI := api.Group("/subjects")
		{
			subjectAPI.POST("", middleware.RequireRole(schoolService, "admin", "super_admin"), subjectHandler.Create)
			subjectAPI.GET("", subjectHandler.FindAll)
			subjectAPI.GET("/:id", subjectHandler.GetByID)
			subjectAPI.GET("/school/:schoolCode", middleware.RequireSchoolMember(schoolService), subjectHandler.GetBySchool)
			subjectAPI.GET("/school/:schoolCode/:subjectCode", middleware.RequireSchoolMember(schoolService), subjectHandler.GetByCode)
			subjectAPI.PATCH("/:id", middleware.RequireRole(schoolService, "admin", "super_admin"), subjectHandler.Update)
			subjectAPI.DELETE("/:id", middleware.RequireRole(schoolService, "admin", "super_admin"), subjectHandler.Delete)
		}

		rbacAPI := api.Group("/rbac")
		{
			// Roles
			rbacAPI.POST("/roles", middleware.RequireRole(schoolService, "super_admin"), rbacHandler.CreateRole)
			rbacAPI.GET("/roles", rbacHandler.GetAllRoles)
			rbacAPI.GET("/roles/:id", rbacHandler.GetRoleByID)
			rbacAPI.PATCH("/roles/:id", middleware.RequireRole(schoolService, "super_admin"), rbacHandler.UpdateRole)
			rbacAPI.DELETE("/roles/:id", middleware.RequireRole(schoolService, "super_admin"), rbacHandler.DeleteRole)

			// User Roles (Assignments)
			rbacAPI.POST("/user-roles", middleware.RequireRole(schoolService, "admin", "super_admin"), rbacHandler.AssignRole)
			rbacAPI.DELETE("/user-roles", middleware.RequireRole(schoolService, "admin", "super_admin"), rbacHandler.RemoveRole)
			rbacAPI.GET("/user-roles/:schoolUserId", rbacHandler.GetUserRoles)
			rbacAPI.PATCH("/user-roles/:schoolUserId", middleware.RequireRole(schoolService, "admin", "super_admin"), rbacHandler.UpdateUserRoles)

			// Super Admin
			rbacAPI.POST("/super-admin", middleware.RequireRole(schoolService, "super_admin"), rbacHandler.CreateSuperAdmin)
		}

		classAPI := api.Group("/classes")
		{
			classAPI.POST("", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin", "teacher"), classHandler.Create)
			classAPI.GET("", classHandler.FindAll)
			classAPI.GET("/:id", classHandler.GetByID)
			classAPI.PATCH("/:id", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin", "teacher"), classHandler.Update)
			classAPI.DELETE("/:id", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin"), classHandler.Delete)
		}

		subjectClassAPI := api.Group("/subject-classes")
		{
			subjectClassAPI.POST("/assign", middleware.RequireRole(schoolService, "admin", "teacher"), subjectClassHandler.Assign)
			subjectClassAPI.GET("/my-teaching", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "teacher"), subjectClassHandler.GetMyTeaching)
			subjectClassAPI.GET("/class/:classId", middleware.RequireSchoolMember(schoolService), subjectClassHandler.GetByClass)
			subjectClassAPI.GET("/:id", middleware.RequireSchoolMember(schoolService), subjectClassHandler.GetByID)
			subjectClassAPI.PATCH("/:id", middleware.RequireRole(schoolService, "admin", "teacher"), subjectClassHandler.Update)
			subjectClassAPI.DELETE("/:id", middleware.RequireRole(schoolService, "admin"), subjectClassHandler.Unassign)
		}

		enrollmentAPI := api.Group("/enrollments")
		{
			enrollmentAPI.POST("", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin"), enrollmentHandler.Enroll)
			enrollmentAPI.GET("/class/:classId", middleware.RequireSchoolMember(schoolService), enrollmentHandler.GetByClass)
			enrollmentAPI.GET("/member/:schoolUserId", middleware.RequireSchoolMember(schoolService), enrollmentHandler.GetByMember)
			enrollmentAPI.GET("/:id", middleware.RequireSchoolMember(schoolService), enrollmentHandler.GetByID)
			enrollmentAPI.PATCH("/:id", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin"), enrollmentHandler.Update)
			enrollmentAPI.DELETE("/:id", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin"), enrollmentHandler.Unenroll)
		}

		mediaAPI := api.Group("/medias")
		{
			mediaAPI.POST("/upload", mediaHandler.Upload)
			mediaAPI.POST("/metadata", mediaHandler.RecordMetadata)
			mediaAPI.GET("/:id", mediaHandler.GetByID)
			mediaAPI.DELETE("/:id", mediaHandler.Delete)
		}

		materialAPI := api.Group("/materials")
		{
			materialAPI.POST("", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "teacher"), materialHandler.Create)
			materialAPI.GET("", materialHandler.FindAll)
			materialAPI.GET("/:id", materialHandler.GetByID)
			materialAPI.PATCH("/:id", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "teacher"), materialHandler.Update)
			materialAPI.DELETE("/:id", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "teacher", "admin"), materialHandler.Delete)
			materialAPI.POST("/progress", materialHandler.UpdateProgress)
		}

		feedAPI := api.Group("/feeds")
		{
			feedAPI.POST("", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "teacher", "admin"), feedHandler.Create)
			feedAPI.GET("/class/:classId", feedHandler.GetByClass)
			feedAPI.GET("/:id", feedHandler.GetByID)
			feedAPI.PATCH("/:id", feedHandler.Update)
			feedAPI.DELETE("/:id", feedHandler.Delete)
		}

		commentAPI := api.Group("/comments")
		{
			commentAPI.POST("", commentHandler.Create)
			commentAPI.GET("", commentHandler.GetBySource)
			commentAPI.GET("/:id", commentHandler.GetByID)
			commentAPI.PATCH("/:id", commentHandler.Update)
			commentAPI.DELETE("/:id", commentHandler.Delete)
		}

		assignmentAPI := api.Group("/assignments")
		{
			// Categories
			assignmentAPI.POST("/categories", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin"), assignmentHandler.CreateCategory)
			assignmentAPI.GET("/categories/school/:schoolCode", middleware.RequireSchoolMember(schoolService), assignmentHandler.GetCategoriesBySchool)

			// Assignments
			assignmentAPI.POST("", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "teacher"), assignmentHandler.CreateAssignment)
			assignmentAPI.GET("/subject-class/submissions/:subjectClassId", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "teacher"), assignmentHandler.GetSubjectClassSubmissions)
			assignmentAPI.GET("/subject-class/:subjectClassId", assignmentHandler.GetBySubjectClass)
			assignmentAPI.GET("/status/:id", assignmentHandler.GetAssignmentStatus)
			assignmentAPI.GET("/my-submission/:assignmentId", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "student"), assignmentHandler.GetMySubmissionByAssignment)
			assignmentAPI.GET("/:assignmentId", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "teacher"), assignmentHandler.GetSubmissionsByAssignment)
			assignmentAPI.PATCH("/:id", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "teacher"), assignmentHandler.UpdateAssignment)
			assignmentAPI.DELETE("/:id", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "teacher", "admin"), assignmentHandler.DeleteAssignment)

			// Submissions
			assignmentAPI.POST("/submit/:assignmentId", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "student"), assignmentHandler.Submit)
			assignmentAPI.GET("/submit/:submissionId", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "teacher"), assignmentHandler.GetSubmissionByID)
			assignmentAPI.PATCH("/submit/:submissionId", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "student"), assignmentHandler.UpdateSubmission)
			assignmentAPI.DELETE("/submit/:submissionId", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "student"), assignmentHandler.DeleteSubmission)

			// Assessments
			assignmentAPI.POST("/assess/:submissionId", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "teacher"), assignmentHandler.Assess)
			assignmentAPI.PATCH("/assess/:submissionId", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "teacher"), assignmentHandler.UpdateAssessment)
			assignmentAPI.DELETE("/assess/:submissionId", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "teacher"), assignmentHandler.DeleteAssessment)
		}

		gradeAPI := api.Group("/grades")
		{
			gradeAPI.POST("/weights", middleware.RequireRole(schoolService, "admin", "teacher"), gradeHandler.ConfigureWeights)
			gradeAPI.GET("/weights/subject/:subjectId", middleware.RequireSchoolMember(schoolService), gradeHandler.GetWeightsBySubject)
			gradeAPI.GET("/class/:classId/subject/:subjectId", middleware.RequireRole(schoolService, "teacher", "admin"), gradeHandler.GetClassGradeReport)
			gradeAPI.GET("/my-grades/:classId", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "student"), gradeHandler.GetMyGradebookByClass)
		}

		notificationAPI := api.Group("/notifications")
		{
			notificationAPI.GET("", notificationHandler.GetNotifications)
			notificationAPI.GET("/unread-count", notificationHandler.GetUnreadCount)
			notificationAPI.PATCH("/read/:id", notificationHandler.MarkAsRead)
			notificationAPI.PATCH("/read-all", notificationHandler.MarkAllAsRead)
			notificationAPI.DELETE("/:id", notificationHandler.Delete)
		}

		logAPI := api.Group("/logs")
		{
			logAPI.GET("/school/:schoolId", logHandler.GetBySchool)
		}

		dashboardAPI := api.Group("/dashboard")
		{
			dashboardAPI.GET("/student/:userId", dashboardHandler.GetStudentDashboard)
			dashboardAPI.GET("/teacher/:schoolUserId", dashboardHandler.GetTeacherDashboard)
			dashboardAPI.GET("/admin/:schoolId", dashboardHandler.GetAdminDashboard)
		}
	}

	//run server
	r.Run(":" + serverPort())
}

func corsMiddleware() gin.HandlerFunc {
	allowedOrigins := parseAllowedOrigins(os.Getenv("CORS_ALLOWED_ORIGINS"))
	allowedOriginSet := make(map[string]bool, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowedOriginSet[origin] = true
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		isAllowed := allowedOriginSet[origin]

		if isAllowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, SchoolId, schoolid")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Max-Age", "86400")
		}

		if c.Request.Method == "OPTIONS" {
			if !isAllowed {
				c.AbortWithStatus(403)
			} else {
				c.AbortWithStatus(204)
			}
			return
		}

		if !isAllowed && origin != "" {
			c.AbortWithStatus(403)
			return
		}

		c.Next()
	}
}

func parseAllowedOrigins(raw string) []string {
	defaultOrigins := []string{}

	if strings.TrimSpace(raw) == "" {
		return defaultOrigins
	}

	// Use a map to avoid duplicates
	seen := make(map[string]bool)
	origins := make([]string, 0)

	// Add default origins first
	for _, origin := range defaultOrigins {
		if !seen[origin] {
			seen[origin] = true
			origins = append(origins, origin)
		}
	}

	// Parse and add environment-configured origins
	parts := strings.Split(raw, ",")
	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin != "" && !seen[origin] {
			seen[origin] = true
			origins = append(origins, origin)
		}
	}

	return origins
}

func serverPort() string {
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		return "8080"
	}
	return port
}

func buildStorageProvider() (storage.Provider, error) {
	provider := strings.ToLower(strings.TrimSpace(os.Getenv("STORAGE_PROVIDER")))
	if provider == "" || provider == "disabled" {
		return storage.NewDisabledStorage(), nil
	}

	if provider == "supabase" {
		return storage.NewSupabaseStorage(
			os.Getenv("SUPABASE_URL"),
			os.Getenv("SUPABASE_SERVICE_KEY"),
			os.Getenv("SUPABASE_BUCKET"),
			10*1024*1024,
		)
	}

	return nil, fmt.Errorf("unsupported storage provider: %s", provider)
}
