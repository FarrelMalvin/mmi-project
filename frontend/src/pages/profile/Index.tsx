import React, { useState, useEffect, useRef, useCallback } from "react";
import { useAuth } from "../../contexts/AuthContext";
import { api } from "../../lib/api";
import { Button } from "../../components/common/Button";
import { Card, CardContent, CardHeader, CardTitle } from "../../components/common/Card";
import { Label } from "../../components/common/Label";
import { Separator } from "../../components/common/Separator";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, DialogDescription } from "../../components/common/Dialog";
import { toast } from "sonner";
import { User, Briefcase, Building2, MapPin, Edit3, Check, X, Upload, Crop } from "lucide-react";
import Cropper from "react-easy-crop";

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL || "";

// Fungsi untuk me-render hasil crop menjadi gambar 400x150 px
const getCroppedImg = (imageSrc: string, pixelCrop: any): Promise<string> => {
  return new Promise((resolve, reject) => {
    const image = new Image();
    image.src = imageSrc;
    image.onload = () => {
      const canvas = document.createElement("canvas");
      const ctx = canvas.getContext("2d");

      if (!ctx) {
        reject(new Error("Canvas context is not available"));
        return;
      }

      // Pastikan hasil akhir berukuran persis 400x150 pixel
      canvas.width = 400;
      canvas.height = 150;

      // Gambar bagian yang di-crop ke canvas
      ctx.drawImage(
        image,
        pixelCrop.x,
        pixelCrop.y,
        pixelCrop.width,
        pixelCrop.height,
        0,
        0,
        400,
        150
      );

      // Konversi ke base64 (Format pasti PNG)
      resolve(canvas.toDataURL("image/png"));
    };
    image.onerror = (error) => reject(error);
  });
};

export default function ProfilePage() {
  const { user } = useAuth();
  const [profile, setProfile] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  
  // State untuk Dialog & Crop
  const [showSignatureDialog, setShowSignatureDialog] = useState(false);
  const [imageSrc, setImageSrc] = useState<string>(""); // Gambar asli sebelum di-crop
  const [finalSignatureBase64, setFinalSignatureBase64] = useState<string>(""); // Hasil akhir 400x150
  
  // State untuk kontrol react-easy-crop
  const [crop, setCrop] = useState({ x: 0, y: 0 });
  const [zoom, setZoom] = useState(1);
  const [croppedAreaPixels, setCroppedAreaPixels] = useState<any>(null);
  const [isCropping, setIsCropping] = useState(false);
  
  const [uploading, setUploading] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    fetchProfile();
  }, []);

  const fetchProfile = async () => {
    setLoading(true);
    try {
      const res = await api.get("/user/profile");
      let profileData = null;
      if (res.data && res.data.data) {
        profileData = res.data.data;
      } else if (res.data) {
        profileData = res.data;
      }
      setProfile(profileData);
    } catch (err) {
      toast.error("Gagal memuat profil");
    } finally {
      setLoading(false);
    }
  };

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    // VALIDASI: Hanya terima PNG
    if (file.type !== "image/png") {
      toast.error("Format file harus PNG (Transparan/Solid)");
      if (fileInputRef.current) fileInputRef.current.value = "";
      return;
    }

    if (file.size > 5 * 1024 * 1024) {
      toast.error("Ukuran file maksimal 5MB");
      if (fileInputRef.current) fileInputRef.current.value = "";
      return;
    }

    const reader = new FileReader();
    reader.onloadend = () => {
      setImageSrc(reader.result as string);
      setIsCropping(true); // Buka mode potong
      setFinalSignatureBase64(""); 
    };
    reader.readAsDataURL(file);
    
    if (fileInputRef.current) fileInputRef.current.value = "";
  };
  const onCropComplete = useCallback((_:any, croppedAreaPixels: any) => {
    setCroppedAreaPixels(croppedAreaPixels);
  }, []);

  const handleCropSave = async () => {
    try {
      const croppedImageBase64 = await getCroppedImg(imageSrc, croppedAreaPixels);
      setFinalSignatureBase64(croppedImageBase64); 
      setIsCropping(false); 
    } catch (e) {
      toast.error("Gagal memotong gambar");
    }
  };

  const handleUploadSignature = async () => {
    if (!finalSignatureBase64) {
      toast.error("Pilih dan potong gambar terlebih dahulu");
      return;
    }

    setUploading(true);
    try {
      const fetchResponse = await fetch(finalSignatureBase64);
      const blob = await fetchResponse.blob();
      
      const file = new File([blob], "signature_cropped.png", { type: "image/png" });

      
      const formData = new FormData();
      formData.append("signature_file", file);

      // 4. Kirim ke backend menggunakan multipart/form-data
      await api.post("/user/signature", formData, {
        headers: {
          "Content-Type": "multipart/form-data",
        },
      });

      toast.success("Tanda tangan berhasil diperbarui");
      resetSignatureState();
      fetchProfile();
    } catch (err: any) {
      console.error("Upload Error:", err.response?.data);
      toast.error(err.response?.data?.message || err.response?.data?.error || "Gagal mengupload tanda tangan");
    } finally {
      setUploading(false);
    }
  };

  const resetSignatureState = () => {
    setShowSignatureDialog(false);
    setImageSrc("");
    setFinalSignatureBase64("");
    setIsCropping(false);
    setCrop({ x: 0, y: 0 });
    setZoom(1);
  };

  const getInitials = (name: string) => {
    if (!name) return "?";
    const parts = name.split(" ");
    if (parts.length >= 2) return `${parts[0][0]}${parts[1][0]}`.toUpperCase();
    return name.substring(0, 2).toUpperCase();
  };

  if (loading) return <div className="p-8 text-center text-slate-400">Memuat profil...</div>;
  if (!profile) return <div className="p-8 text-center text-red-500">Gagal memuat profil</div>;

  return (
    <div className="max-w-4xl mx-auto py-8 px-4">
      <div className="mb-6">
        <h2 className="text-2xl font-bold text-slate-900 tracking-tight" style={{ fontFamily: "'Plus Jakarta Sans', sans-serif" }}>
          Profil Saya
        </h2>
        <p className="text-slate-500 text-sm mt-1">Informasi akun dan pengaturan</p>
      </div>

      <div className="grid gap-6 md:grid-cols-3">
        {/* Kolom Kiri - Avatar & Jabatan */}
        <div className="md:col-span-1">
          <Card className="border-slate-100 shadow-sm rounded-xl overflow-hidden">
            <CardContent className="p-6">
              <div className="flex flex-col items-center">
                <div className="w-32 h-32 rounded-full bg-gradient-to-br from-blue-500 to-indigo-600 flex items-center justify-center mb-4">
                  <span className="text-4xl font-bold text-white">
                    {getInitials(profile.nama || user?.name || "User")}
                  </span>
                </div>
                <h3 className="text-xl font-bold text-slate-900 text-center mb-1">
                  {profile.nama || user?.name || "-"}
                </h3>
                <div className="inline-flex items-center px-3 py-1 rounded-full bg-blue-50 border border-blue-200 mt-2">
                  <Briefcase className="h-3 w-3 text-blue-600 mr-1.5" />
                  <span className="text-xs font-medium text-blue-700">{profile.jabatan || "-"}</span>
                </div>
              </div>
              <Separator className="my-4" />
              <div className="space-y-3">
                <div className="flex items-center text-sm">
                  <Building2 className="h-4 w-4 text-slate-400 mr-3" />
                  <span className="text-slate-600">{profile.departemen || "-"}</span>
                </div>
                <div className="flex items-center text-sm">
                  <MapPin className="h-4 w-4 text-slate-400 mr-3" />
                  <span className="text-slate-600">{profile.wilayah || "-"}</span>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Kolom Kanan - Data Personal & Tanda Tangan */}
        <div className="md:col-span-2 space-y-6">
          <Card className="border-slate-100 shadow-sm rounded-xl">
            <CardHeader className="border-b border-slate-100 pb-4">
              <CardTitle className="text-base font-semibold flex items-center gap-2">
                <User className="h-4 w-4 text-slate-500" /> Informasi Pengguna
              </CardTitle>
            </CardHeader>
            <CardContent className="p-6">
              <div className="grid gap-6 md:grid-cols-2">
                <div><Label className="text-xs text-slate-500 uppercase">Nama Lengkap</Label><p className="text-sm font-medium mt-1">{profile.nama || "-"}</p></div>
                <div><Label className="text-xs text-slate-500 uppercase">Jabatan</Label><p className="text-sm font-medium mt-1">{profile.jabatan || "-"}</p></div>
                <div><Label className="text-xs text-slate-500 uppercase">Departemen</Label><p className="text-sm font-medium mt-1">{profile.departemen || "-"}</p></div>
                <div><Label className="text-xs text-slate-500 uppercase">Wilayah</Label><p className="text-sm font-medium mt-1">{profile.wilayah || "-"}</p></div>
              </div>
            </CardContent>
          </Card>

          {/* Seksion Tanda Tangan */}
          <Card className="border-slate-100 shadow-sm rounded-xl">
            <CardHeader className="border-b border-slate-100 py-4">
              <div className="flex items-center justify-between">
                <CardTitle className="text-base font-semibold flex items-center gap-2">
                  <Edit3 className="h-4 w-4 text-slate-500" /> Tanda Tangan Digital
                </CardTitle>
                <Button size="sm" variant="outline" className="h-8 gap-2" onClick={() => setShowSignatureDialog(true)}>
                  <Upload className="h-3 w-3" /> {profile.path_tanda_tangan ? "Perbarui" : "Upload"}
                </Button>
              </div>
            </CardHeader>
            <CardContent className="p-6">
              {profile.path_tanda_tangan ? (
                <div className="space-y-3">
                  <div className="border-2 border-dashed border-slate-200 rounded-lg p-6 bg-slate-50 flex items-center justify-center min-h-[180px]">
                    <img
                      src={`${BACKEND_URL.replace(/\/$/, "")}${profile.path_tanda_tangan}`}
                      alt="Tanda Tangan"
                      className="max-h-32 object-contain"
                    />
                  </div>
                </div>
              ) : (
                <div className="border-2 border-dashed border-slate-200 rounded-lg p-8 bg-slate-50 text-center">
                  <div className="flex flex-col items-center">
                    <div className="h-12 w-12 rounded-full bg-slate-200 flex items-center justify-center mb-3">
                      <Edit3 className="h-6 w-6 text-slate-400" />
                    </div>
                    <p className="text-sm font-medium text-slate-700 mb-1">Belum ada tanda tangan</p>
                    <Button size="sm" className="bg-slate-900 hover:bg-slate-800 gap-2 mt-3" onClick={() => setShowSignatureDialog(true)}>
                      <Upload className="h-3 w-3" /> Upload Tanda Tangan
                    </Button>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Dialog Upload & Crop */}
      <Dialog open={showSignatureDialog} onOpenChange={(open) => !open && resetSignatureState()}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>Upload Tanda Tangan</DialogTitle>
            <DialogDescription>Hanya PNG transparan/putih. Gambar akan dipotong menjadi ukuran 400x150 pixel.</DialogDescription>
          </DialogHeader>

          {isCropping ? (
            // Tampilan Crop (Pemotongan)
            <div className="space-y-4 py-2">
              <div className="relative h-64 w-full bg-slate-900 rounded-lg overflow-hidden">
                <Cropper
                  image={imageSrc}
                  crop={crop}
                  zoom={zoom}
                  aspect={400 / 150} // Kunci rasio wajib 400x150
                  onCropChange={setCrop}
                  onZoomChange={setZoom}
                  onCropComplete={onCropComplete}
                />
              </div>
              <div className="flex items-center justify-between px-1">
                <p className="text-xs text-slate-500 font-medium">Geser dan zoom untuk menyesuaikan area.</p>
                <Button size="sm" onClick={handleCropSave} className="bg-blue-600 hover:bg-blue-700 text-white gap-2">
                  <Crop className="h-3.5 w-3.5" /> Potong Gambar
                </Button>
              </div>
            </div>
          ) : (
            // Tampilan Pilihan File / Preview Hasil Crop
            <div className="space-y-4 py-4">
              <div
                className="border-2 border-dashed border-slate-300 rounded-lg p-6 text-center transition cursor-pointer hover:border-slate-400 bg-slate-50"
                onClick={() => !finalSignatureBase64 && fileInputRef.current?.click()}
              >
                {finalSignatureBase64 ? (
                  <div className="space-y-4">
                    <div className="flex flex-col items-center justify-center p-4 bg-white border rounded shadow-sm">
                      <p className="text-xs text-slate-400 mb-2 font-medium">Preview Final (400x150 px)</p>
                      <img src={finalSignatureBase64} alt="Preview Final" className="border border-slate-200" />
                    </div>
                    <Button variant="outline" size="sm" onClick={(e) => { e.stopPropagation(); setImageSrc(""); setFinalSignatureBase64(""); setIsCropping(false); }}>
                      <X className="h-3 w-3 mr-1" /> Ulangi Pemotongan
                    </Button>
                  </div>
                ) : (
                  <div className="flex flex-col items-center py-4">
                    <div className="h-12 w-12 rounded-full bg-white shadow-sm border border-slate-100 flex items-center justify-center mb-3">
                      <Upload className="h-5 w-5 text-blue-500" />
                    </div>
                    <p className="text-sm font-semibold text-slate-700 mb-1">Klik untuk pilih file PNG</p>
                    <p className="text-xs text-slate-500">Maks. 5MB</p>
                  </div>
                )}
                <input ref={fileInputRef} type="file" accept="image/png" className="hidden" onChange={handleFileSelect} />
              </div>
            </div>
          )}

          {!isCropping && (
            <DialogFooter className="pt-2">
              <Button variant="outline" onClick={resetSignatureState} disabled={uploading}>
                Batal
              </Button>
              <Button className="bg-slate-900 hover:bg-slate-800 gap-2" onClick={handleUploadSignature} disabled={!finalSignatureBase64 || uploading}>
                {uploading ? "Mengupload..." : <><Check className="h-4 w-4" /> Simpan</>}
              </Button>
            </DialogFooter>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}