package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"telegram-finance-bot/internal/model"
)

// Format pesan yang didukung, contoh:
//   keluar 50000 makan siang
//   masuk 2000000 gaji bulanan
//   keluar 15000 kopi #jajan
//
// Kata pertama: jenis transaksi (masuk/keluar)
// Kata kedua: jumlah (boleh pakai titik/koma ribuan, contoh 50.000)
// Sisanya: catatan. Kategori opsional ditandai dengan awalan '#'.
var linePattern = regexp.MustCompile(`(?i)^(masuk|keluar)\s+([\d.,]+)\s*(.*)$`)

// ErrFormatTidakDikenali dikembalikan kalau pesan tidak cocok pola apapun.
var ErrFormatTidakDikenali = fmt.Errorf("format pesan tidak dikenali, gunakan contoh: keluar 50000 makan siang #jajan")

// Parse mengubah teks pesan menjadi model.Transaction.
func Parse(text string) (model.Transaction, error) {
	text = strings.TrimSpace(text)
	matches := linePattern.FindStringSubmatch(text)
	if matches == nil {
		return model.Transaction{}, ErrFormatTidakDikenali
	}

	txType := model.TransactionType(strings.ToLower(matches[1]))

	amountStr := strings.NewReplacer(".", "", ",", "").Replace(matches[2])
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return model.Transaction{}, fmt.Errorf("jumlah tidak valid: %q", matches[2])
	}

	rest := strings.TrimSpace(matches[3])
	category, note := extractCategory(rest)

	return model.Transaction{
		Date:     time.Now(),
		Type:     txType,
		Amount:   amount,
		Category: category,
		Note:     note,
	}, nil
}

// extractCategory mencari token '#kategori' di dalam teks, memisahkannya
// dari catatan biasa. Kalau tidak ada tag, kategori default "lainnya".
func extractCategory(rest string) (category, note string) {
	words := strings.Fields(rest)
	var noteWords []string
	category = "lainnya"

	for _, w := range words {
		if strings.HasPrefix(w, "#") && len(w) > 1 {
			category = strings.ToLower(strings.TrimPrefix(w, "#"))
			continue
		}
		noteWords = append(noteWords, w)
	}

	return category, strings.Join(noteWords, " ")
}
