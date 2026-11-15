<?php
session_start();

// Inisialisasi data awal permohonan surat jika belum ada
if (!isset($_SESSION['pengajuan_surat'])) {
    $_SESSION['pengajuan_surat'] = [
        [
            "tiket" => "DS-001",
            "nik" => "3301011234560001",
            "nama" => "Budi Santoso",
            "jenis_surat" => "Surat Keterangan Usaha (SKU)",
            "keperluan" => "Pengajuan KUR Bank",
            "status" => "Selesai"
        ],
        [
            "tiket" => "DS-002",
            "nik" => "3301016543210002",
            "nama" => "Siti Aminah",
            "jenis_surat" => "Surat Keterangan Domisili",
            "keperluan" => "Pendaftaran Sekolah Anak",
            "status" => "Diproses"
        ]
    ];
}

// Proses Form Pengajuan Baru
$pesan_sukses = "";
if (isset($_POST['kirim_pengajuan'])) {
    $nomor_urut = count($_SESSION['pengajuan_surat']) + 1;
    $tiket_baru = "DS-" . str_pad($nomor_urut, 3, "0", STR_PAD_LEFT);

    $data_baru = [
        "tiket" => $tiket_baru,
        "nik" => htmlspecialchars($_POST['nik']),
        "nama" => htmlspecialchars($_POST['nama']),
        "jenis_surat" => htmlspecialchars($_POST['jenis_surat']),
        "keperluan" => htmlspecialchars($_POST['keperluan']),
        "status" => "Diproses"
    ];

    array_unshift($_SESSION['pengajuan_surat'], $data_baru);
    $pesan_sukses = "Pengajuan berhasil dikirim! Nomor Tiket Anda adalah <strong>$tiket_baru</strong>.";
}

// Proses Pencarian Status
$keyword = isset($_GET['cari']) ? strtolower(trim($_GET['cari'])) : '';
$data_tampil = $_SESSION['pengajuan_surat'];

if ($keyword !== '') {
    $data_tampil = array_filter($_SESSION['pengajuan_surat'], function($item) use ($keyword) {
        return strpos(strtolower($item['nik']), $keyword) !== false ||
               strpos(strtolower($item['nama']), $keyword) !== false ||
               strpos(strtolower($item['tiket']), $keyword) !== false;
    });
}
?>

<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Pelayanan Digital Desa</title>
    <link rel="stylesheet" href="style.css">
</head>
<body>

    <header>
        <h1>Sistem Pelayanan Administrasi Desa</h1>
        <p>Melayani Masyarakat dengan Cepat, Transparan, dan Mudah</p>
    </header>

    <nav>
        <a href="index.php#layanan">Jenis Layanan</a>
        <a href="index.php#pengajuan">Ajukan Surat</a>
        <a href="index.php#cek-status">Cek Status Surat</a>
        <a href="admin.php" style="color: #ffd166;">Halaman Admin Desa</a>
    </nav>

    <div class="container">
        
        <?php if ($pesan_sukses): ?>
            <div class="alert"><?php echo $pesan_sukses; ?></div>
        <?php endif; ?>

        <!-- Jenis Layanan -->
        <h2 id="layanan" class="section-title">Jenis Layanan Administrasi</h2>
        <div class="service-grid">
            <div class="service-card">
                <h3>Surat Pengantar KTP/KK</h3>
                <p>Pengantar pembuatan atau pembaruan dokumen kependudukan ke tingkat kecamatan/dukcapil.</p>
            </div>
            <div class="service-card">
                <h3>Surat Keterangan Domisili</h3>
                <p>Keterangan tempat tinggal resmi untuk keperluan administrasi, pekerjaan, maupun sekolah.</p>
            </div>
            <div class="service-card">
                <h3>Surat Keterangan Usaha (SKU)</h3>
                <p>Legalitas usaha mikro/kecil warga desa untuk pengajuan permodalan atau perizinan.</p>
            </div>
            <div class="service-card">
                <h3>Surat Keterangan Tidak Mampu</h3>
                <p>Surat pengantar untuk keperluan beasiswa pendidikan, kesehatan, dan bantuan sosial.</p>
            </div>
        </div>

        <!-- Form Pengajuan -->
        <h2 id="pengajuan" class="section-title">Formulir Pengajuan Surat Online</h2>
        <div class="card-box">
            <form method="POST" action="index.php">
                <div class="form-group">
                    <label>Nomor Induk Kependudukan (NIK):</label>
                    <input type="text" name="nik" maxlength="16" placeholder="Masukkan 16 digit NIK" required>
                </div>
                <div class="form-group">
                    <label>Nama Lengkap (Sesuai KTP):</label>
                    <input type="text" name="nama" placeholder="Masukkan nama lengkap" required>
                </div>
                <div class="form-group">
                    <label>Jenis Surat yang Diajukan:</label>
                    <select name="jenis_surat" required>
                        <option value="">-- Pilih Jenis Layanan --</option>
                        <option value="Surat Pengantar KTP/KK">Surat Pengantar KTP/KK</option>
                        <option value="Surat Keterangan Domisili">Surat Keterangan Domisili</option>
                        <option value="Surat Keterangan Usaha (SKU)">Surat Keterangan Usaha (SKU)</option>
                        <option value="Surat Keterangan Tidak Mampu">Surat Keterangan Tidak Mampu</option>
                    </select>
                </div>
                <div class="form-group">
                    <label>Keterangan / Keperluan:</label>
                    <textarea name="keperluan" rows="3" placeholder="Jelaskan keperluan pembuatan surat..." required></textarea>
                </div>
                <button type="submit" name="kirim_pengajuan" class="btn">Kirim Pengajuan</button>
            </form>
        </div>

        <!-- Cek Status Pengajuan -->
        <h2 id="cek-status" class="section-title">Pantau Status Pengajuan Surat</h2>
        <div class="card-box">
            <form method="GET" action="index.php#cek-status" style="display: flex; gap: 10px; margin-bottom: 15px;">
                <input type="text" name="cari" value="<?php echo htmlspecialchars($keyword); ?>" placeholder="Cari berdasarkan No. Tiket, NIK, atau Nama..." style="flex: 1; padding: 10px; border: 1px solid #ccc; border-radius: 5px;">
                <button type="submit" class="btn">Cari</button>
                <?php if ($keyword !== ''): ?>
                    <a href="index.php#cek-status" class="btn btn-danger">Reset</a>
                <?php endif; ?>
            </form>

            <table>
                <thead>
                    <tr>
                        <th>No. Tiket</th>
                        <th>NIK</th>
                        <th>Nama Pemohon</th>
                        <th>Jenis Surat</th>
                        <th>Status</th>
                    </tr>
                </thead>
                <tbody>
                    <?php if (empty($data_tampil)): ?>
                        <tr><td colspan="5" style="text-align: center;">Data pengajuan tidak ditemukan.</td></tr>
                    <?php else: ?>
                        <?php foreach ($data_tampil as $row): ?>
                            <?php 
                                $badge_class = 'proses';
                                if ($row['status'] === 'Selesai') $badge_class = 'selesai';
                                elseif ($row['status'] === 'Ditolak') $badge_class = 'ditolak';
                            ?>
                            <tr>
                                <td><strong><?php echo $row['tiket']; ?></strong></td>
                                <td><?php echo $row['nik']; ?></td>
                                <td><?php echo $row['nama']; ?></td>
                                <td><?php echo $row['jenis_surat']; ?></td>
                                <td><span class="badge <?php echo $badge_class; ?>"><?php echo $row['status']; ?></span></td>
                            </tr>
                        <?php endforeach; ?>
                    <?php endif; ?>
                </tbody>
            </table>
        </div>

    </div>

    <footer>
        <p>&copy; 2026 Pemerintah Desa Digital. Jam Operasional Balai Desa: Senin - Jumat (08.00 - 15.00 WIB)</p>
    </footer>

</body>
</html>