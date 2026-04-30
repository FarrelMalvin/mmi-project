import { useState, type FormEvent } from "react";
import { useNavigate, Navigate } from "react-router-dom";
import { useAuth } from "../../contexts/AuthContext";
import { Button } from "../../components/common/Button";
import { Input } from "../../components/common/Input";
import { Label } from "../../components/common/Label";
import logoimage from "../../assets/logo.png";
import { toast } from "sonner";
import { Lock, User, ArrowRight, Eye, EyeOff } from "lucide-react"; 
import axios from "axios";

export default function LoginPage() {
  const { user, login } = useAuth();
  const navigate = useNavigate();
  const [nama, setNama] = useState<string>("");
  const [password, setPassword] = useState<string>("");
  const [loading, setLoading] = useState<boolean>(false);
  const [showPassword, setShowPassword] = useState<boolean>(false); 

  if (user) return <Navigate to="/" replace />;

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      await login(nama, password);
      toast.success("Login berhasil!");
      navigate("/");
    } catch (err) {
      if (axios.isAxiosError(err)) {
        toast.error(err.response?.data?.message || "Login gagal");
      } else {
        toast.error("Terjadi kesalahan yang tidak diketahui");
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen w-full flex items-center justify-center bg-slate-50 p-4" data-testid="login-page">
      <div className="w-full max-w-md space-y-8 bg-white p-8 rounded-2xl shadow-sm border border-slate-100">
        <div className="flex flex-col items-center text-center space-y-2">
          <div className="w-16 h-16 flex items-center justify-center mb-2">
            <img 
              src={logoimage} 
              alt="Logo" 
              className="w-full h-full object-contain"
            />
          </div>
          <h1 className="text-2xl font-bold text-slate-900">MOSS</h1>
          <p className="text-slate-500 text-sm">Masukkan nama dan password untuk masuk</p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-5" data-testid="login-form">
          <div className="space-y-2">
            <Label htmlFor="nama" className="text-slate-700 text-sm font-medium">Nama</Label>
            <div className="relative">
              <User className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
              <Input
                id="nama"
                type="text"
                placeholder="Masukkan nama"
                value={nama}
                onChange={(e) => setNama(e.target.value)}
                className="pl-10 h-11"
                required
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="password" className="text-slate-700 text-sm font-medium">Password</Label>
            <div className="relative">
              <Lock className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
              <Input
                id="password"
                type={showPassword ? "text" : "password"} 
                placeholder="Masukkan password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="pl-10 pr-10 h-11" 
                required
              />
              {/* Tombol Toggle Mata */}
              <button
                type="button" 
                onClick={() => setShowPassword(!showPassword)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 focus:outline-none"
              >
                {showPassword ? (
                  <EyeOff className="h-4 w-4" />
                ) : (
                  <Eye className="h-4 w-4" />
                )}
              </button>
            </div>
          </div>

          <Button
            type="submit"
            disabled={loading}
            className="w-full h-11 bg-slate-900 hover:bg-slate-800 text-white rounded-lg font-medium"
          >
            {loading ? (
              <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
            ) : (
              <span className="flex items-center justify-center gap-2">
                Masuk <ArrowRight className="h-4 w-4" />
              </span>
            )}
          </Button>
        </form>
      </div>
    </div>
  );
}