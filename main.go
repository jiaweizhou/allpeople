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

func test(response http.ResponseWriter, request *http.Request) {
	//request.ParseForm()
	fmt.Println(request.RemoteAddr)
	//grabcornid := request.Form.Get("grabcornid")
	//fmt.Println("get grabcornid:" + grabcornid)
	response.Write([]byte(`{'flag':` + `}`))
}
func main() {
	db, err := sql.Open("mysql", "root:123456@tcp(10.10.105.196:3306)/alliance")
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
	//ClearEnvelope(db)
	log.Println("adfgsdfg")
	corns.Serve()
	commodities.Serve()
	http.HandleFunc("/test", test)
	http.HandleFunc("/cornopen", corns.Waitforopen)
	http.HandleFunc("/cornclose", corns.Waitforopen)
	http.HandleFunc("/commodityopen", commodities.Waitforopen)
	http.HandleFunc("/commodityclose", corns.Waitforopen)
	fmt.Println("start serve")
	http.ListenAndServe("0.0.0.0:8888", nil)

}
