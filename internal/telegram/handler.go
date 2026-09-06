package telegram

import (
	"context"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"telegram-finance-bot/internal/config"
	"telegram-finance-bot/internal/parser"
	"telegram-finance-bot/internal/sheets"
)

// Bot membungkus koneksi Telegram dan dependency lain yang dibutuhkan.
type Bot struct {
	api          *tgbotapi.BotAPI
	sheetsClient *sheets.Client
	cfg          *config.Config
}

// New membuat instance Bot baru.
func New(cfg *config.Config, sheetsClient *sheets.Client) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		return nil, fmt.Errorf("gagal konek ke Telegram API: %w", err)
	}
	log.Printf("terhubung sebagai @%s", api.Self.UserName)

	return &Bot{api: api, sheetsClient: sheetsClient, cfg: cfg}, nil
}

// Run menjalankan long polling dan memproses update yang masuk.
// Untuk production sebaiknya diganti ke mode webhook.
func (b *Bot) Run(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case update := <-updates:
			if update.Message == nil {
				continue
			}
			b.handleMessage(ctx, update.Message)
		}
	}
}

func (b *Bot) handleMessage(ctx context.Context, msg *tgbotapi.Message) {
	userID := msg.From.ID

	if !b.cfg.IsAllowed(userID) {
		b.reply(msg.Chat.ID, "Maaf, kamu tidak diizinkan memakai bot ini.")
		return
	}

	if msg.IsCommand() {
		b.handleCommand(ctx, msg)
		return
	}

	tx, err := parser.Parse(msg.Text)
	if err != nil {
		b.reply(msg.Chat.ID, err.Error())
		return
	}

	sheetName := b.cfg.DefaultSheetName
	if sheetName == "" {
		sheetName = tx.Date.Format("January 2006") // contoh: "September 2026"
	}

	if err := b.sheetsClient.EnsureSheetExists(ctx, sheetName); err != nil {
		log.Printf("error ensure sheet: %v", err)
		b.reply(msg.Chat.ID, "Gagal menyiapkan sheet tujuan, coba lagi nanti.")
		return
	}

	if err := b.sheetsClient.AppendTransaction(ctx, sheetName, tx); err != nil {
		log.Printf("error append transaction: %v", err)
		b.reply(msg.Chat.ID, "Gagal mencatat transaksi, coba lagi nanti.")
		return
	}

	confirmation := fmt.Sprintf(
		"Tercatat!\n%s | %s | Rp%.0f\nKategori: %s\nCatatan: %s",
		tx.Date.Format("02 Jan 2006"), tx.Type, tx.Amount, tx.Category, tx.Note,
	)
	b.reply(msg.Chat.ID, confirmation)
}

func (b *Bot) handleCommand(ctx context.Context, msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start":
		b.reply(msg.Chat.ID, "Halo! Kirim transaksi dengan format:\nkeluar 50000 makan siang #jajan\nmasuk 2000000 gaji")
	case "help":
		b.reply(msg.Chat.ID, "Format: <masuk/keluar> <jumlah> <catatan> #kategori\nContoh: keluar 15000 kopi #jajan")
	default:
		b.reply(msg.Chat.ID, "Perintah tidak dikenali. Coba /help")
	}
	_ = time.Now() // placeholder kalau nanti butuh timestamp command
}

func (b *Bot) reply(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("gagal kirim balasan: %v", err)
	}
}
