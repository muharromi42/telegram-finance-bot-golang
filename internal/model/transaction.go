package model

import "time"

// TransactionType membedakan pemasukan dan pengeluaran.
type TransactionType string

const (
	TypeIncome  TransactionType = "masuk"
	TypeExpense TransactionType = "keluar"
)

// Transaction merepresentasikan satu baris catatan keuangan.
// Sesuaikan field ini dengan kolom template spreadsheet kamu.
type Transaction struct {
	Date     time.Time
	Type     TransactionType
	Amount   float64
	Category string
	Note     string
}

// ToRow mengubah Transaction menjadi slice of interface{} sesuai urutan
// kolom di spreadsheet: Tanggal | Jenis | Kategori | Jumlah | Catatan
func (t Transaction) ToRow() []interface{} {
	return []interface{}{
		t.Date.Format("2006-01-02"),
		string(t.Type),
		t.Category,
		t.Amount,
		t.Note,
	}
}
