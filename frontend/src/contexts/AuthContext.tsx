// @refresh reset
import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from "react";
import { api } from "../lib/api"; 

export interface User {
  id: number | string;
  email: string;
  name: string;
  role: string; 
}

interface AuthContextType {
  user: User | null;
  login: (nama: string, password: string) => Promise<any>; 
  logout: () => void;
}

const parseJwt = (token: string) => {
  try {
    const base64Url = token.split('.')[1];
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
    const jsonPayload = decodeURIComponent(window.atob(base64).split('').map(function(c) {
        return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
    }).join(''));
    return JSON.parse(jsonPayload);
  } catch (e) {
    return null;
  }
};

export const AuthContext = createContext<AuthContextType | null>(null);

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) throw new Error("useAuth must be used within an AuthProvider");
  return context;
};

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState<boolean>(true);

  // Ubah menjadi async untuk menangani request refresh
  const loadUserFromToken = useCallback(async () => {
    const token = localStorage.getItem("token");
    if (token) {
      const decoded = parseJwt(token);
      if (decoded && decoded.exp * 1000 > Date.now()) {
        setUser({
          id: decoded.user_id || "",
          email: "", 
          name: decoded.nama || "User",
          role: decoded.jabatan ? decoded.jabatan.toLowerCase() : "pegawai" 
        });
      } else {
        // Blok eksekusi saat token expired (Mencegah logout langsung)
        try {
          const res = await api.post("/auth/refresh", {}, { withCredentials: true });
          const newToken = res.data.data.access_token;
          localStorage.setItem("token", newToken);
          
          const newDecoded = parseJwt(newToken);
          setUser({
            id: newDecoded?.user_id || "",
            email: "", 
            name: newDecoded?.nama || "User",
            role: newDecoded?.jabatan ? newDecoded.jabatan.toLowerCase() : "pegawai" 
          });
        } catch (error) {
          // Jika refresh token juga gagal/expired, baru lakukan pembersihan sesi
          localStorage.removeItem("token");
          setUser(null);
        }
      }
    } else {
      setUser(null);
    }
    setLoading(false);
  }, []);

  useEffect(() => {
    loadUserFromToken();
  }, [loadUserFromToken]);

  const login = async (nama: string, password: string): Promise<any> => {
    const res = await api.post("/auth/login", { nama, password });
    const token = res.data.data.access_token;
    localStorage.setItem("token", token);
    await loadUserFromToken(); // Tambahkan await agar state tersinkronisasi
    return res;
  };

  // Fungsi logout tidak diubah sesuai instruksi
  const logout = (): void => {
    localStorage.removeItem("token");
    setUser(null);
  };

  if (loading) {
    return (
      <div className="h-screen flex items-center justify-center bg-slate-50">
        <div className="w-8 h-8 border-2 border-slate-900 border-t-transparent rounded-full animate-spin" />
      </div>
    );
  }

  return (
    <AuthContext.Provider value={{ user, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}