package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v5"
	echoMiddleware "github.com/labstack/echo/v5/middleware"

	"golang-mmi/internal/config"
	"golang-mmi/internal/handler"
	"golang-mmi/internal/repository"
	"golang-mmi/internal/route"
	"golang-mmi/internal/service"
	"golang-mmi/internal/utils"
	appMiddleware "golang-mmi/internal/middleware"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("File .env tidak ditemukan")
	}

	log.Println("Mencoba melakukan koneksi ke database...")

	db := config.InitDB()
	if db == nil {
		log.Fatal("Gagal terhubung ke database")
	}
	log.Println("Database terhubung")

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Gagal load konfigurasi JWT: %v", err)
	}

	tokenStore := repository.NewTokenRepository(db)
	jwtService := config.NewJWTService(cfg, tokenStore)
	notifManager := utils.NewNotificationManager()

	//REPOSITORY
	repoUser := repository.NewUserRepository(db)
	repoPPD := repository.NewPerjalananDinasRepository(db)
	repoRBS := repository.NewRealisasiBonRepository(db)
	repoDoc := repository.NewDocumentRepository(db)

	//SERVICE
	serviceDoc := service.NewDocumentService(repoDoc)
	servicePPD := service.NewPerjalananDinasService(repoPPD, serviceDoc, notifManager)
	serviceRBS := service.NewRealisasiRBSService(repoRBS, repoPPD, serviceDoc, notifManager)
	serviceAuth := service.NewAuthService(repoUser)
	serviceUser := service.NewUserService(repoUser)

	//HANDLER
	handlerPPD := handler.NewPerjalananDinasHandler(servicePPD)
	handlerRBS := handler.NewRealisasiBonHandler(serviceRBS)
	handlerAuth := handler.NewAuthHandler(jwtService, serviceUser, serviceAuth)
	handlerUser := handler.NewUserHandler(serviceUser)
	notifHandler := handler.NewNotificationHandler(notifManager)

	e := echo.New()
	e.Static("/", "public")

	e.Use(echoMiddleware.RequestLogger())
	e.Use(echoMiddleware.Recover())

	authMiddleware := appMiddleware.RequireRoles(jwtService)

	//ROUTES
	route.RegisterRBSRoutes(e, handlerRBS, jwtService)
	route.RegisterPPDRoutes(e, handlerPPD, jwtService)
	route.RegisterAuth(e, handlerAuth, authMiddleware)
	route.RegisterUserRoutes(e, handlerUser, jwtService)
	route.RegisterNotificationRoutes(e, notifHandler)

	/*log.Println("Memulai proses seeding user...")
    config.SeedUsers(db)
    log.Println("Proses seeding selesai.")
	*/

	log.Println("Server berjalan di http://localhost:8081")
	log.Fatal(e.Start(":8081"))

	
}
