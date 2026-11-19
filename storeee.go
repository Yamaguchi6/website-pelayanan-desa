package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type Pengajuan struct {
	Kode        string    `json:"kode"`
	Nama        string    `json:"nama"`
	NIK         string    `json:"nik"`
	Telepon     string    `json:"telepon"`
	LayananSlug string    `json:"layanan_slug"`
	LayananNama string    `json:"layanan_nama"`
	Keperluan   string    `json:"keperluan"`
	Status      string    `json:"status"`
	Catatan     string    `json:"catatan"`
	Dibuat      time.Time `json:"dibuat"`
	Diperbarui  time.Time `json:"diperbarui"`
}

// Store menyimpan pengajuan di satu berkas JSON. Cukup untuk desa kecil;
// ganti dengan SQLite atau PostgreSQL bila data sudah banyak.
type Store struct {
	mu    sync.Mutex
	path  string
	items []Pengajuan
}

func NewStore(path string) (*Store, error) {
	s := &Store{path: path}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if len(b) > 0 {
		if err := json.Unmarshal(b, &s.items); err != nil {
			return nil, fmt.Errorf("berkas data rusak: %w", err)
		}
	}
	return s, nil
}

func (s *Store) simpan() error {
	b, err := json.MarshalIndent(s.items, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s *Store) kodeBaru() string {
	for {
		b := make([]byte, 3)
		_, _ = rand.Read(b)
		kode := "DS-" + time.Now().Format("20060102") + "-" + strings.ToUpper(hex.EncodeToString(b))
		ada := false
		for _, p := range s.items {
			if p.Kode == kode {
				ada = true
				break
			}
		}
		if !ada {
			return kode
		}
	}
}

// Tambah mengisi kode, status, dan waktu, lalu menyimpan. Mengembalikan kode pengajuan.
func (s *Store) Tambah(p Pengajuan) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	p.Kode = s.kodeBaru()
	p.Status = "Diterima"
	p.Dibuat, p.Diperbarui = now, now
	s.items = append(s.items, p)
	if err := s.simpan(); err != nil {
		s.items = s.items[:len(s.items)-1]
		return "", err
	}
	return p.Kode, nil
}

func (s *Store) Cari(kode string) (Pengajuan, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, p := range s.items {
		if p.Kode == kode {
			return p, true
		}
	}
	return Pengajuan{}, false
}

// Semua mengembalikan salinan, terbaru di atas.
func (s *Store) Semua() []Pengajuan {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Pengajuan, len(s.items))
	copy(out, s.items)
	sort.Slice(out, func(i, j int) bool { return out[i].Dibuat.After(out[j].Dibuat) })
	return out
}

func (s *Store) Ubah(kode, status, catatan string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.items {
		if s.items[i].Kode == kode {
			lama := s.items[i]
			s.items[i].Status = status
			s.items[i].Catatan = catatan
			s.items[i].Diperbarui = time.Now()
			if err := s.simpan(); err != nil {
				s.items[i] = lama
				return err
			}
			return nil
		}
	}
	return errors.New("pengajuan tidak ditemukan")
}
