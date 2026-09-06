# Telegram Finance Bot (Golang + Google Sheets)

## Struktur project

```
cmd/bot/main.go          -> entrypoint aplikasi
internal/config          -> load konfigurasi dari environment variable
internal/model           -> struct Transaction
internal/parser          -> parsing teks pesan menjadi Transaction
internal/sheets          -> wrapper Google Sheets API (append, ensure sheet)
internal/telegram        -> handler update Telegram, routing command & pesan
```

## Setup

### 1. Buat bot Telegram

- Chat ke [@BotFather](https://t.me/BotFather), buat bot baru, catat token-nya.

### 2. Buat Service Account Google Cloud

1. Buka [Google Cloud Console](https://console.cloud.google.com/) → buat project (kalau belum ada).
2. Aktifkan **Google Sheets API** untuk project tersebut.
3. Buat **Service Account** → generate key JSON → simpan sebagai `service-account.json`.
4. Buka spreadsheet template keuanganmu → klik **Share** → tambahkan email service account (contoh: `bot@nama-project.iam.gserviceaccount.com`) dengan akses **Editor**.

### 3. Siapkan environment variable

```bash
cp .env.example .env
# isi TELEGRAM_BOT_TOKEN, GOOGLE_CREDENTIALS_PATH, SPREADSHEET_ID, ALLOWED_USER_IDS
```

### 4. Install dependency & jalankan

```bash
go mod tidy
go run ./cmd/bot
```

## Format pesan yang didukung

```
keluar 50000 makan siang #jajan
masuk 2000000 gaji bulanan #gaji
keluar 15000 kopi
```

Kolom yang ditulis ke sheet: `Tanggal | Jenis | Kategori | Jumlah | Catatan`
