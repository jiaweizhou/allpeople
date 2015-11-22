package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
)

var db *sql.DB

var ch = make(chan int)
var process = map[int]chan int{}

func main() {
	db, err := sql.Open("mysql", "root:@/alliance")
	if err != nil {
		log.Fatalf("Open database error: %s\n", err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	corns := NewCorns(db)
	commodities := NewCommodities(db)
	corns.Serve()
	commodities.Serve()
	http.HandleFunc("/cornopen", corns.Waitforopen)
	http.HandleFunc("/cornclose", corns.Waitforopen)
	http.HandleFunc("/commodityopen", commodities.Waitforopen)
	http.HandleFunc("/commodityclose", corns.Waitforopen)
	fmt.Println("start serve")
	http.ListenAndServe("0.0.0.0:8888", nil)

}
