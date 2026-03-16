package api

import (
	"com.hermes.platform/internal/auth"
	"com.hermes.platform/internal/config"
	"com.hermes.platform/internal/repository"
	"com.hermes.platform/internal/services"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB, cfg *config.Config) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	otelCfg := cfg.GetOTelConfig()
	if otelCfg.Enabled {
		r.Use(otelgin.Middleware(otelCfg.ServiceName))
	}

	r.Use(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	testRepo := repository.NewTestRepository(db)
	testService := services.NewTestService(testRepo)
	testHandler := NewTestHandler(testService)

	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	authService := services.NewAuthService(userRepo, roleRepo)
	authHandler := NewAuthHandler(authService)
	userHandler := NewUserHandler(authService)

	permissionService := services.NewPermissionService(userRepo)

	statsService := services.NewStatsService(db)
	statsHandler := NewStatsHandler(statsService)

	tokenService := services.NewAPITokenService(db)
	tokenHandler := NewTokenHandler(tokenService)

	projectRepo := repository.NewProjectRepository(db)
	projectService := services.NewProjectService(projectRepo)
	projectHandler := NewProjectHandler(projectService)

	api := r.Group("/api")
	{
		authRoutes := api.Group("/auth")
		{
			authRoutes.GET("/rsa-pubkey", authHandler.GetRSAPubKey)
			authRoutes.POST("/register", authHandler.Register)
			authRoutes.POST("/login", authHandler.Login)
			authRoutes.POST("/forgot-password", authHandler.ForgotPassword)
			authRoutes.Use(auth.AuthMiddleware())
			authRoutes.POST("/change-password", authHandler.ChangePassword)
			authRoutes.GET("/profile", authHandler.GetProfile)
			authRoutes.POST("/profile/update", authHandler.UpdateProfile)
		}

		tokenRoutes := api.Group("/tokens")
		tokenRoutes.Use(auth.AuthMiddleware())
		{
			tokenRoutes.POST("", tokenHandler.CreateToken)
			tokenRoutes.GET("", tokenHandler.ListTokens)
			tokenRoutes.POST("/:id/delete", tokenHandler.DeleteToken)
		}

		userRoutes := api.Group("/users")
		userRoutes.Use(auth.AuthMiddleware())
		{
			userRoutes.GET("", auth.PermissionMiddleware(permissionService, "user:manage"), userHandler.ListUsers)
			userRoutes.GET("/:id", auth.PermissionMiddleware(permissionService, "user:manage"), userHandler.GetUser)
			userRoutes.POST("/:id/update", auth.PermissionMiddleware(permissionService, "user:manage"), userHandler.UpdateUser)
			userRoutes.POST("/:id/delete", auth.PermissionMiddleware(permissionService, "user:manage"), userHandler.DeleteUser)
			userRoutes.POST("/:id/assign-roles", auth.PermissionMiddleware(permissionService, "user:manage"), userHandler.AssignRoles)
		}

		statsRoutes := api.Group("/stats")
		statsRoutes.Use(auth.AuthMiddleware())
		{
			statsRoutes.GET("/dashboard", statsHandler.GetDashboardStats)
			statsRoutes.GET("/trend", statsHandler.GetTrendData)
			statsRoutes.GET("/running-tasks", statsHandler.GetRunningTasks)
		}

		testRoutes := api.Group("/")
		testRoutes.Use(auth.AuthMiddleware())
		{
			testTasks := testRoutes.Group("/test-tasks")
			{
				testTasks.POST("", auth.PermissionMiddleware(permissionService, "test_task:create"), testHandler.CreateTestTask)
				testTasks.GET("", auth.PermissionMiddleware(permissionService, "test_task:view"), testHandler.ListTestTasks)
				testTasks.GET("/:id/details", auth.PermissionMiddleware(permissionService, "test_detail:view"), testHandler.ListTestDetailsByTaskID)
				testTasks.GET("/:id/records", auth.PermissionMiddleware(permissionService, "test_record:view"), testHandler.ListTestRecordsByTaskID)
				testTasks.GET("/:id", auth.PermissionMiddleware(permissionService, "test_task:view"), testHandler.GetTestTaskByID)
				testTasks.POST("/:id/update", auth.PermissionMiddleware(permissionService, "test_task:edit"), testHandler.UpdateTestTask)
				testTasks.POST("/:id/delete", auth.PermissionMiddleware(permissionService, "test_task:delete"), testHandler.DeleteTestTask)
				testTasks.GET("/buildid/:buildid/progress", auth.PermissionMiddleware(permissionService, "test_task:view"), testHandler.GetTestTaskProgressByBuildID)
			}

			testDetails := testRoutes.Group("/test-details")
			{
				testDetails.POST("", auth.PermissionMiddleware(permissionService, "test_detail:create"), testHandler.CreateTestDetail)
				testDetails.GET("/:id", auth.PermissionMiddleware(permissionService, "test_detail:view"), testHandler.GetTestDetailByID)
				testDetails.POST("/:id/update", auth.PermissionMiddleware(permissionService, "test_detail:edit"), testHandler.UpdateTestDetail)
				testDetails.POST("/:id/delete", auth.PermissionMiddleware(permissionService, "test_detail:delete"), testHandler.DeleteTestDetail)
				testDetails.GET("/:id/steps", auth.PermissionMiddleware(permissionService, "test_detail:view"), testHandler.ListTestStepDetailsByTestDetailID)
			}

			testSteps := testRoutes.Group("/test-step-details")
			{
				testSteps.POST("", auth.PermissionMiddleware(permissionService, "test_detail:create"), testHandler.CreateTestStepDetail)
				testSteps.GET("/:id", auth.PermissionMiddleware(permissionService, "test_detail:view"), testHandler.GetTestStepDetailByID)
				testSteps.POST("/:id/update", auth.PermissionMiddleware(permissionService, "test_detail:edit"), testHandler.UpdateTestStepDetail)
				testSteps.POST("/:id/delete", auth.PermissionMiddleware(permissionService, "test_detail:delete"), testHandler.DeleteTestStepDetail)
			}

			testRecords := testRoutes.Group("/test-records")
			{
				testRecords.POST("", auth.PermissionMiddleware(permissionService, "test_record:create"), testHandler.CreateTestRecord)
				testRecords.GET("/:id", auth.PermissionMiddleware(permissionService, "test_record:view"), testHandler.GetTestRecordByID)
				testRecords.POST("/:id/update", auth.PermissionMiddleware(permissionService, "test_record:edit"), testHandler.UpdateTestRecord)
				testRecords.POST("/:id/delete", auth.PermissionMiddleware(permissionService, "test_record:delete"), testHandler.DeleteTestRecord)
			}
		}

		projectRoutes := api.Group("/projects")
		projectRoutes.Use(auth.AuthMiddleware())
		{
			projectRoutes.POST("", auth.PermissionMiddleware(permissionService, "project:create"), projectHandler.CreateProject)
			projectRoutes.GET("", auth.PermissionMiddleware(permissionService, "project:view"), projectHandler.ListProjects)
			projectRoutes.GET("/:id", auth.PermissionMiddleware(permissionService, "project:view"), projectHandler.GetProject)
			projectRoutes.POST("/:id/update", auth.PermissionMiddleware(permissionService, "project:edit"), projectHandler.UpdateProject)
			projectRoutes.POST("/:id/delete", auth.PermissionMiddleware(permissionService, "project:delete"), projectHandler.DeleteProject)

			projectRoutes.POST("/:id/versions", auth.PermissionMiddleware(permissionService, "project:edit"), projectHandler.CreateVersion)
			projectRoutes.GET("/:id/versions", auth.PermissionMiddleware(permissionService, "project:view"), projectHandler.ListVersions)
			projectRoutes.GET("/versions/:id", auth.PermissionMiddleware(permissionService, "project:view"), projectHandler.GetVersion)
			projectRoutes.POST("/versions/:id/update", auth.PermissionMiddleware(permissionService, "project:edit"), projectHandler.UpdateVersion)
			projectRoutes.POST("/versions/:id/delete", auth.PermissionMiddleware(permissionService, "project:delete"), projectHandler.DeleteVersion)

			projectRoutes.POST("/versions/:id/test-plans", auth.PermissionMiddleware(permissionService, "test_plan:create"), projectHandler.CreateTestPlan)
			projectRoutes.GET("/versions/:id/test-plans", auth.PermissionMiddleware(permissionService, "test_plan:view"), projectHandler.ListTestPlans)
			projectRoutes.GET("/test-plans/:id", auth.PermissionMiddleware(permissionService, "test_plan:view"), projectHandler.GetTestPlan)
			projectRoutes.POST("/test-plans/:id/update", auth.PermissionMiddleware(permissionService, "test_plan:edit"), projectHandler.UpdateTestPlan)
			projectRoutes.POST("/test-plans/:id/delete", auth.PermissionMiddleware(permissionService, "test_plan:delete"), projectHandler.DeleteTestPlan)

			projectRoutes.POST("/test-plans/:id/test-cases", auth.PermissionMiddleware(permissionService, "test_case:create"), projectHandler.CreateTestCase)
			projectRoutes.GET("/test-plans/:id/test-cases", auth.PermissionMiddleware(permissionService, "test_case:view"), projectHandler.ListTestCases)
			projectRoutes.GET("/test-cases/:id", auth.PermissionMiddleware(permissionService, "test_case:view"), projectHandler.GetTestCase)
			projectRoutes.GET("/test-cases/key/:caseKey", auth.PermissionMiddleware(permissionService, "test_case:view"), projectHandler.GetTestCaseByCaseKey)
			projectRoutes.POST("/test-cases/:id/update", auth.PermissionMiddleware(permissionService, "test_case:edit"), projectHandler.UpdateTestCase)
			projectRoutes.POST("/test-cases/:id/delete", auth.PermissionMiddleware(permissionService, "test_case:delete"), projectHandler.DeleteTestCase)
		}

		dictRoutes := api.Group("/dict")
		dictRoutes.Use(auth.AuthMiddleware())
		{
			dictRoutes.GET("/test-case-statuses", projectHandler.GetTestCaseStatuses)
			dictRoutes.GET("/priorities", projectHandler.GetPriorities)
			dictRoutes.GET("/test-plan-statuses", projectHandler.GetTestPlanStatuses)
			dictRoutes.GET("/version-statuses", projectHandler.GetVersionStatuses)
		}
	}
}
