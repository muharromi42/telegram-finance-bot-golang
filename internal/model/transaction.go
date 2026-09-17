package model

import "time"

// TransactionType harus persis sama dengan value dropdown di kolom Type
// pada sheet "Budget Tracking" (data validation: Income,Expenses,Savings).
type TransactionType string

const (
	TypeIncome   TransactionType = "Income"
	TypeExpenses TransactionType = "Expenses"
	TypeSavings  TransactionType = "Savings"
)

// Transaction merepresentasikan satu baris di sheet "Budget Tracking".
type Transaction struct {
	Date        time.Time
	Type        TransactionType
	Category    string // harus persis sama dengan salah satu opsi dropdown SETUP
	Description string
	Amount      float64
}

