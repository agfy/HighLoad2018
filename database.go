package main

import (
	"archive/zip"
	"database/sql"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/lib/pq"
	"io/ioutil"
	"log"
	"os"
	"runtime"
)

/*
var createTableStr = "CREATE TABLE accounts(" +
	"id             serial NOT NULL," +
	"email          varchar(100) NOT NULL," +
	"fname          varchar(50) NOT NULL," +
	"sname          varchar(50) NOT NULL," +
	"phone          varchar(16) NOT NULL," +
	"sex            boolean NOT NULL," +
	"birth          integer NOT NULL," +
	"country        varchar(50) NOT NULL," +
	"city           varchar(50) NOT NULL," +
	"joined         integer NOT NULL," +
	"status         integer NOT NULL," +
	"interests      varchar(100)[] NOT NULL DEFAULT '{}'::varchar(100)[]," +
	"premium_start  integer NOT NULL," +
	"premium_finish integer NOT NULL," +
	"like_ids       integer[] NOT NULL DEFAULT '{}'::int[]," +
	"like_tss       integer[] NOT NULL DEFAULT '{}'::int[]" +
	");"
*/

var createTableStr = "CREATE TABLE accounts(" +
	"id             serial NOT NULL," +
	"email          text NOT NULL," +
	"fname          text NOT NULL," +
	"sname          text NOT NULL," +
	"phone          text NOT NULL," +
	"sex            boolean NOT NULL," +
	"birth          integer NOT NULL," +
	"country        text NOT NULL," +
	"city           text NOT NULL," +
	"joined         integer NOT NULL," +
	"status         integer NOT NULL," +
	"interests      text[]," +
	"premium_start  integer NOT NULL," +
	"premium_finish integer NOT NULL," +
	"like_ids       integer[]," +
	"like_tss       integer[]" +
	");"

func initializeDataBase() (indexes *Indexes, db *sql.DB) {
	println("initializeDataBase")

	indexes = &Indexes{make(map[uint32]struct{}, 0), make(map[string]struct{}), make(map[string]struct{})}

	//db, err := sql.Open("postgres", "host=0.0.0.0 sslmode=disable")
	db, err := sql.Open("postgres", "postgres://postgres:@localhost/postgres?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(createTableStr)
	if err != nil {
		log.Fatal(err)
	}
	/*
		_, err = db.Exec("SET statement_timeout TO 2000;")
		if err != nil {
			log.Fatal(err)
		}
	*/
	file, err := zip.OpenReader("/tmp/data/data.zip")
	//file, err := zip.OpenReader("C:\\Users\\agfy1\\Downloads\\test_accounts_140119\\data\\data.zip")
	if err != nil {
		log.Fatal(err)
	}
	c := 0

	for _, f := range file.File {
		rc, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}
		bs, _ := ioutil.ReadAll(rc)
		_, _ = jsonparser.ArrayEach(bs, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			c++

			id, _ := jsonparser.GetInt(value, "id")
			indexes.accounts[uint32(id)] = struct{}{}
			email, _ := jsonparser.GetString(value, "email")
			//println("email: " + email)
			indexes.emails[email] = struct{}{}
			fName, _ := jsonparser.GetString(value, "fname")
			sName, _ := jsonparser.GetString(value, "sname")
			phone, _ := jsonparser.GetString(value, "phone")
			indexes.phones[phone] = struct{}{}
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
				//println("interests: " + string(element[:]))
			}, "interests")
			premStart, _ := jsonparser.GetInt(value, "premium", "start")
			premFinish, _ := jsonparser.GetInt(value, "premium", "finish")
			likeIds := make([]uint32, 0)
			likeTss := make([]uint32, 0)
			_, _ = jsonparser.ArrayEach(value, func(element []byte, dataType jsonparser.ValueType, offset int, err error) {
				likeId, _ := jsonparser.GetInt(value, "id")
				likeTs, _ := jsonparser.GetInt(value, "ts")
				likeIds = append(likeIds, uint32(likeId))
				likeTss = append(likeTss, uint32(likeTs))
			}, "likes")

			a := &Account{uint32(id), email, fName, sName, phone, sex == "m", uint32(birth),
				country, city, uint32(joined), status, interests, uint32(premStart),
				uint32(premFinish), likeIds, likeTss}

			err = insertAccount(a, db)
			if err != nil {
				log.Fatal(err)
			}
		}, "accounts")
	}
	runtime.GC()

	_, _ = fmt.Fprintf(os.Stdout, "Accounts: %d\n", c)
	return
}

func insertAccount(acc *Account, db *sql.DB) error {
	//println("insertAccount")
	_, err := db.Exec("INSERT INTO accounts (id, email, fname, sname, phone, sex, birth, country, city, joined, "+
		"status, interests, premium_start, premium_finish, like_ids, like_tss) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, "+
		"$10, $11, $12, $13, $14, $15, $16)", acc.id, acc.email, acc.fName, acc.sName, acc.phone, acc.sex, acc.birth,
		acc.country, acc.city, acc.joined, acc.status, pq.Array(acc.interests), acc.premiumStart, acc.premiumFinish,
		pq.Array(acc.likeIds), pq.Array(acc.likeTss))

	if err != nil {
		return err
	}

	return nil
}

func updateAccount(acc *Account, db *sql.DB) error {
	//println("updateAccount")
	_, err := db.Exec("UPDATE accounts SET email = $1, fname = $2, sname = $3, phone = $4, sex = $5, birth = $6, "+
		"country = $7, city = $8, joined = $9, status = $10, interests = $11, premium_start = $12, premium_finish = $13, "+
		"like_ids = $14, like_tss = $15 WHERE id=$16", acc.email, acc.fName, acc.sName, acc.phone, acc.sex, acc.birth,
		acc.country, acc.city, acc.joined, acc.status, pq.Array(acc.interests), acc.premiumStart, acc.premiumFinish,
		pq.Array(acc.likeIds), pq.Array(acc.likeTss), acc.id)

	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func getAccount(id uint32, db *sql.DB) *Account {
	//println("getAccount")
	row := db.QueryRow("SELECT id, email, fname, sname, phone, sex, birth, country, city, joined, "+
		"status, interests, premium_start, premium_finish, like_ids, like_tss FROM accounts WHERE id=$1", id)

	a := Account{}
	a.interests = make([]string, 0)
	a.likeIds = make([]uint32, 0)
	a.likeTss = make([]uint32, 0)

	var ids pq.Int64Array
	var tss pq.Int64Array

	err := row.Scan(&a.id, &a.email, &a.fName, &a.sName, &a.phone, &a.sex, &a.birth, &a.country, &a.city, &a.joined, &a.status,
		pq.Array(&a.interests), &a.premiumStart, &a.premiumFinish, &ids, &tss)

	for i, num := range ids {
		a.likeIds = append(a.likeIds, uint32(num))
		a.likeTss = append(a.likeTss, uint32(tss[i]))
	}

	if err != nil {
		log.Fatal(err)
	}

	return &a
}

func getLikes(id uint32, db *sql.DB) (*[]uint32, *[]uint32) {
	//println("getLikes")
	row := db.QueryRow("SELECT like_ids, like_tss FROM accounts WHERE id=$1", id)

	likeIds := make([]uint32, 0)
	likeTss := make([]uint32, 0)

	var ids pq.Int64Array
	var tss pq.Int64Array

	err := row.Scan(&ids, &tss)

	for i, num := range ids {
		likeIds = append(likeIds, uint32(num))
		likeTss = append(likeTss, uint32(tss[i]))
	}

	if err != nil {
		log.Fatal(err)
	}

	return &likeIds, &likeTss
}

func updateLikes(id uint32, likeIds *[]uint32, likeTss *[]uint32, db *sql.DB) error {
	//println("updateLikes")
	_, err := db.Exec("UPDATE accounts SET like_ids = $1, like_tss = $2 WHERE id=$3", pq.Array(*likeIds),
		pq.Array(*likeTss), id)

	if err != nil {
		log.Fatal(err)
	}

	return nil
}
