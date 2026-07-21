package main

import (
	"backend/internal/handler"
	"backend/internal/middleware"
	"backend/internal/realtime"
	"backend/internal/repository"
	"backend/internal/service"
	"backend/internal/storage"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

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
	emailService := service.NewEmailServiceFromEnv()
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)
	emailVerificationRepo := repository.NewEmailVerificationRepository(db)
	emailVerificationService := service.NewEmailVerificationService(emailVerificationRepo, userRepo, emailService)
	emailVerificationHandler := handler.NewEmailVerificationHandler(emailVerificationService)
	invitationRepo := repository.NewInvitationRepository(db)
	invitationService := service.NewInvitationService(invitationRepo)
	invitationHandler := handler.NewInvitationHandler(invitationService)

	academicYearRepo := repository.NewAcademicYearRepository(db)
	academicYearService := service.NewAcademicYearService(academicYearRepo, schoolService)
	academicYearHandler := handler.NewAcademicYearHandler(academicYearService, schoolService)

	termRepo := repository.NewTermRepository(db)
	termService := service.NewTermService(termRepo)
	termHandler := handler.NewTermHandler(termService)

	rbacRepo := repository.NewRBACRepository(db)
	rbacService := service.NewRBACService(rbacRepo, userService, schoolRepo)

	schoolUserRepo := repository.NewSchoolUserRepository(db)
	schoolUserService := service.NewSchoolUserService(schoolUserRepo, schoolService)
	schoolUserHandler := handler.NewSchoolUserHandler(schoolUserService, schoolService, rbacService)
	adminSchoolMemberImportService := service.NewAdminSchoolMemberImportService(db, emailService)
	adminSchoolMemberImportHandler := handler.NewAdminSchoolMemberImportHandler(adminSchoolMemberImportService)
	schoolMemberInvitationRepo := repository.NewSchoolMemberInvitationRepository(db)
	schoolMemberInvitationService := service.NewSchoolMemberInvitationService(schoolMemberInvitationRepo, emailService)
	schoolMemberInvitationHandler := handler.NewSchoolMemberInvitationHandler(schoolMemberInvitationService)

	authService := service.NewAuthService(userRepo, schoolUserRepo, emailVerificationService)
	authHandler := handler.NewAuthHandler(authService)

	subjectRepo := repository.NewSubjectRepository(db)
	subjectService := service.NewSubjectService(subjectRepo, schoolService)
	subjectHandler := handler.NewSubjectHandler(subjectService, schoolService)

	rbacHandler := handler.NewRBACHandler(rbacService, schoolUserService)
	superAdminBootstrapService := service.NewSuperAdminBootstrapService(db, schoolRepo, schoolUserRepo, rbacRepo)
	superAdminBootstrapHandler := handler.NewSuperAdminBootstrapHandler(superAdminBootstrapService)

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

	sidebarHub := realtime.NewSidebarHub()
	go sidebarHub.Run()
	sidebarStreamHandler := realtime.NewSidebarStreamHandler(sidebarHub, authService)

	notificationService := service.NewNotificationService(notificationRepo, sidebarHub)
	notificationHandler := handler.NewNotificationHandler(notificationService)

	materialRepo := repository.NewMaterialRepository(db)
	materialService := service.NewMaterialService(materialRepo, attachmentService, mediaRepo, storageProvider, notificationService, subjectClassRepo, enrollmentRepo)
	materialSummaryService := service.NewMaterialSummaryService(attachmentService, storageProvider, service.NewPDFTextExtractor(), service.NewMaterialAIServiceFromEnv())
	materialHandler := handler.NewMaterialHandler(materialService, materialSummaryService, subjectClassService)
	assignmentRepo := repository.NewAssignmentRepository(db)

	studentNoteRepo := repository.NewStudentNoteRepository(db)
	studentNoteService := service.NewStudentNoteService(studentNoteRepo, materialRepo, subjectClassRepo)
	studentNoteHandler := handler.NewStudentNoteHandler(studentNoteService)

	feedRepo := repository.NewFeedRepository(db)
	feedService := service.NewFeedService(feedRepo, attachmentService, notificationService, enrollmentRepo, classRepo, subjectClassRepo, sidebarHub)
	commentRepo := repository.NewCommentRepository(db)
	contentOwnerRepo := repository.NewContentOwnerRepository(db)
	commentService := service.NewCommentService(commentRepo, contentOwnerRepo, notificationService, feedRepo, materialRepo, assignmentRepo, enrollmentRepo, subjectClassRepo)
	feedHandler := handler.NewFeedHandler(feedService, commentService, classService, notificationService)
	commentHandler := handler.NewCommentHandler(commentService)

	chatRepo := repository.NewChatRepository(db)
	chatService := service.NewChatService(chatRepo, mediaRepo)
	chatHub := realtime.NewHub()
	go chatHub.Run()
	chatHandler := handler.NewChatHandler(chatService, chatHub)
	chatWebSocketHandler := realtime.NewWebSocketHandler(chatHub, chatService)

	assignmentService := service.NewAssignmentService(assignmentRepo, attachmentService, mediaRepo, notificationService, enrollmentRepo, db)
	assignmentHandler := handler.NewAssignmentHandler(assignmentService, schoolService, subjectClassService)

	gradeHandler := handler.NewGradeHandler(service.NewGradeService(
		repository.NewAssessmentWeightRepository(db),
		repository.NewGradeRepository(db),
		subjectRepo,
		classRepo,
		userRepo,
	), subjectClassService)

	logRepo := repository.NewLogRepository(db)
	logService := service.NewLogService(logRepo)
	logHandler := handler.NewLogHandler(logService)

	dashboardRepo := repository.NewDashboardRepository(db)
	dashboardService := service.NewDashboardService(dashboardRepo)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)

	activityRepo := repository.NewActivityRepository(db)
	activityService := service.NewActivityService(activityRepo)
	activityHandler := handler.NewActivityHandler(activityService)

	// Initialize RBAC middleware
	middleware.InitRBAC(rbacRepo)

	// Shared per-tenant rate limiter: 20 req/s sustained, burst of 40, keyed
	// by SchoolId header
	rateLimiterStore := middleware.NewInMemoryRateLimiterStore(20, 40, 10*time.Minute)

	requestLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	//router setup
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.StructuredLogger(requestLogger))
	r.Use(corsMiddleware())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	api := r.Group("/api")
	{
		//public routes
		// Rate-limited by IP (no school context exists yet at this point)
		api.POST("/login", middleware.RateLimitPerTenant(rateLimiterStore), authHandler.Login)
		api.POST("/register", middleware.RateLimitPerTenant(rateLimiterStore), authHandler.Register)
		// Not rate-limited: token-gated (guessing is already infeasible given token length/hashing)
		api.GET("/invitations/:token", invitationHandler.GetMetadata)
		api.POST("/invitations/:token/accept", invitationHandler.Accept)
		api.POST("/verify-email", emailVerificationHandler.Verify)
		// Not rate-limited: long-lived SSE/WebSocket connections
		api.GET("/events/sidebar", sidebarStreamHandler.Stream)
		api.GET("/ws/chat", chatWebSocketHandler.Chat)

		//protected routes
		api.Use(middleware.AuthRequired())
		api.Use(middleware.RateLimitPerTenant(rateLimiterStore))

		meAPI := api.Group("/me")
		{
			meAPI.GET("/context", authHandler.GetContext)
			meAPI.POST("/resend-verification", emailVerificationHandler.Resend)
		}

		schoolAPI := api.Group("/schools")
		{
			// Self-service Create School (Phase 1): any authenticated user with a
			// verified email may create a school and becomes its Admin.
			schoolAPI.POST("", middleware.RequireVerifiedUser(userRepo), schoolHandler.CreateSchool)
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
			academicYearAPI.GET("", middleware.RequireSchoolMember(schoolService), academicYearHandler.FindAll)
			academicYearAPI.GET("/:id", middleware.RequireSchoolMember(schoolService), academicYearHandler.GetByID)
			academicYearAPI.GET("/school/:schoolCode", middleware.RequireSchoolMember(schoolService), academicYearHandler.GetBySchool)
			academicYearAPI.PATCH("/:id", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin", "super_admin"), academicYearHandler.Update)
			academicYearAPI.PATCH("/activate/:id", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin", "super_admin"), academicYearHandler.Activate)
			academicYearAPI.PATCH("/deactivate/:id", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin", "super_admin"), academicYearHandler.Deactivate)
			academicYearAPI.DELETE("/:id", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin", "super_admin"), academicYearHandler.Delete)
		}

		termAPI := api.Group("/terms")
		{
			termAPI.POST("", middleware.RequireRole(schoolService, "admin", "super_admin"), termHandler.Create)
			termAPI.GET("", middleware.RequireSchoolMember(schoolService), termHandler.FindAll)
			termAPI.GET("/:id", middleware.RequireSchoolMember(schoolService), termHandler.GetByID)
			termAPI.GET("/academic-year/:academicYearId", middleware.RequireSchoolMember(schoolService), termHandler.GetByAcademicYear)
			termAPI.PATCH("/:id", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin", "super_admin"), termHandler.Update)
			termAPI.PATCH("/activate/:id", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin", "super_admin"), termHandler.Activate)
			termAPI.PATCH("/deactivate/:id", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin", "super_admin"), termHandler.Deactivate)
			termAPI.DELETE("/:id", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin", "super_admin"), termHandler.Delete)
		}

		userAPI := api.Group("/users")
		{
			userAPI.POST("", middleware.RequireSystemSuperAdmin(schoolService), userHandler.Create)
			userAPI.GET("", middleware.RequireSystemSuperAdmin(schoolService), userHandler.FindAll)
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

		adminSchoolMemberImportAPI := api.Group("/admin/school-members/import")
		adminSchoolMemberImportAPI.Use(middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin"))
		{
			adminSchoolMemberImportAPI.POST("/preview", adminSchoolMemberImportHandler.Preview)
			adminSchoolMemberImportAPI.POST("/commit", adminSchoolMemberImportHandler.Commit)
		}

		adminSchoolMemberAPI := api.Group("/admin/school-members")
		adminSchoolMemberAPI.Use(middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin"))
		{
			adminSchoolMemberAPI.GET("", adminSchoolMemberImportHandler.ListMembers)
			adminSchoolMemberAPI.POST("", adminSchoolMemberImportHandler.AddMember)
			adminSchoolMemberAPI.DELETE("/:schoolUserId", adminSchoolMemberImportHandler.RemoveMember)
			adminSchoolMemberAPI.PATCH("/:schoolUserId/restore", adminSchoolMemberImportHandler.RestoreMember)
		}

		schoolMemberInvitationAPI := api.Group("/admin/school-member-invitations")
		schoolMemberInvitationAPI.Use(middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin"))
		{
			schoolMemberInvitationAPI.GET("", schoolMemberInvitationHandler.List)
			schoolMemberInvitationAPI.POST("", schoolMemberInvitationHandler.Create)
			schoolMemberInvitationAPI.PATCH("/:id/revoke", schoolMemberInvitationHandler.Revoke)
		}

		subjectAPI := api.Group("/subjects")
		{
			subjectAPI.POST("", middleware.RequireRole(schoolService, "admin", "super_admin"), subjectHandler.Create)
			subjectAPI.GET("", middleware.RequireSchoolMember(schoolService), subjectHandler.FindAll)
			subjectAPI.GET("/:id", middleware.RequireSchoolMember(schoolService), subjectHandler.GetByID)
			subjectAPI.GET("/school/:schoolCode", middleware.RequireSchoolMember(schoolService), subjectHandler.GetBySchool)
			subjectAPI.GET("/school/:schoolCode/:subjectCode", middleware.RequireSchoolMember(schoolService), subjectHandler.GetByCode)
			subjectAPI.PATCH("/:id", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin", "super_admin"), subjectHandler.Update)
			subjectAPI.DELETE("/:id", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin", "super_admin"), subjectHandler.Delete)
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
			rbacAPI.GET("/user-roles/:schoolUserId", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin", "super_admin"), rbacHandler.GetUserRoles)
			rbacAPI.PATCH("/user-roles/:schoolUserId", middleware.RequireRole(schoolService, "admin", "super_admin"), rbacHandler.UpdateUserRoles)

			// Super Admin
			rbacAPI.POST("/super-admin", middleware.RequireRole(schoolService, "super_admin"), rbacHandler.CreateSuperAdmin)
		}

		superAdminAPI := api.Group("/super-admin")
		{
			superAdminAPI.POST("/school-bootstrap", middleware.RequireSystemSuperAdmin(schoolService), superAdminBootstrapHandler.BootstrapSchool)
		}

		classAPI := api.Group("/classes")
		{
			classAPI.POST("", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin", "teacher"), classHandler.Create)
			classAPI.GET("", middleware.RequireSchoolMember(schoolService), classHandler.FindAll)
			classAPI.GET("/:id", middleware.RequireSchoolMember(schoolService), classHandler.GetByID)
			classAPI.PATCH("/:id", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin", "teacher"), classHandler.Update)
			classAPI.DELETE("/:id", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin"), classHandler.Delete)
		}

		subjectClassAPI := api.Group("/subject-classes")
		{
			subjectClassAPI.POST("/assign", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin"), subjectClassHandler.Assign)
			subjectClassAPI.GET("/my-teaching", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "teacher"), subjectClassHandler.GetMyTeaching)
			subjectClassAPI.GET("/class/:classId", middleware.RequireSchoolMember(schoolService), subjectClassHandler.GetByClass)
			subjectClassAPI.GET("/:id", middleware.RequireSchoolMember(schoolService), subjectClassHandler.GetByID)
			subjectClassAPI.PATCH("/:id", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin"), subjectClassHandler.Update)
			subjectClassAPI.DELETE("/:id", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin"), subjectClassHandler.Unassign)
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
			mediaAPI.POST("/upload", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin", "teacher", "student"), mediaHandler.Upload)
			mediaAPI.POST("/metadata", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin", "teacher", "student"), mediaHandler.RecordMetadata)
			mediaAPI.GET("/:id", middleware.RequireSchoolMember(schoolService), mediaHandler.GetByID)
			mediaAPI.DELETE("/:id", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin", "teacher", "student"), mediaHandler.Delete)
		}

		materialAPI := api.Group("/materials")
		{
			materialAPI.POST("", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "teacher"), materialHandler.Create)
			materialAPI.GET("", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin", "teacher", "student"), materialHandler.FindAll)
			materialAPI.GET("/:id", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin", "teacher", "student"), materialHandler.GetByID)
			materialAPI.POST("/:materialId/media/:mediaId/summary", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "student"), materialHandler.SummarizeAttachment)
			materialAPI.PATCH("/:id", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "teacher", "admin"), materialHandler.Update)
			materialAPI.DELETE("/:id", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "teacher", "admin"), materialHandler.Delete)
			materialAPI.POST("/progress", materialHandler.UpdateProgress)
		}

		studentNoteAPI := api.Group("/notes")
		studentNoteAPI.Use(middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "student"))
		{
			studentNoteAPI.GET("", studentNoteHandler.GetAccessibleNotes)
			studentNoteAPI.GET("/subject-class/:subjectClassId", studentNoteHandler.GetSubjectClassNotes)
			studentNoteAPI.GET("/material/:materialId", studentNoteHandler.GetMaterialNote)
			studentNoteAPI.PUT("/material/:materialId", studentNoteHandler.SaveMaterialNote)
			studentNoteAPI.DELETE("/material/:materialId", studentNoteHandler.DeleteMaterialNote)
		}

		feedAPI := api.Group("/feeds")
		{
			feedAPI.POST("", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "teacher", "admin"), feedHandler.Create)
			feedAPI.GET("/unread-count", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin", "teacher", "student"), feedHandler.GetUnreadCount)
			feedAPI.PATCH("/read", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin", "teacher", "student"), feedHandler.MarkRead)
			feedAPI.GET("/class/:classId", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin", "teacher", "student"), feedHandler.GetByClass)
			feedAPI.GET("/:id", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin", "teacher", "student"), feedHandler.GetByID)
			feedAPI.PATCH("/:id", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "teacher", "admin"), feedHandler.Update)
			feedAPI.DELETE("/:id", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "teacher", "admin"), feedHandler.Delete)
		}

		commentAPI := api.Group("/comments")
		{
			commentAPI.POST("", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin", "teacher", "student"), commentHandler.Create)
			commentAPI.GET("", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin", "teacher", "student"), commentHandler.GetBySource)
			commentAPI.GET("/:id", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin", "teacher", "student"), commentHandler.GetByID)
			commentAPI.PATCH("/:id", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin", "teacher", "student"), commentHandler.Update)
			commentAPI.DELETE("/:id", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin", "teacher", "student"), commentHandler.Delete)
		}

		chatAPI := api.Group("/chat")
		{
			chatAPI.GET("/rooms", middleware.RequireSchoolMember(schoolService), chatHandler.ListRooms)
			chatAPI.GET("/members", middleware.RequireSchoolMember(schoolService), chatHandler.ListMembers)
			chatAPI.POST("/school/open", middleware.RequireSchoolMember(schoolService), chatHandler.OpenSchoolRoom)
			chatAPI.POST("/dm/open", middleware.RequireSchoolMember(schoolService), chatHandler.OpenDirectMessage)
			chatAPI.POST("/groups", middleware.RequireSchoolMember(schoolService), chatHandler.CreateGroupRoom)
			chatAPI.GET("/groups/:roomId", middleware.RequireSchoolMember(schoolService), chatHandler.GetGroupInfo)
			chatAPI.PATCH("/groups/:roomId", middleware.RequireSchoolMember(schoolService), chatHandler.RenameGroupRoom)
			chatAPI.POST("/groups/:roomId/leave", middleware.RequireSchoolMember(schoolService), chatHandler.LeaveGroupRoom)
			chatAPI.POST("/groups/:roomId/members", middleware.RequireSchoolMember(schoolService), chatHandler.AddGroupMembers)
			chatAPI.DELETE("/groups/:roomId/members/:userId", middleware.RequireSchoolMember(schoolService), chatHandler.RemoveGroupMember)
			chatAPI.GET("/rooms/:roomId/read-summary", middleware.RequireSchoolMember(schoolService), chatHandler.GetReadSummary)
			chatAPI.GET("/rooms/:roomId/messages", middleware.RequireSchoolMember(schoolService), chatHandler.ListMessages)
			chatAPI.POST("/rooms/:roomId/messages", middleware.RequireSchoolMember(schoolService), chatHandler.CreateMessage)
			chatAPI.PATCH("/rooms/:roomId/read", middleware.RequireSchoolMember(schoolService), chatHandler.MarkRead)
		}

		assignmentAPI := api.Group("/assignments")
		{
			// Categories
			assignmentAPI.POST("/categories", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin"), assignmentHandler.CreateCategory)
			assignmentAPI.GET("/categories/school/:schoolCode", middleware.RequireSchoolMember(schoolService), assignmentHandler.GetCategoriesBySchool)

			// Assignments
			assignmentAPI.POST("", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "teacher"), assignmentHandler.CreateAssignment)
			assignmentAPI.GET("/teacher-assignments", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "teacher"), assignmentHandler.GetTeacherAssignmentInbox)
			assignmentAPI.GET("/teacher-submissions", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "teacher"), assignmentHandler.GetTeacherSubmissionInbox)
			assignmentAPI.GET("/student-assignments", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "student"), assignmentHandler.GetStudentAssignmentInbox)
			assignmentAPI.GET("/student/:assignmentId", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "student"), assignmentHandler.GetStudentAssignmentDetail)
			assignmentAPI.GET("/subject-class/submissions/:subjectClassId", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "teacher"), assignmentHandler.GetSubjectClassSubmissions)
			assignmentAPI.GET("/subject-class/:subjectClassId", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin", "teacher", "student"), assignmentHandler.GetBySubjectClass)
			assignmentAPI.GET("/status/:id", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "teacher", "admin"), assignmentHandler.GetAssignmentStatus)
			assignmentAPI.GET("/my-submission/:assignmentId", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "student"), assignmentHandler.GetMySubmissionByAssignment)
			assignmentAPI.GET("/:assignmentId", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "teacher"), assignmentHandler.GetSubmissionsByAssignment)
			assignmentAPI.PATCH("/:id", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "teacher", "admin"), assignmentHandler.UpdateAssignment)
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
			gradeAPI.POST("/weights", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin"), gradeHandler.ConfigureWeights)
			gradeAPI.GET("/weights/subject/:subjectId", middleware.RequireSchoolMember(schoolService), gradeHandler.GetWeightsBySubject)
			gradeAPI.GET("/class/:classId/subject/:subjectId", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "teacher", "admin"), gradeHandler.GetClassGradeReport)
			gradeAPI.GET("/class/:classId/subject/:subjectId/student/:studentId", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "teacher", "admin", "student"), gradeHandler.GetStudentGradeDetail)
			gradeAPI.GET("/class/:classId/student/:studentId/report", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "teacher", "admin", "student"), gradeHandler.GetStudentReport)
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
			logAPI.GET("/school/:schoolId", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin", "super_admin"), logHandler.GetBySchool)
		}

		dashboardAPI := api.Group("/dashboard")
		{
			dashboardAPI.GET("/student/:userId", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "student"), dashboardHandler.GetStudentDashboard)
			dashboardAPI.GET("/teacher/:schoolUserId", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "teacher"), dashboardHandler.GetTeacherDashboard)
			dashboardAPI.GET("/admin/:schoolId", middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "admin"), dashboardHandler.GetAdminDashboard)
			dashboardAPI.GET("/super-admin", middleware.RequireRole(schoolService, "super_admin"), dashboardHandler.GetSuperAdminDashboard)
		}

		activityAPI := api.Group("/academic-activity")
		activityAPI.Use(middleware.RequireSchoolMember(schoolService), middleware.RequireRole(schoolService, "student", "teacher"))
		{
			activityAPI.GET("", activityHandler.GetAcademicActivity)
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
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, SchoolId, schoolid, Active-Role, active-role")
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
