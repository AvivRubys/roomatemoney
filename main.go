package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/codegangsta/negroni"
	"github.com/coopernurse/gorp"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"

	_ "github.com/mattn/go-sqlite3"
)

var dbMapper *gorp.DbMap
var renderer *render.Render
var debug *bool

func main() {
	// Setting up flags
	port := flag.Int("port", 80, "Port on which to bind the server")
	host := flag.String("host", "", "Host on which to listen")
	dbPath := flag.String("db", "app.db", "Path to the db file")
	debug = flag.Bool("debug", false, "Run in debug mode")
	flag.Parse()

	// Setting up the DB and mappings
	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	dbMapper = &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}
	dbMapper.AddTableWithName(Roomate{}, "roomates").SetKeys(true, "ID")
	dbMapper.AddTableWithName(Payment{}, "payments").SetKeys(true, "ID")
	dbMapper.AddTableWithName(Expense{}, "expenses").SetKeys(true, "ID")
	dbMapper.AddTableWithName(RoomateExpense{}, "roomate_expenses").SetKeys(true, "ID")

	if *debug {
		log.Printf("Using sqlite db at %s\n", *dbPath)
		dbMapper.CreateTablesIfNotExists()
		dbMapper.TraceOn("[gorp]", log.New(os.Stdout, "roomatemoney:", log.Lmicroseconds))
	}

	// Setting up the renderer
	renderer = render.New(render.Options{})

	// Setting up the http server
	listenString := fmt.Sprintf("%s:%d", *host, *port)
	router := mux.NewRouter()
	api := router.PathPrefix("/api").Subrouter()

	api.Path("/expense").Methods("GET").HandlerFunc(getExpenses)
	api.Path("/expense/by_roomate").Methods("GET").HandlerFunc(expensesByRoomate)
	api.Path("/roomate").Methods("GET").HandlerFunc(getRoomates)
	api.Path("/payment").Methods("GET").HandlerFunc(getPayments)

	n := negroni.Classic()
	n.UseHandler(router)
	n.Run(listenString)
}

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
