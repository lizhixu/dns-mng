package main

import (
	"log"

	"dns-mng/config"
	"dns-mng/database"
	"dns-mng/handler"
	"dns-mng/middleware"
	"dns-mng/provider"
	"dns-mng/provider/cloudflare"
	"dns-mng/provider/desec"
	"dns-mng/provider/dnshe"
	"dns-mng/provider/dynu"
	"dns-mng/provider/ndjp"
	"dns-mng/provider/tencentcloud"
	"dns-mng/service"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load config
	cfg := config.Load()

	// Init database
	database.Init(cfg.DBPath)
	defer database.Close()

	// Register providers
	provider.Register(dynu.New())
	provider.Register(tencentcloud.New())
	provider.Register(cloudflare.New())
	provider.Register(ndjp.New())
	provider.Register(desec.New())
	provider.Register(dnshe.New())

	// Init services
	userService := service.NewUserService(cfg)
	accountService := service.NewAccountService()
	domainCacheService := service.NewDomainCacheService()
	dnsService := service.NewDNSService(accountService, domainCacheService)
	acmeService := service.NewAcmeService(dnsService)
	logService := service.NewLogService()
	schedulerLogService := service.NewSchedulerLogService()
	notificationService := service.NewNotificationService()
	emailService := service.NewEmailService()

	// Start scheduler for domain expiry notifications
	schedulerService := service.NewSchedulerService(notificationService, emailService, schedulerLogService)
	schedulerService.Start()
	defer schedulerService.Stop()

	// Init handlers
	authHandler := handler.NewAuthHandler(userService, logService)
	accountHandler := handler.NewAccountHandler(accountService, logService)
	dnsHandler := handler.NewDNSHandler(dnsService, logService)
	providerHandler := handler.NewProviderHandler()
	logHandler := handler.NewLogHandler(logService)
	schedulerLogHandler := handler.NewSchedulerLogHandler(schedulerLogService, schedulerService)
	dnsCheckHandler := handler.NewDNSCheckHandler()
	domainCacheHandler := handler.NewDomainCacheHandler(dnsService, logService)
	notificationHandler := handler.NewNotificationHandler(notificationService, emailService, logService)
	acmeHandler := handler.NewAcmeHandler(acmeService)

	// Setup router
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	// Public routes
	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}
		api.GET("/providers", providerHandler.List)

		// External ACME DNS-01 API (HTTP Basic Auth with system user)
		acme := api.Group("/acme")
		acme.Use(middleware.BasicAuthMiddleware(userService))
		{
			acme.POST("/dns01/present", acmeHandler.Present)
			acme.POST("/dns01/cleanup", acmeHandler.Cleanup)
		}
	}

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(cfg))
	{
		// User profile
		protected.GET("/user/profile", authHandler.GetProfile)
		protected.PUT("/user/password", authHandler.UpdatePassword)

		// Operation logs
		protected.GET("/logs", logHandler.GetLogs)

		// Scheduler logs
		protected.GET("/scheduler-logs", schedulerLogHandler.GetSchedulerLogs)
		protected.GET("/scheduler-logs/:taskName", schedulerLogHandler.GetSchedulerLogsByTask)
		protected.POST("/scheduler/trigger", schedulerLogHandler.TriggerManualCheck)

		// All domains
		protected.GET("/domains", dnsHandler.ListAllDomains)
		protected.GET("/domains/refresh", dnsHandler.RefreshAllDomains)

		// Accounts
		protected.GET("/accounts", accountHandler.List)
		protected.POST("/accounts", accountHandler.Create)
		protected.PUT("/accounts/:id", accountHandler.Update)
		protected.DELETE("/accounts/:id", accountHandler.Delete)

		// DNS
		protected.GET("/accounts/:id/domains", dnsHandler.ListDomains)
		protected.GET("/accounts/:id/domains/refresh", dnsHandler.RefreshDomains)
		protected.GET("/accounts/:id/domains/:domainId", dnsHandler.GetDomain)
		protected.PUT("/accounts/:id/domains/:domainId/cache", domainCacheHandler.UpdateDomainCache)

		// Domain cache batch operations
		protected.POST("/cache/batch", domainCacheHandler.BatchUpdateDomainCache)
		protected.DELETE("/cache/batch", domainCacheHandler.BatchDeleteDomainCache)
		protected.GET("/cache/stats", domainCacheHandler.GetCacheStats)

		// Notification settings
		protected.GET("/accounts/:id/domains/:domainId/notification", notificationHandler.GetNotificationSetting)
		protected.PUT("/accounts/:id/domains/:domainId/notification", notificationHandler.UpdateNotificationSetting)
		protected.GET("/notifications", notificationHandler.GetAllNotificationSettings)

		// Email configuration
		protected.GET("/email/config", notificationHandler.GetEmailConfig)
		protected.PUT("/email/config", notificationHandler.UpdateEmailConfig)
		protected.POST("/email/test", notificationHandler.TestEmailConfig)

		protected.GET("/accounts/:id/domains/:domainId/records", dnsHandler.ListRecords)
		protected.POST("/accounts/:id/domains/:domainId/records", dnsHandler.CreateRecord)
		protected.PUT("/accounts/:id/domains/:domainId/records/:recordId", dnsHandler.UpdateRecord)
		protected.DELETE("/accounts/:id/domains/:domainId/records/:recordId", dnsHandler.DeleteRecord)

		// DNS Check
		protected.POST("/dns/check", dnsCheckHandler.CheckDNS)
	}

	log.Printf("Server starting on :%s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
