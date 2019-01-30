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

var indexes, db = initializeDataBase()

func addAccount(ctx *fasthttp.RequestCtx) {
	a := Account{}
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
			emailExist = true
		case bytes.Equal(key, []byte("fname")) && dataType == jsonparser.String:
			a.fName = string(value)
		case bytes.Equal(key, []byte("sname")) && dataType == jsonparser.String:
			a.sName = string(value)
		case bytes.Equal(key, []byte("phone")) && dataType == jsonparser.String:
			a.phone = string(value)
			if _, ok := indexes.phones[a.phone]; ok {
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
			a.interests = make([]string, 0)
			_, _ = jsonparser.ArrayEach(value, func(element []byte, dataType jsonparser.ValueType, offset int, err error) {
				a.interests = append(a.interests, string(element[:]))
			})
		case bytes.Equal(key, []byte("premium")) && dataType == jsonparser.Object:
			premStart, err := jsonparser.GetInt(value, "start")
			if err != nil {
				return nil
			}
			a.premiumStart = uint32(premStart)

			premFinish, err := jsonparser.GetInt(value, "finish")
			if err != nil {
				return nil
			}
			a.premiumFinish = uint32(premFinish)
		case bytes.Equal(key, []byte("likes")) && dataType == jsonparser.Array:
			a.likeIds = make([]uint32, 0)
			a.likeTss = make([]uint32, 0)
			_, _ = jsonparser.ArrayEach(value, func(element []byte, dataType jsonparser.ValueType, offset int, err error) {
				likeId, _ := jsonparser.GetInt(value, "id")
				likeTs, _ := jsonparser.GetInt(value, "ts")
				a.likeIds = append(a.likeIds, uint32(likeId))
				a.likeTss = append(a.likeTss, uint32(likeTs))
			})
		}

		return nil
	})
	if !idExist || !emailExist || !sexExist || !birthExist || !joinedExist || !statusExist || errorExist {
		ctx.SetStatusCode(400)
	} else {
		indexes.accounts[a.id] = struct{}{}
		indexes.emails[a.email] = struct{}{}
		indexes.phones[a.phone] = struct{}{}

		ctx.SetBody([]byte("{}"))
		ctx.SetStatusCode(201)
		err := insertAccount(&a, db)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func editAccount(id uint32, ctx *fasthttp.RequestCtx) {
	if _, ok := indexes.accounts[id]; !ok {
		ctx.SetStatusCode(404)
		return
	}

	accountCopy := getAccount(id, db)
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
			if _, ok := indexes.emails[accountCopy.email]; ok || !strings.Contains(accountCopy.email, "@") {
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
			if _, ok := indexes.phones[accountCopy.phone]; ok {
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
			accountCopy.interests = make([]string, 0)
			_, _ = jsonparser.ArrayEach(value, func(element []byte, dataType jsonparser.ValueType, offset int, err error) {
				accountCopy.interests = append(accountCopy.interests, string(element[:]))
			})
		case bytes.Equal(key, []byte("premium")):
			if dataType != jsonparser.Object {
				errorExist = true
				return nil
			}
			premStart, err := jsonparser.GetInt(value, "start")
			if err != nil {
				return nil
			}
			accountCopy.premiumStart = uint32(premStart)

			premFinish, err := jsonparser.GetInt(value, "finish")
			if err != nil {
				return nil
			}
			accountCopy.premiumFinish = uint32(premFinish)
		case bytes.Equal(key, []byte("likes")):
			if dataType != jsonparser.Array {
				errorExist = true
				return nil
			}
			accountCopy.likeIds = make([]uint32, 0)
			accountCopy.likeTss = make([]uint32, 0)
			_, _ = jsonparser.ArrayEach(value, func(element []byte, dataType jsonparser.ValueType, offset int, err error) {
				likeId, err1 := jsonparser.GetInt(value, "id")
				likeTs, err2 := jsonparser.GetInt(value, "ts")
				if err1 != nil || err2 != nil {
					errorExist = true
					return
				}
				accountCopy.likeIds = append(accountCopy.likeIds, uint32(likeId))
				accountCopy.likeTss = append(accountCopy.likeTss, uint32(likeTs))
			})
		}

		return nil
	})

	if errorExist {
		ctx.SetStatusCode(400)
	} else {
		updateAccount(accountCopy, db)
		indexes.emails[accountCopy.email] = struct{}{}
		indexes.phones[accountCopy.phone] = struct{}{}

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
			if _, ok := indexes.accounts[like.liker]; !ok {
				errorExist = true
				return
			}
			if _, ok := indexes.accounts[like.likee]; !ok {
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
			likeIds, likeTss := getLikes(like.liker, db)
			/*
				if accountCopy.likeIds == nil {
					accountCopy.likeIds = make([]uint32, 0)
				}
				if accountCopy.likeTss == nil {
					accountCopy.likeTss = make([]uint32, 0)
				}
			*/
			*likeIds = append(*likeIds, uint32(like.likee))
			*likeTss = append(*likeTss, uint32(like.ts))

			updateLikes(like.liker, likeIds, likeTss, db)
		}

		ctx.SetBody([]byte("{}"))
		ctx.SetStatusCode(202)
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	println("fasthttp")
	s := fasthttp.Server{
		Handler: handler,
	}

	println("ListenAndServe")
	err := s.ListenAndServe("0.0.0.0:80")
	if err != nil {
		log.Fatalf("error in ListenAndServe: %s", err)
	}

	println("exit")
	//db.Close()
}

func handler(ctx *fasthttp.RequestCtx) {
	path := ctx.Path()
	method := ctx.Method()
	//length := len(path)

	if bytes.Equal(method, []byte("POST")) {
		if bytes.Equal(path[9:], []byte("/new/")) {
			println("addAccount")
			addAccount(ctx)
		} else if bytes.Equal(path[9:], []byte("/likes/")) {
			println("addLikes")
			addLikes(ctx)
		} else {
			if id, err := strconv.Atoi(string(path[10 : len(path)-1])); err == nil {
				println("editAccount")
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
