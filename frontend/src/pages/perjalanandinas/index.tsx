import React, { useState, useEffect, useCallback } from "react";
import { useAuth } from "../../contexts/AuthContext";
import { api } from "../../lib/api";
import { Button } from "../../components/common/Button";
import { Card, CardContent, CardHeader, CardTitle } from "../../components/common/Card";
import { Input } from "../../components/common/Input";
import { Label } from "../../components/common/Label";
import { Textarea } from "../../components/common/TextArea";
import { Badge } from "../../components/common/Badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../../components/common/Tabs";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "../../components/common/Select";
import { Separator } from "../../components/common/Separator";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, DialogDescription } from "../../components/common/Dialog";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "../../components/common/Table";
import { toast } from "sonner";
import { 
  Plus, CheckCircle2, XCircle, RotateCcw, Download, Trash2, 
  UploadCloud, FileText, Loader2, Eye, Edit, Save, Calendar
} from "lucide-react";

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL || "";

// ===================== HELPER =====================
const fmt = (n: number) => new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", minimumFractionDigits: 0 }).format(n);

const formatDate = (dateStr: string) => {
  if (!dateStr) return "-";
  
  const datePart = dateStr.substring(0, 10); 
  const months = ["Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Ags", "Sep", "Okt", "Nov", "Des"];
  
  if (/^\d{4}-\d{2}-\d{2}$/.test(datePart)) {
    const [year, month, day] = datePart.split("-");
    return `${parseInt(day, 10)} ${months[parseInt(month, 10) - 1]} ${year}`;
  }
  
  if (/^\d{2}-\d{2}-\d{4}$/.test(datePart)) {
    const [day, month, year] = datePart.split("-");
    return `${parseInt(day, 10)} ${months[parseInt(month, 10) - 1]} ${year}`;
  }

  return new Date(dateStr.replace(" ", "T")).toLocaleDateString("id-ID", { 
    day: "numeric", 
    month: "short", 
    year: "numeric" 
  });
};

const formatPeriodeCompact = (periodeString: string) => {
  if (!periodeString) return "-";
  const parts = periodeString.split(" s/d ");
  if (parts.length !== 2) return periodeString;
  const [start, end] = parts;
  return `${formatDate(start)} - ${formatDate(end)}`;
};

// Helper Format Naming Dokumen
const formatDocName = (doc: any, prefix: string) => {
  let numStr = "";

  if (prefix === "Pengajuan Perjalanan Dinas") {
    numStr = doc.nomor_tipe_dokumen || doc.id || "";
  } else if (prefix === "Bon Sementara") {
   numStr = doc.nomor_dokumen || doc.id || "";
  } else {
   numStr = doc.nomor_dokumen || doc.nomor_tipe_dokumen || doc.id || "";
  }

  const cleanNum = numStr.toString().replace(/_/g, '/');
  return `${prefix} - ${cleanNum}`;
};

function StatusBadge({ status }: { status: string }) {
  const isDeclined = status?.toLowerCase().includes("ditolak");
  return (
    <Badge variant="outline" className={`font-medium ${isDeclined ? 'bg-red-50 border-red-200 text-red-700' : 'bg-slate-50 text-slate-700 border-slate-300'}`}>
      {status || "Unknown"}
    </Badge>
  );
}

function PaginationControls({ meta, onPageChange }: { meta: any, onPageChange: (p: number) => void }) {
  if (!meta || meta.total_page <= 1) return null;
  return (
    <div className="flex items-center justify-between px-4 py-3 border-t border-slate-100 bg-slate-50/50">
      <span className="text-xs text-slate-500">Total {meta.total_data || 0} data</span>
      <div className="flex items-center gap-2">
        <Button variant="outline" size="sm" disabled={meta.page <= 1} onClick={() => onPageChange(meta.page - 1)} className="h-8 text-xs">Prev</Button>
        <span className="text-xs font-medium">Hal {meta.page} / {meta.total_page}</span>
        <Button variant="outline" size="sm" disabled={meta.page >= meta.total_page} onClick={() => onPageChange(meta.page + 1)} className="h-8 text-xs">Next</Button>
      </div>
    </div>
  );
}

// ===================== CREATE PPD DIALOG =====================
function CreatePPDDialog({ open, onClose, onSuccess }: { open: boolean; onClose: () => void; onSuccess: () => void; }) {
  const [loading, setLoading] = useState(false);
  const [bonForm, setBonForm] = useState<any>({ 
    tujuan: "", keperluan: "", url_dokumen: "",
    tanggal_berangkat: "", tanggal_kembali: "",
    rincian_hotel: { nama_hotel: "", check_in: "", check_out: "", harga: 0, },
    rincian_transportasi: [], 
    rincian_tambahan: [{ kategori: "Konsumsi", keterangan: "", kuantitas: 1, harga: 0 }]
  });

  const hariIni = new Date().toISOString().split("T")[0];

  const handleBerangkatChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value;
    setBonForm((prev: any) => {
      const updates: any = { tanggal_berangkat: val };
      // Reset tanggal kembali jika lebih kecil dari tanggal berangkat yang baru
      if (prev.tanggal_kembali && val > prev.tanggal_kembali) {
        updates.tanggal_kembali = "";
      }
      return { ...prev, ...updates };
    });
  };

  const calcEstimasiTotal = () => {
    let total = 0;
    bonForm.rincian_tambahan.forEach((item: any) => { total += (parseFloat(item.harga) || 0); });
    if (bonForm.rincian_hotel && bonForm.rincian_hotel.harga) {
      const checkin = new Date(bonForm.rincian_hotel.check_in);
      const checkout = new Date(bonForm.rincian_hotel.check_out);
      const nights = Math.ceil((checkout.getTime() - checkin.getTime()) / (1000 * 60 * 60 * 24));
      total += (parseFloat(bonForm.rincian_hotel.harga) || 0) * (nights > 0 ? nights : 1);
    }
    bonForm.rincian_transportasi.forEach((trans: any) => { total += (parseFloat(trans.harga) || 0); });
    return total;
  };

  const addEstimasiItem = () => setBonForm((p: any) => ({ ...p, rincian_tambahan: [...p.rincian_tambahan, { kategori: "Konsumsi", keterangan: "", kuantitas: 1, harga: 0 }] }));
  const removeEstimasiItem = (idx: number) => setBonForm((p: any) => ({ ...p, rincian_tambahan: p.rincian_tambahan.filter((_: any, i: number) => i !== idx) }));
  const updateEstimasiItem = (idx: number, field: string, val: string) => {
    setBonForm((p: any) => {
      const items = [...p.rincian_tambahan];
      items[idx] = { ...items[idx], [field]: field === 'harga' ? parseFloat(val) || 0 : field === 'kuantitas' ? parseInt(val) || 1 : val };
      return { ...p, rincian_tambahan: items };
    });
  };

  const addTransportItem = () => setBonForm((p: any) => ({ ...p, rincian_transportasi: [...p.rincian_transportasi, { kota_asal: "", kota_tujuan: "", jenis_transportasi: "", tipe_perjalanan: "Keberangkatan", jam_berangkat: "", harga: 0 }] }));
  const removeTransportItem = (idx: number) => setBonForm((p: any) => ({ ...p, rincian_transportasi: p.rincian_transportasi.filter((_: any, i: number) => i !== idx) }));
  const updateTransportItem = (idx: number, field: string, val: string) => {
    setBonForm((p: any) => {
      const items = [...p.rincian_transportasi];
      items[idx] = { ...items[idx], [field]: field === 'harga' ? parseFloat(val) || 0 : val };
      return { ...p, rincian_transportasi: items };
    });
  };

  const submitBon = async () => {
    if (!bonForm.tujuan || !bonForm.tanggal_berangkat || !bonForm.tanggal_kembali || !bonForm.keperluan) {
      toast.error("Field tujuan, tanggal, dan keperluan wajib diisi"); return;
    }

    if (bonForm.rincian_tambahan.some((i: any) => !i.kuantitas || parseFloat(i.kuantitas) <= 0 || !i.harga || parseFloat(i.harga) <= 0)) {
      toast.error("Kuantitas dan Harga pada estimasi biaya tidak boleh 0 atau kosong"); return;
    }

    if (bonForm.rincian_transportasi.some((t: any) => !t.harga || parseFloat(t.harga) <= 0)) {
      toast.error("Harga pada rincian transportasi tidak boleh 0 atau kosong"); return;
    }

    setLoading(true);
    try {
      let totalHargaHotel = 0;
      if (bonForm.rincian_hotel.check_in && bonForm.rincian_hotel.check_out && bonForm.rincian_hotel.harga) {
        const checkin = new Date(bonForm.rincian_hotel.check_in);
        const checkout = new Date(bonForm.rincian_hotel.check_out);
        const nights = Math.ceil((checkout.getTime() - checkin.getTime()) / (1000 * 60 * 60 * 24));
        totalHargaHotel = (parseFloat(bonForm.rincian_hotel.harga) || 0) * (nights > 0 ? nights : 1);
      }
      const payload = {
        tujuan: bonForm.tujuan, tanggal_berangkat: `${bonForm.tanggal_berangkat}T00:00:00Z`, tanggal_kembali: `${bonForm.tanggal_kembali}T00:00:00Z`,
        keperluan: bonForm.keperluan, url_dokumen: bonForm.url_dokumen || "",
        rincian_hotel: bonForm.rincian_hotel.nama_hotel ? { ...bonForm.rincian_hotel, harga: parseFloat(bonForm.rincian_hotel.harga) || 0, total_harga: totalHargaHotel, kategori: "Akomodasi" } : null,
        rincian_transportasi: bonForm.rincian_transportasi.map((t: any) => ({ ...t, harga: parseFloat(t.harga) || 0, jam_berangkat: t.jam_berangkat ? `${bonForm.tanggal_berangkat}T${t.jam_berangkat}:00Z` : "" })),
        rincian_tambahan: bonForm.rincian_tambahan.map((item: any) => ({ kategori: item.kategori, keterangan: item.keterangan, kuantitas: parseInt(item.kuantitas) || 1, harga: parseFloat(item.harga) || 0 }))
      };
      await api.post("/ppd", payload);
      toast.success("Perjalanan Dinas berhasil diajukan");
      onSuccess(); onClose();
      setBonForm({ tujuan: "", tanggal_berangkat: "", tanggal_kembali: "", keperluan: "", url_dokumen: "", rincian_hotel: { kota_tujuan: "", nama_hotel: "", check_in: "", check_out: "", harga: 0, pembayaran: "Head Office" }, rincian_transportasi: [], rincian_tambahan: [{ kategori: "Konsumsi", keterangan: "", kuantitas: 1, harga: 0 }] });
    } catch (err: any) {
      toast.error(err.response?.data?.message || "Gagal mengirim pengajuan");
    } finally { setLoading(false); }
  };

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-3xl max-h-[85vh] overflow-y-auto">
        <DialogHeader><DialogTitle>Pengajuan Perjalanan Dinas</DialogTitle></DialogHeader>
        <div className="space-y-5">
          <div className="grid grid-cols-2 gap-3">
            <div><Label>Tujuan *</Label><Input placeholder="Kota tujuan" value={bonForm.tujuan} onChange={e => setBonForm((p: any) => ({ ...p, tujuan: e.target.value }))} /></div>
            <div><Label>Keperluan *</Label><Input placeholder="Tujuan perjalanan" value={bonForm.keperluan} onChange={e => setBonForm((p: any) => ({ ...p, keperluan: e.target.value }))} /></div>
            
            {/* Input Tanggal dengan Validasi Batas Bawah (min) */}
            <div>
              <Label>Tanggal Berangkat *</Label>
              <Input 
                type="date" 
                value={bonForm.tanggal_berangkat} 
                min={hariIni} 
                onChange={handleBerangkatChange} 
              />
            </div>
            <div>
              <Label>Tanggal Kembali *</Label>
              <Input 
                type="date" 
                value={bonForm.tanggal_kembali} 
                min={bonForm.tanggal_berangkat || hariIni} 
                onChange={e => setBonForm((p: any) => ({ ...p, tanggal_kembali: e.target.value }))} 
              />
            </div>
          </div>
          
          <Separator />
          
          {/* AKOMODASI / HOTEL */}
          <div>
            <h4 className="text-sm font-bold mb-3">AKOMODASI / HOTEL (Opsional)</h4>
            <div className="grid grid-cols-12 gap-3">
              <div className="col-span-12 sm:col-span-6"><Label className="text-xs">Nama Hotel</Label><Input placeholder="Nama hotel (kosongkan jika tidak ada)" value={bonForm.rincian_hotel.nama_hotel} onChange={e => setBonForm((p: any) => ({ ...p, rincian_hotel: { ...p.rincian_hotel, nama_hotel: e.target.value } }))} className="h-9" /></div>
              <div className="col-span-12 sm:col-span-4"><Label className="text-xs">Check-In</Label><Input type="date" value={bonForm.rincian_hotel.check_in} onChange={e => setBonForm((p: any) => ({ ...p, rincian_hotel: { ...p.rincian_hotel, check_in: e.target.value } }))} className="h-9" /></div>
              <div className="col-span-12 sm:col-span-4"><Label className="text-xs">Check-Out</Label><Input type="date" value={bonForm.rincian_hotel.check_out} onChange={e => setBonForm((p: any) => ({ ...p, rincian_hotel: { ...p.rincian_hotel, check_out: e.target.value } }))} className="h-9" /></div>
              <div className="col-span-12 sm:col-span-4"><Label className="text-xs">Harga (Rp)</Label><Input type="number" value={bonForm.rincian_hotel.harga || ""} onChange={e => setBonForm((p: any) => ({ ...p, rincian_hotel: { ...p.rincian_hotel, harga: parseFloat(e.target.value) || 0 } }))} className="h-9" /></div>
            </div>
          </div>

          <Separator />

          {/* TRANSPORTASI */}
          <div>
            <div className="flex items-center justify-between mb-3">
              <h4 className="text-sm font-bold">TRANSPORTASI (Opsional)</h4>
              <Button variant="outline" size="sm" onClick={addTransportItem} className="h-7 gap-1"><Plus className="h-3 w-3" />Tambah Rute</Button>
            </div>
            <div className="space-y-2">
              {bonForm.rincian_transportasi.map((item: any, idx: number) => (
                <div key={idx} className="grid grid-cols-12 gap-2 items-end border border-slate-200 rounded-lg p-3">
                  <div className="col-span-12 sm:col-span-4"><Label className="text-xs">Tipe Perjalanan</Label><Select value={item.tipe_perjalanan} onValueChange={v => updateTransportItem(idx, "tipe_perjalanan", v)}><SelectTrigger className="h-9"><SelectValue /></SelectTrigger><SelectContent><SelectItem value="Keberangkatan">Keberangkatan</SelectItem><SelectItem value="Kedatangan">Kedatangan</SelectItem></SelectContent></Select></div>
                  <div className="col-span-12 sm:col-span-4"><Label className="text-xs">Jenis Angkutan</Label><Input placeholder="Pesawat / Kereta dll" value={item.jenis_transportasi} onChange={e => updateTransportItem(idx, "jenis_transportasi", e.target.value)} className="h-9" /></div>
                  <div className="col-span-12 sm:col-span-4"><Label className="text-xs">Harga (Rp)</Label><Input type="number" value={item.harga || ""} onChange={e => updateTransportItem(idx, "harga", e.target.value)} className="h-9" /></div>
                  
                  <div className="col-span-12 sm:col-span-4 mt-2"><Label className="text-xs">Kota Asal</Label><Input placeholder="Dari mana" value={item.kota_asal} onChange={e => updateTransportItem(idx, "kota_asal", e.target.value)} className="h-9" /></div>
                  <div className="col-span-12 sm:col-span-4 mt-2"><Label className="text-xs">Kota Tujuan</Label><Input placeholder="Ke mana" value={item.kota_tujuan} onChange={e => updateTransportItem(idx, "kota_tujuan", e.target.value)} className="h-9" /></div>
                  <div className="col-span-12 sm:col-span-4 mt-2"><Label className="text-xs">Jam Keberangkatan</Label><Input type="time" value={item.jam_berangkat} onChange={e => updateTransportItem(idx, "jam_berangkat", e.target.value)} className="h-9" /></div>
                  
                  <div className="col-span-12 flex justify-end mt-2">
                    <Button variant="ghost" size="sm" className="h-7 gap-1 text-red-500" onClick={() => removeTransportItem(idx)}><Trash2 className="h-3 w-3" />Hapus Rute</Button>
                  </div>
                </div>
              ))}
            </div>
          </div>

          <Separator />

          {/* BIAYA LAINNYA */}
          <div className="flex items-center justify-between"><h4 className="text-sm font-bold">ESTIMASI BIAYA LAINNYA</h4><Button variant="outline" size="sm" onClick={addEstimasiItem} className="h-7 gap-1"><Plus className="h-3 w-3" />Tambah</Button></div>
          <div className="space-y-2">
            {bonForm.rincian_tambahan.map((item: any, idx: number) => (
              <div key={idx} className="grid grid-cols-12 gap-2 items-end border border-slate-200 rounded-lg p-3">
                <div className="col-span-12 sm:col-span-3"><Label className="text-xs">Kategori</Label>
                <Select value={item.kategori} onValueChange={v => updateEstimasiItem(idx, "kategori", v)}>
                  <SelectTrigger className="h-9"><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="Konsumsi">Konsumsi</SelectItem>
                    <SelectItem value="Transportasi">Transportasi</SelectItem>
                    <SelectItem value="BBM">BBM</SelectItem>
                    <SelectItem value="Entertainment">Entertainment</SelectItem>
                    <SelectItem value="Parkir">Parkir</SelectItem>
                    <SelectItem value="Tol">Tol</SelectItem>
                    <SelectItem value="Lain-lain">Lain-lain</SelectItem>
                  </SelectContent>
                </Select></div>
                <div className="col-span-12 sm:col-span-4"><Label className="text-xs">Keterangan</Label><Input placeholder="Keterangan" value={item.keterangan} onChange={e => updateEstimasiItem(idx, "keterangan", e.target.value)} className="h-9" /></div>
                <div className="col-span-6 sm:col-span-2"><Label className="text-xs">Qty</Label><Input type="number" value={item.kuantitas || ""} onChange={e => updateEstimasiItem(idx, "kuantitas", e.target.value)} className="h-9" /></div>
                <div className="col-span-6 sm:col-span-3"><Label className="text-xs">Harga (Rp)</Label><Input type="number" value={item.harga || ""} onChange={e => updateEstimasiItem(idx, "harga", e.target.value)} className="h-9" /></div>
                {bonForm.rincian_tambahan.length > 1 && <div className="col-span-12 flex justify-end mt-2"><Button variant="ghost" size="sm" className="h-7 gap-1 text-red-500" onClick={() => removeEstimasiItem(idx)}><Trash2 className="h-3 w-3" />Hapus</Button></div>}
              </div>
            ))}
          </div>
          <div className="bg-slate-50 p-3 rounded-lg text-right"><p className="text-sm font-bold text-slate-900">TOTAL ESTIMASI KESELURUHAN: {fmt(calcEstimasiTotal())}</p></div>
        </div>
        <DialogFooter><Button variant="outline" onClick={onClose}>Batal</Button><Button onClick={submitBon} disabled={loading} className="bg-slate-900 hover:bg-slate-800">{loading ? "Mengirim..." : "Ajukan Perjalanan"}</Button></DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

// ===================== CREATE REALISASI DIALOG =====================
function CreateRealisasiDialog({ open, onClose, onSuccess }: { open: boolean; onClose: () => void; onSuccess: () => void; }) {
  const [loading, setLoading] = useState(false);
  const [loadingItems, setLoadingItems] = useState(false);
  const [uploadingIdx, setUploadingIdx] = useState<number | null>(null);
  const [options, setOptions] = useState<any[]>([]);
  
  const [realForm, setRealForm] = useState<any>({
    request_ppd_id: "", 
    periode_berangkat: "", 
    periode_kembali: "", 
    bukti_transfer: "", 
    items: [{ 
      tanggal: "", 
      kategori: "Konsumsi", 
      uraian: "", 
      quantity: 1, 
      harga_per_unit: 0, 
      total: 0, 
      bukti: "" 
    }]
  });

  useEffect(() => {
    if (open) {
      api.get("/rbs/options")
        .then((res) => setOptions(res.data?.data || res.data || []))
        .catch(() => toast.error("Gagal memuat daftar referensi bon"));
    } else {
      setRealForm({ 
        request_ppd_id: "", 
        periode_berangkat: "", 
        periode_kembali: "", 
        bukti_transfer: "", 
        items: [{ 
          tanggal: "", 
          kategori: "Konsumsi", 
          uraian: "", 
          quantity: 1, 
          harga_per_unit: 0, 
          total: 0, 
          bukti: "" 
        }] 
      });
    }
  }, [open]);

  const handleBonChange = async (val: string) => {
    setRealForm((p: any) => ({ ...p, request_ppd_id: val }));
    setLoadingItems(true);
    try {
      const res = await api.get(`/ppd/${val}/item`);
      const responseData = res.data?.data || {};
      const fetchedItems = responseData.items || [];
      
      if (Array.isArray(fetchedItems) && fetchedItems.length > 0) {
        const mappedItems = fetchedItems.map((item: any) => ({
          tanggal: "", 
          kategori: item.kategori || "Konsumsi", 
          uraian: item.uraian || item.keterangan || "", 
          quantity: item.qty || item.kuantitas || item.quantity || 1, 
          harga_per_unit: 0, 
          total: 0, 
          bukti: "" 
        }));
        setRealForm((p: any) => ({ ...p, items: mappedItems }));
      } else {
        setRealForm((p: any) => ({ 
          ...p, 
          items: [{ tanggal: "", kategori: "Konsumsi", uraian: "", quantity: 1, harga_per_unit: 0, total: 0, bukti: "" }] 
        }));
      }
    } catch (err) { 
      toast.error("Gagal mengambil rincian item"); 
    } finally { 
      setLoadingItems(false); 
    }
  };

  const handleRealBerangkatChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value;
    setRealForm((p: any) => {
      const updates: any = { periode_berangkat: val };
      if (p.periode_kembali && val > p.periode_kembali) {
        updates.periode_kembali = "";
      }
      return { ...p, ...updates };
    });
  };

  const updateRealItem = (idx: number, field: string, val: string | number) => {
    setRealForm((p: any) => {
      const items = [...p.items]; 
      items[idx] = { ...items[idx], [field]: val };
      if (field === "quantity" || field === "harga_per_unit") { 
        items[idx].total = (parseFloat(items[idx].quantity) || 0) * (parseFloat(items[idx].harga_per_unit) || 0); 
      }
      return { ...p, items };
    });
  };

  const handleFileUpload = async (idx: number, e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    if (file.size > 2 * 1024 * 1024) {
      toast.error("Ukuran file maksimal 2 MB");
      e.target.value = '';
      return;
    }

    setUploadingIdx(idx);
    const formData = new FormData();
    formData.append("struk", file);

    try {
      const res = await api.post("/upload/struk", formData, { headers: { "Content-Type": "multipart/form-data" } });
      const fileUrl = res.data?.data?.url || res.data?.data || res.data?.url || res.data;
      
      setRealForm((p: any) => {
        const items = [...p.items]; 
        items[idx] = { ...items[idx], bukti: fileUrl };
        return { ...p, items };
      });
      toast.success("Struk berhasil diunggah");
    } catch (err) {
      toast.error("Gagal mengunggah struk");
    } finally {
      setUploadingIdx(null);
    }
  };

  const handleTransferUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    if (file.size > 2 * 1024 * 1024) {
      toast.error("Ukuran file maksimal 2 MB");
      e.target.value = '';
      return;
    }

    const formData = new FormData();
    formData.append("struk", file);

    try {
      const res = await api.post("/upload/struk", formData, { headers: { "Content-Type": "multipart/form-data" } });
      const fileUrl = res.data?.data?.url || res.data?.data || res.data?.url || res.data;
      setRealForm((p: any) => ({ ...p, bukti_transfer: fileUrl }));
      toast.success("Bukti transfer berhasil diunggah");
    } catch (err) {
      toast.error("Gagal mengunggah bukti transfer");
    }
  };

  const addRealItem = () => setRealForm((p: any) => ({ 
    ...p, 
    items: [...p.items, { tanggal: "", kategori: "Konsumsi", uraian: "", quantity: 1, harga_per_unit: 0, total: 0, bukti: "" }] 
  }));
  
  const removeRealItem = (idx: number) => setRealForm((p: any) => ({ 
    ...p, 
    items: p.items.filter((_: any, i: number) => i !== idx) 
  }));

  const submitRealisasi = async () => {
    if (!realForm.request_ppd_id || !realForm.periode_berangkat || !realForm.periode_kembali || realForm.items.length === 0) { 
      toast.error("Mohon lengkapi referensi, periode keberangkatan, dan kedatangan"); return; 
    }

    if (realForm.items.some((i: any) => !i.quantity || parseFloat(i.quantity) <= 0 || !i.harga_per_unit || parseFloat(i.harga_per_unit) <= 0)) {
      toast.error("Qty dan Harga/Unit pada rincian aktual tidak boleh 0 atau kosong"); return;
    }

    setLoading(true);
    try {
      const selectedBon = options.find(o => o.id.toString() === realForm.request_ppd_id);
      const nomorBonSementara = selectedBon ? (selectedBon.nomor_tipe_dokumen || selectedBon.nomor_dokumen) : "";
      
      const totalRealisasi = realForm.items.reduce((sum: number, i: any) => sum + (parseFloat(i.total) || 0), 0);
      const estimasiAwal = selectedBon ? (parseFloat(selectedBon.total_estimasi) || 0) : 0;
      const selisih = estimasiAwal - totalRealisasi;

      const payload = {
        request_ppd_id: parseInt(realForm.request_ppd_id),
        total_realisasi: totalRealisasi,
        selisih: selisih,
        periode_berangkat: `${realForm.periode_berangkat}T00:00:00Z`,
        periode_kembali: `${realForm.periode_kembali}T00:00:00Z`,
        nomor_bon_sementara: nomorBonSementara,
        url_bukti_transfer: realForm.bukti_transfer || null,
        items: realForm.items.map((i: any) => ({
          tanggal: i.tanggal ? i.tanggal : null,
          kategori: i.kategori,
          uraian: i.uraian,
          kuantitas: parseInt(i.quantity) || 1,
          harga_unit: parseFloat(i.harga_per_unit) || 0,
          total: parseFloat(i.total) || 0,
          url_struk: i.bukti || ""
        }))
      };

      await api.post("/rbs", payload);
      toast.success("Realisasi berhasil diajukan"); 
      onSuccess(); 
      onClose();
    } catch (err: any) { 
      toast.error(err.response?.data?.message || err.response?.data?.error || "Gagal mengajukan realisasi"); 
    } finally { 
      setLoading(false); 
    }
  };

  const totalRealisasi = realForm.items.reduce((sum: number, i: any) => sum + (parseFloat(i.total) || 0), 0);

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="w-[95vw] max-w-full sm:max-w-5xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Realisasi Bon Sementara</DialogTitle>
          <DialogDescription>Rekap pengeluaran aktual berdasarkan rincian bon sementara</DialogDescription>
        </DialogHeader>
        
        <div className="space-y-4 py-2">
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <div className="sm:col-span-1">
              <Label>Ref. No. Bon Sementara *</Label>
              <Select value={realForm.request_ppd_id} onValueChange={handleBonChange}>
                <SelectTrigger className="w-full"><SelectValue placeholder="Pilih Bon Sementara" /></SelectTrigger>
                <SelectContent>
                  {options.map(opt => (
                    <SelectItem key={opt.id} value={opt.id.toString()}>
                      {opt.nomor_tipe_dokumen || opt.nomor_dokumen} - {opt.tujuan}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div>
              <Label>Berangkat *</Label>
              <Input type="date" value={realForm.periode_berangkat} onChange={handleRealBerangkatChange} />
            </div>
            <div>
              <Label>Kembali *</Label>
              <Input type="date" value={realForm.periode_kembali} min={realForm.periode_berangkat} onChange={e => setRealForm((p: any) => ({ ...p, periode_kembali: e.target.value }))} />
            </div>
          </div>

          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <Label className="text-sm font-bold">Rincian Biaya Aktual</Label>
              <Button variant="outline" size="sm" onClick={addRealItem} className="h-7 gap-1">
                <Plus className="h-3 w-3" /> Tambah
              </Button>
            </div>
            
            {loadingItems ? (
              <div className="text-center py-6 text-slate-500 bg-slate-50 rounded-lg border border-dashed border-slate-200">
                Memuat rincian item...
              </div>
            ) : (
              <div className="space-y-4">
                {realForm.items.map((item: any, idx: number) => (
                  <div key={idx} className="flex flex-col sm:grid sm:grid-cols-12 gap-3 p-3 sm:p-4 border border-slate-200 rounded-xl bg-white shadow-sm relative">
                    
                    {/* Header Mobile - Tombol hapus untuk layar kecil */}
                    <div className="flex justify-between items-center sm:hidden mb-1">
                      <span className="text-xs font-bold text-slate-500 uppercase tracking-wider">Item #{idx + 1}</span>
                      {realForm.items.length > 1 && (
                        <Button variant="ghost" size="sm" className="h-7 w-7 p-0 text-red-500 bg-red-50 rounded-md" onClick={() => removeRealItem(idx)}>
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      )}
                    </div>

                    <div className="sm:col-span-2">
                      <Label className="text-xs text-slate-500 mb-1.5 block">Tanggal</Label>
                      <Input type="date" className="h-9 text-xs" value={item.tanggal} onChange={e => updateRealItem(idx, "tanggal", e.target.value)} />
                    </div>
                    
                    <div className="sm:col-span-2">
                      <Label className="text-xs text-slate-500 mb-1.5 block">Kategori</Label>
                      <Select value={item.kategori} onValueChange={v => updateRealItem(idx, "kategori", v)}>
                        <SelectTrigger className="h-9 text-xs"><SelectValue /></SelectTrigger>
                        <SelectContent>
                          <SelectItem value="Konsumsi">Konsumsi</SelectItem>
                          <SelectItem value="Transportasi">Transportasi</SelectItem>
                          <SelectItem value="Akomodasi">Akomodasi</SelectItem>
                          <SelectItem value="BBM">BBM</SelectItem>
                          <SelectItem value="Entertainment">Entertainment</SelectItem>
                          <SelectItem value="Parkir">Parkir</SelectItem>
                          <SelectItem value="Tol">Tol</SelectItem>
                          <SelectItem value="Lain-lain">Lain-lain</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                    
                    <div className="sm:col-span-3">
                      <Label className="text-xs text-slate-500 mb-1.5 block">Uraian</Label>
                      <Input placeholder="Detail pengeluaran" className="h-9 text-xs" value={item.uraian} onChange={e => updateRealItem(idx, "uraian", e.target.value)} />
                    </div>
                    
                    <div className="grid grid-cols-2 gap-3 sm:col-span-2">
                      <div>
                        <Label className="text-xs text-slate-500 mb-1.5 block">Qty</Label>
                        <Input type="number" className="h-9 text-xs" value={item.quantity ?? ""} onChange={e => updateRealItem(idx, "quantity", e.target.value)} />
                      </div>
                      <div>
                        <Label className="text-xs text-slate-500 mb-1.5 block">Harga/Unit</Label>
                        <Input type="number" className="h-9 text-xs" value={item.harga_per_unit ?? ""} onChange={e => updateRealItem(idx, "harga_per_unit", e.target.value)} />
                      </div>
                    </div>
                    
                    <div className="sm:col-span-2">
                      <Label className="text-xs text-slate-500 mb-1.5 block">Total</Label>
                      <Input disabled className="h-9 text-xs bg-slate-50 font-semibold text-slate-900" value={fmt(item.total)} />
                    </div>
                    
                    <div className="sm:col-span-1 flex flex-row sm:flex-col items-end justify-between sm:justify-end gap-2 mt-2 sm:mt-0">
                      <div className="flex-1 sm:w-full">
                        <Label className="text-xs text-slate-500 mb-1.5 sm:hidden block">Struk / Bukti</Label>
                        <label className={`flex items-center justify-center h-9 w-full sm:w-9 border border-slate-300 rounded-md hover:bg-slate-50 transition-colors cursor-pointer ${uploadingIdx === idx ? 'opacity-50' : ''} ${item.bukti ? 'bg-emerald-50 border-emerald-200' : ''}`}>
                          {uploadingIdx === idx ? <Loader2 className="h-4 w-4 animate-spin text-slate-400" /> : <UploadCloud className={`h-4 w-4 ${item.bukti ? 'text-emerald-600' : 'text-slate-500'}`} />}
                          <span className={`ml-2 text-xs sm:hidden font-medium ${item.bukti ? 'text-emerald-700' : 'text-slate-600'}`}>{item.bukti ? 'Terupload' : 'Upload File'}</span>
                          <input type="file" accept="image/*,.pdf" className="hidden" disabled={uploadingIdx === idx} onChange={e => handleFileUpload(idx, e)} />
                        </label>
                      </div>
                      
                      {realForm.items.length > 1 && (
                        <Button variant="ghost" size="icon" className="hidden sm:flex h-9 w-9 text-red-500 hover:bg-red-50" onClick={() => removeRealItem(idx)}>
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}
            <div className="bg-slate-50 p-3 rounded-lg flex justify-between items-center mt-4">
              <span className="text-sm text-slate-600 font-medium">Total Realisasi Aktual:</span>
              <span className="text-base font-bold text-slate-900">{fmt(totalRealisasi)}</span>
            </div>
          </div>
          
          <div>
            <Label>Upload Bukti Transfer Sisa Dana (Opsional)</Label>
            <label className="mt-1 border-2 border-dashed border-slate-300 rounded-lg p-4 text-center hover:bg-slate-50 transition cursor-pointer flex flex-col items-center gap-1">
              <UploadCloud className={`h-5 w-5 ${realForm.bukti_transfer ? 'text-emerald-500' : 'text-slate-400'}`} />
              <p className="text-xs text-slate-500">{realForm.bukti_transfer ? "✓ Bukti berhasil dipilih" : "Klik untuk upload bukti transfer (Maks 2MB)"}</p>
              <input type="file" accept="image/*,.pdf" className="hidden" onChange={handleTransferUpload} />
            </label>
          </div>
        </div>
        
        <DialogFooter className="flex-col sm:flex-row gap-2">
          <Button variant="outline" onClick={onClose} className="w-full sm:w-auto">Batal</Button>
          <Button onClick={submitRealisasi} disabled={loading} className="w-full sm:w-auto bg-emerald-600 hover:bg-emerald-700 text-white">
            {loading ? "Menyimpan..." : "Simpan Realisasi"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

// ===================== UNIVERSAL VIEW (Table & Logic) =====================
function UniversalView({ userRole }: { userRole: string }) {
  const isPegawai = userRole === "pegawai";
  const [tab, setTab] = useState<string>(isPegawai ? "bon" : "approval");
  const [loading, setLoading] = useState(true);
  
  const [historyBons, setHistoryBons] = useState<any[]>([]);
  const [ppdHistPage, setPpdHistPage] = useState(1);
  const [ppdHistMeta, setPpdHistMeta] = useState({ page: 1, total_page: 1, total_data: 0 });

  const [historyRealisasi, setHistoryRealisasi] = useState<any[]>([]);
  const [rbsHistPage, setRbsHistPage] = useState(1);
  const [rbsHistMeta, setRbsHistMeta] = useState({ page: 1, total_page: 1, total_data: 0 });

  const [pendingBons, setPendingBons] = useState<any[]>([]);
  const [ppdPendPage, setPpdPendPage] = useState(1);
  const [ppdPendMeta, setPpdPendMeta] = useState({ page: 1, total_page: 1, total_data: 0 });

  const [pendingRealisasi, setPendingRealisasi] = useState<any[]>([]);
  const [rbsPendPage, setRbsPendPage] = useState(1);
  const [rbsPendMeta, setRbsPendMeta] = useState({ page: 1, total_page: 1, total_data: 0 });
  
  const [showCreateBon, setShowCreateBon] = useState(false);
  const [showCreateReal, setShowCreateReal] = useState(false);
  const [showDecline, setShowDecline] = useState<any | null>(null);
  const [showResubmit, setShowResubmit] = useState<any | null>(null);
  const [showDetail, setShowDetail] = useState<any | null>(null);
  const [declineReason, setDeclineReason] = useState("");
  const [viewImage, setViewImage] = useState<string | null>(null);

  // EDIT STATE
  const [isEditMode, setIsEditMode] = useState(false);
  const [editForm, setEditForm] = useState<any>(null);
  const [isSaving, setIsSaving] = useState(false);
  const [editClickCount, setEditClickCount] = useState(0);

  const [filterMonth, setFilterMonth] = useState("");
  const [filterYear, setFilterYear] = useState("");

  const fetchHistory = useCallback(async () => {
    setLoading(true);
    try {
      const paramsPpd = { page: ppdHistPage, limit: 10 };
      const paramsRbs = { page: rbsHistPage, limit: 10, ...(userRole === "hrga" && { month: filterMonth, year: filterYear }) };

      const [ppdRes, rbsRes] = await Promise.all([
        api.get("/ppd", { params: paramsPpd }).catch(() => ({ data: { data: [], meta: {} } })),
        api.get("/rbs", { params: paramsRbs }).catch(() => ({ data: { data: [], meta: {} } }))
      ]);

      setHistoryBons(ppdRes.data?.data || []);
      setPpdHistMeta(ppdRes.data?.meta || { page: 1, total_page: 1, total_data: 0 });

      setHistoryRealisasi(rbsRes.data?.data || []);
      setRbsHistMeta(rbsRes.data?.meta || { page: 1, total_page: 1, total_data: 0 });
    } finally { setLoading(false); }
  }, [ppdHistPage, rbsHistPage, filterMonth, filterYear, userRole]);

  const fetchPending = useCallback(async () => {
    if (isPegawai) return;
    try {
      const [ppdRes, rbsRes] = await Promise.all([
        api.get("/ppd/pending", { params: { page: ppdPendPage, limit: 10 } }).catch(() => ({ data: { data: [], meta: {} } })),
        api.get("/rbs/pending", { params: { page: rbsPendPage, limit: 10 } }).catch(() => ({ data: { data: [], meta: {} } }))
      ]);

      setPendingBons(ppdRes.data?.data || []);
      setPpdPendMeta(ppdRes.data?.meta || { page: 1, total_page: 1, total_data: 0 });

      setPendingRealisasi(rbsRes.data?.data || []);
      setRbsPendMeta(rbsRes.data?.meta || { page: 1, total_page: 1, total_data: 0 });
    } catch (err) {}
  }, [isPegawai, ppdPendPage, rbsPendPage]);

  const refetchAll = () => { fetchHistory(); fetchPending(); };
  useEffect(() => { fetchHistory(); }, [fetchHistory]);
  useEffect(() => { fetchPending(); }, [fetchPending]);

  const openDetail = async (row: any, type: "ppd" | "rbs") => {
    setShowDetail({ ...row, type, isLoading: true });
    setIsEditMode(false);
    setEditClickCount(0);
    try {
      const res = await api.get(`/${type}/${row.id}`);
      const data = res.data?.data || res.data;
      
      if (type === "rbs" && Array.isArray(data)) {
        setShowDetail({ ...row, items: data, type, isLoading: false });
      } else {
        setShowDetail({ ...row, ...data, type, isLoading: false });
        if (type === "ppd") {
          setEditForm({
            ...data,
            rincian_tambahan: data.rincian_tambahan || [],
            rincian_transportasi: data.rincian_transportasi || [],
            rincian_hotel: data.rincian_hotel || { nama_hotel: "", check_in: "", check_out: "", harga: 0 }
          });
        }
      }
    } catch { toast.error(`Gagal memuat detail`); setShowDetail(null); }
  };

  const approve = async (id: number, type: "ppd" | "rbs") => {
    try { await api.patch(`/${type}/${id}/approve`, {}); toast.success("Disetujui"); setShowDetail(null); refetchAll(); } catch { toast.error("Gagal menyetujui dokumen"); }
  };
  
  const decline = async () => {
    if (!declineReason.trim()) { toast.error("Alasan wajib diisi"); return; }
    try { await api.patch(`/${showDecline.type}/${showDecline.id}/decline`, { catatan: declineReason }); toast.success("Ditolak"); setShowDecline(null); setShowDetail(null); setDeclineReason(""); refetchAll(); } catch { toast.error("Gagal menolak dokumen"); }
  };
  
  const handleResubmit = async (id: number, type: string) => {
    try { await api.patch(`/${type}/${id}/resubmit`); toast.success("Diajukan ulang"); setShowResubmit(null); refetchAll(); } catch { toast.error("Gagal mengajukan ulang"); }
  };
  
  const downloadPdf = async (id: number, type: "ppd" | "rbs", name: string, pdfType: "rpd" | "bs" = "rpd") => {
    try {
      const token = localStorage.getItem("token");
      const endpoint = type === "ppd" ? (pdfType === "bs" ? `/ppd/${id}/download/bs` : `/ppd/${id}/download`) : `/rbs/${id}/download`;
      const res = await fetch(`${BACKEND_URL}/api/v1${endpoint}`, { headers: { Authorization: `Bearer ${token}` } });
      if (!res.ok) throw new Error("Gagal mengunduh file");
      const url = window.URL.createObjectURL(await res.blob()); 
      const a = document.createElement("a"); 
      a.href = url; 
      a.download = `${name}.pdf`; 
      a.click(); 
      window.URL.revokeObjectURL(url);
    } catch { toast.error("Gagal mendownload PDF"); }
  };
  
  const downloadExcel = async () => {
    try {
      const token = localStorage.getItem("token");
      const params = new URLSearchParams();
      if (filterMonth) params.append('month', filterMonth);
      if (filterYear) params.append('year', filterYear);
      
      const res = await fetch(`${BACKEND_URL}/api/v1/rbs/download/excel?${params.toString()}`, { headers: { Authorization: `Bearer ${token}` }});
      if (!res.ok) throw new Error("Failed to download");
      
      // Ambil filename dari Content-Disposition header jika tersedia, fallback dengan format dari backend
      let filename = "Rekap_Realisasi_Bon_Sementara_Semua.xlsx";
      const disposition = res.headers.get('content-disposition');
      if (disposition && disposition.includes('filename=')) {
        const matches = disposition.match(/filename="?([^"]+)"?/);
        if (matches && matches[1]) {
           filename = matches[1];
        }
      } else {
        if (filterMonth && filterYear) {
          filename = `Rekap_Realisasi_Bon_Sementara_${filterMonth.padStart(2, '0')}_${filterYear}.xlsx`;
        } else if (filterYear) {
          filename = `Rekap_Realisasi_Bon_Sementara_${filterYear}.xlsx`;
        }
      }

      const blob = await res.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url; 
      a.download = filename; 
      a.click();
      window.URL.revokeObjectURL(url);
    } catch { toast.error("Gagal download Excel"); }
  };

  // --- LOGIKA EDIT PPD (HRGA) ---
  const saveEdit = async () => {
    if (!editForm) return;

    if (editClickCount >= 5) {
      toast.error("Batas percobaan klik tercapai (Maks 5x). Batalkan edit lalu coba lagi.");
      return;
    }

    if ((editForm.rincian_tambahan || []).some((i: any) => !(i.kuantitas ?? i.qty) || parseFloat(i.kuantitas ?? i.qty) <= 0 || !(i.harga ?? i.harga_unit) || parseFloat(i.harga ?? i.harga_unit) <= 0)) {
      toast.error("Kuantitas dan Harga pada estimasi biaya tidak boleh 0 atau kosong"); return;
    }

    if ((editForm.rincian_transportasi || []).some((t: any) => !t.harga || parseFloat(t.harga) <= 0)) {
      toast.error("Harga pada rincian transportasi tidak boleh 0 atau kosong"); return;
    }

    setIsSaving(true);
    setEditClickCount((prev) => prev + 1);

    try {
      let totalHargaHotel = 0;
      if (editForm.rincian_hotel?.check_in && editForm.rincian_hotel?.check_out && editForm.rincian_hotel?.harga) {
        const checkin = new Date(editForm.rincian_hotel.check_in);
        const checkout = new Date(editForm.rincian_hotel.check_out);
        const nights = Math.ceil((checkout.getTime() - checkin.getTime()) / (1000 * 60 * 60 * 24));
        totalHargaHotel = (parseFloat(editForm.rincian_hotel.harga) || 0) * (nights > 0 ? nights : 1);
      }

      const rincianTambahan = (editForm.rincian_tambahan || []).map((item: any) => {
        const data: any = {
          kategori: item.kategori || "Lainnya",
          keterangan: item.keterangan || item.uraian,
          kuantitas: parseInt(item.kuantitas || item.qty) || 1,
          harga: parseFloat(item.harga || item.harga_unit) || 0
        };
        if (item.id) data.id = item.id;
        else if (item.ID) data.id = item.ID;

        return data;
      });

      const rincianTransport = (editForm.rincian_transportasi || []).map((t: any) => {
          let jamBerangkat = t.jam_berangkat;
          if (jamBerangkat && jamBerangkat.length === 5) {
                jamBerangkat = `${showDetail.tanggal_berangkat?.slice(0,10)}T${jamBerangkat}:00Z`;
          }
          const data: any = {
            kota_asal: t.kota_asal,
            kota_tujuan: t.kota_tujuan,
            jenis_transportasi: t.jenis_transportasi,
            tipe_perjalanan: t.tipe_perjalanan,
            jam_berangkat: jamBerangkat,
            harga: parseFloat(t.harga) || 0
          };
          if (t.id) data.id = t.id;
          else if (t.ID) data.id = t.ID;

          return data;
      });

      let hotelData = null;
      if (editForm.rincian_hotel?.nama_hotel) {
         hotelData = {
            ...editForm.rincian_hotel, 
            harga: parseFloat(editForm.rincian_hotel.harga) || 0, 
            total_harga: totalHargaHotel 
         };
         if (editForm.rincian_hotel.id) hotelData.id = editForm.rincian_hotel.id;
         else if (editForm.rincian_hotel.ID) hotelData.id = editForm.rincian_hotel.ID;
      }

      const payload = {
        tujuan: showDetail.tujuan,
        tanggal_berangkat: showDetail.tanggal_berangkat,
        tanggal_kembali: showDetail.tanggal_kembali,
        keperluan: showDetail.keperluan,
        rincian_hotel: hotelData,
        rincian_transportasi: rincianTransport,
        rincian_tambahan: rincianTambahan
      };

      await api.put(`/ppd/${showDetail.id}/edit`, payload);
      toast.success("Perubahan data berhasil disimpan");
      setIsEditMode(false);
      setEditClickCount(0);
      openDetail(showDetail, "ppd");
      refetchAll(); 
    } catch (error: any) {
      toast.error(error.response?.data?.message || "Gagal menyimpan perubahan");
    } finally {
      setIsSaving(false);
    }
  };

  const updateEditFormEstimasi = (idx: number, field: string, value: any) => {
      setEditForm((prev: any) => {
          const newItems = [...prev.rincian_tambahan];
          newItems[idx] = { ...newItems[idx], [field]: value };
          return { ...prev, rincian_tambahan: newItems };
      });
  };
  
  const updateEditFormTransport = (idx: number, field: string, value: any) => {
    setEditForm((prev: any) => {
        const newItems = [...prev.rincian_transportasi];
        newItems[idx] = { ...newItems[idx], [field]: value };
        return { ...prev, rincian_transportasi: newItems };
    });
  };

  const years = []; const currentYear = new Date().getFullYear(); for (let y = currentYear; y >= currentYear - 5; y--) years.push(y);
  const months = [{ value: "1", label: "Januari" }, { value: "2", label: "Februari" }, { value: "3", label: "Maret" }, { value: "4", label: "April" }, { value: "5", label: "Mei" }, { value: "6", label: "Juni" }, { value: "7", label: "Juli" }, { value: "8", label: "Agustus" }, { value: "9", label: "September" }, { value: "10", label: "Oktober" }, { value: "11", label: "November" }, { value: "12", label: "Desember" }];

  return (
    <div className="space-y-6">
      <Tabs value={tab} onValueChange={setTab}>
        <div className="flex flex-col md:flex-row items-start md:items-center justify-between gap-4 mb-4">
          <div className="w-full md:w-auto overflow-x-auto pb-2 -mb-2 no-scrollbar">
            <TabsList className="inline-flex min-w-max">
              {!isPegawai && <TabsTrigger value="approval">Menunggu Approval</TabsTrigger>}
              <TabsTrigger value="bon">{isPegawai ? "Perjalanan Dinas" : "Riwayat Pengajuan"}</TabsTrigger>
              <TabsTrigger value="realisasi">Realisasi Bon</TabsTrigger>
            </TabsList>
          </div>
          <div className="flex shrink-0 gap-2 w-full md:w-auto justify-start md:justify-end">
            {tab === "bon" && (<Button onClick={() => setShowCreateBon(true)} className="w-full md:w-auto bg-slate-900 hover:bg-slate-800 gap-2"><Plus className="h-4 w-4" /> Ajukan Perjalanan Dinas</Button>)}
            {tab === "realisasi" && ["pegawai", "atasan", "hrga"].includes(userRole) && (<Button onClick={() => setShowCreateReal(true)} className="w-full md:w-auto bg-slate-900 hover:bg-slate-800 gap-2"><Plus className="h-4 w-4" /> Buat Realisasi</Button>)}
          </div>
        </div>

        {/* TAB: APPROVAL */}
        {!isPegawai && (
          <TabsContent value="approval" className="space-y-6">
            <Card className="border-slate-100 shadow-sm rounded-xl">
              <CardHeader className="pb-3"><CardTitle className="text-base font-semibold flex items-center gap-2">Persetujuan Perjalanan Dinas {ppdPendMeta.total_data > 0 && <Badge className="bg-amber-100 text-amber-700 border-amber-200">{ppdPendMeta.total_data}</Badge>}</CardTitle></CardHeader>
              <CardContent className="p-0">
                <div className="overflow-x-auto">
                  <Table className="min-w-[800px]">
                    <TableHeader><TableRow className="bg-slate-50/50"><TableHead>No. Dokumen</TableHead><TableHead>Pemohon</TableHead><TableHead>Tujuan</TableHead><TableHead className="text-right">Estimasi</TableHead><TableHead className="text-center">Aksi</TableHead></TableRow></TableHeader>
                    <TableBody>
                      {loading ? (<TableRow><TableCell colSpan={5} className="text-center py-10">Memuat...</TableCell></TableRow>) : pendingBons.length === 0 ? (<TableRow><TableCell colSpan={5} className="text-center py-10 text-slate-400">Kosong</TableCell></TableRow>) : (
                        pendingBons.map(b => (
                          <TableRow key={b.id} className="cursor-pointer hover:bg-slate-50" onClick={() => openDetail(b, "ppd")}>
                            <TableCell className="font-mono text-xs">{b.nomor_tipe_dokumen || b.nomor_dokumen || "-"}</TableCell><TableCell>{b.nama}</TableCell><TableCell>{b.tujuan}</TableCell><TableCell className="text-right font-medium">{fmt(b.total_estimasi || 0)}</TableCell>
                            <TableCell onClick={e => e.stopPropagation()}>
                              <div className="flex justify-center gap-1">
                                <Button size="sm" className="h-7 bg-emerald-600 hover:bg-emerald-700" onClick={() => approve(b.id, "ppd")}><CheckCircle2 className="h-3 w-3 mr-1" />Setujui</Button>
                                <Button size="sm" variant="outline" className="h-7 text-red-600" onClick={() => setShowDecline({ ...b, type: "ppd" })}><XCircle className="h-3 w-3 mr-1" />Tolak</Button>
                              </div>
                            </TableCell>
                          </TableRow>
                        ))
                      )}
                    </TableBody>
                  </Table>
                </div>
                <PaginationControls meta={ppdPendMeta} onPageChange={setPpdPendPage} />
              </CardContent>
            </Card>

            <Card className="border-slate-100 shadow-sm rounded-xl">
              <CardHeader className="pb-3"><CardTitle className="text-base font-semibold flex items-center gap-2">Persetujuan Realisasi Bon {rbsPendMeta.total_data > 0 && <Badge className="bg-amber-100 text-amber-700 border-amber-200">{rbsPendMeta.total_data}</Badge>}</CardTitle></CardHeader>
              <CardContent className="p-0">
                <div className="overflow-x-auto">
                  <Table className="min-w-[800px]">
                    <TableHeader><TableRow className="bg-slate-50/50"><TableHead>Ref. No. Bon</TableHead><TableHead>Pemohon</TableHead><TableHead>Periode</TableHead><TableHead className="text-right">Total Realisasi</TableHead><TableHead className="text-right">Selisih</TableHead><TableHead className="text-center">Aksi</TableHead></TableRow></TableHeader>
                    <TableBody>
                      {loading ? (<TableRow><TableCell colSpan={6} className="text-center py-10">Memuat...</TableCell></TableRow>) : pendingRealisasi.length === 0 ? (<TableRow><TableCell colSpan={6} className="text-center py-10 text-slate-400">Kosong</TableCell></TableRow>) : (
                        pendingRealisasi.map(r => (
                          <TableRow key={r.id} className="cursor-pointer hover:bg-slate-50" onClick={() => openDetail(r, "rbs")}>
                            <TableCell className="text-sm font-mono">{r.nomor_bon_sementara || "-"}</TableCell><TableCell className="text-sm">{r.nama}</TableCell>
                            <TableCell className="text-sm">
                              {r.periode ? (
                                <div className="inline-flex items-center gap-1.5 px-2 py-1 bg-emerald-50 border border-emerald-200 rounded text-xs">
                                  <Calendar className="h-3 w-3 text-emerald-600 shrink-0" />
                                  <span className="text-emerald-900 font-medium whitespace-nowrap">{formatPeriodeCompact(r.periode)}</span>
                                </div>
                              ) : "-"}
                            </TableCell>
                            <TableCell className="text-sm text-right font-medium">{fmt(r.total_realisasi || 0)}</TableCell><TableCell className={`text-sm text-right font-medium ${(r.selisih || 0) >= 0 ? "text-emerald-600" : "text-red-600"}`}>{fmt(r.selisih || 0)}</TableCell>
                            <TableCell onClick={e => e.stopPropagation()}>
                              <div className="flex justify-center gap-1"><Button size="sm" className="h-7 bg-emerald-600 hover:bg-emerald-700" onClick={() => approve(r.id, "rbs")}><CheckCircle2 className="h-3 w-3 mr-1" />Setujui</Button><Button size="sm" variant="outline" className="h-7 text-red-600" onClick={() => setShowDecline({ ...r, type: "rbs" })}><XCircle className="h-3 w-3 mr-1" />Tolak</Button></div>
                            </TableCell>
                          </TableRow>
                        ))
                      )}
                    </TableBody>
                  </Table>
                </div>
                <PaginationControls meta={rbsPendMeta} onPageChange={setRbsPendPage} />
              </CardContent>
            </Card>
          </TabsContent>
        )}

        {/* TAB: PPD */}
        <TabsContent value="bon">
          <Card className="border-slate-100 shadow-sm rounded-xl">
            <CardHeader className="pb-3 border-b border-slate-100"><CardTitle className="text-base font-semibold">Riwayat Pengajuan</CardTitle></CardHeader>
            <CardContent className="p-0">
              <div className="overflow-x-auto">
                <Table className="min-w-[800px]"><TableHeader><TableRow className="bg-slate-50/50"><TableHead>No. Dokumen</TableHead>{!isPegawai && <TableHead>Pemohon</TableHead>}<TableHead>Tujuan</TableHead><TableHead className="text-right">Total Estimasi</TableHead><TableHead className="text-center">Status</TableHead><TableHead className="text-right">Aksi</TableHead></TableRow></TableHeader>
                  <TableBody>
                    {loading ? (<TableRow><TableCell colSpan={7} className="text-center py-10">Memuat...</TableCell></TableRow>) : historyBons.length === 0 ? (<TableRow><TableCell colSpan={7} className="text-center py-10 text-slate-400">Kosong</TableCell></TableRow>) : (
                      historyBons.map(b => (
                        <TableRow key={b.id} className="cursor-pointer hover:bg-slate-50" onClick={() => openDetail(b, "ppd")}>
                          <TableCell className="text-sm font-mono">{b.nomor_tipe_dokumen || b.nomor_dokumen || "-"}</TableCell>
                          {!isPegawai && <TableCell className="text-sm">{b.nama}</TableCell>}
                          <TableCell className="text-sm"><p>{b.tujuan}</p><p className="text-xs text-slate-500 truncate max-w-[150px]">{b.keperluan}</p></TableCell><TableCell className="text-sm text-right font-medium">{fmt(b.total_estimasi || 0)}</TableCell><TableCell className="text-center"><StatusBadge status={b.status} /></TableCell>
                          <TableCell onClick={e => e.stopPropagation()}>
                            <div className="flex justify-end gap-1 flex-wrap">
                              {isPegawai && b.status?.toLowerCase().includes("ditolak") && (<Button variant="ghost" size="sm" className="h-7 text-blue-600" onClick={() => setShowResubmit({ ...b, type: "ppd" })}><RotateCcw className="h-3 w-3 mr-1" />Ulang</Button>)}
                              {b.is_downloadable && b.status !== "Draft" && !b.status?.toLowerCase().includes("ditolak") && (<Button variant="ghost" size="sm" className="h-7 text-blue-600" onClick={() => downloadPdf(b.id, "ppd", formatDocName(b, "Pengajuan Perjalanan Dinas"), "rpd")}><FileText className="h-3 w-3 mr-1" />PPD</Button>)}
                              {b.status === "Selesai" && (<Button variant="ghost" size="sm" className="h-7 text-emerald-600" onClick={() => downloadPdf(b.id, "ppd", formatDocName(b, "Bon Sementara"), "bs")}><Download className="h-3 w-3 mr-1" />Bon</Button>)}
                            </div>
                          </TableCell>
                        </TableRow>
                      ))
                    )}
                  </TableBody>
                </Table>
              </div>
              <PaginationControls meta={ppdHistMeta} onPageChange={setPpdHistPage} />
            </CardContent>
          </Card>
        </TabsContent>

        {/* TAB: REALISASI */}
        <TabsContent value="realisasi">
          <Card className="border-slate-100 shadow-sm rounded-xl">
            <CardHeader className="pb-3 border-b border-slate-100">
              <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-3">
                <CardTitle className="text-base font-semibold">Data Realisasi</CardTitle>
                {userRole === "hrga" && (
                  <div className="flex flex-wrap items-center gap-2 w-full sm:w-auto">
                    <Select value={filterMonth} onValueChange={setFilterMonth}><SelectTrigger className="w-[110px] h-8 text-xs"><SelectValue placeholder="Bulan" /></SelectTrigger><SelectContent>{months.map(m => <SelectItem key={m.value} value={m.value}>{m.label}</SelectItem>)}</SelectContent></Select>
                    <Select value={filterYear} onValueChange={setFilterYear}><SelectTrigger className="w-[90px] h-8 text-xs"><SelectValue placeholder="Tahun" /></SelectTrigger><SelectContent>{years.map(y => <SelectItem key={y} value={String(y)}>{y}</SelectItem>)}</SelectContent></Select>
                    {(filterMonth || filterYear) && (<Button variant="ghost" size="sm" className="h-8 text-xs" onClick={() => { setFilterMonth(""); setFilterYear(""); }}>Reset</Button>)}
                    <Button size="sm" className="h-8 gap-1 bg-emerald-600 hover:bg-emerald-700 text-white ml-auto sm:ml-0" onClick={downloadExcel}><Download className="h-3 w-3" />Excel</Button>
                  </div>
                )}
              </div>
            </CardHeader>
            <CardContent className="p-0">
              <div className="overflow-x-auto">
                <Table className="min-w-[700px]"><TableHeader><TableRow className="bg-slate-50/50"><TableHead>Ref. No. Bon</TableHead>{!isPegawai && <TableHead>Pemohon</TableHead>}<TableHead className="w-48">Periode</TableHead><TableHead className="text-right">Total Realisasi</TableHead><TableHead className="text-right">Selisih</TableHead><TableHead className="text-center">Status</TableHead><TableHead className="text-right">Aksi</TableHead></TableRow></TableHeader>
                  <TableBody>
                    {loading ? (<TableRow><TableCell colSpan={7} className="text-center py-10">Memuat...</TableCell></TableRow>) : historyRealisasi.length === 0 ? (<TableRow><TableCell colSpan={7} className="text-center py-10 text-slate-400">Kosong</TableCell></TableRow>) : (
                      historyRealisasi.map(r => (
                        <TableRow key={r.id} className="cursor-pointer hover:bg-slate-50" onClick={() => openDetail(r, "rbs")}>
                          <TableCell className="text-sm font-mono">{r.nomor_bon_sementara || "-"}</TableCell>
                          {!isPegawai && <TableCell className="text-sm">{r.nama}</TableCell>}
                          <TableCell className="text-sm">
                            {r.periode ? (
                              <div className="inline-flex items-center gap-1.5 px-2 py-1 bg-emerald-50 border border-emerald-200 rounded text-xs">
                                <Calendar className="h-3 w-3 text-emerald-600 shrink-0" />
                                <span className="text-emerald-900 font-medium whitespace-nowrap">
                                  {formatPeriodeCompact(r.periode)}
                                </span>
                              </div>
                            ) : (
                              "-"
                            )}
                          </TableCell>
                          <TableCell className="text-sm text-right font-medium">{fmt(r.total_realisasi || 0)}</TableCell><TableCell className={`text-sm text-right font-medium ${(r.selisih || 0) >= 0 ? "text-emerald-600" : "text-red-600"}`}>{fmt(r.selisih || 0)}</TableCell><TableCell className="text-center"><StatusBadge status={r.status} /></TableCell>
                          <TableCell className="text-right" onClick={e => e.stopPropagation()}>{r.status === "Selesai" && (<Button variant="ghost" size="sm" className="h-7 text-emerald-600" onClick={() => downloadPdf(r.id, "rbs", formatDocName(r, "Realisasi Bon Sementara"))}><Download className="h-3 w-3 mr-1" />PDF</Button>)}</TableCell>
                        </TableRow>
                      ))
                    )}
                  </TableBody>
                </Table>
              </div>
              <PaginationControls meta={rbsHistMeta} onPageChange={setRbsHistPage} />
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* DIALOGS */}
      <CreatePPDDialog open={showCreateBon} onClose={() => setShowCreateBon(false)} onSuccess={refetchAll} />
      <CreateRealisasiDialog open={showCreateReal} onClose={() => setShowCreateReal(false)} onSuccess={refetchAll} />

      <Dialog open={!!showDecline} onOpenChange={() => { setShowDecline(null); setDeclineReason(""); }}>
        <DialogContent>
          <DialogHeader><DialogTitle>Tolak Pengajuan</DialogTitle><DialogDescription>Berikan alasan penolakan</DialogDescription></DialogHeader>
          <Textarea placeholder="Alasan..." value={declineReason} onChange={e => setDeclineReason(e.target.value)} />
          <DialogFooter><Button variant="outline" onClick={() => { setShowDecline(null); setDeclineReason(""); }}>Batal</Button><Button className="bg-red-600 hover:bg-red-700 text-white" onClick={decline}>Tolak</Button></DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={!!showResubmit} onOpenChange={() => setShowResubmit(null)}>
        <DialogContent>
          <DialogHeader><DialogTitle>Ajukan Ulang Dokumen?</DialogTitle></DialogHeader>
          <DialogFooter><Button variant="outline" onClick={() => setShowResubmit(null)}>Batal</Button><Button className="bg-blue-600 hover:bg-blue-700 text-white" onClick={() => handleResubmit(showResubmit?.id, showResubmit?.type)}>Ajukan Ulang</Button></DialogFooter>
        </DialogContent>
      </Dialog>

      {/* IMAGE VIEWER DIALOG UNTUK MELIHAT STRUK */}
      <Dialog open={!!viewImage} onOpenChange={() => setViewImage(null)}>
        <DialogContent className="sm:max-w-2xl bg-white/95 backdrop-blur">
          <DialogHeader><DialogTitle>Lihat Struk</DialogTitle></DialogHeader>
          <div className="flex justify-center p-4">
            <img src={viewImage || ""} alt="Struk" className="max-h-[60vh] max-w-full rounded border border-slate-200 object-contain shadow-sm" />
          </div>
          <DialogFooter><Button onClick={() => setViewImage(null)}>Tutup</Button></DialogFooter>
        </DialogContent>
      </Dialog>

      {/* DETAIL DIALOG */}
      <Dialog open={!!showDetail} onOpenChange={() => { setShowDetail(null); setIsEditMode(false); setEditClickCount(0); }}>
        <DialogContent className="w-[95vw] max-w-full sm:max-w-4xl max-h-[90vh] overflow-y-auto">
          {showDetail?.isLoading ? (
            <div className="flex flex-col items-center justify-center py-20"><Loader2 className="h-8 w-8 animate-spin text-slate-400 mb-4" /><p className="text-slate-500">Memuat rincian...</p></div>
          ) : (
            <>
              <DialogHeader>
                <div className="flex items-center justify-between pr-8">
                    <div>
                        <DialogTitle>
                            {showDetail?.type === "ppd" 
                                ? (isEditMode ? "Edit Perjalanan Dinas" : "Detail Perjalanan Dinas") 
                                : "Detail Realisasi Bon"}
                        </DialogTitle>
                        <DialogDescription>Nomor: {showDetail?.nomor_tipe_dokumen || showDetail?.nomor_dokumen || showDetail?.nomor_dokumen_referensi || "-"}</DialogDescription>
                    </div>
                    {/* Header Action: Switch ke Edit jika role HRGA dan sedang di tab approval PPD */}
                    {showDetail?.type === "ppd" && userRole === "hrga" && tab === "approval" && !isEditMode && (
                        <Button variant="outline" size="sm" onClick={() => setIsEditMode(true)} className="gap-1">
                            <Edit className="h-4 w-4"/> Edit Data
                        </Button>
                    )}
                </div>
              </DialogHeader>

              {/* DETAIL & EDIT PPD */}
              {showDetail?.type === "ppd" && (
                <div className="space-y-4">
                  {/* Info Dasar DIBUAT READONLY */}
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 text-sm">
                    <div className="flex flex-col"><span className="text-xs text-slate-500 font-medium">Pemohon</span> <span className="text-slate-900 font-medium break-words">{showDetail.nama || "-"}</span></div>
                    <div className="flex flex-col"><span className="text-xs text-slate-500 font-medium">Status</span> <span className="mt-0.5"><StatusBadge status={showDetail.status} /></span></div>
                    <div className="flex flex-col"><span className="text-xs text-slate-500 font-medium">Tujuan</span> <span className="text-slate-900 break-words">{showDetail.tujuan || "-"}</span></div>
                    <div className="flex flex-col"><span className="text-xs text-slate-500 font-medium">Keperluan</span> <span className="text-slate-900 break-words">{showDetail.keperluan || "-"}</span></div>
                  </div>
                  
                  {/* PERIODE - MODERN FORMAT */}
                  {!isEditMode && (
                      <div className="bg-gradient-to-r from-blue-50 to-indigo-50 border border-blue-200 rounded-lg p-3 sm:p-4">
                        <div className="flex flex-col sm:flex-row items-start sm:items-center gap-3">
                          <div className="flex items-center gap-3">
                            <div className="h-10 w-10 rounded-full bg-blue-100 flex items-center justify-center shrink-0">
                              <Calendar className="h-5 w-5 text-blue-600" />
                            </div>
                            <div>
                              <p className="text-xs text-blue-600 font-medium uppercase tracking-wide">
                                Periode Perjalanan
                              </p>
                              <p className="text-sm text-slate-700 font-semibold mt-0.5">
                                {formatDate(showDetail.tanggal_berangkat || showDetail.periode_berangkat)}{" "}
                                <span className="text-blue-400 mx-1">→</span>{" "}
                                {formatDate(showDetail.tanggal_kembali || showDetail.periode_kembali)}
                              </p>
                            </div>
                          </div>
                        </div>
                      </div>
                  )}

                  {/* Akomodasi */}
                  {(!isEditMode && showDetail.rincian_hotel && showDetail.rincian_hotel.nama_hotel) || isEditMode ? (
                    <>
                        <Separator />
                        <div>
                            <div className="flex justify-between mb-3">
                                <h4 className="font-semibold text-slate-700">Akomodasi / Hotel</h4>
                            </div>
                            {isEditMode ? (
                                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-sm">
                                    <div><Label>Hotel</Label><Input value={editForm?.rincian_hotel?.nama_hotel || ""} onChange={(e) => setEditForm({...editForm, rincian_hotel: {...editForm.rincian_hotel, nama_hotel: e.target.value}})} className="h-8" /></div>
                                   <div><Label>Harga (Rp)</Label><Input type="number" value={editForm?.rincian_hotel?.harga ?? ""} onChange={(e) => setEditForm({...editForm, rincian_hotel: {...editForm.rincian_hotel, harga: e.target.value}})} className="h-8" /></div>
                                </div>
                            ) : (
                                <div className="flex flex-col gap-2 text-sm p-3 bg-slate-50 rounded-lg border border-slate-100">
                                    <div className="flex justify-between gap-2"><span className="text-slate-500 shrink-0">Hotel</span> <span className="font-medium text-slate-900 text-right break-words">{showDetail.rincian_hotel.nama_hotel}</span></div>
                                    <div className="flex justify-between border-t border-slate-200 mt-1 pt-2 gap-2"><span className="text-slate-600 font-semibold shrink-0">Total Harga</span> <span className="font-bold text-slate-900 text-right">{fmt(showDetail.rincian_hotel.total_harga)}</span></div>
                                </div>
                            )}
                        </div>
                    </>
                  ) : null}

                  {/* Transportasi */}
                  {(!isEditMode && showDetail.rincian_transportasi && showDetail.rincian_transportasi.length > 0) || isEditMode ? (
                    <>
                        <Separator />
                        <div>
                            <div className="flex flex-wrap items-center justify-between gap-2 mb-3">
                                <h4 className="font-semibold text-slate-700">Transportasi</h4>
                                {isEditMode && <Button size="sm" variant="outline" className="h-7 text-xs" onClick={() => setEditForm({...editForm, rincian_transportasi: [...(editForm.rincian_transportasi||[]), { kota_asal: "", kota_tujuan: "", jenis_transportasi: "", tipe_perjalanan: "Keberangkatan", jam_berangkat: "", harga: 0 }]})}>Tambah Rute</Button>}
                            </div>
                            
                            {isEditMode ? (
                                editForm?.rincian_transportasi?.map((trans: any, idx: number) => (
                                    <div key={idx} className="flex flex-col sm:grid sm:grid-cols-12 gap-3 text-sm mb-4 pb-4 border-b border-slate-100 last:border-0 last:pb-0 last:mb-0">
                                        <div className="sm:col-span-3"><Label className="text-xs">Jenis</Label><Input value={trans.jenis_transportasi} onChange={(e) => updateEditFormTransport(idx, 'jenis_transportasi', e.target.value)} className="h-8 text-xs" /></div>
                                        <div className="sm:col-span-3"><Label className="text-xs">Asal</Label><Input value={trans.kota_asal} onChange={(e) => updateEditFormTransport(idx, 'kota_asal', e.target.value)} className="h-8 text-xs" /></div>
                                        <div className="sm:col-span-3"><Label className="text-xs">Tujuan</Label><Input value={trans.kota_tujuan} onChange={(e) => updateEditFormTransport(idx, 'kota_tujuan', e.target.value)} className="h-8 text-xs" /></div>
                                        <div className="sm:col-span-2"><Label className="text-xs">Harga</Label><Input type="number" value={trans.harga ?? ""} onChange={(e) => updateEditFormTransport(idx, 'harga', e.target.value)} className="h-8 text-xs" /></div>
                                        <div className="sm:col-span-1 flex items-end justify-end"><Button size="icon" variant="ghost" className="h-8 w-8 text-red-500" onClick={() => setEditForm({...editForm, rincian_transportasi: editForm.rincian_transportasi.filter((_:any, i:number) => i !== idx)})}><Trash2 className="h-4 w-4"/></Button></div>
                                    </div>
                                ))
                            ) : (
                                <div className="space-y-3">
                                    {showDetail.rincian_transportasi.map((trans: any, idx: number) => (
                                        <div key={idx} className="flex flex-col gap-2 text-sm p-3 bg-slate-50 rounded-lg border border-slate-100">
                                            <div className="flex justify-between gap-2"><span className="text-slate-500 shrink-0">Tipe</span> <span className="font-medium text-slate-900 text-right break-words">{trans.tipe_perjalanan}</span></div>
                                            <div className="flex justify-between gap-2"><span className="text-slate-500 shrink-0">Jenis</span> <span className="font-medium text-slate-900 text-right break-words">{trans.jenis_transportasi}</span></div>
                                            <div className="flex justify-between gap-2"><span className="text-slate-500 shrink-0">Rute</span> <span className="font-medium text-slate-900 text-right break-words">{trans.kota_asal} → {trans.kota_tujuan}</span></div>
                                            <div className="flex justify-between border-t border-slate-200 mt-1 pt-2 gap-2"><span className="text-slate-600 font-semibold shrink-0">Harga</span> <span className="font-bold text-slate-900 text-right">{fmt(trans.harga)}</span></div>
                                        </div>
                                    ))}
                                </div>
                            )}
                        </div>
                    </>
                  ) : null}

                  {/* Rincian Tambahan (Estimasi) */}
                  {(!isEditMode && showDetail.rincian_tambahan && showDetail.rincian_tambahan.length > 0) || isEditMode ? (
                    <>
                        <Separator />
                        <div>
                            <div className="flex flex-wrap items-center justify-between gap-2 mb-3">
                                <h4 className="font-semibold text-slate-700">Estimasi Biaya Tambahan</h4>
                                {isEditMode && <Button size="sm" variant="outline" className="h-7 text-xs" onClick={() => setEditForm({...editForm, rincian_tambahan: [...(editForm.rincian_tambahan||[]), { kategori: "Lainnya", keterangan: "", kuantitas: 1, harga: 0 }]})}>Tambah Biaya</Button>}
                            </div>

                            <div className="space-y-3">
                                {isEditMode ? (
                                    editForm?.rincian_tambahan?.map((item: any, idx: number) => (
                                        <div key={idx} className="flex flex-col sm:grid sm:grid-cols-12 gap-3 mb-4 pb-4 border-b border-slate-100 last:border-0 last:pb-0 last:mb-0">
                                            <div className="sm:col-span-5"><Label className="text-xs sm:hidden">Keterangan</Label><Input placeholder="Keterangan" value={item.keterangan || item.uraian || ""} onChange={(e) => updateEditFormEstimasi(idx, 'keterangan', e.target.value)} className="h-8 text-xs" /></div>
                                            <div className="sm:col-span-2"><Label className="text-xs sm:hidden">Qty</Label><Input type="number" placeholder="Qty" value={item.kuantitas ?? item.qty ?? ""} onChange={(e) => updateEditFormEstimasi(idx, 'kuantitas', e.target.value)} className="h-8 text-xs" /></div>
                                            <div className="sm:col-span-4"><Label className="text-xs sm:hidden">Harga</Label><Input type="number" placeholder="Harga" value={item.harga ?? item.harga_unit ?? ""} onChange={(e) => updateEditFormEstimasi(idx, 'harga', e.target.value)} className="h-8 text-xs" /></div>
                                            <div className="sm:col-span-1 flex items-end justify-end"><Button size="icon" variant="ghost" className="h-8 w-8 text-red-500 w-full sm:w-auto" onClick={() => setEditForm({...editForm, rincian_tambahan: editForm.rincian_tambahan.filter((_:any, i:number) => i !== idx)})}><Trash2 className="h-4 w-4 mx-auto"/></Button></div>
                                        </div>
                                    ))
                                ) : (
                                    <div className="space-y-3">
                                        {showDetail.rincian_tambahan.map((item: any, idx: number) => (
                                            <div key={idx} className="flex flex-col gap-2 text-sm p-3 bg-slate-50 rounded-lg border border-slate-100">
                                                <div className="flex justify-between gap-2"><span className="text-slate-500 shrink-0">Kategori</span> <span className="font-medium text-slate-900 text-right break-words">{item.kategori || "-"}</span></div>
                                                <div className="flex justify-between gap-2"><span className="text-slate-500 shrink-0">Uraian</span> <span className="font-medium text-slate-900 text-right break-words">{item.keterangan || item.uraian}</span></div>
                                                <div className="flex justify-between gap-2"><span className="text-slate-500 shrink-0">Qty</span> <span className="font-medium text-slate-900 text-right">{item.kuantitas || item.qty}</span></div>
                                                <div className="flex justify-between border-t border-slate-200 mt-1 pt-2 gap-2"><span className="text-slate-600 font-semibold shrink-0">Harga</span> <span className="font-bold text-slate-900 text-right">{fmt(item.harga || item.harga_unit || 0)}</span></div>
                                            </div>
                                        ))}
                                    </div>
                                )}
                                
                                {!isEditMode && (
                                    <div className="mt-4 bg-slate-900 text-white rounded-lg p-3 sm:p-4 shadow-sm">
                                        <div className="flex justify-between items-center">
                                            <span className="text-sm font-semibold text-slate-200">TOTAL ESTIMASI</span>
                                            <span className="text-lg sm:text-xl font-bold">{fmt(showDetail.total_estimasi)}</span>
                                        </div>
                                    </div>
                                )}
                            </div>
                        </div>
                    </>
                  ) : null}
                </div>
              )}

              {/* DETAIL RBS */}
              {showDetail?.type === "rbs" && (
                 <div className="space-y-4">
                  {/* Info Dasar */}
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 text-sm">
                    <div className="flex flex-col"><span className="text-xs text-slate-500 font-medium">Ref. Bon</span> <span className="text-slate-900 font-medium break-words">{showDetail.no_ref_bon_sementara || showDetail.nomor_bon_sementara || showDetail.nomor_dokumen_referensi || "-"}</span></div>
                    <div className="flex flex-col"><span className="text-xs text-slate-500 font-medium">Pemohon</span> <span className="text-slate-900 font-medium break-words">{showDetail.pemohon || showDetail.nama || "-"}</span></div>
                    <div className="flex flex-col sm:col-span-2"><span className="text-xs text-slate-500 font-medium">Status</span> <span className="mt-0.5"><StatusBadge status={showDetail.status} /></span></div>
                  </div>
                  
                  {/* PERIODE - MODERN FORMAT */}
                  <div className="bg-gradient-to-r from-emerald-50 to-teal-50 border border-emerald-200 rounded-lg p-3 sm:p-4">
                    <div className="flex flex-col sm:flex-row items-start sm:items-center gap-3">
                      <div className="flex items-center gap-3">
                        <div className="h-10 w-10 rounded-full bg-emerald-100 flex items-center justify-center shrink-0">
                          <Calendar className="h-5 w-5 text-emerald-600" />
                        </div>
                        <div>
                          <p className="text-xs text-emerald-600 font-medium uppercase tracking-wide">
                            Periode Realisasi
                          </p>
                          <p className="text-sm text-slate-700 font-semibold mt-0.5 break-words">
                            {showDetail.periode 
                                ? formatPeriodeCompact(showDetail.periode) 
                                : `${formatDate(showDetail.tanggal_berangkat)} - ${formatDate(showDetail.tanggal_kedatangan)}`}
                          </p>
                        </div>
                      </div>
                    </div>
                  </div>
                  
                  {/* RINCIAN BIAYA AKTUAL (RESPONSIVE CARD) */}
                  {(showDetail.rincian_realisasi || showDetail.items)?.length > 0 && (
                    <>
                        <Separator />
                        <div>
                            <h4 className="font-semibold text-slate-700 mb-3">Rincian Biaya Aktual</h4>
                            <div className="space-y-3">
                                {(showDetail.rincian_realisasi || showDetail.items).map((item: any, idx: number) => { 
                                    const strukUrl = item.url_struk || item.bukti; 
                                    return (
                                        <div key={idx} className="flex flex-col gap-2 text-sm p-3 bg-slate-50 rounded-lg border border-slate-100">
                                            <div className="flex justify-between gap-2"><span className="text-slate-500 shrink-0">Tanggal</span> <span className="font-medium text-slate-900 text-right">{item.tanggal_transaksi ? formatDate(item.tanggal_transaksi) : "-"}</span></div>
                                            <div className="flex justify-between gap-2 items-start"><span className="text-slate-500 shrink-0">Kategori</span> <span className="text-right"><Badge variant="outline" className="text-xs font-normal">{item.kategori || "-"}</Badge></span></div>
                                            <div className="flex justify-between gap-2"><span className="text-slate-500 shrink-0">Uraian</span> <span className="font-medium text-slate-900 text-right break-words">{item.uraian}</span></div>
                                            <div className="flex justify-between border-t border-slate-200 mt-1 pt-2 gap-2"><span className="text-slate-600 font-semibold shrink-0">Total Harga</span> <span className="font-bold text-slate-900 text-right">{fmt(item.total_harga || item.total)}</span></div>
                                            {strukUrl && (
                                                <div className="flex justify-end mt-1">
                                                    <Button variant="outline" size="sm" className="h-8 text-blue-600 gap-1 text-xs w-full sm:w-auto bg-white" onClick={() => setViewImage(strukUrl.startsWith("http") || strukUrl.startsWith("data:") ? strukUrl : `${BACKEND_URL}${strukUrl}`)}>
                                                        <Eye className="h-3.5 w-3.5" /> Lihat Struk
                                                    </Button>
                                                </div>
                                            )}
                                        </div>
                                    );
                                })}
                            </div>
                        </div>
                    </>
                  )}

                  {/* REKAP TOTAL */}
                  <div className="bg-slate-50 p-3 sm:p-4 rounded-lg space-y-2 mt-4 text-sm border border-slate-200">
                    <div className="flex justify-between gap-2">
                        <span className="text-slate-600 shrink-0">Total Realisasi Aktual</span>
                        <span className="font-bold text-slate-900 text-right">
                            {fmt(showDetail.total_realisasi ?? (showDetail.rincian_realisasi || []).reduce((sum: number, item: any) => sum + (item.total_harga || 0), 0))}
                        </span>
                    </div>
                    {(showDetail.uang_muka !== undefined || showDetail.total_estimasi !== undefined) && (
                        <div className="flex justify-between border-b border-slate-200 pb-2 gap-2">
                            <span className="text-slate-600 shrink-0">Uang Muka (Estimasi)</span>
                            <span className="font-medium text-slate-900 text-right">{fmt(showDetail.uang_muka || showDetail.total_estimasi || 0)}</span>
                        </div>
                    )}
                    {(showDetail.selisih !== undefined || showDetail.sisa_bon !== undefined) && (
                        <div className={`flex justify-between pt-1 font-bold gap-2 ${((showDetail.selisih || showDetail.sisa_bon) || 0) >= 0 ? "text-emerald-600" : "text-red-600"}`}>
                            <span className="shrink-0">Selisih / Sisa Bon</span>
                            <span className="text-right">{fmt(showDetail.selisih || showDetail.sisa_bon || 0)}</span>
                        </div>
                    )}
                  </div>
                </div>
              )}

              {/* RIWAYAT PERSETUJUAN */}
              {!isEditMode && (showDetail?.riwayat_persetujuan?.length > 0 || showDetail?.approval_history?.length > 0) && (
                <>
                    <Separator />
                    <div>
                        <h4 className="font-semibold text-slate-700 mb-3">Riwayat Persetujuan</h4>
                        <div className="space-y-2">
                            {(showDetail.riwayat_persetujuan || showDetail.approval_history).map((h: any, idx: number) => {
                                const statusTindakan = h.tindakan || h.status || (h.action === "approved" ? "Disetujui" : "Ditolak");
                                const isTolak = statusTindakan.toLowerCase().includes("tolak") || h.action === "declined";
                                return (
                                <div key={idx} className={`flex flex-col sm:flex-row items-start sm:items-center justify-between text-sm p-3 rounded-lg border ${isTolak ? "bg-red-50 border-red-100" : "bg-emerald-50 border-emerald-100"}`}>
                                    <div>
                                        <span className={`font-semibold ${isTolak ? "text-red-700" : "text-emerald-700"}`}>
                                            {h.nama || h.nama_approver || h.by}
                                        </span>
                                        <span className="text-slate-500 ml-2 text-xs uppercase tracking-wider">
                                            ({h.jabatan || h.role_approver || h.role})
                                        </span>
                                        {h.catatan && (
                                            <p className="text-xs text-slate-600 mt-1 italic">"{h.catatan}"</p>
                                        )}
                                    </div>
                                    <div className="text-left sm:text-right mt-2 sm:mt-0">
                                        <span className={`font-medium ${isTolak ? "text-red-600" : "text-emerald-600"}`}>
                                            {statusTindakan}
                                        </span>
                                        {(h.tanggal || h.at || h.waktu_disetujui) && (
                                            <p className="text-xs text-slate-500 mt-0.5">
                                                {h.waktu_disetujui 
                                                  ? formatDate(h.waktu_disetujui) 
                                                  : h.tanggal 
                                                  ? formatDate(h.tanggal) 
                                                  : h.at?.slice(0, 10)}
                                            </p>
                                        )}
                                    </div>
                                </div>
                            )})}
                        </div>
                    </div>
                </>
              )}

              <DialogFooter className="mt-6 border-t pt-4 flex flex-col sm:flex-row items-center sm:justify-between w-full gap-3">
                <div className="flex gap-2 w-full sm:w-auto">
                    {/* Action Khusus Approval & Edit */}
                    {showDetail?.type === "ppd" && userRole === "hrga" && tab === "approval" && (
                        isEditMode ? (
                            <Button 
                              className={`${editClickCount >= 5 ? 'bg-slate-400 cursor-not-allowed' : 'bg-emerald-600 hover:bg-emerald-700'} text-white gap-1 w-full sm:w-auto`} 
                              onClick={saveEdit} 
                              disabled={isSaving || editClickCount >= 5}
                            >
                                {isSaving ? "Menyimpan..." : editClickCount >= 5 ? "Batas Klik Tercapai" : <><Save className="h-4 w-4"/> Simpan Perubahan</>}
                            </Button>
                        ) : (
                            <>
                                <Button className="bg-emerald-600 hover:bg-emerald-700 text-white gap-1 flex-1 sm:flex-none" onClick={() => approve(showDetail.id, "ppd")}>
                                    <CheckCircle2 className="h-4 w-4"/> Setuju
                                </Button>
                                <Button variant="outline" className="text-red-600 border-red-200 hover:bg-red-50 gap-1 flex-1 sm:flex-none" onClick={() => setShowDecline({ ...showDetail, type: "ppd" })}>
                                    <XCircle className="h-4 w-4"/> Tolak
                                </Button>
                            </>
                        )
                    )}
                    
                    {/* Tombol Download Konsisten di Detail Dialog */}
                    {!isEditMode && showDetail?.status === "Selesai" && showDetail?.type === "rbs" && (
                        <Button className="bg-emerald-600 hover:bg-emerald-700 text-white gap-1 w-full sm:w-auto" onClick={() => downloadPdf(showDetail.id, "rbs", formatDocName(showDetail, "Realisasi Bon Sementara"))}>
                            <Download className="h-4 w-4" /> Download RBS
                        </Button>
                    )}
                </div>

                <div className="flex flex-col sm:flex-row gap-2 w-full sm:w-auto">
                    {isEditMode ? (
                         <Button variant="outline" onClick={() => { setIsEditMode(false); setEditForm(null); setEditClickCount(0); }} className="w-full sm:w-auto">Batal Edit</Button>
                    ) : (
                        <>
                            <Button variant="outline" onClick={() => setShowDetail(null)} className="w-full sm:w-auto">Tutup</Button>
                            {showDetail?.is_downloadable && showDetail?.type === "ppd" && showDetail?.status !== "Draft" && !showDetail?.status?.toLowerCase().includes("ditolak") && (
                                <Button className="bg-blue-600 hover:bg-blue-700 text-white gap-1 w-full sm:w-auto" onClick={() => downloadPdf(showDetail.id, "ppd", formatDocName(showDetail, "Pengajuan Perjalanan Dinas"), "rpd")}>
                                    <FileText className="h-4 w-4" /> Download PPD
                                </Button>
                            )}
                        </>
                    )}
                </div>
              </DialogFooter>
            </>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}

// ===================== ROOT PAGE =====================
export default function BonPage() {
  const { user } = useAuth();
  const userRole = (user?.role || "pegawai").toLowerCase(); 
  const roleTitles: Record<string, { title: string; subtitle: string }> = {
    pegawai: { title: "Perjalanan Dinas & Realisasi", subtitle: "Ajukan perjalanan dinas dan pantau proses bon" },
    atasan: { title: "Pengajuan & Persetujuan Atasan", subtitle: "Ajukan perjalanan dinas Anda serta setujui pengajuan dari tim" },
    hrga: { title: "Pengajuan & Verifikasi HRGA", subtitle: "Ajukan perjalanan dinas Anda serta verifikasi kesesuaian operasional" },
    direktur: { title: "Persetujuan Direktur", subtitle: "Persetujuan tahap akhir pengajuan perjalanan dinas" },
    finance: { title: "Persetujuan Finance", subtitle: "Verifikasi akhir dan proses pencairan dana bon" }
  };
  const info = roleTitles[userRole] || { title: "Sistem Perjalanan Dinas", subtitle: "Manajemen dokumen dan pengajuan" };

  return (
    <div className="max-w-7xl mx-auto py-6 px-4 sm:px-6 lg:px-8">
      <div className="mb-6"><h2 className="text-2xl font-bold text-slate-900 tracking-tight">{info.title}</h2><p className="text-slate-500 text-sm mt-1">{info.subtitle}</p></div>
      <UniversalView userRole={userRole} />
    </div>
  );
}