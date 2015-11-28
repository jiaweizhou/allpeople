package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Grabcommodities struct {
	Id             int `orm:"auto"`
	Picture        string
	Details        string
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
	Worth          int
}

type Grabcommodityrecords struct {
	Id              int `orm:"auto"`
	Userid          int
	Grabcommodityid int
	Count           int
	Numbers         string
	Type            int
	Created_at      int
}

type Commodities struct {
	db      *sql.DB
	process map[int]chan int
}

func NewCommodities(db *sql.DB) *Commodities {
	return &Commodities{
		db:      db,
		process: make(map[int]chan int),
	}
}
func (c *Commodities) Waitforopen(response http.ResponseWriter, request *http.Request) {
	request.ParseForm()

	grabcommodityid := request.Form.Get("grabcommodityid")
	id, err := strconv.Atoi(grabcommodityid)
	fmt.Println(grabcommodityid)

	result, err := c.Getactivity(id)
	if err != nil {
		response.Write([]byte("{'flag':0}"))
		return
	}
	go func() {
		c.process[result.Id] = make(chan int)
		c.Open(result, c.process[result.Id])
	}()
	response.Write([]byte("{'flag':1}"))
}
func (c *Commodities) Serve() {
	result, err := c.Getactivities()
	if err != nil {
		log.Println("get Getactivities:" + err.Error())
	}
	for _, v := range result {
		go func(v *Grabcommodities) {
			c.process[v.Id] = make(chan int)
			c.Open(v, c.process[v.Id])
		}(v)
	}
}
func (c *Commodities) Open(activity *Grabcommodities, end chan int) {
	log.Println("start open title:", activity.Title, " version:", activity.Version)
	log.Println(time.Unix(int64(activity.End_at), 0).Sub(time.Now()).String())
	//fmt.Println(time.Unix(int64(activity.End_at), 0).Sub(time.Now()).String())
	ch := time.Tick(time.Unix(int64(activity.End_at), 0).Sub(time.Now()))
	defer close(end)
	select {
	case <-ch:
		records, numbers, err := c.Getrecords(activity.Id)
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
		number := total%activity.Needed + 10000001
		winnerrecord := numbers[strconv.Itoa(number)]

		_, err = c.db.Exec(fmt.Sprintf("update grabcommodities set islotteried=%d,winneruserid=%d,winnerrecordid=%d,winnernumber=%d where id = %d", 1, winnerrecord.Userid, winnerrecord.Id, number, activity.Id))
		//_, err = c.db.Exec(fmt.Sprintf("insert into grabcommodities(picture,pictures,details,title,version,needed,remain,created_at,date,end_at,islotteried,winneruserid,foruser,kind) select picture,pictures,details,title,version+1,needed,needed,%d,%d,0,0,0,0,kind from grabcorns where id = %d", time.Now().Unix(), time.Now().Unix(), activity.Id))
		//		form := url.Values{}
		//		form.Add("picture", activity.Picture)   //"http://7xoc8r.com2.z0.glb.qiniucdn.com/corns/20071224162158623_2.jpg")
		//		form.Add("pictures", activity.Pictures) //"http://7xoc8r.com2.z0.glb.qiniucdn.com/corns/20071224162158623_2.jpg http://7xoc8r.com2.z0.glb.qiniucdn.com/corns/2007822154648385_2.jpg http://7xoc8r.com2.z0.glb.qiniucdn.com/corns/2531170_193356481000_2.jpg http://7xoc8r.com2.z0.glb.qiniucdn.com/corns/5528723_101453638160_2.jpg http://7xoc8r.com2.z0.glb.qiniucdn.com/corns/6348-12011120200785.jpg http://7xoc8r.com2.z0.glb.qiniucdn.com/corns/8337244_105659585000_2.jpg")
		//		form.Add("details", activity.Details)
		//		form.Add("title", activity.Title)
		//		form.Add("version", strconv.Itoa(activity.Version+1))
		//		form.Add("needed", strconv.Itoa(activity.Needed))
		//		form.Add("date", fmt.Sprint(time.Now().Unix()))
		//		form.Add("kind", strconv.Itoa(activity.Kind))
		//		form.Add("worth", strconv.Itoa(activity.Worth))
		//		response, err := http.PostForm("http://183.129.190.82:50001/v1/grabcommodities/create", form)
		//		if err != nil {
		//			log.Println("create grabcommodities err:" + err.Error())
		//		} else {
		//			defer response.Body.Close()
		//			tt, _ := ioutil.ReadAll(response.Body)
		//			log.Println("create grabcommodities:" + string(tt))
		//			log.Println("open success")
		//		}

	case <-end:
		return
	}
}
func (c *Commodities) Getactivities() ([]*Grabcommodities, error) {
	result := []*Grabcommodities{}
	rows, err := c.db.Query(fmt.Sprintf("select id,picture,pictures,details,title,version,date,needed,end_at,kind,worth from grabcommodities where islotteried = 0 and end_at!=0 and foruser=0"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	//rows, _ := db.Query("select * from grabcommodities")

	for rows.Next() {
		one := &Grabcommodities{}
		err := rows.Scan(&one.Id, &one.Picture, &one.Pictures, &one.Details, &one.Title, &one.Version, &one.Date, &one.Needed, &one.End_at, &one.Kind, &one.Worth)
		result = append(result, one)
		if err != nil {
			log.Println("Get Grabcommodities err:" + err.Error())
			return result, err
		}
	}
	return result, nil
}

func (c *Commodities) Getrecords(id int) ([]*Grabcommodityrecords, map[string]*Grabcommodityrecords, error) {
	result := []*Grabcommodityrecords{}
	rows, err := c.db.Query(fmt.Sprintf("select id,numbers,userid,created_at from grabcommodityrecords where grabcommodityid = %d order by grabcommodityrecords.created_at desc", id))
	if err != nil {
		log.Println("Get Grabcommodityrecords err:" + err.Error())
		return nil, nil, err
	}
	defer rows.Close()

	//rows, _ := db.Query("select * from grabcommodities")
	numbers := map[string]*Grabcommodityrecords{}
	for rows.Next() {
		one := &Grabcommodityrecords{}
		err := rows.Scan(&one.Id, &one.Numbers, &one.Userid, &one.Created_at)
		result = append(result, one)
		thisnumbers := strings.Split(one.Numbers, " ")
		for _, v := range thisnumbers {
			numbers[v] = one
		}
		if err != nil {
			log.Println("Get Grabcommodityrecords err:" + err.Error())
			return result, numbers, err
		}
	}
	return result, numbers, nil
}

func (c *Commodities) Getactivity(id int) (*Grabcommodities, error) {
	rows, err := c.db.Query(fmt.Sprintf("select id,picture,pictures,details,title,version,date,needed,end_at,kind,worth from grabcommodities where id = %d", id))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer rows.Close()
	one := &Grabcommodities{}
	//rows, _ := db.Query("select * from grabcommodities")

	for rows.Next() {
		err := rows.Scan(&one.Id, &one.Picture, &one.Pictures, &one.Details, &one.Title, &one.Version, &one.Date, &one.Needed, &one.End_at, &one.Kind, &one.Worth)
		if err != nil {
			//fmt.Println(err)
			log.Println("Get Grabcommodity:" + err.Error())
			return nil, err
		}
		return one, nil
	}
	return nil, fmt.Errorf("not found")
}
