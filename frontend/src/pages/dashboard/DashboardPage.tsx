import { useAuth } from "../../contexts/AuthContext";
import { useNavigate } from "react-router-dom";
import { 
  Receipt, 
} from "lucide-react";
import type { LucideIcon } from "lucide-react";

interface AppItem {
  path: string;
  label: string;
  icon: LucideIcon;
  gradient: string;
}

const appItems: AppItem[] = [
  { path: "/bon", label: "Bon", icon: Receipt, gradient: "from-orange-400 to-amber-500" },
];

export default function DashboardPage() {
  const { user } = useAuth();
  const navigate = useNavigate();

  return (
    <div className="flex flex-col items-center justify-center min-h-[calc(100vh-8rem)]">
      <div className="text-center mb-12 animate-fade-in">
        <h1 className="text-2xl md:text-3xl font-bold text-slate-900">
          Selamat Datang, {user?.name}
        </h1>
        <p className="text-slate-400 mt-2 text-sm">Pilih fitur untuk memulai</p>
      </div>

      {/* PERUBAHAN: Mengganti grid menjadi flex dan justify-center agar item ke tengah */}
      <div className="flex flex-wrap justify-center gap-6 md:gap-8 animate-fade-in stagger-1">
        {appItems.map((item) => {
          const Icon = item.icon;
          return (
            <button
              key={item.path}
              onClick={() => navigate(item.path)}
              className="flex flex-col items-center gap-3 group outline-none"
            >
              <div className={`w-20 h-20 md:w-24 md:h-24 rounded-2xl bg-gradient-to-br ${item.gradient} flex items-center justify-center shadow-md group-hover:shadow-xl group-hover:scale-105 group-active:scale-95 transition-all duration-200`}>
                <Icon className="h-9 w-9 md:h-11 md:w-11 text-white" strokeWidth={1.6} />
              </div>
              <span className="text-sm font-semibold text-slate-700 group-hover:text-slate-900 transition-colors">
                {item.label}
              </span>
            </button>
          );
        })}
      </div>
    </div>
  );
}