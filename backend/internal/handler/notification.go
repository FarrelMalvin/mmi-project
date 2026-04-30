package handler

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v5"
	"golang-mmi/internal/utils"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Izinkan semua CORS
	},
}

// Struct Handler untuk menyimpan instance Manager
type NotificationHandler struct {
	Manager *utils.NotificationManager
}

// Constructor Handler
func NewNotificationHandler(m *utils.NotificationManager) *NotificationHandler {
	return &NotificationHandler{Manager: m}
}

// Method untuk menangani koneksi WebSocket
func (h *NotificationHandler) HandleWS(c *echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	userRole := c.QueryParam("role")
	if userRole == "" {
		userRole = "Guest"
	}
	userID := uint(1) // Dummy ID, sesuaikan dengan JWT nanti

	// Gunakan Manager dari struct
	h.Manager.AddClient(ws, userRole, userID)
	defer h.Manager.RemoveClient(ws)

	for {
		if _, _, err := ws.ReadMessage(); err != nil {
			c.Logger().Error("Client disconnected", "error", err)
			break
		}
	}
	return nil
}

func (h *NotificationHandler) TriggerNotif(c *echo.Context) error {
	targetRole := c.QueryParam("target")
	if targetRole == "" {
		targetRole = "HRGA"
	}

	go h.Manager.SendToRole(targetRole, "Ada dokumen baru yang perlu persetujuan Anda!")

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Notifikasi berhasil dikirim ke " + targetRole,
	})
}