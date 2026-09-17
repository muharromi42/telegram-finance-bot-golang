package telegram

import (
	"context"
	"fmt"
	"log"

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
		b.handleCommand(msg)
		return
	}

	tx, err := parser.Parse(msg.Text)
	if err != nil {
		b.reply(msg.Chat.ID, err.Error())
		return
	}

	if err := b.sheetsClient.AppendTransaction(ctx, tx); err != nil {
		log.Printf("error append transaction: %v", err)
		b.reply(msg.Chat.ID, "Gagal mencatat transaksi, coba lagi nanti.")
		return
	}

	confirmation := fmt.Sprintf(
		"Tercatat ke Budget Tracking!\n%s | %s | Rp%.0f\nKategori: %s\nCatatan: %s",
		tx.Date.Format("02 Jan 2006"), tx.Type, tx.Amount, tx.Category, tx.Description,
	)
	b.reply(msg.Chat.ID, confirmation)
}

func (b *Bot) handleCommand(msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start":
		b.reply(msg.Chat.ID, "Halo! Format pesan: <type> <kategori> <jumlah> <catatan>\nContoh: expenses food 50000 makan siang\nKetik /help untuk daftar lengkap kategori.")
	case "help":
		b.reply(msg.Chat.ID, helpText)
	default:
		b.reply(msg.Chat.ID, "Perintah tidak dikenali. Coba /help")
	}
}

func (b *Bot) reply(chatID int64, text string) {
	m := tgbotapi.NewMessage(chatID, text)
	if _, err := b.api.Send(m); err != nil {
		log.Printf("gagal kirim balasan: %v", err)
	}
}

const helpText = `Format: <type> <kategori> <jumlah> <catatan>

Type yang didukung: income, expenses, savings

Kategori Income: gaji, sampingan, bisnis, investasi, sekali, lain
Kategori Expenses: rumah, makan, utilitas, perawatan, asuransi, transport, belanja, kesehatan, cicilan, hiburan, liburan, hadiah, lain
Kategori Savings: umum, investasi, darurat, cadangan, bisnis

Contoh:
expenses makan 50000 makan siang
income gaji 5000000 gaji bulan ini
savings darurat 200000 nabung darurat`
