package route

import (
    "github.com/labstack/echo/v5"
    
    // Pastikan nama depannya konsisten!
    "golang-mmi/internal/config"
    "golang-mmi/internal/handler"
)

// ---------------------------------------------------------
// 1. RUTE AUTH (Sudah Sempurna)
// ---------------------------------------------------------
func RegisterAuth(e *echo.Echo, h *handler.AuthHandler, authMiddleware echo.MiddlewareFunc) {
    authGroup := e.Group("/api/v1/auth")

    authGroup.POST("/login", h.Login)
    authGroup.POST("/refresh", h.Refresh)
    authGroup.POST("/logout", h.Logout, authMiddleware)
}

// ---------------------------------------------------------
// 2. RUTE PPD (Perbaikan Parameter)
// ---------------------------------------------------------
// UBAH: Parameter ketiga sekarang adalah jwtService
func RegisterPPDRoutes(e *echo.Echo, h *handler.PerjalananDinasHandler, jwtService *config.JWTService) {
    ppdGroup := e.Group("/api/v1/ppd")

    allowAll := config.RequireRoles(jwtService, "Pegawai", "Atasan", "Direktur", "HRGA", "Finance")
    ppdGroup.GET("", h.GetRiwayatPerjalananDinas, allowAll)
    ppdGroup.POST("", h.CreatePengajuanPerjalanaDinas, allowAll)
    ppdGroup.GET("/:id", h.GetListPerjalananDetail, allowAll)
    

    allowApprover := config.RequireRoles(jwtService, "Atasan", "Direktur", "HRGA", "Finance")
    ppdGroup.PATCH("/:id/approve", h.ApprovePerjalananDinas, allowApprover)
    ppdGroup.PATCH("/:id/decline", h.DeclinePerjalananDinas, allowApprover)
	ppdGroup.GET("/pending", h.GetListPendingPerjalananDinas, allowAll)
}

// ---------------------------------------------------------
// 3. RUTE RBS (Masih Pukul Rata)
// ---------------------------------------------------------
func RegisterRBSRoutes(e *echo.Echo, h *handler.RealisasiBonsHandler, authMiddleware echo.MiddlewareFunc) {
    rbsGroup := e.Group("/api/v1/rbs")
    
    // Catatan: Ini berarti SEMUA role yang dilempar dari main.go bisa melakukan Approve/Decline RBS.
    // Jika kamu ingin membatasinya seperti PPD, rute ini harus di-refactor nanti.
    rbsGroup.Use(authMiddleware)

    rbsGroup.GET("", h.GetListRBS)
    rbsGroup.POST("", h.CreateRealisasiBon)
    rbsGroup.GET("/dropdown-ppd", h.GetDropdownPPD)
    rbsGroup.GET("/:id", h.GetListRBSDetail)
    rbsGroup.PATCH("/:id/approve", h.ApproveRBS)
    rbsGroup.PATCH("/:id/decline", h.DeclineRBS)
    rbsGroup.GET("/pending", h.GetListPendingRBS)
}