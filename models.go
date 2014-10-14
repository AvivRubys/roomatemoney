package main

import "time"

// Roomate is used to hold roomate data
type Roomate struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Payment represents a payment between two roomates
type Payment struct {
	ID            int       `json:"id"`
	Amount        int       `json:"amount"`
	Description   string    `json:"description"`
	Date          time.Time `json:"date"`
	FromRoomateID int       `json:"from_roomate_id"`
	ToRoomateID   int       `json:"to_roomate_id"`
}

// Expense represents a shared expense
type Expense struct {
	ID          int              `json:"id"`
	Description string           `json:"description"`
	Amount      int              `json:"amount"`
	Date        time.Time        `json:"date"`
	Details     []RoomateExpense `json:"details" db:"-"`
}

// RoomateExpense represents a single roomate's part in an expense
type RoomateExpense struct {
	ID        int `json:"id"`
	ExpenseID int `json:"expense_id"`
	RoomateID int `json:"roomate_id"`
	Amount    int `json:"amount"`
}
