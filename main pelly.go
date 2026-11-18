package main

import (
	"crypto/subtle"
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
	_ "time/tzdata"
)

//go:embed templates/*.html static/*
var content embed.FS

var (
	nikRe     = regexp.MustCompile(`^\d{16}$`)
	telpRe    = regexp.MustCompile(`^(\+62|0)\d{8,13}$`)
	alurLayan = []string{"Diterima", "Diverifikasi", "Diproses", "Siap diambil", "Selesai"}
	bulan     = []string{"Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
)

type App struct {
	tpl       map[string]*template.Template
	store     *Store
	desa      Desa
	adminUser string
	adminPass string
}

type Langkah struct {
	Nama  string
	Kelas string
}

func main() {
	store, err := NewStore(env("DATA_FILE", "data/pengajuan.json"))
	if err != nil {
		log.Fatal(err)
	}
	a := &App{store: store, desa: muatDesa(), adminUser: env("ADMIN_USER", "admin"), adminPass: os.Getenv("ADMIN_PASS")}
	if err := a.muatTemplate(); err != nil {
		log.Fatal(err)
	}
	if a.adminPass == "" {
		log.Println("ADMIN_PASS belum diatur: halaman /admin dinonaktifkan")
	}

	static, _ := fs.Sub(content, "static")
	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(static)))
	mux.HandleFunc("GET /{$}", a.beranda)
	mux.HandleFunc("GET /layanan", a.layanan)
	mux.HandleFunc("GET /ajukan", a.formAjukan)
	mux.HandleFunc("POST /ajukan", a.kirimAjukan)
	mux.HandleFunc("GET /status", a.status)
	mux.HandleFunc("GET /admin", a.admin(a.adminDaftar))
	mux.HandleFunc("POST /admin/ubah", a.admin(a.adminUbah))

	srv := &http.Server{
		Addr:              ":" + env("PORT", "8080"),
		Handler:           amankan(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Printf("Layanan %s berjalan di http://localhost%s", a.desa.Nama, srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

func (a *App) muatTemplate() error {
	wib, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return err
	}
	funcs := template.FuncMap{
		"tgl": func(t time.Time) string {
			t = t.In(wib)
			return t.Format("2") + " " + bulan[t.Month()-1] + " " + t.Format("2006, 15.04") + " WIB"
		},
	}
	a.tpl = map[string]*template.Template{}
	for _, p := range []string{"home", "layanan", "ajukan", "status", "admin"} {
		t, err := template.New(p).Funcs(funcs).ParseFS(content, "templates/layout.html", "templates/"+p+".html")
		if err != nil {
			return err
		}
		a.tpl[p] = t
	}
	return nil
}

func amankan(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "same-origin")
		h.Set("Content-Security-Policy", "default-src 'self'; style-src 'self' https://fonts.googleapis.com; font-src https://fonts.gstatic.com; img-src 'self' data:")
		next.ServeHTTP(w, r)
	})
}

func (a *App) render(w http.ResponseWriter, page, judul string, kode int, data map[string]any) {
	data["Title"] = judul
	data["Desa"] = a.desa
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(kode)
	if err := a.tpl[page].ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("render %s: %v", page, err)
	}
}

func (a *App) beranda(w http.ResponseWriter, r *http.Request) {
	a.render(w, "home", "Beranda", http.StatusOK, map[string]any{
		"Layanan":    daftarLayanan,
		"Pengumuman": daftarPengumuman,
	})
}

func (a *App) layanan(w http.ResponseWriter, r *http.Request) {
	a.render(w, "layanan", "Layanan", http.StatusOK, map[string]any{"Layanan": daftarLayanan})
}

func (a *App) formAjukan(w http.ResponseWriter, r *http.Request) {
	form := map[string]string{"layanan": r.URL.Query().Get("layanan")}
	a.tampilForm(w, form, nil, http.StatusOK)
}

func (a *App) tampilForm(w http.ResponseWriter, form, errs map[string]string, kode int) {
	a.render(w, "ajukan", "Ajukan layanan", kode, map[string]any{
		"Layanan": daftarLayanan, "Form": form, "Errors": errs,
	})
}

func (a *App) kirimAjukan(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Data terlalu besar", http.StatusRequestEntityTooLarge)
		return
	}
	pembersih := strings.NewReplacer(" ", "", "-", "")
	form := map[string]string{
		"nama":      strings.TrimSpace(r.FormValue("nama")),
		"nik":       pembersih.Replace(r.FormValue("nik")),
		"telepon":   pembersih.Replace(r.FormValue("telepon")),
		"layanan":   r.FormValue("layanan"),
		"keperluan": strings.TrimSpace(r.FormValue("keperluan")),
	}
	errs := map[string]string{}
	if len([]rune(form["nama"])) < 3 || len([]rune(form["nama"])) > 100 {
		errs["nama"] = "Tulis nama lengkap sesuai KTP (3 sampai 100 huruf)."
	}
	if !nikRe.MatchString(form["nik"]) {
		errs["nik"] = "NIK harus 16 angka."
	}
	if !telpRe.MatchString(form["telepon"]) {
		errs["telepon"] = "Tulis nomor telepon yang bisa dihubungi, misalnya 081234567890."
	}
	l, ok := cariLayanan(form["layanan"])
	if !ok {
		errs["layanan"] = "Pilih salah satu jenis layanan."
	}
	if len([]rune(form["keperluan"])) > 500 {
		errs["keperluan"] = "Keperluan maksimal 500 karakter."
	}
	if len(errs) > 0 {
		a.tampilForm(w, form, errs, http.StatusUnprocessableEntity)
		return
	}

	kode, err := a.store.Tambah(Pengajuan{
		Nama: form["nama"], NIK: form["nik"], Telepon: form["telepon"],
		LayananSlug: l.Slug, LayananNama: l.Nama, Keperluan: form["keperluan"],
	})
	if err != nil {
		log.Printf("simpan pengajuan: %v", err)
		errs["umum"] = "Pengajuan belum tersimpan. Coba lagi beberapa saat."
		a.tampilForm(w, form, errs, http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/status?baru=1&kode="+url.QueryEscape(kode), http.StatusSeeOther)
}

func (a *App) status(w http.ResponseWriter, r *http.Request) {
	kode := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("kode")))
	data := map[string]any{"Kode": kode, "Baru": r.URL.Query().Get("baru") == "1"}
	code := http.StatusOK
	if kode != "" {
		if p, ok := a.store.Cari(kode); ok {
			data["Hasil"] = p
			data["Langkah"] = langkah(p.Status)
			data["Ditolak"] = p.Status == "Ditolak"
		} else {
			code = http.StatusNotFound
		}
	}
	a.render(w, "status", "Status pengajuan", code, data)
}

func langkah(status string) []Langkah {
	idx := -1
	for i, s := range alurLayan {
		if s == status {
			idx = i
		}
	}
	out := make([]Langkah, len(alurLayan))
	for i, s := range alurLayan {
		k := ""
		if i < idx {
			k = "selesai"
		} else if i == idx {
			k = "sekarang"
		}
		out[i] = Langkah{s, k}
	}
	return out
}

// ---- Admin ----

func (a *App) admin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if a.adminPass == "" {
			http.NotFound(w, r)
			return
		}
		u, p, ok := r.BasicAuth()
		okU := subtle.ConstantTimeCompare([]byte(u), []byte(a.adminUser)) == 1
		okP := subtle.ConstantTimeCompare([]byte(p), []byte(a.adminPass)) == 1
		if !ok || !okU || !okP {
			w.Header().Set("WWW-Authenticate", `Basic realm="Admin Desa", charset="UTF-8"`)
			http.Error(w, "Perlu login", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func (a *App) adminDaftar(w http.ResponseWriter, r *http.Request) {
	a.render(w, "admin", "Admin", http.StatusOK, map[string]any{
		"Items":    a.store.Semua(),
		"Statuses": append(append([]string{}, alurLayan...), "Ditolak"),
	})
}

func (a *App) adminUbah(w http.ResponseWriter, r *http.Request) {
	if o := r.Header.Get("Origin"); o != "" {
		if u, err := url.Parse(o); err != nil || u.Host != r.Host {
			http.Error(w, "Asal permintaan tidak valid", http.StatusForbidden)
			return
		}
	}
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Data tidak valid", http.StatusBadRequest)
		return
	}
	status := r.FormValue("status")
	valid := status == "Ditolak"
	for _, s := range alurLayan {
		if s == status {
			valid = true
		}
	}
	if !valid {
		http.Error(w, "Status tidak dikenal", http.StatusBadRequest)
		return
	}
	catatan := strings.TrimSpace(r.FormValue("catatan"))
	if len([]rune(catatan)) > 300 {
		http.Error(w, "Catatan maksimal 300 karakter", http.StatusBadRequest)
		return
	}
	if err := a.store.Ubah(r.FormValue("kode"), status, catatan); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}