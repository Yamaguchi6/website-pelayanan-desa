package main

import "os"

// Desa berisi identitas desa. Ubah lewat variabel lingkungan atau langsung di sini.
type Desa struct {
	Nama      string
	Kecamatan string
	Kabupaten string
	Alamat    string
	Telepon   string
	Email     string
	Jam       []string
}

func muatDesa() Desa {
	return Desa{
		Nama:      env("DESA_NAMA", "Desa Sukamaju"),
		Kecamatan: env("DESA_KECAMATAN", "Kecamatan Contoh"),
		Kabupaten: env("DESA_KABUPATEN", "Kabupaten Contoh"),
		Alamat:    env("DESA_ALAMAT", "Jl. Balai Desa No. 1"),
		Telepon:   env("DESA_TELEPON", "0812-0000-0000"),
		Email:     env("DESA_EMAIL", "desa@example.id"),
		Jam: []string{
			"Senin sampai Kamis: 08.00 - 15.00",
			"Jumat: 08.00 - 11.30",
			"Sabtu dan Minggu: tutup",
		},
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// Layanan adalah jenis surat yang bisa diajukan warga.
// Syarat dan estimasi hanya contoh: sesuaikan dengan aturan desa Anda.
type Layanan struct {
	Slug      string
	Nama      string
	Deskripsi string
	Estimasi  string
	Syarat    []string
}

var daftarLayanan = []Layanan{
	{"domisili", "Surat Keterangan Domisili", "Menyatakan tempat tinggal Anda saat ini.", "1 hari kerja",
		[]string{"Fotokopi KTP", "Fotokopi Kartu Keluarga", "Surat pengantar RT/RW"}},
	{"sktm", "Surat Keterangan Tidak Mampu", "Untuk keperluan bantuan, beasiswa, atau layanan kesehatan.", "2 hari kerja",
		[]string{"Fotokopi KTP", "Fotokopi Kartu Keluarga", "Surat pengantar RT/RW", "Foto rumah tampak depan"}},
	{"usaha", "Surat Keterangan Usaha", "Menyatakan Anda menjalankan usaha di wilayah desa.", "2 hari kerja",
		[]string{"Fotokopi KTP", "Fotokopi Kartu Keluarga", "Foto tempat usaha"}},
	{"pengantar-ktp-kk", "Surat Pengantar KTP/KK", "Pengantar pembuatan atau perubahan KTP dan Kartu Keluarga ke kecamatan.", "1 hari kerja",
		[]string{"Fotokopi KTP atau KK lama", "Dokumen pendukung perubahan data (akta, surat nikah, dan sejenisnya)"}},
	{"kelahiran", "Surat Keterangan Kelahiran", "Pengantar pembuatan akta kelahiran.", "1 hari kerja",
		[]string{"Surat keterangan lahir dari bidan atau rumah sakit", "Fotokopi KTP kedua orang tua", "Fotokopi Kartu Keluarga", "Fotokopi surat nikah"}},
	{"kematian", "Surat Keterangan Kematian", "Pengantar pembuatan akta kematian dan pengurusan waris.", "1 hari kerja",
		[]string{"Fotokopi KTP almarhum/almarhumah", "Fotokopi Kartu Keluarga", "KTP pelapor dan 2 orang saksi"}},
	{"skck", "Surat Pengantar SKCK", "Pengantar pembuatan Surat Keterangan Catatan Kepolisian.", "1 hari kerja",
		[]string{"Fotokopi KTP", "Fotokopi Kartu Keluarga", "Pas foto 4x6 (2 lembar)"}},
	{"pindah", "Surat Keterangan Pindah", "Pengantar pindah domisili ke desa atau kota lain.", "2 hari kerja",
		[]string{"Fotokopi KTP", "Fotokopi Kartu Keluarga", "Alamat tujuan pindah"}},
}

func cariLayanan(slug string) (Layanan, bool) {
	for _, l := range daftarLayanan {
		if l.Slug == slug {
			return l, true
		}
	}
	return Layanan{}, false
}

// Pengumuman tampil di beranda. Ganti isinya sesuai kabar terbaru desa.
type Pengumuman struct {
	Judul   string
	Isi     string
	Tanggal string
}

var daftarPengumuman = []Pengumuman{
	{"Pelayanan Jumat dipersingkat", "Kantor desa melayani sampai pukul 11.30 setiap hari Jumat.", "10 November 2026"},
	{"Pembaruan data Kartu Keluarga", "Warga yang datanya berubah dapat mengajukan surat pengantar lewat situs ini.", "3 November 2026"},
}