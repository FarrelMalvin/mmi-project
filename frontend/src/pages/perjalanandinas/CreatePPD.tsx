import React, { useState } from "react";
import { api } from "../../lib/api";
import { getApiErrorMessage } from "../../lib/error";
import { Button } from "../../components/common/Button";
import { Input } from "../../components/common/Input";
import { Label } from "../../components/common/Label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../../components/common/Select";
import { Separator } from "../../components/common/Separator";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "../../components/common/Dialog";
import { toast } from "sonner";
import { Plus, Trash2 } from "lucide-react";

const fmt = (n: number) =>
  new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
  }).format(n);

export default function CreatePPDDialog({
  open,
  onClose,
  onSuccess,
}: {
  open: boolean;
  onClose: () => void;
  onSuccess: () => void;
}) {
  const [loading, setLoading] = useState(false);

  const [bonForm, setBonForm] = useState<any>({
    tujuan: "",
    keperluan: "",
    url_dokumen: "",
    tanggal_berangkat: "",
    tanggal_kembali: "",
    rincian_hotel: {
      nama_hotel: "",
      check_in: "",
      check_out: "",
      harga_per_malam: 0,
    },
    rincian_transportasi: [],
    rincian_tambahan: [
      {
        kategori: "Konsumsi",
        keterangan: "",
        kuantitas: 1,
        harga: 0,
      },
    ],
  });

  const hariIni = new Date().toISOString().split("T")[0];

  const calculateNights = (checkIn: string, checkOut: string) => {
    if (!checkIn || !checkOut) return 0;

    const start = new Date(checkIn);
    const end = new Date(checkOut);

    if (isNaN(start.getTime()) || isNaN(end.getTime())) return 0;

    const diff = Math.ceil(
      (end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24)
    );

    return diff > 0 ? diff : 0;
  };

  const calculateHotelTotal = () => {
    if (!bonForm.rincian_hotel?.nama_hotel) return 0;

    const nights = calculateNights(
      bonForm.rincian_hotel.check_in,
      bonForm.rincian_hotel.check_out
    );

    const hargaPerMalam =
      parseFloat(bonForm.rincian_hotel.harga_per_malam) || 0;

    return hargaPerMalam * nights;
  };

  const handleBerangkatChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value;

    setBonForm((prev: any) => {
      const updates: any = { tanggal_berangkat: val };

      if (prev.tanggal_kembali && val > prev.tanggal_kembali) {
        updates.tanggal_kembali = "";
      }

      if (prev.rincian_hotel.check_in && val > prev.rincian_hotel.check_in) {
        updates.rincian_hotel = {
          ...prev.rincian_hotel,
          check_in: "",
          check_out: "",
        };
      }

      return { ...prev, ...updates };
    });
  };

  const calcEstimasiTotal = () => {
    let total = 0;

    bonForm.rincian_tambahan.forEach((item: any) => {
      const harga = parseFloat(item.harga) || 0;
      const qty = parseInt(item.kuantitas) || 1;

      total += harga * qty;
    });

    total += calculateHotelTotal();

    bonForm.rincian_transportasi.forEach((trans: any) => {
      total += parseFloat(trans.harga) || 0;
    });

    return total;
  };

  const addEstimasiItem = () =>
    setBonForm((p: any) => ({
      ...p,
      rincian_tambahan: [
        ...p.rincian_tambahan,
        {
          kategori: "Konsumsi",
          keterangan: "",
          kuantitas: 1,
          harga: 0,
        },
      ],
    }));

  const removeEstimasiItem = (idx: number) =>
    setBonForm((p: any) => ({
      ...p,
      rincian_tambahan: p.rincian_tambahan.filter(
        (_: any, i: number) => i !== idx
      ),
    }));

  const updateEstimasiItem = (idx: number, field: string, val: string) => {
    setBonForm((p: any) => {
      const items = [...p.rincian_tambahan];

      let updatedVal: any = val;
      if (field === "harga") updatedVal = val === "" ? "" : parseFloat(val);
      if (field === "kuantitas") updatedVal = val === "" ? "" : parseInt(val);

      items[idx] = {
        ...items[idx],
        [field]: updatedVal,
      };

      return {
        ...p,
        rincian_tambahan: items,
      };
    });
  };

  const addTransportItem = () => {
    if (bonForm.rincian_transportasi.length >= 2) {
      toast.error("Maksimal 2 rute transportasi");
      return;
    }

    setBonForm((p: any) => ({
      ...p,
      rincian_transportasi: [
        ...p.rincian_transportasi,
        {
          kota_asal: "",
          kota_tujuan: "",
          jenis_transportasi: "",
          nomor_kendaraan: "",
          uraian_lainnya: "",
          tipe_perjalanan: "Keberangkatan",
          jam_berangkat: "",
          harga: 0,
        },
      ],
    }));
  };

  const removeTransportItem = (idx: number) =>
    setBonForm((p: any) => ({
      ...p,
      rincian_transportasi: p.rincian_transportasi.filter(
        (_: any, i: number) => i !== idx
      ),
    }));

  const updateTransportItem = (idx: number, field: string, val: string) => {
    setBonForm((p: any) => {
      const items = [...p.rincian_transportasi];
      const updatedVal = field === "harga" ? (val === "" ? "" : parseFloat(val)) : val;

      items[idx] = {
        ...items[idx],
        [field]: updatedVal,
      };

      if (field === "jenis_transportasi") {
        if (val !== "Kendaraan Dinas" && val !== "Kendaraan Pribadi") {
          items[idx].nomor_kendaraan = "";
        }

        if (val !== "Lain-lain") {
          items[idx].uraian_lainnya = "";
        }
      }

      return {
        ...p,
        rincian_transportasi: items,
      };
    });
  };

  const resetForm = () => {
    setBonForm({
      tujuan: "",
      tanggal_berangkat: "",
      tanggal_kembali: "",
      keperluan: "",
      url_dokumen: "",
      rincian_hotel: {
        nama_hotel: "",
        check_in: "",
        check_out: "",
        harga_per_malam: 0,
      },
      rincian_transportasi: [],
      rincian_tambahan: [
        {
          kategori: "Konsumsi",
          keterangan: "",
          kuantitas: 1,
          harga: 0,
        },
      ],
    });
  };

  const submitBon = async () => {
    if (
      !bonForm.tujuan ||
      !bonForm.tanggal_berangkat ||
      !bonForm.tanggal_kembali ||
      !bonForm.keperluan
    ) {
      toast.error("Field tujuan, tanggal, dan keperluan wajib diisi");
      return;
    }

    if (bonForm.tanggal_kembali < bonForm.tanggal_berangkat) {
      toast.error("Tanggal kembali tidak boleh kurang dari tanggal berangkat");
      return;
    }

    if (bonForm.rincian_hotel.nama_hotel) {
      if (!bonForm.rincian_hotel.check_in || !bonForm.rincian_hotel.check_out) {
        toast.error("Check-In dan Check-Out hotel wajib diisi jika mengisi hotel");
        return;
      }

      if (
        !bonForm.rincian_hotel.harga_per_malam ||
        parseFloat(bonForm.rincian_hotel.harga_per_malam) <= 0
      ) {
        toast.error("Harga per malam hotel wajib diisi dan harus lebih dari 0");
        return;
      }

      if (bonForm.rincian_hotel.check_in < bonForm.tanggal_berangkat) {
        toast.error("Check-In hotel tidak boleh kurang dari tanggal berangkat");
        return;
      }

      if (bonForm.rincian_hotel.check_out > bonForm.tanggal_kembali) {
        toast.error("Check-Out hotel tidak boleh melebihi tanggal kembali");
        return;
      }

      if (bonForm.rincian_hotel.check_out <= bonForm.rincian_hotel.check_in) {
        toast.error("Tanggal Check-Out hotel harus lebih besar dari Check-In");
        return;
      }
    }

    if (
      bonForm.rincian_tambahan.some(
        (i: any) =>
          i.kuantitas === "" ||
          !i.kuantitas ||
          parseFloat(i.kuantitas) <= 0 ||
          i.harga === "" ||
          i.harga == null ||
          parseFloat(i.harga) < 0
      )
    ) {
      toast.error("Kuantitas dan Harga pada estimasi biaya tidak boleh kosong atau tidak valid");
      return;
    }

    if (
      bonForm.rincian_transportasi.some(
        (t: any) =>
          !t.jenis_transportasi ||
          !t.tipe_perjalanan ||
          !t.kota_asal ||
          !t.kota_tujuan
      )
    ) {
      toast.error("Data transportasi belum lengkap");
      return;
    }

    if (
      bonForm.rincian_transportasi.some(
        (t: any) =>
          (t.jenis_transportasi === "Kendaraan Dinas" ||
            t.jenis_transportasi === "Kendaraan Pribadi") &&
          !t.nomor_kendaraan
      )
    ) {
      toast.error("Nomor Kendaraan wajib diisi untuk Kendaraan Dinas atau Pribadi");
      return;
    }

    if (
      bonForm.rincian_transportasi.some(
        (t: any) => t.jenis_transportasi === "Lain-lain" && !t.uraian_lainnya
      )
    ) {
      toast.error("Uraian wajib diisi jika jenis angkutan memilih Lain-lain");
      return;
    }

    if (
      bonForm.rincian_transportasi.some(
        (t: any) =>
          t.harga === "" || t.harga == null || parseFloat(t.harga) < 0
      )
    ) {
      toast.error("Harga pada rincian transportasi tidak boleh kosong atau kurang dari 0");
      return;
    }

    setLoading(true);

    try {
      const totalHargaHotel = calculateHotelTotal();

      const payload = {
        tujuan: bonForm.tujuan,
        tanggal_berangkat: `${bonForm.tanggal_berangkat}T00:00:00Z`,
        tanggal_kembali: `${bonForm.tanggal_kembali}T00:00:00Z`,
        keperluan: bonForm.keperluan,
        url_dokumen: bonForm.url_dokumen || "",

        rincian_hotel: bonForm.rincian_hotel.nama_hotel
          ? {
              nama_hotel: bonForm.rincian_hotel.nama_hotel,
              check_in: `${bonForm.rincian_hotel.check_in}T00:00:00Z`,
              check_out: `${bonForm.rincian_hotel.check_out}T00:00:00Z`,
              harga_per_malam:
                parseFloat(bonForm.rincian_hotel.harga_per_malam) || 0,
              harga_total: totalHargaHotel,
            }
          : null,

        rincian_transportasi: bonForm.rincian_transportasi.map((t: any) => ({
          tipe_perjalanan: t.tipe_perjalanan,
          kota_asal: t.kota_asal,
          kota_tujuan: t.kota_tujuan,
          jenis_transportasi:
            t.jenis_transportasi === "Lain-lain"
              ? `Lain-lain - ${t.uraian_lainnya}`
              : t.jenis_transportasi,
          harga: parseFloat(t.harga) || 0,
          jam_berangkat: t.jam_berangkat
            ? `${bonForm.tanggal_berangkat}T${t.jam_berangkat}:00Z`
            : "",
          nomor_kendaraan:
            t.jenis_transportasi === "Kendaraan Dinas" ||
            t.jenis_transportasi === "Kendaraan Pribadi"
              ? t.nomor_kendaraan
              : "",
        })),

        rincian_tambahan: bonForm.rincian_tambahan.map((item: any) => ({
          kategori: item.kategori,
          keterangan: item.keterangan,
          kuantitas: parseInt(item.kuantitas) || 1,
          harga: parseFloat(item.harga) || 0,
        })),
      };

      await api.post("/ppd", payload);

      toast.success("Perjalanan Dinas berhasil diajukan");

      onSuccess();
      onClose();
      resetForm();
    } catch (err: any) {
      toast.error(getApiErrorMessage(err, "Gagal mengirim pengajuan"));
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-3xl max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Pengajuan Perjalanan Dinas</DialogTitle>
        </DialogHeader>

        <div className="space-y-5">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div>
              <Label>Tujuan *</Label>
              <Input
                placeholder="Kota tujuan"
                value={bonForm.tujuan}
                onChange={(e) =>
                  setBonForm((p: any) => ({
                    ...p,
                    tujuan: e.target.value,
                  }))
                }
              />
            </div>

            <div>
              <Label>Keperluan *</Label>
              <Input
                placeholder="Tujuan perjalanan"
                value={bonForm.keperluan}
                onChange={(e) =>
                  setBonForm((p: any) => ({
                    ...p,
                    keperluan: e.target.value,
                  }))
                }
              />
            </div>

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
                onChange={(e) =>
                  setBonForm((p: any) => ({
                    ...p,
                    tanggal_kembali: e.target.value,
                  }))
                }
              />
            </div>
          </div>

          <Separator />

          <div>
            <h4 className="text-sm font-bold mb-3">
              AKOMODASI / HOTEL (Opsional)
            </h4>

            <div className="grid grid-cols-12 gap-3">
              <div className="col-span-12 sm:col-span-6">
                <Label className="text-xs">Nama Hotel</Label>
                <Input
                  placeholder="Nama hotel (kosongkan jika tidak ada)"
                  value={bonForm.rincian_hotel.nama_hotel}
                  onChange={(e) =>
                    setBonForm((p: any) => ({
                      ...p,
                      rincian_hotel: {
                        ...p.rincian_hotel,
                        nama_hotel: e.target.value,
                      },
                    }))
                  }
                  className="h-9"
                />
              </div>

              <div className="col-span-12 sm:col-span-3">
                <Label className="text-xs">Check-In</Label>
                <Input
                  type="date"
                  value={bonForm.rincian_hotel.check_in}
                  min={bonForm.tanggal_berangkat || hariIni}
                  max={bonForm.tanggal_kembali || ""}
                  onChange={(e) =>
                    setBonForm((p: any) => ({
                      ...p,
                      rincian_hotel: {
                        ...p.rincian_hotel,
                        check_in: e.target.value,
                      },
                    }))
                  }
                  className="h-9"
                />
              </div>

              <div className="col-span-12 sm:col-span-3">
                <Label className="text-xs">Check-Out</Label>
                <Input
                  type="date"
                  value={bonForm.rincian_hotel.check_out}
                  min={
                    bonForm.rincian_hotel.check_in ||
                    bonForm.tanggal_berangkat ||
                    hariIni
                  }
                  max={bonForm.tanggal_kembali || ""}
                  onChange={(e) =>
                    setBonForm((p: any) => ({
                      ...p,
                      rincian_hotel: {
                        ...p.rincian_hotel,
                        check_out: e.target.value,
                      },
                    }))
                  }
                  className="h-9"
                />
              </div>

              <div className="col-span-12 sm:col-span-4">
                <Label className="text-xs">Harga per Malam (Rp)</Label>
                <Input
                  type="number"
                  value={
                    bonForm.rincian_hotel.harga_per_malam === 0
                      ? ""
                      : bonForm.rincian_hotel.harga_per_malam ?? ""
                  }
                  onChange={(e) => {
                    const val = e.target.value;

                    setBonForm((p: any) => ({
                      ...p,
                      rincian_hotel: {
                        ...p.rincian_hotel,
                        harga_per_malam:
                          val === "" ? "" : parseFloat(val),
                      },
                    }));
                  }}
                  className="h-9"
                />
              </div>

              {bonForm.rincian_hotel.nama_hotel && (
                <div className="col-span-12 bg-slate-50 border border-slate-200 rounded-lg p-3 text-sm space-y-1">
                  <div className="flex justify-between">
                    <span className="text-slate-600">Jumlah Malam</span>
                    <span className="font-medium">
                      {calculateNights(
                        bonForm.rincian_hotel.check_in,
                        bonForm.rincian_hotel.check_out
                      )}{" "}
                      malam
                    </span>
                  </div>

                  <div className="flex justify-between">
                    <span className="text-slate-600">Total Hotel</span>
                    <span className="font-bold text-slate-900">
                      {fmt(calculateHotelTotal())}
                    </span>
                  </div>
                </div>
              )}
            </div>
          </div>

          <Separator />

          <div>
            <div className="flex items-center justify-between mb-3">
              <h4 className="text-sm font-bold">TRANSPORTASI (Opsional)</h4>

              {bonForm.rincian_transportasi.length < 2 && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={addTransportItem}
                  className="h-7 gap-1"
                >
                  <Plus className="h-3 w-3" />
                  Tambah Rute
                </Button>
              )}
            </div>

            <div className="space-y-2">
              {bonForm.rincian_transportasi.map((item: any, idx: number) => (
                <div
                  key={idx}
                  className="grid grid-cols-12 gap-2 items-end border border-slate-200 rounded-lg p-3"
                >
                  <div className="col-span-12 sm:col-span-4">
                    <Label className="text-xs">Tipe Perjalanan</Label>
                    <Select
                      value={item.tipe_perjalanan}
                      onValueChange={(v) =>
                        updateTransportItem(idx, "tipe_perjalanan", v)
                      }
                    >
                      <SelectTrigger className="h-9">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="Keberangkatan">
                          Keberangkatan
                        </SelectItem>
                        <SelectItem value="Kedatangan">Kedatangan</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="col-span-12 sm:col-span-4">
                    <Label className="text-xs">Jenis Angkutan</Label>
                    <Select
                      value={item.jenis_transportasi}
                      onValueChange={(v) =>
                        updateTransportItem(idx, "jenis_transportasi", v)
                      }
                    >
                      <SelectTrigger className="h-9">
                        <SelectValue placeholder="Pilih Angkutan" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="Pesawat">Pesawat</SelectItem>
                        <SelectItem value="Kereta Api">Kereta Api</SelectItem>
                        <SelectItem value="Bus">Bus</SelectItem>
                        <SelectItem value="Kendaraan Dinas">
                          Kendaraan Dinas
                        </SelectItem>
                        <SelectItem value="Kendaraan Pribadi">
                          Kendaraan Pribadi
                        </SelectItem>
                        <SelectItem value="Lain-lain">Lain-lain</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="col-span-12 sm:col-span-4">
                    <Label className="text-xs">Harga (Rp)</Label>
                    <Input
                      type="number"
                      value={item.harga === 0 ? "" : item.harga ?? ""}
                      onChange={(e) =>
                        updateTransportItem(idx, "harga", e.target.value)
                      }
                      className="h-9"
                    />
                  </div>

                  <div className="col-span-12 sm:col-span-4 mt-2">
                    <Label className="text-xs">Kota Asal</Label>
                    <Input
                      placeholder="Dari mana"
                      value={item.kota_asal}
                      onChange={(e) =>
                        updateTransportItem(idx, "kota_asal", e.target.value)
                      }
                      className="h-9"
                    />
                  </div>

                  <div className="col-span-12 sm:col-span-4 mt-2">
                    <Label className="text-xs">Kota Tujuan</Label>
                    <Input
                      placeholder="Ke mana"
                      value={item.kota_tujuan}
                      onChange={(e) =>
                        updateTransportItem(idx, "kota_tujuan", e.target.value)
                      }
                      className="h-9"
                    />
                  </div>

                  <div className="col-span-12 sm:col-span-4 mt-2">
                    <Label className="text-xs">Jam Keberangkatan</Label>
                    <Input
                      type="time"
                      value={item.jam_berangkat}
                      onChange={(e) =>
                        updateTransportItem(
                          idx,
                          "jam_berangkat",
                          e.target.value
                        )
                      }
                      className="h-9"
                    />
                  </div>

                  {(item.jenis_transportasi === "Kendaraan Dinas" ||
                    item.jenis_transportasi === "Kendaraan Pribadi") && (
                    <div className="col-span-12 sm:col-span-4 mt-2">
                      <Label className="text-xs">Nomor Kendaraan</Label>
                      <Input
                        placeholder="Contoh: B 1234 CD"
                        value={item.nomor_kendaraan || ""}
                        onChange={(e) =>
                          updateTransportItem(
                            idx,
                            "nomor_kendaraan",
                            e.target.value
                          )
                        }
                        className="h-9 border-blue-300 focus-visible:ring-blue-500"
                      />
                    </div>
                  )}

                  {item.jenis_transportasi === "Lain-lain" && (
                    <div className="col-span-12 sm:col-span-4 mt-2">
                      <Label className="text-xs">Uraian Kendaraan</Label>
                      <Input
                        placeholder="Sebutkan jenis kendaraan"
                        value={item.uraian_lainnya || ""}
                        onChange={(e) =>
                          updateTransportItem(
                            idx,
                            "uraian_lainnya",
                            e.target.value
                          )
                        }
                        className="h-9 border-blue-300 focus-visible:ring-blue-500"
                      />
                    </div>
                  )}

                  <div className="col-span-12 flex justify-end mt-2">
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-7 gap-1 text-red-500"
                      onClick={() => removeTransportItem(idx)}
                    >
                      <Trash2 className="h-3 w-3" />
                      Hapus Rute
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          </div>

          <Separator />

          <div className="flex items-center justify-between">
            <h4 className="text-sm font-bold">ESTIMASI BIAYA LAINNYA</h4>
            <Button
              variant="outline"
              size="sm"
              onClick={addEstimasiItem}
              className="h-7 gap-1"
            >
              <Plus className="h-3 w-3" />
              Tambah
            </Button>
          </div>

          <div className="space-y-2">
            {bonForm.rincian_tambahan.map((item: any, idx: number) => (
              <div
                key={idx}
                className="grid grid-cols-12 gap-2 items-end border border-slate-200 rounded-lg p-3"
              >
                <div className="col-span-12 sm:col-span-3">
                  <Label className="text-xs">Kategori</Label>
                  <Select
                    value={item.kategori}
                    onValueChange={(v) =>
                      updateEstimasiItem(idx, "kategori", v)
                    }
                  >
                    <SelectTrigger className="h-9">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="Konsumsi">Konsumsi</SelectItem>
                      <SelectItem value="Transportasi">Transportasi</SelectItem>
                      <SelectItem value="BBM">BBM</SelectItem>
                      <SelectItem value="Entertainment">
                        Entertainment
                      </SelectItem>
                      <SelectItem value="Parkir">Parkir</SelectItem>
                      <SelectItem value="Tol">Tol</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                <div className="col-span-12 sm:col-span-4">
                  <Label className="text-xs">Keterangan</Label>
                  <Input
                    placeholder="Keterangan"
                    value={item.keterangan}
                    onChange={(e) =>
                      updateEstimasiItem(idx, "keterangan", e.target.value)
                    }
                    className="h-9"
                  />
                </div>

                <div className="col-span-6 sm:col-span-2">
                  <Label className="text-xs">Qty</Label>
                  <Input
                    type="number"
                    value={item.kuantitas === 0 ? "" : item.kuantitas ?? ""}
                    onChange={(e) =>
                      updateEstimasiItem(idx, "kuantitas", e.target.value)
                    }
                    className="h-9"
                  />
                </div>

                <div className="col-span-6 sm:col-span-3">
                  <Label className="text-xs">Harga (Rp)</Label>
                  <Input
                    type="number"
                    value={item.harga === 0 ? "" : item.harga ?? ""}
                    onChange={(e) =>
                      updateEstimasiItem(idx, "harga", e.target.value)
                    }
                    className="h-9"
                  />
                </div>

                {bonForm.rincian_tambahan.length > 1 && (
                  <div className="col-span-12 flex justify-end mt-2">
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-7 gap-1 text-red-500"
                      onClick={() => removeEstimasiItem(idx)}
                    >
                      <Trash2 className="h-3 w-3" />
                      Hapus
                    </Button>
                  </div>
                )}
              </div>
            ))}
          </div>

          <div className="bg-slate-50 p-3 rounded-lg text-right">
            <p className="text-sm font-bold text-slate-900">
              TOTAL ESTIMASI KESELURUHAN: {fmt(calcEstimasiTotal())}
            </p>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            Batal
          </Button>
          <Button
            onClick={submitBon}
            disabled={loading}
            className="bg-slate-900 hover:bg-slate-800"
          >
            {loading ? "Mengirim..." : "Ajukan Perjalanan"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}