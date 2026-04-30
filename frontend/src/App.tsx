import React from "react";
import "@/App.css";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { Toaster } from "./components/common/Sonner";
import { AuthProvider, useAuth } from "./contexts/AuthContext";
import DashboardLayout from "./pages/dashboard/DashboardLayout";
const LoginPage = React.lazy(() => import("./pages/login/LoginPage"));
const DashboardPage = React.lazy(() => import("./pages/dashboard/DashboardPage"));
const PerjalananDinasPage = React.lazy(() => import("./pages/perjalanandinas"));
const ProfilePage = React.lazy(() => import("./pages/profile/Index"));

function ProtectedRoute({ children }: { children: React.ReactElement }) {
  const { user } = useAuth();
  if (!user) return <Navigate to="/login" replace />;
  return children;
}

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route
            path="/"
            element={
              <ProtectedRoute>
                <DashboardLayout />
              </ProtectedRoute>
            }
          >
            <Route index element={<DashboardPage />} />
            <Route path="bon" element={<PerjalananDinasPage />} />
            <Route path="profile" element={<ProfilePage />} />
          </Route>
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
        <Toaster position="top-right" richColors />
      </AuthProvider>
    </BrowserRouter>
  );
}