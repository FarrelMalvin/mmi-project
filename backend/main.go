package main

import (
	"log"
	"net/http"
	"runtime"
	"runtime/debug"

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
	_ "net/http/pprof"
)

func main() {

	runtime.GOMAXPROCS(1)

	debug.SetMemoryLimit(512 * 1024 * 1024)

	go func() {
		log.Println("Memulai pprof di http://localhost:6060")
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
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
	serviceUpload := &service.UploadImpl{}

	//HANDLER
	handlerPPD := handler.NewPerjalananDinasHandler(servicePPD)
	handlerRBS := handler.NewRealisasiBonHandler(serviceRBS)
	handlerAuth := handler.NewAuthHandler(jwtService, serviceUser, serviceAuth)
	handlerUser := handler.NewUserHandler(serviceUser)
	notifHandler := handler.NewNotificationHandler(notifManager)
	handlerUpload := handler.NewUploadHandler(serviceUpload)

	e := echo.New()
	
	e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
	AllowOrigins: []string{"http://localhost:5173"},
	AllowCredentials: true,
	AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodPatch, http.MethodOptions},
	}))

	e.Use(echoMiddleware.RequestLogger())
	e.Use(echoMiddleware.Recover())

	authMiddleware := appMiddleware.RequireRoles(jwtService)

	e.Static("/storage", "storage")
	e.Static("/public", "public")

	//ROUTES
	route.RegisterRBSRoutes(e, handlerRBS, jwtService)
	route.RegisterPPDRoutes(e, handlerPPD, jwtService)
	route.RegisterAuth(e, handlerAuth, authMiddleware)
	route.RegisterUserRoutes(e, handlerUser, jwtService)
	route.RegisterNotificationRoutes(e, notifHandler)
	route.RegisterUploadRoutes(e, handlerUpload, jwtService)

	/*log.Println("Memulai proses seeding user...")
    config.SeedUsers(db)
    log.Println("Proses seeding selesai.")
	*/

	log.Println("Server berjalan di http://localhost:8081")
	log.Fatal(e.Start(":8081"))

	
}
