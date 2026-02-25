package main

import (
	"log"

	"dns-mng/config"
	"dns-mng/database"
	"dns-mng/handler"
	"dns-mng/middleware"
	"dns-mng/provider"
	"dns-mng/provider/dynu"
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

	// Init services
	userService := service.NewUserService(cfg)
	accountService := service.NewAccountService()
	dnsService := service.NewDNSService(accountService)
	logService := service.NewLogService()

	// Init handlers
	authHandler := handler.NewAuthHandler(userService, logService)
	accountHandler := handler.NewAccountHandler(accountService, logService)
	dnsHandler := handler.NewDNSHandler(dnsService, logService)
	providerHandler := handler.NewProviderHandler()
	logHandler := handler.NewLogHandler(logService)

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

		// All domains
		protected.GET("/domains", dnsHandler.ListAllDomains)

		// Accounts
		protected.GET("/accounts", accountHandler.List)
		protected.POST("/accounts", accountHandler.Create)
		protected.PUT("/accounts/:id", accountHandler.Update)
		protected.DELETE("/accounts/:id", accountHandler.Delete)

		// DNS
		protected.GET("/accounts/:id/domains", dnsHandler.ListDomains)
		protected.GET("/accounts/:id/domains/:domainId", dnsHandler.GetDomain)
		protected.GET("/accounts/:id/domains/:domainId/records", dnsHandler.ListRecords)
		protected.POST("/accounts/:id/domains/:domainId/records", dnsHandler.CreateRecord)
		protected.PUT("/accounts/:id/domains/:domainId/records/:recordId", dnsHandler.UpdateRecord)
		protected.DELETE("/accounts/:id/domains/:domainId/records/:recordId", dnsHandler.DeleteRecord)
	}

	log.Printf("Server starting on :%s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
