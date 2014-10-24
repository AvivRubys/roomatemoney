package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

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
	api.Path("/expense").Methods("POST", "PUT").HandlerFunc(newExpense)
	api.Path("/expense/by_roomate").Methods("GET").HandlerFunc(expensesByRoomate)
	api.Path("/roomate").Methods("GET").HandlerFunc(getRoomates)
	api.Path("/payment").Methods("GET").HandlerFunc(getPayments)
	api.Path("/payment").Methods("POST", "PUT").HandlerFunc(newPayment)

	n := negroni.Classic()
	n.UseHandler(router)
	n.Run(listenString)
}
