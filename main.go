package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	"golang-mmi/internal/config"
	"golang-mmi/internal/handler"
	"golang-mmi/internal/repository"
	"golang-mmi/internal/route"
	"golang-mmi/internal/service"
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

	//REPOSITORY
	repoUser := repository.NewUserRepository(db)
	repoPPD := repository.NewPerjalananDinasRepository(db)
	repoRBS := repository.NewRealisasiBonRepository(db)

	//SERVICE
	servicePPD := service.NewPerjalananDinasService(repoPPD)
	serviceRBS := service.NewRealisasiRBSService(repoRBS, servicePPD)
	serviceAuth := service.NewUserService(repoUser)

	//HANDLER
	handlerPPD := handler.NewPerjalananDinasHandler(servicePPD)
	handlerRBS := handler.NewRealisasiBonHandler(serviceRBS)
	handlerAuth := handler.NewAuthHandler(jwtService, serviceAuth)

	e := echo.New()

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	authMiddleware := config.RequireRoles(jwtService)

	//ROUTES
	route.RegisterRBSRoutes(e, handlerRBS, authMiddleware)
	route.RegisterPPDRoutes(e, handlerPPD, jwtService)
	route.RegisterAuth(e, handlerAuth, authMiddleware)

	/*log.Println("Memulai proses seeding user...")
    config.SeedUsers(db)
    log.Println("Proses seeding selesai.")
	*/

	log.Println("Server berjalan di http://localhost:8081")
	log.Fatal(e.Start(":8081"))
}
