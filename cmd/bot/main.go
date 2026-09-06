package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"telegram-finance-bot/internal/config"
	"telegram-finance-bot/internal/sheets"
	"telegram-finance-bot/internal/telegram"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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
