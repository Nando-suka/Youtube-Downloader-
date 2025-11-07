# YouTube Downloader

Aplikasi web berbasis Go untuk mengunduh audio dari YouTube dan platform lainnya menggunakan yt-dlp.

## Fitur

- 🔍 Pencarian video YouTube menggunakan YouTube Data API
- ⬇️ Download audio dalam format MP3
- 🚦 Rate limiting untuk mencegah abuse
- 🔑 Multi API key support dengan rotasi otomatis
- 🎨 Interface web yang sederhana

## Persyaratan

- Go 1.21 atau lebih baru
- yt-dlp executable (download dari [yt-dlp GitHub](https://github.com/yt-dlp/yt-dlp))
- YouTube Data API v3 key (dapatkan dari [Google Cloud Console](https://console.cloud.google.com/apis/credentials))

## Instalasi

1. Clone repository ini:
```bash
git clone https://github.com/username/Youtube_downloader.git
cd Youtube_downloader
```

2. Install dependencies:
```bash
go mod download
```

3. Download yt-dlp:
   - **Windows**: Download `yt-dlp.exe` dan letakkan di folder project atau di PATH
   - **Linux/Mac**: Install melalui package manager atau download dari GitHub

4. Buat file `.env` berdasarkan `.env.example`:
```bash
cp .env.example .env
```

5. Edit file `.env` dan isi dengan API keys Anda:
```
YOUTUBE_API_KEY_MAIN=your_api_key_here
YTDLP_PATH=yt-dlp.exe
SERVER_PORT=8080
```

## Cara Menjalankan

```bash
go run cmd/server/main.go
```

Server akan berjalan di `http://localhost:8080` (atau port yang dikonfigurasi di `.env`).

## Konfigurasi

File `.env` mendukung konfigurasi berikut:

- `YOUTUBE_API_KEY_MAIN`: API key utama untuk YouTube Data API
- `YOUTUBE_API_KEY_BACKUP_1` sampai `YOUTUBE_API_KEY_BACKUP_5`: API key cadangan (opsional)
- `YTDLP_PATH`: Path ke executable yt-dlp
- `SERVER_PORT`: Port untuk server web (default: 8080)
- `RATE_LIMIT_REQUESTS`: Jumlah request maksimal per window (default: 100)
- `RATE_LIMIT_WINDOW`: Durasi window rate limiting dalam detik (default: 60)
- `TEMP_DIR`: Direktori untuk file temporary (default: ./tmp)

## Struktur Project

```
.
├── cmd/
│   └── server/
│       └── main.go          # Entry point aplikasi
├── config/
│   ├── config.go            # Konfigurasi aplikasi
│   └── keys.go              # API key manager
├── internal/
│   └── handlers/
│       ├── download.go      # Handler untuk download
│       ├── home.go          # Handler untuk halaman utama
│       └── search.go        # Handler untuk pencarian
├── middleware/
│   └── ratelimit.go         # Rate limiting middleware
├── staticDua/
│   ├── index.html           # Frontend HTML
│   └── style.css            # Styling
├── utils/
│   └── security.go          # Utility keamanan
└── tmp/                     # Direktori temporary (auto-generated)
```

## API Endpoints

- `GET /` - Halaman utama
- `GET /search?q=<query>` - Pencarian video YouTube
- `POST /download` - Download audio dari URL

## Keamanan

⚠️ **PENTING**: Jangan pernah commit file `.env` atau `env` yang berisi API keys ke repository. File-file ini sudah di-ignore di `.gitignore`.

## Platform yang Didukung

- YouTube (youtube.com, youtu.be)
- SoundCloud (soundcloud.com)

## Lisensi

MIT License - lihat file [LICENSE](LICENSE) untuk detail.

## Kontribusi

Kontribusi sangat diterima! Silakan buat issue atau pull request.

## Catatan

- Pastikan yt-dlp selalu up-to-date untuk kompatibilitas terbaik
- YouTube API memiliki quota harian, gunakan backup keys jika diperlukan
- File temporary akan otomatis dihapus setelah download selesai

