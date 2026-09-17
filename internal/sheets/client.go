package sheets

import (
	"context"
	"fmt"

	"google.golang.org/api/option"
	sheetsapi "google.golang.org/api/sheets/v4"

	"telegram-finance-bot/internal/model"
)

// sheetName adalah nama tab tetap tempat mencatat transaksi.
const sheetName = "Budget Tracking"

// firstDataRow adalah baris pertama tempat data transaksi dimulai
// (baris 1-8 dipakai untuk judul & ringkasan, baris 9 adalah header kolom).
const firstDataRow = 10

// Client membungkus sheets.Service dari Google API.
type Client struct {
	svc           *sheetsapi.Service
	spreadsheetID string
}

// NewClient membuat client baru menggunakan service account JSON.
func NewClient(ctx context.Context, credentialsPath, spreadsheetID string) (*Client, error) {
	svc, err := sheetsapi.NewService(ctx, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		return nil, fmt.Errorf("gagal membuat sheets client: %w", err)
	}
	return &Client{svc: svc, spreadsheetID: spreadsheetID}, nil
}

// nextEmptyRow mencari baris kosong pertama di kolom C, dimulai dari
// firstDataRow. Kita tidak pakai Values.Append karena sheet ini punya
// kolom "spacer" kosong (D dan F) di antara Date/Type/Category, yang
// membuat algoritma auto-detect tabel bawaan Sheets API salah menebak
// batas tabel (lihat penjelasan di README/percakapan).
func (c *Client) nextEmptyRow(ctx context.Context) (int, error) {
	readRange := fmt.Sprintf("%s!C%d:C3000", sheetName, firstDataRow)
	resp, err := c.svc.Spreadsheets.Values.Get(c.spreadsheetID, readRange).Context(ctx).Do()
	if err != nil {
		return 0, fmt.Errorf("gagal membaca kolom Date untuk cari baris kosong: %w", err)
	}
	// resp.Values hanya berisi baris yang tidak kosong secara berurutan
	// dari awal range, jadi baris kosong berikutnya = firstDataRow + jumlah baris terisi.
	return firstDataRow + len(resp.Values), nil
}

// AppendTransaction menulis satu baris transaksi ke sheet "Budget Tracking"
// dengan menargetkan kolom secara eksplisit (C, E, dan G:I terpisah),
// supaya tidak bergantung pada auto-detect tabel yang rawan salah karena
// ada kolom spacer kosong (D, F) di antara header.
//
// Kolom J (Balance) SENGAJA tidak disentuh karena template sudah mengisi
// formula Balance untuk setiap baris sampai baris 3000.
func (c *Client) AppendTransaction(ctx context.Context, tx model.Transaction) error {
	row, err := c.nextEmptyRow(ctx)
	if err != nil {
		return err
	}

	dateRange := fmt.Sprintf("%s!C%d", sheetName, row)
	typeRange := fmt.Sprintf("%s!E%d", sheetName, row)
	restRange := fmt.Sprintf("%s!G%d:I%d", sheetName, row, row) // Category, Description, Amount (berdempetan, aman)

	data := []*sheetsapi.ValueRange{
		{Range: dateRange, Values: [][]interface{}{{tx.Date.Format("2006-01-02")}}},
		{Range: typeRange, Values: [][]interface{}{{string(tx.Type)}}},
		{Range: restRange, Values: [][]interface{}{{tx.Category, tx.Description, tx.Amount}}},
	}

	batchReq := &sheetsapi.BatchUpdateValuesRequest{
		ValueInputOption: "USER_ENTERED",
		Data:             data,
	}

	_, err = c.svc.Spreadsheets.Values.BatchUpdate(c.spreadsheetID, batchReq).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("gagal menulis ke sheet %q baris %d: %w", sheetName, row, err)
	}
	return nil
}
