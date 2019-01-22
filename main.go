package main

import (
	"bytes"
	"github.com/buger/jsonparser"
	"github.com/valyala/fasthttp"
	"log"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var db = initializeSchema()

func addAccount(ctx *fasthttp.RequestCtx) {
	a := &Account{}
	body := ctx.Request.Body()

	var idExist, emailExist, sexExist, birthExist, joinedExist, statusExist, errorExist bool
	_ = jsonparser.ObjectEach(body, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		if errorExist {
			return nil
		}

		switch {
		case bytes.Equal(key, []byte("id")) && dataType == jsonparser.Number:
			i, err := jsonparser.GetInt(value)
			if err != nil {
				return nil
			}
			a.id = uint32(i)
			idExist = true
		case bytes.Equal(key, []byte("email")) && dataType == jsonparser.String:
			a.email = string(value)
			/*
				if _, ok := db.emails[a.email]; ok {
					return nil
				}
			*/
			emailExist = true
		case bytes.Equal(key, []byte("fname")) && dataType == jsonparser.String:
			a.fName = string(value)
		case bytes.Equal(key, []byte("sname")) && dataType == jsonparser.String:
			a.sName = string(value)
		case bytes.Equal(key, []byte("phone")) && dataType == jsonparser.String:
			a.phone = string(value)
			if _, ok := db.phones[a.phone]; ok {
				errorExist = true
				return nil
			}
		case bytes.Equal(key, []byte("sex")) && dataType == jsonparser.String:
			sexStr, _ := jsonparser.ParseString(value)
			if sexStr != "m" && sexStr != "f" {
				errorExist = true
				return nil
			}
			a.sex = sexStr == "m"
			sexExist = true
		case bytes.Equal(key, []byte("birth")) && dataType == jsonparser.Number:
			b, err := jsonparser.GetInt(value)
			if err != nil {
				return nil
			}
			a.birth = uint32(b)
			birthExist = true
		case bytes.Equal(key, []byte("country")) && dataType == jsonparser.String:
			a.country = string(value)
		case bytes.Equal(key, []byte("city")) && dataType == jsonparser.String:
			a.city = string(value)
		case bytes.Equal(key, []byte("joined")) && dataType == jsonparser.Number:
			j, err := jsonparser.GetInt(value)
			if err != nil {
				return nil
			}
			a.joined = uint32(j)
			joinedExist = true
		case bytes.Equal(key, []byte("status")) && dataType == jsonparser.String:
			statusStr, _ := jsonparser.ParseString(value)
			switch statusStr {
			case "свободны":
				a.status = 0
			case "заняты":
				a.status = 1
			case "всё сложно":
				a.status = 2
			default:
				errorExist = true
				return nil
			}

			statusExist = true
		case bytes.Equal(key, []byte("interests")) && dataType == jsonparser.Array:
			interests := make([]string, 0)
			_, _ = jsonparser.ArrayEach(value, func(element []byte, dataType jsonparser.ValueType, offset int, err error) {
				interests = append(interests, string(element[:]))
			})

			a.interests = &interests
		case bytes.Equal(key, []byte("premium")) && dataType == jsonparser.Object:
			premStart, err := jsonparser.GetInt(value, "start")
			if err != nil {
				return nil
			}
			a.premium.start = uint32(premStart)

			premFinish, err := jsonparser.GetInt(value, "finish")
			if err != nil {
				return nil
			}
			a.premium.finish = uint32(premFinish)
		case bytes.Equal(key, []byte("likes")) && dataType == jsonparser.Array:
			likes := make([]Like, 0)
			_, _ = jsonparser.ArrayEach(value, func(element []byte, dataType jsonparser.ValueType, offset int, err error) {
				likeId, _ := jsonparser.GetInt(value, "id")
				likeTs, _ := jsonparser.GetInt(value, "ts")
				likes = append(likes, Like{uint32(likeId), uint32(likeTs)})
			})

			a.likes = &likes
			//likesExist = true
		}

		return nil
	})
	if !idExist || !emailExist || !sexExist || !birthExist || !joinedExist || !statusExist || errorExist {
		ctx.SetStatusCode(400)
	} else {
		db.emails[a.email] = struct{}{}
		db.phones[a.phone] = struct{}{}

		ctx.SetBody([]byte("{}"))
		ctx.SetStatusCode(201)
		db.accounts[a.id] = a
	}
}

func editAccount(id uint32, ctx *fasthttp.RequestCtx) {
	if _, ok := db.accounts[id]; !ok {
		ctx.SetStatusCode(404)
		return
	}

	accountCopy := *db.accounts[id]
	body := ctx.Request.Body()

	var errorExist bool
	_ = jsonparser.ObjectEach(body, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		if errorExist {
			return nil
		}

		switch {
		case bytes.Equal(key, []byte("email")):
			if dataType != jsonparser.String {
				errorExist = true
				return nil
			}
			accountCopy.email = string(value)
			if _, ok := db.emails[accountCopy.email]; ok || !strings.Contains(accountCopy.email, "@") {
				errorExist = true
				return nil
			}
		case bytes.Equal(key, []byte("fname")):
			if dataType != jsonparser.String {
				errorExist = true
				return nil
			}
			accountCopy.fName = string(value)
		case bytes.Equal(key, []byte("sname")):
			if dataType != jsonparser.String {
				errorExist = true
				return nil
			}
			accountCopy.sName = string(value)
		case bytes.Equal(key, []byte("phone")):
			if dataType != jsonparser.String {
				errorExist = true
				return nil
			}
			accountCopy.phone = string(value)
			if _, ok := db.phones[accountCopy.phone]; ok {
				errorExist = true
				return nil
			}
		case bytes.Equal(key, []byte("sex")):
			if dataType != jsonparser.String {
				errorExist = true
				return nil
			}
			sexStr, _ := jsonparser.ParseString(value)
			if sexStr != "m" && sexStr != "f" {
				errorExist = true
				return nil
			}
			accountCopy.sex = sexStr == "m"
		case bytes.Equal(key, []byte("birth")):
			if dataType != jsonparser.Number {
				errorExist = true
				return nil
			}
			b, err := jsonparser.GetInt(value)
			if err != nil {
				return nil
			}
			accountCopy.birth = uint32(b)
		case bytes.Equal(key, []byte("country")):
			if dataType != jsonparser.String {
				errorExist = true
				return nil
			}
			accountCopy.country = string(value)
		case bytes.Equal(key, []byte("city")):
			if dataType != jsonparser.String {
				errorExist = true
				return nil
			}
			accountCopy.city = string(value)
		case bytes.Equal(key, []byte("joined")):
			if dataType != jsonparser.Number {
				errorExist = true
				return nil
			}
			j, err := jsonparser.GetInt(value)
			if err != nil {
				return nil
			}
			accountCopy.joined = uint32(j)
		case bytes.Equal(key, []byte("status")):
			if dataType != jsonparser.String {
				errorExist = true
				return nil
			}
			statusStr, _ := jsonparser.ParseString(value)
			switch statusStr {
			case "свободны":
				accountCopy.status = 0
			case "заняты":
				accountCopy.status = 1
			case "всё сложно":
				accountCopy.status = 2
			default:
				errorExist = true
				return nil
			}
		case bytes.Equal(key, []byte("interests")):
			if dataType != jsonparser.Array {
				errorExist = true
				return nil
			}
			interests := make([]string, 0)
			_, _ = jsonparser.ArrayEach(value, func(element []byte, dataType jsonparser.ValueType, offset int, err error) {
				interests = append(interests, string(element[:]))
			})

			accountCopy.interests = &interests
		case bytes.Equal(key, []byte("premium")):
			if dataType != jsonparser.Object {
				errorExist = true
				return nil
			}
			premStart, err := jsonparser.GetInt(value, "start")
			if err != nil {
				return nil
			}
			accountCopy.premium.start = uint32(premStart)

			premFinish, err := jsonparser.GetInt(value, "finish")
			if err != nil {
				return nil
			}
			accountCopy.premium.finish = uint32(premFinish)
		case bytes.Equal(key, []byte("likes")):
			if dataType != jsonparser.Array {
				errorExist = true
				return nil
			}
			likes := make([]Like, 0)
			_, _ = jsonparser.ArrayEach(value, func(element []byte, dataType jsonparser.ValueType, offset int, err error) {
				likeId, err1 := jsonparser.GetInt(value, "id")
				likeTs, err2 := jsonparser.GetInt(value, "ts")
				if err1 != nil || err2 != nil {
					errorExist = true
					return
				}
				likes = append(likes, Like{uint32(likeId), uint32(likeTs)})
			})

			accountCopy.likes = &likes
		}

		return nil
	})

	if errorExist {
		ctx.SetStatusCode(400)
	} else {
		db.accounts[id] = &accountCopy
		db.emails[accountCopy.email] = struct{}{}
		db.phones[accountCopy.phone] = struct{}{}

		ctx.SetBody([]byte("{}"))
		ctx.SetStatusCode(202)
	}
}

func addLikes(ctx *fasthttp.RequestCtx) {
	body := ctx.Request.Body()
	var errorExist bool
	now := time.Now().Unix()

	type requestLike struct {
		liker uint32
		likee uint32
		ts    uint32
	}
	requestLikes := make([]*requestLike, 0)

	_, _ = jsonparser.ArrayEach(body, func(element []byte, dataType jsonparser.ValueType, offset int, err error) {
		var likerExist, likeeExist, tsExist bool
		var like requestLike

		if errorExist {
			return
		}

		_ = jsonparser.ObjectEach(element, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
			switch {
			case bytes.Equal(key, []byte("liker")) && dataType == jsonparser.Number:
				i, err := jsonparser.GetInt(value)
				if err != nil {
					errorExist = true
					return nil
				}
				like.liker = uint32(i)
				likerExist = true
			case bytes.Equal(key, []byte("likee")) && dataType == jsonparser.Number:
				i, err := jsonparser.GetInt(value)
				if err != nil {
					errorExist = true
					return nil
				}
				like.likee = uint32(i)
				likeeExist = true
			case bytes.Equal(key, []byte("ts")) && dataType == jsonparser.Number:
				i, err := jsonparser.GetInt(value)
				if err != nil || i > now {
					errorExist = true
					return nil
				}
				like.ts = uint32(i)
				tsExist = true
			}

			return nil
		})

		if !likeeExist || !likeeExist || !tsExist {
			errorExist = true
			return
		} else {
			if _, ok := db.accounts[like.liker]; !ok {
				errorExist = true
				return
			}
			if _, ok := db.accounts[like.likee]; !ok {
				errorExist = true
				return
			}

			requestLikes = append(requestLikes, &like)
		}

	}, "likes")

	if errorExist {
		ctx.SetStatusCode(400)
		return
	} else {
		for _, like := range requestLikes {
			if db.accounts[like.liker].likes == nil {
				likes := make([]Like, 0)
				(*db.accounts[like.liker]).likes = &likes
			}
			*db.accounts[like.liker].likes = append(*db.accounts[like.liker].likes, Like{like.likee, like.ts})
		}

		ctx.SetBody([]byte("{}"))
		ctx.SetStatusCode(202)
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	s := fasthttp.Server{
		Handler: handler,
	}

	err := s.ListenAndServe("0.0.0.0:80")
	if err != nil {
		log.Fatalf("error in ListenAndServe: %s", err)
	}
}

func handler(ctx *fasthttp.RequestCtx) {
	path := ctx.Path()
	method := ctx.Method()
	//length := len(path)

	if bytes.Equal(method, []byte("POST")) {
		if bytes.Equal(path[9:], []byte("/new/")) {
			addAccount(ctx)
		} else if bytes.Equal(path[9:], []byte("/likes/")) {
			addLikes(ctx)
		} else {
			if id, err := strconv.Atoi(string(path[10 : len(path)-1])); err == nil {
				editAccount(uint32(id), ctx)
			} else {
				ctx.SetStatusCode(404)
			}
		}
	} else if bytes.Equal(method, []byte("GET")) {
		ctx.SetStatusCode(404)
	} else {
		ctx.SetStatusCode(404)
	}
}
