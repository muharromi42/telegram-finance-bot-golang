package sheets

import (
	"context"
	"fmt"

	"google.golang.org/api/option"
	sheetsapi "google.golang.org/api/sheets/v4"

	"telegram-finance-bot/internal/model"
)

// Client membungkus sheets.Service dari Google API.
type Client struct {
	svc           *sheetsapi.Service
	spreadsheetID string
}

// NewClient membuat client baru menggunakan service account JSON.
// credentialsPath: path ke file JSON service account.
// Spreadsheet tujuan harus sudah di-share (akses Editor) ke email
// service account tersebut.
func NewClient(ctx context.Context, credentialsPath, spreadsheetID string) (*Client, error) {
	svc, err := sheetsapi.NewService(ctx, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		return nil, fmt.Errorf("gagal membuat sheets client: %w", err)
	}
	return &Client{svc: svc, spreadsheetID: spreadsheetID}, nil
}

// AppendTransaction menambahkan satu baris transaksi ke sheet/tab tertentu.
// sheetName harus sudah ada sebagai tab di spreadsheet, contoh "September 2026".
func (c *Client) AppendTransaction(ctx context.Context, sheetName string, tx model.Transaction) error {
	valueRange := &sheetsapi.ValueRange{
		Values: [][]interface{}{tx.ToRow()},
	}

	// range "SheetName!A:E" berarti "tambahkan di baris kosong pertama
	// pada kolom A sampai E", Sheets API otomatis mencari baris terakhir.
	writeRange := fmt.Sprintf("%s!A:E", sheetName)

	_, err := c.svc.Spreadsheets.Values.Append(c.spreadsheetID, writeRange, valueRange).
		ValueInputOption("USER_ENTERED").
		InsertDataOption("INSERT_ROWS").
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("gagal menulis ke sheet %q: %w", sheetName, err)
	}
	return nil
}

// EnsureSheetExists mengecek apakah tab dengan nama tertentu sudah ada.
// Kalau belum, sheet baru dibuat dengan header default.
func (c *Client) EnsureSheetExists(ctx context.Context, sheetName string) error {
	spreadsheet, err := c.svc.Spreadsheets.Get(c.spreadsheetID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("gagal membaca metadata spreadsheet: %w", err)
	}

	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == sheetName {
			return nil // sudah ada
		}
	}

	// Sheet belum ada, buat baru
	addSheetReq := &sheetsapi.Request{
		AddSheet: &sheetsapi.AddSheetRequest{
			Properties: &sheetsapi.SheetProperties{Title: sheetName},
		},
	}
	batchReq := &sheetsapi.BatchUpdateSpreadsheetRequest{
		Requests: []*sheetsapi.Request{addSheetReq},
	}
	if _, err := c.svc.Spreadsheets.BatchUpdate(c.spreadsheetID, batchReq).Context(ctx).Do(); err != nil {
		return fmt.Errorf("gagal membuat sheet baru %q: %w", sheetName, err)
	}

	// Tulis header kolom di baris pertama
	header := &sheetsapi.ValueRange{
		Values: [][]interface{}{{"Tanggal", "Jenis", "Kategori", "Jumlah", "Catatan"}},
	}
	_, err = c.svc.Spreadsheets.Values.Update(c.spreadsheetID, sheetName+"!A1", header).
		ValueInputOption("USER_ENTERED").
		Context(ctx).
		Do()
	return err
}
