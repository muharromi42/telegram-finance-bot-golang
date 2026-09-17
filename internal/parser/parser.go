package parser

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"telegram-finance-bot/internal/model"
)

// categoryAliases memetakan kata kunci singkat (yang diketik user) ke nama
// kategori PERSIS seperti yang ada di dropdown SETUP pada spreadsheet.
// Ini penting supaya data yang ditulis bot valid terhadap data validation
// yang sudah ada di sheet "Budget Tracking".
var categoryAliases = map[model.TransactionType]map[string]string{
	model.TypeIncome: {
		"gaji":       "Employment",
		"employment": "Employment",
		"sampingan":  "Side Hustles",
		"sidehustle": "Side Hustles",
		"bisnis":     "Business Income ", // sesuai SETUP (ada trailing space)
		"business":   "Business Income ",
		"investasi":  "Investment Income ", // ada trailing space di SETUP
		"investment": "Investment Income ",
		"sekali":     "One-time Income",
		"onetime":    "One-time Income",
		"lain":       "Other Income",
		"other":      "Other Income",
	},
	model.TypeExpenses: {
		"rumah":         "Housing",
		"housing":       "Housing",
		"makan":         "Food & Groceries",
		"food":          "Food & Groceries",
		"utilitas":      "Utilities",
		"utilities":     "Utilities",
		"perawatan":     "Personal Care",
		"personal":      "Personal Care",
		"asuransi":      "Insurances",
		"insurance":     "Insurances",
		"transport":     "Transportation",
		"transportasi":  "Transportation",
		"belanja":       "Shopping ", // sesuai SETUP (ada trailing space)
		"shopping":      "Shopping ",
		"kesehatan":     "Health/Medical",
		"health":        "Health/Medical",
		"cicilan":       "Debt Payments ", // ada trailing space di SETUP
		"debt":          "Debt Payments ",
		"hiburan":       "Entertainment",
		"entertainment": "Entertainment",
		"liburan":       "Vacation/Travelling",
		"vacation":      "Vacation/Travelling",
		"hadiah":        "Gifts/Donations",
		"donasi":        "Gifts/Donations",
		"gift":          "Gifts/Donations",
		"lain":          "Other Expenses",
		"other":         "Other Expenses",
	},
	model.TypeSavings: {
		"umum":       "General Savings",
		"general":    "General Savings",
		"investasi":  "Investments",
		"investment": "Investments",
		"darurat":    "Emergency Fund",
		"emergency":  "Emergency Fund",
		"cadangan":   "Sinking Fund",
		"sinking":    "Sinking Fund",
		"bisnis":     "Business Investment",
		"business":   "Business Investment",
	},
}

// typeAliases memetakan kata yang diketik user ke TransactionType resmi.
var typeAliases = map[string]model.TransactionType{
	"income":   model.TypeIncome,
	"masuk":    model.TypeIncome,
	"pemasukan": model.TypeIncome,
	"gaji":     model.TypeIncome,

	"expenses":    model.TypeExpenses,
	"expense":     model.TypeExpenses,
	"keluar":      model.TypeExpenses,
	"pengeluaran": model.TypeExpenses,

	"savings": model.TypeSavings,
	"saving":  model.TypeSavings,
	"nabung":  model.TypeSavings,
	"tabungan": model.TypeSavings,
}

// Format pesan yang didukung:
//   <type> <kategori> <jumlah> <catatan...>
// Contoh:
//   expenses food 50000 makan siang
//   income gaji 5000000 gaji bulan ini
//   savings darurat 200000 nabung darurat
func Parse(text string) (model.Transaction, error) {
	fields := strings.Fields(strings.TrimSpace(text))
	if len(fields) < 3 {
		return model.Transaction{}, fmt.Errorf(
			"format tidak lengkap. Contoh: expenses food 50000 makan siang")
	}

	txType, ok := typeAliases[strings.ToLower(fields[0])]
	if !ok {
		return model.Transaction{}, fmt.Errorf(
			"jenis %q tidak dikenali. Gunakan: income, expenses, atau savings", fields[0])
	}

	aliasMap := categoryAliases[txType]
	category, ok := aliasMap[strings.ToLower(fields[1])]
	if !ok {
		return model.Transaction{}, fmt.Errorf(
			"kategori %q tidak dikenali untuk jenis %s. Coba /help untuk lihat daftar kategori",
			fields[1], txType)
	}

	amountStr := strings.NewReplacer(".", "", ",", "").Replace(fields[2])
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return model.Transaction{}, fmt.Errorf("jumlah tidak valid: %q", fields[2])
	}

	description := ""
	if len(fields) > 3 {
		description = strings.Join(fields[3:], " ")
	}

	return model.Transaction{
		Date:        time.Now(),
		Type:        txType,
		Category:    category,
		Description: description,
		Amount:      amount,
	}, nil
}
