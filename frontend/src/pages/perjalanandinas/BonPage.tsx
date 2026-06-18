import { useState, useEffect, useCallback, Suspense, lazy } from "react";
import { useAuth } from "../../contexts/AuthContext";
import { api } from "../../lib/api";
import { getApiErrorMessage } from "../../lib/error";
import { Button } from "../../components/common/Button";
import { Card, CardContent, CardHeader, CardTitle } from "../../components/common/Card";
import { Input } from "../../components/common/Input";
import { Textarea } from "../../components/common/TextArea";
import { Badge } from "../../components/common/Badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../../components/common/Tabs";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "../../components/common/Select";
import { Separator } from "../../components/common/Separator";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, DialogDescription } from "../../components/common/Dialog";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "../../components/common/Table";
import { toast } from "sonner";
import { 
  Plus, CheckCircle2, XCircle, RotateCcw, Download, 
  FileText, Loader2, Eye, Edit, Save, Calendar
} from "lucide-react";

const CreatePPDDialog = lazy(() => import("./CreatePPD"));
const CreateRealisasiDialog = lazy(() => import("./CreateRealisasi"));

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
  return new Date(dateStr.replace(" ", "T")).toLocaleDateString("id-ID", { day: "numeric", month: "short", year: "numeric" });
};

const formatPeriodeCompact = (periodeString: string) => {
  if (!periodeString) return "-";
  const parts = periodeString.split(" s/d ");
  if (parts.length !== 2) return periodeString;
  const [start, end] = parts;
  return `${formatDate(start)} - ${formatDate(end)}`;
};

const formatDocName = (doc: any, prefix: string) => {
  let numStr = "";
  if (prefix === "Pengajuan Perjalanan Dinas") numStr = doc.nomor_tipe_dokumen || doc.id || "";
  else if (prefix === "Bon Sementara") numStr = doc.nomor_dokumen || doc.id || "";
  else numStr = doc.nomor_dokumen || doc.nomor_tipe_dokumen || doc.id || "";
  
  const cleanNum = numStr.toString().replace(/_/g, '/');
  return `${prefix} - ${cleanNum}`;
};

const resolveFileUrl = (url?: string) => {
  if (!url) return "";
  if (url.startsWith("http") || url.startsWith("data:")) {
    return url;
  }
  return `${BACKEND_URL}${url}`;
};

const isPdfFile = (url?: string) => {
  if (!url) return false;
  const cleanUrl = url.toLowerCase().split("?")[0];
  return cleanUrl.endsWith(".pdf") || url.startsWith("data:application/pdf");
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

// ===================== UNIVERSAL VIEW =====================
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
  
  const [viewFile, setViewFile] = useState<{ url: string; title: string; } | null>(null);

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
      
      if (type === "rbs") {
        const itemsList = Array.isArray(data) ? data : (data.rincian_realisasi || data.items || []);
        setShowDetail({ ...row, ...data, items: itemsList, type, isLoading: false });
        
        setEditForm({
          ...data,
          items: itemsList
        });
      } else {
        setShowDetail({ ...row, ...data, type, isLoading: false });
        setEditForm({
          ...data,
          rincian_tambahan: data.rincian_tambahan || [],
          rincian_transportasi: data.rincian_transportasi || [],
          rincian_hotel: data.rincian_hotel ? { 
            ...data.rincian_hotel, 
            harga_per_malam: data.rincian_hotel.harga_per_malam || data.rincian_hotel.harga || 0 
          } : { nama_hotel: "", check_in: "", check_out: "", harga_per_malam: 0 }
        });
      }
    } catch { toast.error(`Gagal memuat detail`); setShowDetail(null); }
  };

  const approve = async (id: number, type: "ppd" | "rbs") => {
    try { await api.patch(`/${type}/${id}/approve`, {}); toast.success("Disetujui"); setShowDetail(null); refetchAll(); } catch (err) { toast.error(getApiErrorMessage(err, "Gagal menyetujui dokumen")); }
  };
  
  const decline = async () => {
    if (!declineReason.trim()) { toast.error("Alasan wajib diisi"); return; }
    try { await api.patch(`/${showDecline.type}/${showDecline.id}/decline`, { catatan: declineReason }); toast.success("Ditolak"); setShowDecline(null); setShowDetail(null); setDeclineReason(""); refetchAll(); } catch (err) { toast.error(getApiErrorMessage(err, "Gagal menolak dokumen")); }
  };
  
  const handleResubmit = async (id: number, type: string) => {
    try { await api.patch(`/${type}/${id}/resubmit`); toast.success("Diajukan ulang"); setShowResubmit(null); refetchAll(); } catch (err) { toast.error(getApiErrorMessage(err, "Gagal mengajukan ulang")); }
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
    } catch (err) { toast.error(getApiErrorMessage(err, "Gagal mendownload PDF")); }
  };
  
  const downloadExcel = async () => {
    try {
      const token = localStorage.getItem("token");
      const params = new URLSearchParams();
      if (filterMonth) params.append('month', filterMonth);
      if (filterYear) params.append('year', filterYear);
      
      const res = await fetch(`${BACKEND_URL}/api/v1/rbs/download/excel?${params.toString()}`, { headers: { Authorization: `Bearer ${token}` }});
      if (!res.ok) throw new Error("Failed to download");
      
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
    } catch (err) { toast.error(getApiErrorMessage(err, "Gagal download Excel")); }
  };

  // --- LOGIKA EDIT PPD (HRGA) ---
  const saveEditPpd = async () => {
    if (!editForm) return;
    if (editClickCount >= 5) { toast.error("Batas percobaan klik tercapai (Maks 5x). Batalkan edit lalu coba lagi."); return; }

    if ((editForm.rincian_tambahan || []).some((i: any) => !(i.kuantitas ?? i.qty) || parseFloat(i.kuantitas ?? i.qty) <= 0)) {
      toast.error("Kuantitas pada estimasi biaya tidak boleh 0 atau kosong"); return;
    }

    setIsSaving(true);
    setEditClickCount((prev) => prev + 1);

    try {
      let totalHargaHotel = 0;
      
      if (editForm.rincian_hotel?.check_in && editForm.rincian_hotel?.check_out && editForm.rincian_hotel?.harga_per_malam) {
        const checkin = new Date(editForm.rincian_hotel.check_in).toISOString();
        const checkout = new Date(editForm.rincian_hotel.check_out).toISOString();
        const nights = Math.ceil((new Date(checkout).getTime() - new Date(checkin).getTime()) / (1000 * 60 * 60 * 24));
        totalHargaHotel = (parseFloat(editForm.rincian_hotel.harga_per_malam) || 0) * (nights > 0 ? nights : 0);
      }

      const rincianTambahan = (editForm.rincian_tambahan || []).map((item: any) => {
        const data: any = { kategori: item.kategori || "Lainnya", keterangan: item.keterangan || item.uraian, kuantitas: parseInt(item.kuantitas || item.qty) || 1, harga: parseFloat(item.harga || item.harga_unit) || 0 };
        if (item.id) data.id = item.id; else if (item.ID) data.id = item.ID;
        return data;
      });

      const rincianTransport = (editForm.rincian_transportasi || []).map((t: any) => {
          let jamBerangkat = t.jam_berangkat;
          if (jamBerangkat && jamBerangkat.length === 5) jamBerangkat = `${showDetail.tanggal_berangkat?.slice(0,10)}T${jamBerangkat}:00Z`;
          const data: any = { kota_asal: t.kota_asal, kota_tujuan: t.kota_tujuan, jenis_transportasi: t.jenis_transportasi, tipe_perjalanan: t.tipe_perjalanan, jam_berangkat: jamBerangkat, harga: parseFloat(t.harga) || 0 };
          if (t.nomor_kendaraan) data.nomor_kendaraan = t.nomor_kendaraan;
          if (t.id) data.id = t.id; else if (t.ID) data.id = t.ID;
          return data;
      });

      let hotelData = null;
      if (editForm.rincian_hotel?.nama_hotel) {
         hotelData = { ...editForm.rincian_hotel, harga_per_malam: parseFloat(editForm.rincian_hotel.harga_per_malam) || 0, harga_total: totalHargaHotel };
         if (editForm.rincian_hotel.id) hotelData.id = editForm.rincian_hotel.id; else if (editForm.rincian_hotel.ID) hotelData.id = editForm.rincian_hotel.ID;
      }

      const payload = {
        tujuan: showDetail.tujuan, tanggal_berangkat: showDetail.tanggal_berangkat, tanggal_kembali: showDetail.tanggal_kembali,
        keperluan: showDetail.keperluan, rincian_hotel: hotelData, rincian_transportasi: rincianTransport, rincian_tambahan: rincianTambahan
      };

      await api.put(`/ppd/${showDetail.id}/edit`, payload);
      toast.success("Perubahan data berhasil disimpan");
      setIsEditMode(false); setEditClickCount(0);
      openDetail(showDetail, "ppd"); refetchAll(); 
    } catch (error: any) {
      toast.error(getApiErrorMessage(error, "Gagal menyimpan perubahan"));
    } finally { setIsSaving(false); }
  };

  // --- LOGIKA EDIT RBS (HRGA) ---
  const saveEditRbs = async () => {
    if (!editForm) return;
    if (editClickCount >= 5) { toast.error("Batas percobaan klik tercapai (Maks 5x). Batalkan edit lalu coba lagi."); return; }

    setIsSaving(true);
    setEditClickCount((prev) => prev + 1);

    try {     
      const itemsPayload = (editForm.items || []).map((i: any) => {
        const total = parseFloat(i.total_harga ?? i.total ?? i.harga_unit ?? i.harga ?? 0) || 0;

        return {
          id: i.id || i.ID,
          kuantitas: parseInt(i.kuantitas ?? i.qty ?? 1) || 1,
          harga_unit: total,
          total_harga: total,
          total,
        };
      });

      await api.put(`/rbs/${showDetail.id}/edit`, { rincian_tambahan: itemsPayload });
      toast.success("Perubahan realisasi berhasil disimpan");
      setIsEditMode(false); setEditClickCount(0);
      openDetail(showDetail, "rbs"); refetchAll(); 
    } catch (error: any) {
      toast.error(getApiErrorMessage(error, "Gagal menyimpan perubahan realisasi"));
    } finally { setIsSaving(false); }
  };

  // State Updaters
  const updateEditFormEstimasi = (idx: number, field: string, value: any) => {
      setEditForm((prev: any) => { const newItems = [...prev.rincian_tambahan]; newItems[idx] = { ...newItems[idx], [field]: value }; return { ...prev, rincian_tambahan: newItems }; });
  };
  const updateEditFormTransport = (idx: number, field: string, value: any) => {
    setEditForm((prev: any) => { const newItems = [...prev.rincian_transportasi]; newItems[idx] = { ...newItems[idx], [field]: value }; return { ...prev, rincian_transportasi: newItems }; });
  };
  const updateEditFormRbsItem = (idx: number, field: string, value: any) => {
    setEditForm((prev: any) => {
        const newItems = [...(prev.items || [])];
        newItems[idx] = { ...newItems[idx], [field]: value };
        return { ...prev, items: newItems };
    });
  };

  const calcLivePpdTotal = () => {
      let t = 0;
      const dataToCalc = isEditMode ? editForm : showDetail;
      if (!dataToCalc) return 0;
      
      (dataToCalc.rincian_tambahan || []).forEach((i: any) => {
        const harga = parseFloat(i.harga ?? i.harga_unit) || 0;
        const qty = parseInt(i.kuantitas ?? i.qty) || 1;
        t += harga * qty;
      });
      (dataToCalc.rincian_transportasi || []).forEach((tItem: any) => t += parseFloat(tItem.harga) || 0);
      
      let n = 0;
      if (dataToCalc.rincian_hotel?.check_in && dataToCalc.rincian_hotel?.check_out) {
          const ci = new Date(dataToCalc.rincian_hotel.check_in).getTime();
          const co = new Date(dataToCalc.rincian_hotel.check_out).getTime();
          if (!isNaN(ci) && !isNaN(co)) n = Math.ceil((co - ci) / (1000 * 60 * 60 * 24));
      }
      t += (parseFloat(dataToCalc.rincian_hotel?.harga_per_malam ?? dataToCalc.rincian_hotel?.harga) || 0) * (n > 0 ? n : 0);
      return t;
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
            {tab === "bon" && ["pegawai", "atasan", "hrga"].includes(userRole) && (<Button onClick={() => setShowCreateBon(true)} className="w-full md:w-auto bg-slate-900 hover:bg-slate-800 gap-2"><Plus className="h-4 w-4" /> Ajukan Perjalanan Dinas</Button>)}
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
                    <TableHeader><TableRow className="bg-slate-50/50"><TableHead>Ref. No. Bon</TableHead><TableHead>Pemohon</TableHead><TableHead>Periode</TableHead><TableHead className="text-right">Total Realisasi</TableHead><TableHead className="text-right whitespace-nowrap min-w-[120px]">Selisih</TableHead><TableHead className="text-center">Aksi</TableHead></TableRow></TableHeader>
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
                            <TableCell className="text-sm text-right font-medium">{fmt(r.total_realisasi || 0)}</TableCell><TableCell className={`text-sm text-right font-medium whitespace-nowrap ${(r.selisih || 0) >= 0 ? "text-emerald-600" : "text-red-600"}`}>{fmt(r.selisih || 0)}</TableCell>
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
                <Table className="min-w-[800px]"><TableHeader><TableRow className="bg-slate-50/50"><TableHead>No. Dokumen</TableHead>{!isPegawai && <TableHead>Pemohon</TableHead>}<TableHead>Tujuan</TableHead><TableHead className="text-right">Total Estimasi</TableHead><TableHead className="text-center">Status</TableHead><TableHead className="text-center w-[180px]">Aksi</TableHead></TableRow></TableHeader>
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
                <Table className="min-w-[700px]"><TableHeader><TableRow className="bg-slate-50/50"><TableHead>Ref. No. Bon</TableHead>{!isPegawai && <TableHead>Pemohon</TableHead>}<TableHead className="w-48">Periode</TableHead><TableHead className="text-right">Total Realisasi</TableHead><TableHead className="text-right whitespace-nowrap min-w-[120px]">Selisih</TableHead><TableHead className="text-center">Status</TableHead><TableHead className="text-center w-[180px]">Aksi</TableHead></TableRow></TableHeader>
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

      {/* DIALOGS (LAZY LOADED & CONDITIONALLY RENDERED) */}
      {showCreateBon && (
        <Suspense fallback={<div className="fixed inset-0 z-50 flex items-center justify-center bg-black/20"><Loader2 className="h-8 w-8 animate-spin text-white" /></div>}>
          <CreatePPDDialog open={showCreateBon} onClose={() => setShowCreateBon(false)} onSuccess={refetchAll} />
        </Suspense>
      )}

      {showCreateReal && (
        <Suspense fallback={<div className="fixed inset-0 z-50 flex items-center justify-center bg-black/20"><Loader2 className="h-8 w-8 animate-spin text-white" /></div>}>
          <CreateRealisasiDialog open={showCreateReal} onClose={() => setShowCreateReal(false)} onSuccess={refetchAll} />
        </Suspense>
      )}

      {/* OTHER DIALOGS */}
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

      {/* VIEW FILE DIALOG (PDF/IMAGE) */}
      <Dialog open={!!viewFile} onOpenChange={() => setViewFile(null)}>
        <DialogContent className="sm:max-w-3xl bg-white/95 backdrop-blur">
          <DialogHeader>
            <DialogTitle>{viewFile?.title || "Lihat File"}</DialogTitle>
          </DialogHeader>

          <div className="flex justify-center p-4">
            {viewFile?.url && isPdfFile(viewFile.url) ? (
              <iframe 
                src={viewFile.url} 
                className="w-full h-[60vh] rounded border border-slate-200 shadow-sm" 
                title={viewFile.title} 
              />
            ) : (
              <img 
                src={viewFile?.url || ""} 
                alt={viewFile?.title || "Lampiran"} 
                className="max-h-[60vh] max-w-full rounded border border-slate-200 object-contain shadow-sm" 
              />
            )}
          </div>

          <DialogFooter>
            {viewFile?.url && (
              <Button
                variant="outline"
                onClick={() => window.open(viewFile.url, "_blank")}
              >
                Buka di Tab Baru
              </Button>
            )}
            <Button onClick={() => setViewFile(null)}>Tutup</Button>
          </DialogFooter>
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
                                : (isEditMode ? "Edit Realisasi Bon" : "Detail Realisasi Bon")}
                        </DialogTitle>
                        <DialogDescription>Nomor: {showDetail?.nomor_tipe_dokumen || showDetail?.nomor_dokumen || showDetail?.nomor_dokumen_referensi || "-"}</DialogDescription>
                    </div>
                    {/* Header Action: Switch ke Edit jika role HRGA dan sedang di tab approval PPD atau RBS */}
                    {(showDetail?.type === "ppd" || showDetail?.type === "rbs") && userRole === "hrga" && tab === "approval" && !isEditMode && (
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
                  {showDetail?.rincian_hotel && (showDetail.rincian_hotel.nama_hotel || isEditMode) && (
                    <>
                        <Separator />
                        <div>
                            <div className="flex justify-between mb-3">
                                <h4 className="font-semibold text-slate-700">Akomodasi / Hotel</h4>
                            </div>
                            <div className={`flex flex-col gap-2 text-sm p-3 rounded-lg border ${isEditMode ? 'bg-white border-slate-200 shadow-sm' : 'bg-slate-50 border-slate-100'}`}>
                                <div className="flex justify-between gap-2 items-center">
                                    <span className="text-slate-500 shrink-0">Hotel</span>
                                    {isEditMode ? (
                                        <Input value={editForm?.rincian_hotel?.nama_hotel || ""} onChange={(e) => setEditForm({...editForm, rincian_hotel: {...editForm.rincian_hotel, nama_hotel: e.target.value}})} className="h-8 w-48 text-right text-xs" />
                                    ) : (
                                        <span className="font-medium text-slate-900 text-right break-words">{showDetail.rincian_hotel.nama_hotel}</span>
                                    )}
                                </div>
                                <div className="flex justify-between gap-2 items-center mt-2">
                                    <span className="text-slate-500 shrink-0">Check-in</span>
                                    {isEditMode ? (
                                        <Input type="date" value={editForm?.rincian_hotel?.check_in?.substring(0,10) || ""} onChange={(e) => setEditForm({...editForm, rincian_hotel: {...editForm.rincian_hotel, check_in: e.target.value}})} className="h-8 w-36 text-xs" />
                                    ) : (
                                        <span className="font-medium text-slate-900 text-right">{formatDate(showDetail.rincian_hotel.check_in)}</span>
                                    )}
                                </div>
                                <div className="flex justify-between gap-2 items-center mt-2">
                                    <span className="text-slate-500 shrink-0">Check-out</span>
                                    {isEditMode ? (
                                        <Input type="date" value={editForm?.rincian_hotel?.check_out?.substring(0,10) || ""} onChange={(e) => setEditForm({...editForm, rincian_hotel: {...editForm.rincian_hotel, check_out: e.target.value}})} className="h-8 w-36 text-xs" />
                                    ) : (
                                        <span className="font-medium text-slate-900 text-right">{formatDate(showDetail.rincian_hotel.check_out)}</span>
                                    )}
                                </div>
                                <div className="flex justify-between border-t border-slate-100 mt-1 pt-2 gap-2 items-center">
                                    <span className="text-slate-600 font-semibold shrink-0">Harga/Malam</span>
                                    {isEditMode ? (
                                        <Input type="number" className="h-8 text-right w-32 font-bold text-slate-900" value={editForm?.rincian_hotel?.harga_per_malam ?? ""} onChange={(e) => setEditForm({...editForm, rincian_hotel: {...editForm.rincian_hotel, harga_per_malam: e.target.value}})} />
                                    ) : (
                                        <span className="font-bold text-slate-900 text-right">{fmt(showDetail.rincian_hotel.harga_per_malam || showDetail.rincian_hotel.harga)}</span>
                                    )}
                                </div>
                                {!isEditMode && (
                                <div className="flex justify-between mt-1 gap-2 items-center">
                                    <span className="text-slate-600 font-semibold shrink-0">Total Harga Hotel</span>
                                    <span className="font-bold text-slate-900 text-right">{fmt(showDetail.rincian_hotel.harga_total || showDetail.rincian_hotel.total_harga || 0)}</span>
                                </div>
                                )}
                            </div>
                        </div>
                    </>
                  )}

                  {/* Transportasi */}
                  {showDetail?.rincian_transportasi && showDetail.rincian_transportasi.length > 0 && (
                    <>
                        <Separator />
                        <div>
                            <div className="flex flex-wrap items-center justify-between gap-2 mb-3">
                                <h4 className="font-semibold text-slate-700">Transportasi</h4>
                            </div>

                            <div className="space-y-3">
                                {(isEditMode ? editForm?.rincian_transportasi : showDetail.rincian_transportasi)?.map((trans: any, idx: number) => (
                                    <div key={idx} className={`flex flex-col gap-2 text-sm p-3 rounded-lg border ${isEditMode ? 'bg-white border-slate-200 shadow-sm' : 'bg-slate-50 border-slate-100'}`}>
                                        <div className="flex justify-between gap-2 items-center">
                                            <span className="text-slate-500 shrink-0">Tipe</span>
                                            <span className="font-medium text-slate-900 text-right break-words">{trans.tipe_perjalanan}</span>
                                        </div>
                                        <div className="flex justify-between gap-2 items-center">
                                            <span className="text-slate-500 shrink-0">Jenis</span>
                                            <span className="font-medium text-slate-900 text-right break-words">{trans.jenis_transportasi}</span>
                                        </div>
                                        {trans.nomor_kendaraan && (
                                          <div className="flex justify-between gap-2 items-center">
                                              <span className="text-slate-500 shrink-0">No. Kendaraan</span>
                                              <span className="font-medium text-slate-900 text-right break-words">{trans.nomor_kendaraan}</span>
                                          </div>
                                        )}
                                        <div className="flex justify-between gap-2 items-center">
                                            <span className="text-slate-500 shrink-0">Rute</span>
                                            <span className="font-medium text-slate-900 text-right break-words">{trans.kota_asal} → {trans.kota_tujuan}</span>
                                        </div>
                                        <div className="flex justify-between border-t border-slate-100 mt-1 pt-2 gap-2 items-center">
                                            <span className="text-slate-600 font-semibold shrink-0">Harga</span>
                                            {isEditMode ? (
                                                <Input type="number" className="h-8 text-right w-32 font-bold text-slate-900" value={trans.harga ?? ""} onChange={(e) => updateEditFormTransport(idx, 'harga', e.target.value)} />
                                            ) : (
                                                <span className="font-bold text-slate-900 text-right">{fmt(trans.harga)}</span>
                                            )}
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </div>
                    </>
                  )}

                  {/* Rincian Tambahan (Estimasi) */}
                  {showDetail?.rincian_tambahan && showDetail.rincian_tambahan.length > 0 && (
                    <>
                        <Separator />
                        <div>
                            <div className="flex flex-wrap items-center justify-between gap-2 mb-3">
                                <h4 className="font-semibold text-slate-700">Estimasi Biaya Tambahan</h4>
                            </div>

                            <div className="space-y-3">
                                {(isEditMode ? editForm?.rincian_tambahan : showDetail.rincian_tambahan)?.map((item: any, idx: number) => (
                                    <div key={idx} className={`flex flex-col gap-2 text-sm p-3 rounded-lg border ${isEditMode ? 'bg-white border-slate-200 shadow-sm' : 'bg-slate-50 border-slate-100'}`}>
                                        <div className="flex justify-between gap-2 items-center">
                                            <span className="text-slate-500 shrink-0">Kategori</span>
                                            <span className="font-medium text-slate-900 text-right break-words">{item.kategori || "-"}</span>
                                        </div>
                                        <div className="flex justify-between gap-2 items-center">
                                            <span className="text-slate-500 shrink-0">Uraian</span>
                                            <span className="font-medium text-slate-900 text-right break-words">{item.keterangan || item.uraian}</span>
                                        </div>
                                        <div className="flex justify-between gap-2 items-center">
                                            <span className="text-slate-500 shrink-0">Qty</span>
                                            {isEditMode ? (
                                                <Input type="number" value={item.kuantitas ?? item.qty ?? ""} onChange={(e) => updateEditFormEstimasi(idx, 'kuantitas', e.target.value)} className="h-8 w-20 text-right text-xs" />
                                            ) : (
                                                <span className="font-medium text-slate-900 text-right">{item.kuantitas || item.qty}</span>
                                            )}
                                        </div>
                                        <div className="flex justify-between gap-2 items-center">
                                            <span className="text-slate-500 shrink-0">Harga/Item</span>
                                            {isEditMode ? (
                                                <Input type="number" className="h-8 text-right w-32 font-bold text-slate-900" value={item.harga ?? item.harga_unit ?? ""} onChange={(e) => updateEditFormEstimasi(idx, 'harga', e.target.value)} />
                                            ) : (
                                                <span className="font-medium text-slate-900 text-right">{fmt(item.harga || item.harga_unit || 0)}</span>
                                            )}
                                        </div>
                                        {!isEditMode && (
                                            <div className="flex justify-between border-t border-slate-100 mt-1 pt-2 gap-2 items-center">
                                                <span className="text-slate-600 font-semibold shrink-0">Total</span>
                                                <span className="font-bold text-slate-900 text-right">{fmt((item.harga || item.harga_unit || 0) * (item.kuantitas || item.qty || 1))}</span>
                                            </div>
                                        )}
                                    </div>
                                ))}

                                <div className={`mt-4 rounded-lg p-3 sm:p-4 shadow-sm border ${isEditMode ? 'bg-slate-50 border-slate-200' : 'bg-slate-900 text-white border-transparent'}`}>
                                    <div className="flex justify-between items-center">
                                        <span className={`text-sm font-semibold ${isEditMode ? 'text-slate-600' : 'text-slate-200'}`}>TOTAL ESTIMASI</span>
                                        <span className={`text-lg sm:text-xl font-bold ${isEditMode ? 'text-slate-900' : 'text-white'}`}>
                                            {isEditMode ? fmt(calcLivePpdTotal()) : fmt(showDetail.total_estimasi)}
                                        </span>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </>
                  )}
                </div>
              )}

              {/* DETAIL & EDIT RBS (Dikelompokkan Per Hari) */}
              {showDetail?.type === "rbs" && (() => {
                  const rawRbsItems = isEditMode ? (editForm?.items || []) : (showDetail?.items || []);

                  // Mengelompokkan berdasarkan tanggal
                  const groupedItems: Record<string, any[]> = {};
                  rawRbsItems.forEach((item: any, idx: number) => {
                      const dateKey = item.tanggal_transaksi || item.tanggal || "Tanpa Tanggal";
                      if (!groupedItems[dateKey]) groupedItems[dateKey] = [];
                      // Simpan index asli (_idx) agar bisa di-update pada editForm array
                      groupedItems[dateKey].push({ ...item, _idx: idx });
                  });

                  // Kalkulasi Total Keseluruhan
                  const grandTotal = rawRbsItems.reduce((acc: number, cur: any) => acc + (parseFloat(cur.total_harga ?? cur.total ?? cur.harga) || 0), 0);
                  const uangEstimasi = parseFloat(showDetail.uang_muka ?? showDetail.total_estimasi ?? 0) || 0;
                  const selisihAktual = uangEstimasi - grandTotal;
                  return (
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

                        {/* RINCIAN BIAYA AKTUAL (Grouped Per Hari) */}
                        <Separator />
                        <div>
                            <h4 className="font-semibold text-slate-700 mb-3">Rincian Biaya Aktual</h4>
                            {Object.keys(groupedItems).length === 0 ? (
                                <p className="text-sm text-slate-500 text-center py-6 bg-slate-50 rounded-lg border border-dashed border-slate-200">Tidak ada rincian aktual.</p>
                            ) : (
                                <div className="space-y-5">
                                    {/* Sort Date Key agar berurutan */}
                                    {Object.keys(groupedItems).sort().map((dateKey) => {
                                        const items = groupedItems[dateKey];
                                        const dayTotal = items.reduce((acc, cur) => acc + (parseFloat(cur.total_harga ?? cur.total ?? cur.harga) || 0), 0);

                                        return (
                                            <div key={dateKey} className="border border-slate-200 rounded-lg overflow-hidden">
                                                <div className="bg-slate-100 p-3 border-b border-slate-200 flex flex-col sm:flex-row sm:justify-between sm:items-center gap-1">
                                                    <span className="font-semibold text-slate-800 text-sm">
                                                        {dateKey === "Tanpa Tanggal" ? dateKey : formatDate(dateKey)}
                                                    </span>
                                                    <span className="font-bold text-slate-900 text-sm">
                                                        Total Harian: {fmt(dayTotal)}
                                                    </span>
                                                </div>
                                                <div className="p-3 bg-slate-50 space-y-3">
                                                    {items.map((item) => { 
                                                        const strukUrl = item.url_struk || item.bukti; 
                                                        return (
                                                            <div key={item._idx} className="flex flex-col gap-2 text-sm p-3 bg-white rounded-lg border border-slate-200 shadow-sm">
                                                                <div className="flex justify-between gap-2 items-start"><span className="text-slate-500 shrink-0">Kategori</span> <span className="text-right"><Badge variant="outline" className="text-xs font-normal">{item.kategori || "-"}</Badge></span></div>
                                                                <div className="flex justify-between gap-2"><span className="text-slate-500 shrink-0">Uraian</span> <span className="font-medium text-slate-900 text-right break-words">{item.uraian}</span></div>
                                                                <div className="flex justify-between border-t border-slate-100 mt-1 pt-2 gap-2 items-center">
                                                                    <span className="text-slate-600 font-semibold shrink-0">Harga / Total</span> 
                                                                    {isEditMode ? (
                                                                        <Input 
                                                                            type="number" 
                                                                            className="h-8 text-right w-32 font-bold text-slate-900" 
                                                                            value={item.total_harga ?? item.total ?? item.harga ?? ""} 
                                                                            onChange={(e) => updateEditFormRbsItem(item._idx, 'total_harga', e.target.value)} 
                                                                        />
                                                                    ) : (
                                                                        <span className="font-bold text-slate-900 text-right">{fmt(item.total_harga || item.total || item.harga || 0)}</span>
                                                                    )}
                                                                </div>

                                                                {!isEditMode && strukUrl && (
                                                                    <div className="flex justify-end mt-2">
                                                                        <Button 
                                                                          variant="outline" 
                                                                          size="sm" 
                                                                          className="h-8 text-blue-600 gap-1 text-xs w-full sm:w-auto bg-slate-50" 
                                                                          onClick={() => setViewFile({ title: "Lihat Struk", url: resolveFileUrl(strukUrl) })}
                                                                        >
                                                                            <Eye className="h-3.5 w-3.5" /> Lihat Struk
                                                                        </Button>
                                                                    </div>
                                                                )}
                                                            </div>
                                                        );
                                                    })}
                                                </div>
                                            </div>
                                        );
                                    })}
                                </div>
                            )}
                        </div>

                        {/* REKAP TOTAL GRAND TOTAL */}
                        <div className="bg-slate-50 p-3 sm:p-4 rounded-lg space-y-2 mt-4 text-sm border border-slate-200">
                          <div className="flex justify-between gap-2">
                              <span className="text-slate-600 shrink-0">Total Realisasi Keseluruhan</span>
                              <span className="font-bold text-slate-900 text-right text-base">
                                  {fmt(grandTotal)}
                              </span>
                          </div>
                          {(showDetail.uang_muka !== undefined || showDetail.total_estimasi !== undefined) && (
                              <div className="flex justify-between border-b border-slate-200 pb-2 gap-2">
                                  <span className="text-slate-600 shrink-0">Uang Muka (Estimasi Awal)</span>
                                  <span className="font-medium text-slate-900 text-right">{fmt(showDetail.uang_muka || showDetail.total_estimasi || 0)}</span>
                              </div>
                          )}
                          {(showDetail.selisih !== undefined || showDetail.sisa_bon !== undefined) && !isEditMode && (
                              <div className={`flex justify-between pt-1 font-bold gap-2 ${((selisihAktual) || 0) >= 0 ? "text-emerald-600" : "text-red-600"}`}>
                                  <span className="shrink-0">Selisih / Sisa Bon</span>
                                  <span className="text-right">{fmt(selisihAktual || 0)}</span>
                              </div>
                          )}
                        </div>

                        {/* TAMPILAN BUKTI TRANSFER (DILUAR KOTAK REKAP) */}
                        {!isEditMode && (showDetail.url_bukti_transfer || showDetail.bukti_transfer) && (
                          <div className="bg-blue-50 border border-blue-200 rounded-lg p-3 sm:p-4">
                            <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
                              <div>
                                <p className="text-xs text-blue-600 font-medium uppercase tracking-wide">Bukti Transfer Sisa Dana</p>
                                <p className="text-sm text-slate-700 mt-0.5">File bukti transfer tersedia</p>
                              </div>
                              <Button
                                variant="outline"
                                size="sm"
                                className="h-8 text-blue-600 gap-1 bg-white w-full sm:w-auto"
                                onClick={() => setViewFile({
                                  title: "Lihat Bukti Transfer",
                                  url: resolveFileUrl(showDetail.url_bukti_transfer || showDetail.bukti_transfer)
                                })}
                              >
                                <Eye className="h-3.5 w-3.5" /> Lihat Bukti Transfer
                              </Button>
                            </div>
                          </div>
                        )}
                      </div>
                  );
              })()}

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
                {/* BAGIAN KIRI */}
                <div className="flex gap-2 w-full sm:w-auto">
                    {(showDetail?.type === "ppd" || showDetail?.type === "rbs") && userRole === "hrga" && tab === "approval" && (
                        isEditMode ? (
                            <Button className={`${editClickCount >= 5 ? 'bg-slate-400 cursor-not-allowed' : 'bg-emerald-600 hover:bg-emerald-700'} text-white gap-1 w-full sm:w-auto`} onClick={showDetail.type === "ppd" ? saveEditPpd : saveEditRbs} disabled={isSaving || editClickCount >= 5}>
                                {isSaving ? "Menyimpan..." : editClickCount >= 5 ? "Batas Klik Tercapai" : <><Save className="h-4 w-4"/> Simpan Perubahan</>}
                            </Button>
                        ) : (
                            <>
                                <Button className="bg-emerald-600 hover:bg-emerald-700 text-white gap-1 flex-1 sm:flex-none" onClick={() => approve(showDetail.id, showDetail.type)}><CheckCircle2 className="h-4 w-4"/> Setuju</Button>
                                <Button variant="outline" className="text-red-600 border-red-200 hover:bg-red-50 gap-1 flex-1 sm:flex-none" onClick={() => setShowDecline({ ...showDetail, type: showDetail.type })}><XCircle className="h-4 w-4"/> Tolak</Button>
                            </>
                        )
                    )}
                </div>

                {/* BAGIAN KANAN */}
                <div className="flex flex-col sm:flex-row gap-2 w-full sm:w-auto justify-end">
                    {isEditMode ? (
                        <Button variant="outline" onClick={() => { setIsEditMode(false); setEditForm(null); setEditClickCount(0); }} className="w-full sm:w-auto">Batal Edit</Button>
                    ) : (
                        <>
                            <Button variant="outline" onClick={() => setShowDetail(null)} className="w-full sm:w-auto">Tutup</Button>
                            
                            {/* Download PPD */}
                            {showDetail?.is_downloadable && showDetail?.type === "ppd" && showDetail?.status !== "Draft" && !showDetail?.status?.toLowerCase().includes("ditolak") && (
                                <Button className="bg-blue-600 hover:bg-blue-700 text-white gap-1 w-full sm:w-auto" onClick={() => downloadPdf(showDetail.id, "ppd", formatDocName(showDetail, "Pengajuan Perjalanan Dinas"), "rpd")}>
                                    <FileText className="h-4 w-4" /> Download PPD
                                </Button>
                            )}

                            {/* Download Bon */}
                            {showDetail?.type === "ppd" && showDetail?.status === "Selesai" && (
                                <Button className="bg-emerald-600 hover:bg-emerald-700 text-white gap-1 w-full sm:w-auto" onClick={() => downloadPdf(showDetail.id, "ppd", formatDocName(showDetail, "Bon Sementara"), "bs")}>
                                    <Download className="h-4 w-4" /> Download Bon
                                </Button>
                            )}

                            {/* Download RBS */}
                            {showDetail?.type === "rbs" && showDetail?.status === "Selesai" && (
                                <Button className="bg-emerald-600 hover:bg-emerald-700 text-white gap-1 w-full sm:w-auto" onClick={() => downloadPdf(showDetail.id, "rbs", formatDocName(showDetail, "Realisasi Bon Sementara"))}>
                                    <Download className="h-4 w-4" /> Download RBS
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