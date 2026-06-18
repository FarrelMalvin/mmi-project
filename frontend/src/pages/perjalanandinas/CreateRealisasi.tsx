import React, { useState, useEffect } from "react";
import { api } from "../../lib/api";
import { getApiErrorMessage } from "../../lib/error";
import { Button } from "../../components/common/Button";
import { Input } from "../../components/common/Input";
import { Label } from "../../components/common/Label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "../../components/common/Select";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, DialogDescription } from "../../components/common/Dialog";
import { toast } from "sonner";
import { Plus, Trash2, Loader2, UploadCloud } from "lucide-react";

const fmt = (n: number) => new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", minimumFractionDigits: 0 }).format(n);

export default function CreateRealisasiDialog({ open, onClose, onSuccess }: { open: boolean; onClose: () => void; onSuccess: () => void; }) {
  const [loading, setLoading] = useState(false);
  const [loadingItems, setLoadingItems] = useState(false);
  const [uploadingIdx] = useState<number | null>(null);
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
      bukti: "" ,
      struk_file: null,
      struk_preview: null
    }]
  });

  useEffect(() => {
    if (open) {
      api.get("/rbs/options")
        .then((res) => setOptions(res.data?.data || res.data || []))
        .catch((err) => toast.error(getApiErrorMessage(err, "Gagal memuat daftar referensi bon")));
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
      toast.error(getApiErrorMessage(err, "Gagal mengambil rincian item")); 
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

  const handleFileUpload = (idx: number, e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    if (file.size > 2 * 1024 * 1024) {
      toast.error("Ukuran file maksimal 2 MB");
      e.target.value = "";
      return;
    }

    setRealForm((p: any) => {
      const items = [...p.items];

      items[idx] = {
        ...items[idx],
        bukti: file.name,
        struk_file: file,
      };

      return {
        ...p,
        items,
      };
    });

    toast.success("Struk dipilih.");
  };

  const handleTransferUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    if (file.size > 2 * 1024 * 1024) {
      toast.error("Ukuran file maksimal 2 MB");
      e.target.value = "";
      return;
    }

    setRealForm((p: any) => ({
      ...p,
      bukti_transfer: file.name,
      bukti_transfer_file: file,
    }));

    toast.success("Bukti transfer dipilih.");
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
    if (
      !realForm.request_ppd_id ||
      !realForm.periode_berangkat ||
      !realForm.periode_kembali ||
      realForm.items.length === 0
    ) {
      toast.error("Mohon lengkapi referensi, periode keberangkatan, dan kedatangan");
      return;
    }

    if (
      realForm.items.some(
        (i: any) =>
          i.tanggal &&
          (i.tanggal < realForm.periode_berangkat ||
            i.tanggal > realForm.periode_kembali)
      )
    ) {
      toast.error("Tanggal item pengeluaran harus berada di antara periode berangkat dan kembali");
      return;
    }

    if (
      realForm.items.some(
        (i: any) =>
          i.quantity === "" ||
          i.quantity == null ||
          parseFloat(i.quantity) <= 0 ||
          i.harga_per_unit === "" ||
          i.harga_per_unit == null ||
          parseFloat(i.harga_per_unit) <= 0
      )
    ) {
      toast.error("Qty dan Harga/Unit pada rincian aktual tidak boleh 0 atau kosong");
      return;
    }

    if (realForm.items.some((i: any) => !i.struk_file)) {
      toast.error("Struk/Bukti wajib dipilih untuk setiap item rincian aktual");
      return;
    }

    setLoading(true);

    try {
      const selectedBon = options.find(
        (o) => o.id.toString() === realForm.request_ppd_id
      );

      const nomorBonSementara = selectedBon
        ? selectedBon.nomor_tipe_dokumen || selectedBon.nomor_dokumen
        : "";

      const totalRealisasi = realForm.items.reduce(
        (sum: number, i: any) => sum + (parseFloat(i.total) || 0),
        0
      );

      const estimasiAwal = selectedBon
        ? parseFloat(selectedBon.total_estimasi) || 0
        : 0;

      const selisih = estimasiAwal - totalRealisasi;

      const payload = {
        request_ppd_id: parseInt(realForm.request_ppd_id),
        total_realisasi: totalRealisasi,
        selisih: selisih,
        periode_berangkat: `${realForm.periode_berangkat}T00:00:00Z`,
        periode_kembali: `${realForm.periode_kembali}T00:00:00Z`,
        nomor_bon_sementara: nomorBonSementara,

        url_bukti_transfer: null,
        bukti_transfer_field: realForm.bukti_transfer_file ? "bukti_transfer" : "",

        items: realForm.items.map((i: any, idx: number) => ({
          tanggal: i.tanggal ? `${i.tanggal}T00:00:00Z` : null,
          kategori: i.kategori,
          uraian: i.uraian,
          kuantitas: parseInt(i.quantity) || 1,
          harga_unit: parseFloat(i.harga_per_unit) || 0,
          total: parseFloat(i.total) || 0,

          struk_field: i.struk_file ? `struk_${idx}` : "",
        })),
      };

      const formData = new FormData();
      formData.append("payload", JSON.stringify(payload));

      realForm.items.forEach((item: any, idx: number) => {
        if (item.struk_file) {
          formData.append(`struk_${idx}`, item.struk_file);
        }
      });

      if (realForm.bukti_transfer_file) {
        formData.append("bukti_transfer", realForm.bukti_transfer_file);
      }

      await api.post("/rbs", formData, {
        headers: {
          "Content-Type": "multipart/form-data",
        },
      });

      toast.success("Realisasi berhasil diajukan");
      onSuccess();
      onClose();
    } catch (err: any) {
      toast.error(getApiErrorMessage(err, "Gagal mengajukan realisasi"));
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
                      <Input 
                        type="date" 
                        className="h-9 text-xs" 
                        value={item.tanggal} 
                        min={realForm.periode_berangkat || undefined}
                        max={realForm.periode_kembali || undefined}
                        onChange={e => updateRealItem(idx, "tanggal", e.target.value)} 
                      />
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
                        {/* Menambahkan tanda bintang (*) agar terlihat wajib */}
                        <Label className="text-xs text-slate-500 mb-1.5 sm:hidden block">Struk / Bukti *</Label>
                        <label className={`flex items-center justify-center h-9 w-full sm:w-9 border ${!item.bukti ? 'border-red-300 bg-red-50' : 'border-slate-300 hover:bg-slate-50'} rounded-md transition-colors cursor-pointer ${uploadingIdx === idx ? 'opacity-50' : ''} ${item.bukti ? 'bg-emerald-50 border-emerald-200' : ''}`}>
                          {uploadingIdx === idx ? <Loader2 className="h-4 w-4 animate-spin text-slate-400" /> : <UploadCloud className={`h-4 w-4 ${item.bukti ? 'text-emerald-600' : (!item.bukti ? 'text-red-400' : 'text-slate-500')}`} />}
                          <span className={`ml-2 text-xs sm:hidden font-medium ${item.bukti ? 'text-emerald-700' : 'text-red-500'}`}>{item.bukti ? 'Terupload' : 'Wajib Upload'}</span>
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