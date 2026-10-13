# Rencana Implementasi QRIS BTN

Tanggal: 9 Oktober 2026

Status: Rencana berdasarkan pemeriksaan source dan dokumen. Koneksi BTN dan perangkat belum diuji dalam penyusunan rencana ini.

## 1. Keputusan Cakupan

| Bagian | Keputusan |
|---|---|
| Backend | Integrasi pada `D:\applications\gbsPOS\gbs-pos-api` |
| Android | Fokus module `:app` pada `D:\applications\pos` |
| Perangkat POS | Sunmi D3 Pro dengan customer display |
| QRIS MPM | Periksa dan benahi implementasi yang sudah ada, lalu lengkapi Notify, Refund, dan Report |
| QRIS CPM | Implementasikan pembayaran nyata untuk fitur scan QR pelanggan |
| QRIS TAP | Ditunda; dibahas kembali setelah perangkat EDC/reader NFC dan kontrak integrasi BTN tersedia |
| EDC/ECR | Tahap terakhir, setelah integrasi SNAP MPM dan CPM selesai |
| Fuel POS | Di luar cakupan rencana ini |

Urutan pengerjaan:

```text
Konfirmasi kontrak BTN
  -> Fondasi transaksi dan kesiapan MPM
  -> Backend CPM
  -> Android CPM
  -> Notify, Refund, dan Report MPM/CPM
  -> UAT dan aktivasi SNAP
  -> EDC/ECR
```

Rencana ini tidak menetapkan implementasi QRIS TAP pada NFC bawaan Sunmi. Dokumentasi SUNMI menyebut akses NFC D3 Pro melalui API native Android, tetapi keputusan saat ini adalah menunda TAP.

## 2. Dokumen Acuan

Dokumen bank berikut berada di root repository backend. Nomor halaman mengacu pada halaman PDF.

| Dokumen | Bagian yang digunakan |
|---|---|
| `Dokumen Spesifikasi Katalog BTN SNAP QRIS v1.23 [Unpublished].pdf` | Token: hlm. 13-14; Generate: 15-16; Decode: 17-18; Query: 19-21; Notify: 22-23; Refund: 24-26; Report: 26-29 |
| `Dokumen Spesifikasi SNAP QRIS CPM v1.05.pdf` | Decode: hlm. 15-16; Payment: 17-19; Query: 20-22; Cancel: 22-25; Refund: 25-28 |
| `POS - EDC v25.1.1.1_BTN.pdf` | Tahap terakhir: serial interface, frame, checksum, QRIS inquiry, dan recovery |
| `Panduan ECR YOKKE (2025).pdf` | Aktivasi ECR pada EDC terpilih; bukan spesifikasi protokol pesan |
| `ISO 8583 ITM - EDC V5r2.7 - 13.08.2024 (1) (1).pdf` | Referensi komunikasi EDC ke host BTN; tidak menjadi implementasi POS langsung pada tahap SNAP |
| `Dokumen Spesifikasi SNAP QRIS TAP v1.00 (4).pdf` | Referensi tersimpan untuk pekerjaan TAP yang ditunda |

## 3. Baseline Implementasi

Temuan berikut berdasarkan source yang diperiksa pada 9 Oktober 2026. Pemeriksaan graph dilengkapi pembacaan source karena metadata index menunjukkan perubahan file. Baseline ini bukan bukti bahwa integrasi sudah lulus UAT BTN.

### 3.1 Backend

Path relatif terhadap `D:\applications\gbsPOS`.

| File | Kondisi yang ditemukan |
|---|---|
| `gbs-pos-api/internal/service/btn_snap_client.go` | Token B2B, signature RSA/HMAC, Generate MPM, Query MPM, timestamp WIB, dan external ID sudah tersedia |
| `gbs-pos-api/internal/dto/btn_snap_qris_dto.go` | DTO token, Generate/Query MPM, dan snapshot checkout tersedia |
| `gbs-pos-api/internal/service/qris_direct_service.go` | Jalur BTN memvalidasi checkout; Query memeriksa nominal, currency, merchant, terminal, referensi, dan waktu bayar; confirm lokal ditolak untuk provider BTN |
| `gbs-pos-api/internal/repository/qris_transaction_repo.go` | Referensi BTN dan pembaruan status bersyarat tersedia; status pending tidak menimpa transaksi yang sudah paid/refunded |
| `gbs-pos-api/internal/service/order_service.go` | Finalisasi `QRIS` memiliki guard BTN, tetapi `QRIS_SCAN` belum dicakup guard tersebut; void pesanan belum memanggil refund BTN |
| `gbs-pos-api/internal/model/model.go` | Tabel `pos_qris_transactions` sudah menyimpan referensi, provider, snapshot, nominal, status, dan waktu pembayaran |
| `gbs-pos-api/internal/config/config.go` | Konfigurasi kredensial SNAP serta satu merchant/TID bank tersedia |
| `gbs-pos-api/internal/router/qris_direct_route.go` | Kontrak convert/status/confirm/cancel internal Android tersedia |

Pada source integrasi yang diperiksa, belum ditemukan operasi CPM serta MPM Notify, Refund, dan Report. Decode MPM belum tersedia di client BTN yang diperiksa; untuk checkout merchant yang menampilkan QR, operasi ini tidak menjadi prasyarat pembayaran.

### 3.2 Android

Path relatif terhadap `D:\applications\pos\app\src\main\java\com\gis\gbs\posapplication\`.

| File | Kondisi yang ditemukan |
|---|---|
| `data/remote/dto/QrisDto.kt` | Request memakai default `provider = "QRIS_DIRECT"` |
| `data/remote/qris/QrisPaymentManager.kt` | Generate, polling, penyimpanan transaksi lokal, serta pemulihan melalui order ID tersedia; inisialisasi saat ini menghapus transaksi BTN yang belum paid |
| `ui/screen/payment/QrisScanDialog.kt` | Menangkap input scanner, tetapi menghasilkan sukses setelah delay; input manual menggunakan placeholder `MANUAL_INPUT` |
| `ui/screen/payment/PaymentDialog.kt` | Hasil dialog scan belum meneruskan payload QR ke proses pembayaran bank |
| `ui/viewmodel/cart/CartViewModel.kt` | `processQrisScanPayment()` masih demo dan menyimpan pesanan tanpa transaksi CPM BTN |
| `ui/viewmodel/checkout/CheckoutViewModel.kt` | Snapshot, generate, status, dan finalisasi MPM tersedia |
| `data/repository/OrderRepositoryImpl.kt` | `finalizeQrisOrder()` memfinalisasi ke backend sebelum menyimpan hasil lokal |

Temuan yang harus dibereskan sebelum aktivasi:

1. Android mengirim `QRIS_DIRECT` secara default, sedangkan backend memakai nilai ini untuk melewati jalur BTN. Keberadaan client BTN tidak berarti checkout Android sedang menggunakan BTN.
2. Transaksi BTN pending/unknown harus dipulihkan setelah restart, bukan dihapus saat manager dibuat.
3. Status sukses fitur scan pelanggan harus berasal dari pembayaran bank, bukan timer.
4. Semua jalur finalisasi, termasuk cart, checkout, dan bulk sync, harus memvalidasi pembayaran CPM yang terkait.
5. Pembatalan asinkron yang gagal tidak boleh diikuti penghapusan state pembayaran aktif. Periksa juga jalur cancel/reset pada `ui/screen/cart/CartScreen.kt`.

Kedua repository mempunyai perubahan lokal pada saat pemeriksaan. Saat implementasi dimulai, tinjau diff yang ada agar pekerjaan pengguna tetap terjaga.

## 4. Aturan Transaksi

- Backend menentukan merchant, merchant PAN, TID bank, dan provider yang diizinkan berdasarkan konfigurasi dan identitas terminal POS yang terautentikasi.
- Bedakan ID perangkat POS dengan TID bank; jangan menganggap keduanya sama.
- Simpan ID checkout dan snapshot tervalidasi sebelum request pembayaran. Pesanan final dan pengurangan stok dilakukan setelah pembayaran terverifikasi.
- Simpan percobaan pembayaran beserta referensi sebelum request bank untuk mendukung recovery saat respons hilang.
- Satu checkout hanya boleh memiliki satu percobaan pembayaran yang masih aktif/unknown. Gunakan transaksi database dan constraint untuk menangani request bersamaan.
- Nilai uang pada request bank berbentuk string desimal dua angka, misalnya `10000.00`. Gunakan perhitungan minor unit yang presisi di batas integrasi dan validasi rentang sesuai schema database.
- Gunakan token/signature yang sudah ada. Private key dan API secret tetap di server.
- `X-EXTERNAL-ID` adalah identitas request, berbeda dari referensi transaksi. Gunakan nilai alfanumerik 16 karakter dan pastikan keunikannya sesuai kontrak harian BTN.
- Timeout atau respons tidak dapat diverifikasi menghasilkan `UNKNOWN`, lalu Query memakai referensi original. Jangan otomatis mengirim pembayaran baru atau beralih ke QR lokal.
- Tombol cek pembayaran meminta verifikasi, bukan mengubah status menjadi paid secara manual.
- Simpan jenis alur pada transaksi. Pemrosesan transaksi lama memakai alur yang tersimpan, bukan konfigurasi provider terbaru.
- Batasi akses status, cancel, dan refund menurut kepemilikan merchant/terminal dan hak pengguna.
- Jangan mencatat payload QR pelanggan, token, signature, secret, atau customer PAN lengkap dalam log. Persistensi payload sensitif hanya jika dibutuhkan kontrak, dengan pembatasan akses dan retensi.
- Pembayaran sudah terverifikasi tetapi finalisasi pesanan gagal harus tetap dapat diselesaikan memakai pembayaran yang sama.

### Status pembayaran

| Kondisi | Status internal |
|---|---|
| Menunggu hasil, initiated, paying, pending | `PENDING`, dengan status bank original tetap disimpan |
| Bank mengonfirmasi berhasil dan data pembayaran cocok | `PAID` |
| Bank mengonfirmasi pembatalan | `CANCELLED` |
| Bank mengonfirmasi gagal | `FAILED` |
| MPM expired sesuai respons v1.23 dan waktu berlaku tersimpan | `EXPIRED` |
| Timeout, hasil ambigu, atau respons tidak cocok | `UNKNOWN` |
| Refund bank terkonfirmasi | `REFUNDED` untuk full refund |

`responseCode` sukses tidak selalu berarti dana sudah terbayar. Validasi status transaksi dan seluruh data yang disyaratkan kontrak. Penanganan `not found` harus mengikuti masing-masing alur; aturan expired MPM tidak boleh disalin langsung ke CPM.

## 5. Tahap 0: Konfirmasi Kontrak BTN

- [ ] Konfirmasi akses UAT dan scope MPM/CPM untuk kredensial partner.
- [ ] Konfirmasi base URL production, MID, TID, merchant PAN, nama merchant, lokasi, dan channel.
- [ ] Konfirmasi format payload CPM yang diberikan aplikasi pelanggan dan apakah Decode wajib sebelum Payment.
- [ ] Konfirmasi arti status sukses Payment CPM dan field bukti yang harus diverifikasi; contoh dokumen menggunakan nilai placeholder `transactionStatus`.
- [ ] Konfirmasi respons Query untuk pending, not found, dan hasil pembayaran yang timeout.
- [ ] Konfirmasi masa berlaku transaksi CPM serta perilaku Cancel ketika pembayaran sudah berhasil.
- [ ] Konfirmasi dukungan full/partial refund dan cara memeriksa hasil refund saat respons hilang.
- [ ] Konfirmasi autentikasi Notify MPM: penggunaan token/signature, kunci verifikasi bank, toleransi waktu, replay, dan retry.
- [ ] Konfirmasi bentuk `detailData` Report, pagination, serta rentang waktu yang diperbolehkan.

Ketidaksesuaian CPM yang perlu dijawab:

| Field | Masalah pada dokumen | Keputusan implementasi |
|---|---|---|
| `partnerReferenceNo` | Decode dibatasi 16, Payment 20, beberapa contoh lebih panjang | Referensi decode tersendiri maksimal 16; payment maksimal 20. Gunakan 16 karakter bila bank mengizinkan strategi seragam; hubungan antaroperasi mengikuti kontrak BTN |
| `referenceNo` | Tabel menyebut 16, contoh lebih panjang | Jangan memotong referensi respons; gunakan nilai original untuk Query/Cancel/Refund |
| `terminalId` | Tabel maksimal 8, contoh ada yang lebih panjang | Validasi request menurut kontrak TID CPM; jangan memotong TID MPM agar tampak cocok |
| `expiryTime` | Format ISO 8601, contoh timestamp tidak valid | Gunakan format WIB yang valid sesuai definisi; konfirmasi sumber waktu kedaluwarsanya |
| `reffId` dan amount | Contoh memakai `reffid` dan typo `transactioAmount` | Tetapkan nama field berdasarkan respons UAT; alias hanya untuk variasi yang benar-benar dikonfirmasi |
| `additionalInfo` | Parent ditandai optional tetapi beberapa child mandatory | Kirim child wajib beserta parent object yang diperlukan |

Permintaan bank menggunakan JSON valid, bukan menyalin kesalahan sintaks contoh PDF.

## 6. Tahap 1: Fondasi dan Kesiapan MPM

### Backend

- [x] Validasi pilihan provider; request arbitrer tidak boleh melewati jalur bank yang diwajibkan server.
- [x] Pertahankan endpoint MPM yang sudah digunakan Android, dengan pemilihan BTN yang benar.
- [x] Pisahkan respons penolakan definitif bank dari hasil transport yang unknown.
- [x] Tambahkan kontrol satu percobaan aktif per checkout, termasuk kasus dua request bersamaan.
- [x] Pastikan semua alias confirm/confirm-pending tetap menolak perubahan paid lokal untuk transaksi bank.
- [x] Pertahankan penanganan `4045101` sebagai belum ditemukan pembayaran, bukan paid.
- [x] Pertahankan respons expired `2005100` + status `07` sesuai v1.23; timer lokal saja tidak menentukan finalitas bank.
- [x] Pastikan referensi pembayaran, checkout, nominal, merchant, terminal, dan currency diverifikasi sebelum finalisasi.

### Android

- [x] Hilangkan default provider lokal pada checkout BTN; gunakan keputusan konfigurasi server.
- [x] Pulihkan transaksi pending/unknown saat startup dan lanjutkan Query dengan identitas yang sama.
- [x] Pertahankan snapshot checkout agar perubahan keranjang tidak mengubah pesanan yang sudah dibayar.
- [x] Simpan percobaan lokal sebelum request, dan periksa keberhasilan persistensinya untuk kebutuhan recovery.
- [x] Hapus transaksi aktif hanya setelah finalisasi pesanan berhasil atau hasil terminal bank sudah ditangani.
- [x] Tangani cancel/reset secara berurutan; jangan membuang state ketika pembatalan gagal.
- [x] Pastikan cart dan checkout mengikuti guard yang sama.

Kriteria selesai: checkout Android benar-benar menggunakan BTN, berhasil dibayar melalui UAT, dapat dipulihkan setelah restart, dan tidak membuat pembayaran/pesanan ganda.

## 7. Tahap 2: Backend CPM

Alur:

```text
Android kirim checkout + payload scan
  -> Backend validasi dan simpan percobaan
  -> Decode CPM, bila diwajibkan kontrak
  -> Payment CPM
  -> Verifikasi hasil atau Query CPM
  -> Simpan status bank
  -> Android finalisasi pesanan melalui backend
```

### Endpoint bank

| Operasi | Method dan path | Service code |
|---|---|---|
| Decode | `POST /snap/v1.0/qr/qr-cpm-decode` | `00` |
| Payment | `POST /snap/v1.0/qr/qr-cpm-payment` | `60` |
| Query | `POST /snap/v1.0/qr/qr-cpm-query` | `61` |
| Cancel | `POST /snap/v1.0/qr/qr-cpm-cancel` | `62` |
| Refund | `POST /snap/v1.0/qr/qr-cpm-refund` | `80` |

CPM v1.05 tidak memiliki Payment Notify; pemeriksaan lanjutan menggunakan Query.

### Pekerjaan

- [ ] Tambahkan DTO CPM terpisah dari DTO MPM.
- [ ] Tambahkan metode konkret Decode/Payment/Query/Cancel/Refund pada client BTN yang sudah ada; reuse token, signing, dan `net/http`.
- [ ] Tambahkan service CPM dengan aturan status dan validasi per operasi.
- [ ] Simpan `transactionDateTime` original, TID bank, merchant PAN, scanner/device ID, referensi, dan snapshot yang dibutuhkan Query/Cancel/Refund.
- [ ] Kirim `merchantLocation` uppercase sesuai CPM v1.05.
- [ ] Periksa panjang field, QR maksimum 512 sesuai kontrak, nominal, currency, timestamp, dan expiry sebelum request bank.
- [ ] Gunakan data merchant server, bukan nama/lokasi/PAN yang bebas dikirim Android.
- [ ] Perluas updater status berdasarkan jenis alur tanpa membuat updater CPM memanggil Query MPM.
- [ ] Perluas guard `OrderService.Create()` agar `QRIS_SCAN` memerlukan transaksi CPM paid yang cocok, termasuk jalur bulk sync.
- [ ] Tetapkan pengulangan finalisasi sebagai operasi idempotent dan cegah satu pembayaran digunakan untuk pesanan lain.
- [ ] Tambahkan migrasi SQL berikutnya dan selaraskan model AutoMigrate. Tentukan nomor migrasi saat implementasi.

Kontrak internal usulan, belum merupakan endpoint yang tersedia:

| Endpoint | Fungsi |
|---|---|
| `POST /v1/btn-qris/cpm/payments` | Membuat percobaan CPM dari order ID, snapshot, dan payload scan |
| `GET /v1/btn-qris/cpm/payments/:id` | Memeriksa transaksi melalui Query sesuai kebutuhan |
| `GET /v1/btn-qris/cpm/orders/:orderId` | Memulihkan percobaan ketika respons pembuatan hilang |
| `POST /v1/btn-qris/cpm/payments/:id/cancel` | Meminta cancel bank dan memverifikasi hasil |
| `POST /v1/btn-qris/cpm/payments/:id/refund` | Meminta refund dengan referensi operasi yang tersimpan |

Decode dapat diorkestrasi service dalam endpoint pembuatan pembayaran. Tidak perlu endpoint Android tersendiri jika UI tidak membutuhkan langkah decode terpisah.

Kriteria selesai: setiap operasi mengikuti payload UAT, hasil unknown dapat direkonsiliasi, dan backend menolak finalisasi sebelum bukti pembayaran cocok.

## 8. Tahap 3: Android CPM

- [ ] Tambahkan DTO internal dan metode API CPM.
- [ ] Ganti demo `processQrisScanPayment()` dengan pembuatan percobaan bank menggunakan order ID stabil dan snapshot.
- [ ] Teruskan payload sebenarnya dari `QrisScanDialog` melalui `PaymentDialog` ke proses pembayaran.
- [ ] Hapus sukses berbasis delay dan placeholder `MANUAL_INPUT`.
- [ ] Pertahankan huruf besar/kecil payload; jangan mengubah string QR pelanggan.
- [ ] Gunakan scanner input yang sudah tersedia. Untuk kamera, gunakan dependency ZXing yang sudah terpasang jika perangkat mendukung.
- [ ] Tampilkan state menunggu scan, memproses, pending, unknown, berhasil, gagal, dan dibatalkan.
- [ ] Tombol coba lagi pada hasil unknown memeriksa percobaan yang sama, bukan menginisiasi pembayaran baru.
- [ ] Simpan identitas percobaan dan snapshot untuk pemulihan setelah restart; hindari penyimpanan QR pelanggan setelah tidak dibutuhkan.
- [ ] Gunakan jalur finalisasi backend-first yang sudah ada. Kegagalan finalisasi tidak boleh meminta pelanggan membayar ulang.
- [ ] Sinkronkan nominal dan status pada customer display; cetak struk setelah finalisasi berhasil.
- [ ] Aktifkan pilihan CPM hanya jika konfigurasi server untuk terminal mendukungnya.

Kriteria selesai: scan pelanggan memanggil BTN, semua state dapat ditangani, restart tidak kehilangan percobaan, dan pesanan sukses hanya muncul setelah bank serta backend memvalidasi pembayaran.

## 9. Tahap 4: Notify, Refund, dan Report

### Notify MPM

- [ ] Tambahkan `POST /snap/v1/qr/qr-mpm-notify` pada URL publik yang didaftarkan ke BTN.
- [ ] Terapkan autentikasi bank sesuai kontrak; token login kasir bukan autentikasi callback BTN.
- [ ] Verifikasi signature terhadap body request yang sesuai aturan SNAP dan toleransi timestamp yang disepakati.
- [ ] Cocokkan referensi, merchant, terminal, nominal, currency, dan status dengan transaksi tersimpan.
- [ ] Tangani retry/replay serta balapan Notify dan Query secara idempotent.
- [ ] Jika data callback belum cukup untuk memverifikasi pembayaran, lakukan Query, bukan menetapkan paid dari data yang belum lengkap.
- [ ] Kirim acknowledgment sesuai SNAP, termasuk `2005200` setelah pemrosesan yang diperlukan berhasil.

### Refund MPM dan CPM

| Alur | Path bank | Service code |
|---|---|---|
| MPM | `/snap/v1.0/qr/qr-mpm-refund` | `78` |
| CPM | `/snap/v1.0/qr/qr-cpm-refund` | `80` |

- [ ] Simpan operasi refund dan `partnerRefundNo` sebelum request bank.
- [ ] Validasi otorisasi, transaksi asal, nominal, dan jumlah yang sudah direfund.
- [ ] Untuk MPM, gunakan `generateTime`, waktu pembayaran, dan TID yang tersimpan; untuk CPM, kirim scanner/device ID dan `refundDate` sesuai kontrak.
- [ ] Mulai dengan full refund yang disepakati; partial refund diaktifkan setelah aturan nominal dan status dikonfirmasi BTN.
- [ ] Respons refund hilang tetap unknown; cegah refund baru sampai hasil operasi original dapat diperiksa melalui mekanisme BTN yang dikonfirmasi.
- [ ] Pisahkan status pengembalian dana dari void/retur stok. Perubahan stok mengikuti tindakan pesanan yang disepakati dan dijalankan satu kali.

### Report

- [ ] Tambahkan `POST /snap/v1.0/qr/qr-mpm-report` dengan service code `99`.
- [ ] Gunakan `additionalInfo.transactionType = QRMPM` atau `QRCPM` sesuai kontrak.
- [ ] Terapkan validasi periode, pagination, serta filter merchant yang diotorisasi.
- [ ] Cocokkan laporan dengan percobaan internal untuk investigasi transaksi unknown/selisih.
- [ ] Sediakan akses laporan dan refund melalui histori/operasional `:app` yang sudah ada tanpa memperluas pekerjaan ke CMS baru.

Kriteria selesai: callback aman dan idempotent, refund memiliki bukti bank, dan laporan bisa dipakai untuk rekonsiliasi.

## 10. Tahap 5: UAT dan Aktivasi SNAP

| Area | Skenario yang harus dibuktikan |
|---|---|
| Signing dan token | Request ditandatangani dari body yang dikirim; expiry/token invalid ditangani tanpa pengulangan pembayaran yang tidak aman |
| MPM | Generate, belum dibayar `4045101`, sukses, expired v1.23, Notify berulang, dan Notify/Query bersamaan |
| CPM | Decode, payload invalid/expired, payment sukses/pending/ditolak, Query, Cancel, dan Refund |
| Nominal/referensi | Backend menolak mismatch nominal, currency, merchant, terminal, order, dan referensi |
| Konkurensi | Dua request untuk checkout yang sama tidak menghasilkan dua inisiasi bank |
| Jaringan | Respons pembayaran hilang, backend restart, aplikasi restart, dan Query setelah timeout |
| Finalisasi | Bank paid tetapi penyimpanan order gagal; retry tidak menagih ulang atau mengurangi stok dua kali |
| Akses | Pengguna/terminal lain tidak dapat membaca atau mengubah transaksi yang tidak diotorisasi |
| Android | Scanner/kamera yang dipakai, state gagal/unknown, customer display, dan cetak ulang struk |
| Refund | Permintaan berulang, respons hilang, dan pemisahan refund bank dari retur/void pesanan |

Gunakan test Go yang sudah ada, termasuk `gbs-pos-api/internal/service/btn_snap_client_test.go`, lalu tambahkan regresi untuk perilaku pembayaran yang berubah. Test otomatis dan fixture adalah pemeriksaan lokal, bukan bukti UAT bank.

Saat implementasi, jalankan test paket backend yang terdampak serta build/test module `:app` dengan Gradle wrapper. Uji PostgreSQL untuk constraint/konkurensi yang tidak dapat dibuktikan oleh SQLite in-memory. Perubahan Room/Hilt/KSP mengikuti pemeriksaan build yang diwajibkan repository Android.

Aktivasi production setelah bukti UAT tersedia, pemetaan merchant/terminal benar, callback terdaftar, dan prosedur recovery/refund sudah dapat digunakan.

## 11. Tahap Terakhir: EDC/ECR

Tahap ini dimulai setelah SNAP MPM/CPM selesai dan EDC sudah dipilih. ECR tidak menjadi prasyarat integrasi CPM melalui API BTN.

### Prasyarat perangkat

- [ ] Tentukan model EDC, firmware/aplikasi bank, kabel, dan metode koneksi.
- [ ] Konfirmasi protokol ECR pada perangkat tersebut dengan BTN/YOKKE.
- [ ] Uji USB host, permission, VID/PID, interface, endpoint baca/tulis, baud rate, parity, data bits, dan stop bits pada Sunmi D3 Pro.
- [ ] Konfirmasi timeout, ACK/NAK, retry, serta cara inquiry/recovery.

Spesifikasi POS-EDC BTN menyebut Ingenico Move2500, USB 9600 dan RS232 115200. Nilai ini harus dicocokkan dengan perangkat yang dipilih. Panduan aktivasi YOKKE tidak membuktikan bahwa semua EDC YOKKE memakai frame BTN tersebut.

### Implementasi

- [ ] Buat codec frame `STX + Version + Type + Subtype + Data + ETX + XOR checksum` sesuai firmware.
- [ ] Hitung checksum dari byte kedua sampai ETX, tidak termasuk STX. Uji contoh checksum `06` dalam dokumen.
- [ ] Tambahkan transport dua arah dan parser frame parsial pada Android.
- [ ] Implementasikan QRIS Payment (`3C/30`), Last Transaction (`3C/32`), Any Transaction (`3C/33`), Echo (`3B/30`), Resend (`3B/31`), dan Check Paper (`3B/32`).
- [ ] Tambahkan Sale (`31/30`), Void (`32/30`), dan Card Info (`34/30`) sesuai kebutuhan serta dukungan EDC; parser harus mengikuti format respons tiap command.
- [ ] Simpan percobaan sebelum mengirim command dan serialisasikan pembayaran pada satu koneksi EDC.
- [ ] ACK `06` berarti pesan diterima, bukan pembayaran sukses.
- [ ] Bedakan format nominal ECR dari ISO 8583: contoh ECR Rp5.000 memakai `000000005000`; DE4 ISO 8583 memakai dua digit desimal implisit.
- [ ] Saat kabel putus setelah inisiasi, lakukan inquiry/resend; jangan otomatis mengirim command pembayaran baru.
- [ ] Cocokkan hasil inquiry dengan percobaan aktif berdasarkan identitas dan referensi yang tersedia. Nominal sama atau last transaction saja belum cukup untuk membuktikan transaksi tersebut milik checkout saat ini.
- [ ] Kirim hasil EDC ke endpoint backend terautentikasi, ikat ke percobaan tersimpan, dan finalisasi secara idempotent. Konfirmasi mekanisme rekonsiliasi/verifikasi bank untuk transaksi ECR sebelum aktivasi; jangan menganggap referensi ECR dapat langsung dipakai Query SNAP.
- [ ] Reuse pola USB permission dari `data/printer/UsbPrinterManager.kt`, dengan koneksi khusus EDC yang mendukung baca/tulis dan tidak mengambil alih interface printer.

Kriteria selesai: transaksi perangkat nyata dapat diproses dan dipulihkan tanpa pembayaran ganda; printer tetap bekerja; referensi transaksi tersimpan untuk rekonsiliasi.

Komunikasi ISO 8583 ke host tetap ditangani EDC pada skema ini. Implementasi ISO 8583 langsung hanya dibahas jika BTN menetapkan skema host-to-host yang berbeda.

## 12. Ringkasan Deliverable

| Tahap | Hasil |
|---|---|
| 0 | Kontrak payload/status dan akses UAT MPM/CPM terkonfirmasi |
| 1 | Checkout MPM memakai BTN dan recovery transaksi bekerja |
| 2 | Backend CPM beserta validasi finalisasi pembayaran |
| 3 | Fitur scan pelanggan `:app` melakukan transaksi CPM nyata |
| 4 | Notify MPM, Refund MPM/CPM, dan Report untuk operasional |
| 5 | Bukti UAT SNAP dan konfigurasi aktivasi production |
| Terakhir | Integrasi EDC/ECR pada perangkat terpilih |

Langkah awal implementasi: konfirmasi kontrak pada Tahap 0, lalu benahi pemilihan provider dan pemulihan transaksi MPM pada Tahap 1. QRIS TAP tetap ditunda.
