package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Grabcorns struct {
	Id             int `orm:"auto"`
	Picture        string
	Kind           int
	Title          string
	Version        int
	Needed         int
	Remain         int
	Created_at     int
	Date           int
	End_at         int
	Islotteried    int
	Winneruserid   int
	Winnerrecordid int
	Winnernumber   int
	Foruser        int
	Pictures       string
}

type Grabcornrecords struct {
	Id         int `orm:"auto"`
	Userid     int
	Grabcornid int
	Count      int
	Numbers    string
	Type       int
	Created_at int
}

type Corns struct {
	db      *sql.DB
	process map[int]chan int
}

func NewCorns(db *sql.DB) *Corns {
	return &Corns{
		db:      db,
		process: make(map[int]chan int),
	}
}
func (c *Corns) Waitforopen(response http.ResponseWriter, request *http.Request) {
	request.ParseForm()

	grabcornid := request.Form.Get("grabcornid")
	id, err := strconv.Atoi(grabcornid)
	fmt.Println(grabcornid)

	result, err := c.Getactivity(id)
	if err != nil {
		http.Error(response, "{'flag':0}", 500)
		return
	}
	go func() {
		c.process[result.Id] = make(chan int)
		c.Open(result, c.process[result.Id])
	}()
	response.Write([]byte("{'flag':1}"))
}
func (c *Corns) Serve() {
	result, err := c.Getactivities()
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range result {
		go func() {
			c.process[v.Id] = make(chan int)
			c.Open(v, c.process[v.Id])
		}()
	}
}
func (c *Corns) Open(grabcorn *Grabcorns, end chan int) {
	fmt.Println(time.Unix(int64(grabcorn.End_at), 0).Sub(time.Now()).String())
	ch := time.Tick(time.Unix(int64(grabcorn.End_at), 0).Sub(time.Now()))
	defer close(end)
	select {
	case <-ch:
		//default:
		records, numbers, err := c.Getrecords(grabcorn.Id)
		if err != nil {
			fmt.Println(err)
			return
		}
		times := 50
		if len(records) < 50 {
			times = len(records)
		}
		total := 0
		for i := 0; i < times; i++ {
			total += records[i].Created_at
		}
		number := total%grabcorn.Needed + 10000001
		winnerrecord := numbers[strconv.Itoa(number)]

		_, err = c.db.Exec(fmt.Sprintf("update grabcorns set islotteried=%d,winneruserid=%d,winnerrecordid=%d,winnernumber=%d where id = %d", 1, winnerrecord.Userid, winnerrecord.Id, number, grabcorn.Id))
		//_, err = c.db.Exec(fmt.Sprintf("insert into grabcorns(picture,pictures,title,version,needed,remain,created_at,date,end_at,islotteried,winneruserid,foruser,kind) select picture,pictures,title,version+1,needed,needed,%d,%d,0,0,0,0,kind from grabcorns where id = %d", time.Now().Unix(), time.Now().Unix(), grabcorn.Id))
		if err != nil {
			fmt.Println("kaijiangshibai" + err.Error())
		}
		form := url.Values{}
		form.Add("picture", "http://7xoc8r.com2.z0.glb.qiniucdn.com/corns/20071224162158623_2.jpg")
		form.Add("pictures", "http://7xoc8r.com2.z0.glb.qiniucdn.com/corns/20071224162158623_2.jpg http://7xoc8r.com2.z0.glb.qiniucdn.com/corns/2007822154648385_2.jpg http://7xoc8r.com2.z0.glb.qiniucdn.com/corns/2531170_193356481000_2.jpg http://7xoc8r.com2.z0.glb.qiniucdn.com/corns/5528723_101453638160_2.jpg http://7xoc8r.com2.z0.glb.qiniucdn.com/corns/6348-12011120200785.jpg http://7xoc8r.com2.z0.glb.qiniucdn.com/corns/8337244_105659585000_2.jpg")
		form.Add("title", grabcorn.Title)
		form.Add("version", strconv.Itoa(grabcorn.Version+1))
		form.Add("needed", strconv.Itoa(grabcorn.Needed))
		form.Add("date", fmt.Sprint(time.Now().Unix()))
		form.Add("kind", strconv.Itoa(grabcorn.Kind))
		response, err := http.PostForm("http://localhost/alliance/web/v1/grabcorns/create", form)
		if err != nil {
			fmt.Println("kaijiangshibai" + err.Error())
		} else {
			defer response.Body.Close()
			tt, _ := ioutil.ReadAll(response.Body)
			fmt.Println(string(tt))
		}

	case <-end:
		return
	}
}
func (c *Corns) Getactivities() ([]*Grabcorns, error) {
	result := []*Grabcorns{}
	rows, err := c.db.Query(fmt.Sprintf("select id,picture,pictures,title,version,date,needed,end_at,kind from grabcorns where islotteried = 0 and end_at!=0 and foruser=0"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	//rows, _ := db.Query("select * from grabcommodities")

	for rows.Next() {
		one := &Grabcorns{}
		err := rows.Scan(&one.Id, &one.Picture, &one.Pictures, &one.Title, &one.Version, &one.Date, &one.Needed, &one.End_at, &one.Kind)
		result = append(result, one)
		if err != nil {
			log.Fatal(err)
		}
	}
	return result, nil
}

func (c *Corns) Getrecords(id int) ([]*Grabcornrecords, map[string]*Grabcornrecords, error) {
	result := []*Grabcornrecords{}
	rows, err := c.db.Query(fmt.Sprintf("select id,numbers,userid,created_at from grabcornrecords where grabcornid = %d order by grabcornrecords.created_at desc", id))
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	//rows, _ := db.Query("select * from grabcommodities")
	numbers := map[string]*Grabcornrecords{}
	for rows.Next() {
		one := &Grabcornrecords{}
		err := rows.Scan(&one.Id, &one.Numbers, &one.Userid, &one.Created_at)
		result = append(result, one)
		thisnumbers := strings.Split(one.Numbers, " ")
		for _, v := range thisnumbers {
			numbers[v] = one
		}
		if err != nil {
			log.Fatal(err)
		}
	}
	return result, numbers, nil
}

func (c *Corns) Getactivity(id int) (*Grabcorns, error) {
	rows, err := c.db.Query(fmt.Sprintf("select id,picture,pictures,title,version,date,needed,end_at,kind from grabcorns where id = %d", id))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	one := &Grabcorns{}
	//rows, _ := db.Query("select * from grabcommodities")

	for rows.Next() {
		err := rows.Scan(&one.Id, &one.Picture, &one.Pictures, &one.Title, &one.Version, &one.Date, &one.Needed, &one.End_at, &one.Kind)
		if err != nil {
			log.Fatal(err)
		}
		return one, nil
	}
	return nil, fmt.Errorf("not found")
}