package main

import (
	"archive/zip"
	"fmt"
	"github.com/buger/jsonparser"
	"io/ioutil"
	"log"
	"os"
	//"os/exec"
	"runtime"
)

func initializeSchema() (db *Schema) {
	db = &Schema{make(map[uint32]*Account, 0), make(map[string]struct{}), make(map[string]struct{})}

	file, err := zip.OpenReader("/tmp/data/data.zip")
	//file, err := zip.OpenReader("C:\\Users\\agfy1\\Downloads\\test_accounts_140119\\data\\data.zip")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	for _, f := range file.File {
		rc, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}
		bs, _ := ioutil.ReadAll(rc)
		_, _ = jsonparser.ArrayEach(bs, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			id, _ := jsonparser.GetInt(value, "id")
			email, _ := jsonparser.GetString(value, "email")
			db.emails[email] = struct{}{}
			fName, _ := jsonparser.GetString(value, "fname")
			sName, _ := jsonparser.GetString(value, "sname")
			phone, _ := jsonparser.GetString(value, "phone")
			db.phones[phone] = struct{}{}
			sex, _ := jsonparser.GetString(value, "sex")
			birth, _ := jsonparser.GetInt(value, "birth")
			country, _ := jsonparser.GetString(value, "country ")
			city, _ := jsonparser.GetString(value, "city")
			joined, _ := jsonparser.GetInt(value, "joined")
			statusStr, _ := jsonparser.GetString(value, "status")
			var status int8 = 0
			switch statusStr {
			case "свободны":
				status = 0
			case "заняты":
				status = 1
			case "всё сложно":
				status = 2
			}
			interests := make([]string, 0)
			_, _ = jsonparser.ArrayEach(value, func(element []byte, dataType jsonparser.ValueType, offset int, err error) {
				interests = append(interests, string(element[:]))
			}, "interests")
			premStart, _ := jsonparser.GetInt(value, "premium", "start")
			premFinish, _ := jsonparser.GetInt(value, "premium", "finish")
			likes := make([]Like, 0)
			_, _ = jsonparser.ArrayEach(value, func(element []byte, dataType jsonparser.ValueType, offset int, err error) {
				likeId, _ := jsonparser.GetInt(value, "id")
				likeTs, _ := jsonparser.GetInt(value, "ts")
				likes = append(likes, Like{uint32(likeId), uint32(likeTs)})
			}, "likes")

			a := &Account{uint32(id), email, fName, sName, phone, sex == "m", uint32(birth),
				country, city, uint32(joined), status, &interests, Premium{uint32(premStart),
					uint32(premFinish)}, &likes}
			db.accounts[uint32(id)] = a
		}, "accounts")
	}
	runtime.GC()

	_, _ = fmt.Fprintf(os.Stdout, "Accounts: %d\n", len(db.accounts))
	return
}
