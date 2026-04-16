package route

import (
    "github.com/labstack/echo/v5"
   
    "golang-mmi/internal/middleware"
    "golang-mmi/internal/handler"
    "golang-mmi/internal/config"
    "golang-mmi/internal/constant"
)

func RegisterAuth(e *echo.Echo, h *handler.AuthHandler, authMiddleware echo.MiddlewareFunc) {
    authGroup := e.Group("/api/v1/auth")

    authGroup.POST("/login", h.Login)
    authGroup.POST("/refresh", h.Refresh)
    authGroup.POST("/logout", h.Logout, authMiddleware)
}

func RegisterPPDRoutes(e *echo.Echo, h *handler.PerjalananDinasHandler, jwtService *config.JWTService) {
    ppdGroup := e.Group("/api/v1/ppd")

    allowAll := middleware.RequireRoles(jwtService, "Pegawai", "Atasan", "Direktur", "HRGA", "Finance")
    ppdGroup.GET("", h.GetRiwayatPerjalananDinas, allowAll)
    ppdGroup.POST("", h.CreatePengajuanPerjalanaDinas, allowAll)
    ppdGroup.GET("/:id", h.GetPerjalananDetail, allowAll)
    ppdGroup.GET("/:id/item", h.GetItemsByPPDID, allowAll)
    

    allowApprover := middleware.RequireRoles(jwtService, "Atasan", "Direktur", "HRGA", "Finance")
    ppdGroup.PATCH("/:id/approve", h.ApprovePerjalananDinas, allowApprover)
    ppdGroup.PATCH("/:id/decline", h.DeclinePerjalananDinas, allowApprover)
    ppdGroup.GET("/:id/download", h.GeneratePPDPDF, allowAll)
    ppdGroup.GET("/:id/download/bs", h.GenerateBSPDF, allowAll)
	ppdGroup.GET("/pending", h.GetListPendingPerjalananDinas, allowAll)
}

func RegisterRBSRoutes(e *echo.Echo, h *handler.RealisasiBonsHandler, jwtService *config.JWTService) {
    rbsGroup := e.Group("/api/v1/rbs")

    allowAll := middleware.RequireRoles(jwtService, "Pegawai", "Atasan", "Direktur", "HRGA", "Finance")
    rbsGroup.GET("", h.GetListRBS, allowAll)
    rbsGroup.POST("", h.CreateRealisasiBon, allowAll)
     rbsGroup.GET("/options", h.GetDropdownPPD, allowAll)
    rbsGroup.GET("/:id", h.GetListRBSDetail, allowAll)
   
    

    allowApprover := middleware.RequireRoles(jwtService, "Atasan", "Direktur", "HRGA", "Finance")
    rbsGroup.PATCH("/:id/approve", h.ApproveRBS, allowApprover)
    rbsGroup.PATCH("/:id/decline", h.DeclineRBS, allowApprover)
	rbsGroup.GET("/pending", h.GetListPendingRBS, allowAll)
    rbsGroup.GET("/:id/download", h.GenerateRBSPDF, allowAll)
    rbsGroup.GET("/download/excel", h.DownloadExcel, allowAll)
}

func RegisterUserRoutes(e *echo.Echo, h *handler.UserHandler, jwtService *config.JWTService) {
    userGroup := e.Group("/api/v1/user")
    allowAll := middleware.RequireRoles(jwtService, "Pegawai", "Atasan", "Direktur", "HRGA", "Finance")
    userGroup.POST("/signature", h.UpdateSignature, allowAll)
}

func RegisterNotificationRoutes(e *echo.Echo, h *handler.NotificationHandler) {
	e.GET("/ws", h.HandleWS)

	e.POST("/trigger", h.TriggerNotif)
}

func RegisterUploadRoutes(e *echo.Echo, h *handler.UploadHandler, jwtService *config.JWTService) {
    upload := e.Group("/api/v1/upload")
    upload.Use(middleware.RequireRoles(jwtService,
        constant.JabatanPegawai,
        constant.JabatanAtasan,
        constant.JabatanHRGA,
    ))
    upload.POST("/struk", h.UploadStruk)
}