package utils

import (

	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v5"
	
)

type ClientInfo struct {
	Jabatan string
	UserID  uint
}

type NotificationManager struct {
	clients map[*websocket.Conn]ClientInfo
	mutex   sync.RWMutex
}

func NewNotificationManager() *NotificationManager {
	return &NotificationManager{
		clients: make(map[*websocket.Conn]ClientInfo),
	}
}

func (m *NotificationManager) AddClient(conn *websocket.Conn, jabatan string, userID uint) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.clients[conn] = ClientInfo{Jabatan: jabatan, UserID: userID}
	log.Printf("WS: Client masuk (Jabatan: %s, UserID: %d)", jabatan, userID)
}

func (m *NotificationManager) RemoveClient(conn *websocket.Conn) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.clients, conn)
	conn.Close()
	log.Printf("WS: Client terputus. Sisa client: %d", len(m.clients))
}

func (m *NotificationManager) SendToRole(targetRole string, message string) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for conn, info := range m.clients {
		if info.Jabatan == targetRole {
			// Write JSON Message
			err := conn.WriteJSON(map[string]string{
				"type":    "NOTIFIKASI",
				"message": message,
			})
			if err != nil {
				log.Println("Gagal mengirim WS:", err)
			}
		}
	}
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true 
		},
	}
	
	notifManager = NewNotificationManager()
)

func wsHandler(c *echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	// Simulasi: Ambil Role dari URL (misal: ws://localhost:1323/ws?jabatan=HRGA)
	userRole := c.QueryParam("jabatan")
	if userRole == "" {
		userRole = "Guest"
	}
	userID := uint(1) // Dummy ID

	// Daftarkan klien ke Manager
	notifManager.AddClient(ws, userRole, userID)
	defer notifManager.RemoveClient(ws)

	// Loop untuk menjaga koneksi tetap terbuka dan mendeteksi jika klien terputus
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			c.Logger().Error("Client disconnected", "error", err)
			break // Keluar dari loop jika ada error/terputus
		}
	}
	return nil
}

// Simulasi Endpoint untuk men-trigger Notifikasi
func triggerNotif(c *echo.Context) error {
	// Contoh: Atasan approve, trigger notif ke HRGA
	targetRole := c.QueryParam("target")
	if targetRole == "" {
		targetRole = "HRGA"
	}

	// Proses pengiriman berjalan di background agar API langsung merespons
	go notifManager.SendToRole(targetRole, "Ada dokumen baru yang perlu persetujuan Anda!")

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Notifikasi berhasil dikirim ke " + targetRole,
	})
}
