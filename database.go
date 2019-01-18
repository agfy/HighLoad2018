package main

import (
	"fmt"
	"github.com/buger/jsonparser"
	"io/ioutil"
	"os"
	//"os/exec"
	"runtime"
	"strconv"
)

func initializeSchema() (db*Schema) {
	db = &Schema{make(map[uint]*Account, 0)}
	/*
	_, err := exec.Command("sh","-c", "unzip /tmp/data/data.zip -d /tmp/base/").Output()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stdout, err.Error())
	}
*/
	id := 1
	var fileName = "/tmp/base/accounts_" + strconv.Itoa(id) + ".json"
	//var fileName = "C:\\Users\\agfy1\\Downloads\\test_accounts_140119\\data\\data\\accounts_" + strconv.Itoa(id) + ".json"

	for fileExists(fileName) {
		dat, _ := ioutil.ReadFile(fileName)
		c := 0
		_, _ = jsonparser.ArrayEach(dat, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			c++
			id, _ := jsonparser.GetInt(value, "id")
			email, _ := jsonparser.GetString(value, "email")
			fName, _ := jsonparser.GetString(value, "fname")
			sName, _ := jsonparser.GetString(value, "sname")
			phone, _ := jsonparser.GetString(value, "phone")
			sex, _ := jsonparser.GetString(value, "sex")
			birth , _ := jsonparser.GetInt(value, "birth")
			country  , _ := jsonparser.GetString(value, "country ")
			city , _ := jsonparser.GetString(value, "city")
			joined , _ := jsonparser.GetInt(value, "joined")
			statusStr, _ := jsonparser.GetString(value, "status")
			var status int8 = 0
			switch statusStr {
			case "свободны":
				status = 0
			case "заняты":
				status = 1
			case  "всё сложно":
				status = 2
			}
			interests := make([]string, 0)
			_, _ = jsonparser.ArrayEach(value, func(element []byte, dataType jsonparser.ValueType, offset int, err error) {
				interests = append(interests, string(element[:]))
			}, "interests")
			premStart , _ := jsonparser.GetInt(value, "premium", "start")
			premFinish , _ := jsonparser.GetInt(value, "premium", "finish")
			likes := make([]Like, 0)
			_, _ = jsonparser.ArrayEach(value, func(element []byte, dataType jsonparser.ValueType, offset int, err error) {
				likeId, _ := jsonparser.GetInt(value, "id")
				likeTs , _ := jsonparser.GetInt(value, "ts")
				likes = append(likes, Like{uint32(likeId), uint32(likeTs)})
			}, "likes")

			a := &Account{uint32(id),email,fName,sName, phone, sex == "m",uint32(birth),
				country, city, uint32(joined), status, &interests, Premium{uint32(premStart),
					uint32(premFinish)}, &likes}
			db.accounts[uint(id)] = a
		}, "accounts")
		id++
		fileName = "/tmp/base/users_" + strconv.Itoa(id) + ".json"
		//fileName = "C:\\Users\\agfy1\\Downloads\\test_accounts_140119\\data\\data\\accounts_" + strconv.Itoa(id) + ".json"

	}
	runtime.GC()

	_, _ = fmt.Fprintf(os.Stdout, "Accounts: %d\n", len(db.accounts))
	return
}

