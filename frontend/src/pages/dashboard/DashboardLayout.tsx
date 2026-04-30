import { useState, useEffect, useCallback } from "react";
import { Outlet, NavLink, useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "../../contexts/AuthContext";
import { api } from "../../lib/api";
import { Button } from "../../components/common/Button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "../../components/common/DropdownMenu"; 
import {
  LayoutGrid,
  Receipt,
  Bell,
  LogOut,
  User,
  ChevronRight,
  CheckCircle2,
  XCircle,
  Home,
} from "lucide-react";

// Import gambar logo
import logoimage from "../../assets/logo.png";

interface Notification {
  id: number | string;
  message: string;
  is_read: boolean;
  type: string;
  created_at: string;
}

const pageNames: Record<string, string> = {
  "/": "Dashboard",
  "/bon": "Bon / Reimbursement",
  "/profile": "Profil Saya",
};

const roleLabels: Record<string, string> = { 
  pegawai: "Pegawai", 
  atasan: "Atasan", 
  finance: "Finance",
  hrga: "HRGA",
  direktur: "Direktur" 
};

const roleBadgeColors: Record<string, string> = {
  pegawai: "bg-blue-50 text-blue-700 border-blue-200",
  atasan: "bg-amber-50 text-amber-700 border-amber-200",
  finance: "bg-emerald-50 text-emerald-700 border-emerald-200",
  hrga: "bg-indigo-50 text-indigo-700 border-indigo-200",
  direktur: "bg-purple-50 text-purple-700 border-purple-200",
};

export default function DashboardLayout() {
  const { user, logout } = useAuth();
  const location = useLocation();
  const navigate = useNavigate();
  const [notifications, setNotifications] = useState<Notification[]>([]);

  const userRole = user?.role || "pegawai";

  const fetchNotificationHistory = useCallback(async () => {
    try {
      const res = await api.options<Notification[]>("/notifications");
      if (Array.isArray(res.data)) {
        setNotifications(res.data);
      }
    } catch { /* abaikan jika endpoint belum tersedia di backend */ }
  }, []);

  useEffect(() => {
    fetchNotificationHistory();

    const backendUrl = import.meta.env.VITE_BACKEND_URL;
    const wsUrl = backendUrl.replace(/^http/, 'ws') + `/ws?role=${userRole}`;
    
    const ws = new WebSocket(wsUrl);

    ws.onopen = () => console.log(`WebSocket Terhubung [Role: ${userRole}]`);

    ws.onmessage = (event) => {
      try {
        const newNotif: Notification = JSON.parse(event.data);
        setNotifications((prev) => [newNotif, ...prev]);
      } catch (error) {
        const newNotif: Notification = {
          id: Date.now(),
          message: event.data,
          is_read: false,
          type: "info",
          created_at: new Date().toISOString()
        };
        setNotifications((prev) => [newNotif, ...prev]);
      }
    };

    ws.onerror = (error) => console.error("WebSocket Error:", error);
    ws.onclose = () => console.log("WebSocket Terputus");

    return () => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.close();
      }
    };
  }, [userRole, fetchNotificationHistory]);

  const unreadCount = notifications.filter((n) => !n.is_read).length;

  const markAllRead = async () => {
    try {
      await api.put("/notifications/read-all");
      setNotifications((prev) => prev.map((n) => ({ ...n, is_read: true })));
    } catch { /* silent */ }
  };

  const currentPage = pageNames[location.pathname] || "Dashboard";
  const isHome = location.pathname === "/";
  
  const roleLabel = roleLabels[userRole] || userRole;
  const roleColor = roleBadgeColors[userRole] || "bg-slate-50 text-slate-700 border-slate-200";

  return (
    <div className="flex flex-col h-screen bg-slate-50" data-testid="dashboard-layout">
      <header className="h-14 bg-white border-b border-slate-200 flex items-center justify-between px-3 lg:px-5 shrink-0 z-50" data-testid="topbar">
        <div className="flex items-center gap-1 md:gap-2">
          
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                className="h-9 w-9 text-slate-500 hover:text-slate-900 hover:bg-slate-100 focus-visible:ring-0 focus-visible:ring-offset-0"
              >
                <LayoutGrid className="h-5 w-5" />
              </Button>
            </DropdownMenuTrigger>
            
            <DropdownMenuContent align="start" className="w-56 p-2">
              <div className="px-2 py-1.5 text-[10px] font-bold text-slate-400 uppercase tracking-wider mb-1">
                Menu Navigasi
              </div>
              <DropdownMenuItem 
                onClick={() => navigate("/")} 
                className={`gap-3 cursor-pointer p-3 rounded-lg ${isHome ? "bg-slate-100 font-semibold text-slate-900" : "text-slate-600 hover:text-slate-900 hover:bg-slate-50"}`}
              >
                <Home className={`h-5 w-5 ${isHome ? "text-slate-900" : "text-slate-500"}`} /> 
                <span className="text-sm">Dashboard</span>
              </DropdownMenuItem>
              <DropdownMenuItem 
                onClick={() => navigate("/bon")} 
                className={`gap-3 cursor-pointer p-3 rounded-lg mt-1 ${location.pathname === "/bon" ? "bg-slate-100 font-semibold text-slate-900" : "text-slate-600 hover:text-slate-900 hover:bg-slate-50"}`}
              >
                <Receipt className={`h-5 w-5 ${location.pathname === "/bon" ? "text-slate-900" : "text-slate-500"}`} /> 
                <span className="text-sm">Bon / Reimbursement</span>
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>

          <NavLink to="/" className="flex items-center gap-2 mr-1 ml-1" data-testid="nav-logo">
            {/* Bagian Logo Diubah */}
            <div className="w-8 h-8 flex items-center justify-center">
              <img 
                src={logoimage} 
                alt="Logo MOSS" 
                className="w-full h-full object-contain"
              />
            </div>
            <span className="text-slate-900 font-bold text-sm hidden sm:block" style={{ fontFamily: "'Plus Jakarta Sans', sans-serif" }}>
              MOSS
            </span>
          </NavLink>

          {!isHome && (
            <div className="flex items-center gap-1 text-sm ml-1">
              <ChevronRight className="h-3.5 w-3.5 text-slate-300" />
              <span className="text-slate-500 font-medium">{currentPage}</span>
            </div>
          )}

          <div className={`text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-md border ml-2 hidden sm:inline-block ${roleColor}`}>
            {roleLabel}
          </div>
        </div>

        <div className="flex items-center gap-1">
          
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" className="relative h-9 w-9 text-slate-500 hover:text-slate-900 focus-visible:ring-0 focus-visible:ring-offset-0">
                <Bell className="h-[18px] w-[18px]" />
                {unreadCount > 0 && (
                  <span className="absolute -top-0.5 -right-0.5 w-4 h-4 bg-red-500 text-white text-[10px] font-bold rounded-full flex items-center justify-center">
                    {unreadCount > 9 ? "9+" : unreadCount}
                  </span>
                )}
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-80">
              <div className="flex items-center justify-between p-3 border-b border-slate-100">
                <span className="text-sm font-semibold text-slate-900" style={{ fontFamily: "'Plus Jakarta Sans', sans-serif" }}>Notifikasi</span>
                {unreadCount > 0 && (
                  <button onClick={markAllRead} className="text-xs text-blue-600 hover:text-blue-700 font-medium">
                    Tandai semua dibaca
                  </button>
                )}
              </div>
              <div className="max-h-64 overflow-y-auto">
                {notifications.length === 0 ? (
                  <div className="p-4 text-center text-sm text-slate-400">Tidak ada notifikasi</div>
                ) : (
                  notifications.slice(0, 10).map((notif) => (
                    <div
                      key={notif.id}
                      className={`p-3 border-b border-slate-50 last:border-0 ${!notif.is_read ? "bg-blue-50/50" : ""}`}
                    >
                      <div className="flex items-start gap-2">
                        {notif.type === "approved" ? (
                          <CheckCircle2 className="h-4 w-4 text-emerald-500 mt-0.5 shrink-0" />
                        ) : notif.type === "info" ? (
                          <Bell className="h-4 w-4 text-blue-500 mt-0.5 shrink-0" />
                        ) : (
                          <XCircle className="h-4 w-4 text-red-500 mt-0.5 shrink-0" />
                        )}
                        <div>
                          <p className="text-sm text-slate-700 leading-snug">{notif.message}</p>
                          <p className="text-xs text-slate-400 mt-1">{new Date(notif.created_at).toLocaleDateString("id-ID")}</p>
                        </div>
                      </div>
                    </div>
                  ))
                )}
              </div>
            </DropdownMenuContent>
          </DropdownMenu>

          <div className="w-px h-6 bg-slate-200 mx-1 hidden sm:block" />

          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" className="flex items-center gap-2 h-9 px-2 hover:bg-slate-100 focus-visible:ring-0 focus-visible:ring-offset-0">
                <div className="w-7 h-7 bg-slate-900 rounded-full flex items-center justify-center">
                  <span className="text-white text-xs font-semibold">
                    {user?.name?.charAt(0)?.toUpperCase() || "U"}
                  </span>
                </div>
                <span className="text-sm font-medium text-slate-700 hidden sm:block max-w-[120px] truncate">
                  {user?.name}
                </span>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-52">
              <div className="px-3 py-2 border-b border-slate-100">
                <p className="text-sm font-semibold text-slate-900">{user?.name}</p>
                <p className="text-xs text-slate-500">{user?.email}</p>
                <div className={`inline-block text-[10px] font-bold uppercase tracking-wider px-1.5 py-0.5 rounded border mt-1 ${roleColor}`}>
                  {roleLabel}
                </div>
              </div>
              <DropdownMenuItem 
                className="gap-2 cursor-pointer py-2.5 mt-1 rounded-lg" 
                onClick={() => navigate('/profile')} 
              >
                <User className="h-4 w-4 text-slate-500" /> Profil
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem className="gap-2 text-red-600 cursor-pointer py-2.5 rounded-lg" onClick={logout}>
                <LogOut className="h-4 w-4" /> Keluar
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </header>

      <main className="flex-1 overflow-y-auto">
        <div className="p-4 md:p-6 lg:p-8 max-w-7xl mx-auto">
          <Outlet />
        </div>
      </main>
    </div>
  );
}