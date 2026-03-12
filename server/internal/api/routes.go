package api

import (
	"com.hermes.platform/internal/auth"
	"com.hermes.platform/internal/repository"
	"com.hermes.platform/internal/services"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

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

	api := r.Group("/api")
	{
		authRoutes := api.Group("/auth")
		{
			authRoutes.POST("/register", authHandler.Register)
			authRoutes.POST("/login", authHandler.Login)
			authRoutes.POST("/forgot-password", authHandler.ForgotPassword)
			authRoutes.Use(auth.AuthMiddleware())
			authRoutes.POST("/change-password", authHandler.ChangePassword)
			authRoutes.GET("/profile", authHandler.GetProfile)
			authRoutes.PUT("/profile", authHandler.UpdateProfile)
		}

		userRoutes := api.Group("/users")
		userRoutes.Use(auth.AuthMiddleware())
		{
			userRoutes.GET("", auth.PermissionMiddleware(permissionService, "user:manage"), userHandler.ListUsers)
			userRoutes.GET("/:id", auth.PermissionMiddleware(permissionService, "user:manage"), userHandler.GetUser)
			userRoutes.PUT("/:id", auth.PermissionMiddleware(permissionService, "user:manage"), userHandler.UpdateUser)
			userRoutes.DELETE("/:id", auth.PermissionMiddleware(permissionService, "user:manage"), userHandler.DeleteUser)
			userRoutes.PUT("/:id/roles", auth.PermissionMiddleware(permissionService, "user:manage"), userHandler.AssignRoles)
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
				testTasks.PUT("/:id", auth.PermissionMiddleware(permissionService, "test_task:edit"), testHandler.UpdateTestTask)
				testTasks.DELETE("/:id", auth.PermissionMiddleware(permissionService, "test_task:delete"), testHandler.DeleteTestTask)
				testTasks.GET("/buildid/:buildid/progress", auth.PermissionMiddleware(permissionService, "test_task:view"), testHandler.GetTestTaskProgressByBuildID)
			}

			testDetails := testRoutes.Group("/test-details")
			{
				testDetails.POST("", auth.PermissionMiddleware(permissionService, "test_detail:create"), testHandler.CreateTestDetail)
				testDetails.GET("/:id", auth.PermissionMiddleware(permissionService, "test_detail:view"), testHandler.GetTestDetailByID)
				testDetails.PUT("/:id", auth.PermissionMiddleware(permissionService, "test_detail:edit"), testHandler.UpdateTestDetail)
				testDetails.DELETE("/:id", auth.PermissionMiddleware(permissionService, "test_detail:delete"), testHandler.DeleteTestDetail)
				testDetails.GET("/:id/steps", auth.PermissionMiddleware(permissionService, "test_detail:view"), testHandler.ListTestStepDetailsByTestDetailID)
			}

			testSteps := testRoutes.Group("/test-step-details")
			{
				testSteps.POST("", auth.PermissionMiddleware(permissionService, "test_detail:create"), testHandler.CreateTestStepDetail)
				testSteps.GET("/:id", auth.PermissionMiddleware(permissionService, "test_detail:view"), testHandler.GetTestStepDetailByID)
				testSteps.PUT("/:id", auth.PermissionMiddleware(permissionService, "test_detail:edit"), testHandler.UpdateTestStepDetail)
				testSteps.DELETE("/:id", auth.PermissionMiddleware(permissionService, "test_detail:delete"), testHandler.DeleteTestStepDetail)
			}

			testRecords := testRoutes.Group("/test-records")
			{
				testRecords.POST("", auth.PermissionMiddleware(permissionService, "test_record:create"), testHandler.CreateTestRecord)
				testRecords.GET("/:id", auth.PermissionMiddleware(permissionService, "test_record:view"), testHandler.GetTestRecordByID)
				testRecords.PUT("/:id", auth.PermissionMiddleware(permissionService, "test_record:edit"), testHandler.UpdateTestRecord)
				testRecords.DELETE("/:id", auth.PermissionMiddleware(permissionService, "test_record:delete"), testHandler.DeleteTestRecord)
			}
		}
	}
}
