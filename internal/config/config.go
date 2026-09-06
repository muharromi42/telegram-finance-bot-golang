package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config menampung semua konfigurasi yang dibutuhkan bot.
type Config struct {
	// Token bot Telegram, didapat dari @BotFather
	TelegramBotToken string

	// Path ke file JSON service account Google (untuk akses Sheets API)
	GoogleCredentialsPath string

	// ID spreadsheet tujuan (bagian di URL sheet: .../d/<SPREADSHEET_ID>/edit)
	SpreadsheetID string

	// Nama sheet/tab default tempat menulis transaksi, misal "September 2026"
	// Kalau kosong, bot akan pakai nama bulan berjalan otomatis.
	DefaultSheetName string

	// Daftar user ID Telegram yang diizinkan memakai bot (whitelist)
	AllowedUserIDs map[int64]bool
}

// Load membaca konfigurasi dari environment variable.
// Wajib diisi: TELEGRAM_BOT_TOKEN, GOOGLE_CREDENTIALS_PATH, SPREADSHEET_ID, ALLOWED_USER_IDS
func Load() (*Config, error) {
	cfg := &Config{
		TelegramBotToken:      os.Getenv("TELEGRAM_BOT_TOKEN"),
		GoogleCredentialsPath: os.Getenv("GOOGLE_CREDENTIALS_PATH"),
		SpreadsheetID:         os.Getenv("SPREADSHEET_ID"),
		DefaultSheetName:      os.Getenv("DEFAULT_SHEET_NAME"),
		AllowedUserIDs:        map[int64]bool{},
	}

	if cfg.TelegramBotToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN belum diset")
	}
	if cfg.GoogleCredentialsPath == "" {
		return nil, fmt.Errorf("GOOGLE_CREDENTIALS_PATH belum diset")
	}
	if cfg.SpreadsheetID == "" {
		return nil, fmt.Errorf("SPREADSHEET_ID belum diset")
	}

	rawIDs := os.Getenv("ALLOWED_USER_IDS") // contoh: "123456789,987654321"
	if rawIDs == "" {
		return nil, fmt.Errorf("ALLOWED_USER_IDS belum diset, minimal isi user ID Telegram kamu sendiri")
	}
	for _, part := range strings.Split(rawIDs, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("ALLOWED_USER_IDS mengandung nilai tidak valid: %q", part)
		}
		cfg.AllowedUserIDs[id] = true
	}

	return cfg, nil
}

// IsAllowed mengecek apakah user ID boleh memakai bot.
func (c *Config) IsAllowed(userID int64) bool {
	return c.AllowedUserIDs[userID]
}
