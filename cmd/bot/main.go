package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"telegram-finance-bot/internal/config"
	"telegram-finance-bot/internal/sheets"
	"telegram-finance-bot/internal/telegram"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Load file .env kalau ada (untuk development lokal).
	// Di production biasanya env var di-set langsung oleh sistem/container,
	// jadi tidak masalah kalau file .env tidak ditemukan.
	if err := godotenv.Load(); err != nil {
		log.Println("info: file .env tidak ditemukan, lanjut pakai environment variable sistem")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	sheetsClient, err := sheets.NewClient(ctx, cfg.GoogleCredentialsPath, cfg.SpreadsheetID)
	if err != nil {
		log.Fatalf("sheets client error: %v", err)
	}

	bot, err := telegram.New(cfg, sheetsClient)
	if err != nil {
		log.Fatalf("telegram bot error: %v", err)
	}

	log.Println("bot berjalan, tekan Ctrl+C untuk berhenti")
	if err := bot.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("bot berhenti dengan error: %v", err)
	}
}
