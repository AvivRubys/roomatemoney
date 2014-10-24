package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

func getExpenses(w http.ResponseWriter, r *http.Request) {
	var res []Expense
	_, err := dbMapper.Select(&res, "SELECT * FROM expenses")
	if err != nil {
		// TODO: Return something instead of panic
		log.Panic(err)
	}

	// Iterating using index cause range with two args copies the element
	// can be converted to IN (?,?,?,?) at some point, but I cba
	for i := range res {
		expense := &res[i]
		_, err = dbMapper.Select(&expense.Details,
			"SELECT * FROM roomate_expenses WHERE ExpenseID = ?", expense.ID)
		if err != nil {
			// TODO: Return something instead of panic
			log.Panic(err)
		}
	}

	renderer.JSON(w, http.StatusOK, res)
}

func getRoomates(w http.ResponseWriter, r *http.Request) {
	var res []Roomate
	_, err := dbMapper.Select(&res, "SELECT * FROM roomates")
	if err != nil {
		// TODO: Return something instead of panic
		log.Panic(err)
	}

	renderer.JSON(w, http.StatusOK, res)
}

func getPayments(w http.ResponseWriter, r *http.Request) {
	var res []Payment
	_, err := dbMapper.Select(&res, "SELECT * FROM payments")
	if err != nil {
		// TODO: Return something instead of panic
		log.Panic(err)
	}

	renderer.JSON(w, http.StatusOK, res)
}

func expensesByRoomate(w http.ResponseWriter, r *http.Request) {
	var expensesByID []struct {
		ID     int
		Amount int
	}

	// Query explanation: Take the total expenses sum, divide by number of roomates
	// deduct the sum of all the actual money the current roomate has spent
	// deduct the sum of the payments the roomate received
	// and add the sum of all the payments he has sent
	_, err := dbMapper.Select(&expensesByID,
		`SELECT r.id ID,
      SUM(e.amount)/(SELECT COUNT(1) from roomates) -
      SUM(re.amount) -
      IFNULL((SELECT SUM(p.amount) FROM payments p WHERE p.FromRoomateID = r.ID), 0) +
      IFNULL((SELECT SUM(p.amount) FROM payments p WHERE p.ToRoomateID = r.ID), 0) AMOUNT
     FROM expenses e,
     roomates r LEFT JOIN roomate_expenses re ON r.id = re.RoomateID
     GROUP BY r.id`)
	if err != nil {
		// TODO: Return something instead of panic
		log.Panic(err)
	}
	expensesByIDMap := make(map[string]int)
	for _, elem := range expensesByID {
		expensesByIDMap[strconv.Itoa(elem.ID)] = elem.Amount
	}

	renderer.JSON(w, http.StatusOK, expensesByIDMap)
}

func newExpense(w http.ResponseWriter, r *http.Request) {
	expense := Expense{}
	err := json.NewDecoder(r.Body).Decode(expense)
	if err != nil {
		// TODO: Return something instead of panic
		log.Panic(err)
	}

	trans, err := dbMapper.Begin()
	if err != nil {
		log.Panic(err)
	}
	err = trans.Insert(&expense)
	if err != nil {
		log.Panic(err)
	}

	for i := range expense.Details {
		err = trans.Insert(&expense.Details[i])
		if err != nil {
			log.Panic(err)
		}
	}

	err = trans.Commit()
	if err != nil {
		log.Panic(err)
	}
}

func newPayment(w http.ResponseWriter, r *http.Request) {
	payment := Payment{}
	err := json.NewDecoder(r.Body).Decode(&payment)
	if err != nil {
		// TODO: Return something instead of panic
		log.Panic(err)
	}

	err = dbMapper.Insert(&payment)
	if err != nil {
		log.Panic(err)
	}
}
