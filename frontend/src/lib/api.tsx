import axios, { type AxiosInstance } from "axios";

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL || "";

export const api: AxiosInstance = axios.create({ 
  baseURL: `${BACKEND_URL}/api/v1`,
  withCredentials: true 
});

let isRefreshing = false;
let failedQueue: any[] = [];

const processQueue = (error: any, token: string | null = null) => {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error);
    } else {
      prom.resolve(token);
    }
  });
  failedQueue = [];
};

// Request Interceptor: Menempelkan token ke setiap request
api.interceptors.request.use((config) => {
  const token = localStorage.getItem("token");
  if (token && config.headers) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Response Interceptor: Menangani error 401 dan Refresh Token
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    // Jika error 401 dan bukan request retry
    if (error.response?.status === 401 && !originalRequest._retry) {
      
      // Jika proses refresh sedang berjalan, masukkan request ke antrean
      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject });
        })
          .then((token) => {
            originalRequest.headers.Authorization = `Bearer ${token}`;
            return api(originalRequest);
          })
          .catch((err) => Promise.reject(err));
      }

      originalRequest._retry = true;
      isRefreshing = true;

      console.log("🔄 Mencoba refresh token...");

      try {
        // Menggunakan axios standar (bukan 'api') agar bersih dari header lama
        const res = await axios.post(
          `${BACKEND_URL}/api/v1/auth/refresh`, 
          {}, 
          { 
            withCredentials: true,
            headers: { 
              'Authorization': '' // Kosongkan seperti di Postman
            } 
          }
        );

        const newAccessToken = res.data?.data?.access_token;

        if (newAccessToken) {
          localStorage.setItem("token", newAccessToken);
          
          processQueue(null, newAccessToken);
          
          originalRequest.headers.Authorization = `Bearer ${newAccessToken}`;
          return api(originalRequest);
        }
      } catch (refreshError: any) {
        processQueue(refreshError, null);
        console.error("❌ Refresh token gagal. Sesi berakhir.");
        
        localStorage.removeItem("token");
        window.location.href = "/login";
        return Promise.reject(refreshError);
      } finally {
        isRefreshing = false;
      }
    }

    return Promise.reject(error);
  }
);