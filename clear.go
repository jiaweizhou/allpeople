package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

func ClearEnvelope(db *sql.DB) {
	bnow := time.Now()
	lingcheng := time.Date(bnow.Year(), bnow.Month(), bnow.Day()+1, 0, 0, 0, 0, bnow.Location())
	c := time.After(lingcheng.Sub(time.Now()))
	log.Println(lingcheng.Sub(time.Now()))
	day := time.Hour * 24
	go func() {
		for {
			select {
			case <-c:
			default:
				bnow = time.Now()
				_, err := db.Exec(fmt.Sprintf("update users set isdraw=0"))
				log.Println("update conmplete")
				if err != nil {
					c = time.After(time.Minute)
				} else {
					c = time.After(day - time.Now().Sub(bnow))
				}
			}
		}
	}()
}
