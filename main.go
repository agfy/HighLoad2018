package main

import (
	"bytes"
	"github.com/buger/jsonparser"
	"github.com/valyala/fasthttp"
	"log"
	"runtime"
	"strconv"
)

var db = initializeSchema()

func addAccount(ctx *fasthttp.RequestCtx) {
	a := &Account{}
	body := ctx.Request.Body()

	var idExist, emailExist, sexExist, birthExist, joinedExist, statusExist, interestsExist, premStartExist bool
	var premFinishExist, likesExist, wrongPhone bool
	_ = jsonparser.ObjectEach(body, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
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
			if _, ok := db.emails[a.email]; ok {
				return nil
			}
			db.emails[a.email] = struct{}{}
			emailExist = true
		case bytes.Equal(key, []byte("fname")) && dataType == jsonparser.String:
			a.fName = string(value)
		case bytes.Equal(key, []byte("sname")) && dataType == jsonparser.String:
			a.sName = string(value)
		case bytes.Equal(key, []byte("phone")) && dataType == jsonparser.String:
			a.phone = string(value)
			if _, ok := db.phones[a.phone]; ok {
				wrongPhone = true
				return nil
			}
			db.emails[a.email] = struct{}{}
		case bytes.Equal(key, []byte("sex")) && dataType == jsonparser.String:
			sexStr, _ := jsonparser.ParseString(value)
			if sexStr != "m" && sexStr != "f" {
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
				return nil
			}

			statusExist = true
		case bytes.Equal(key, []byte("interests")) && dataType == jsonparser.Array:
			interests := make([]string, 0)
			_, _ = jsonparser.ArrayEach(value, func(element []byte, dataType jsonparser.ValueType, offset int, err error) {
				interests = append(interests, string(element[:]))
			})

			a.interests = &interests
			interestsExist = true
		case bytes.Equal(key, []byte("premium")) && dataType == jsonparser.Object:
			premStart, err := jsonparser.GetInt(value, "start")
			if err != nil {
				return nil
			}

			a.premium.start = uint32(premStart)
			premStartExist = true

			premFinish, err := jsonparser.GetInt(value, "finish")
			if err != nil {
				return nil
			}

			a.premium.finish = uint32(premFinish)
			premFinishExist = true
		case bytes.Equal(key, []byte("likes")) && dataType == jsonparser.Array:
			likes := make([]Like, 0)
			_, _ = jsonparser.ArrayEach(value, func(element []byte, dataType jsonparser.ValueType, offset int, err error) {
				likeId, _ := jsonparser.GetInt(value, "id")
				likeTs, _ := jsonparser.GetInt(value, "ts")
				likes = append(likes, Like{uint32(likeId), uint32(likeTs)})
			})

			a.likes = &likes
			likesExist = true
		}

		return nil
	})
	if !idExist || !emailExist || !sexExist || !birthExist || !joinedExist || !statusExist || !interestsExist ||
		!premStartExist || !premFinishExist || !likesExist || wrongPhone {
		ctx.SetStatusCode(400)
	} else {
		ctx.SetBody([]byte("{}"))
		ctx.SetStatusCode(201)
		db.accounts[a.id] = a
	}
}

func addLikes(ctx *fasthttp.RequestCtx) {
	body := ctx.Request.Body()
	var errorExist bool

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
					return nil
				}
				like.liker = uint32(i)
				likerExist = true
			case bytes.Equal(key, []byte("likee")) && dataType == jsonparser.Number:
				i, err := jsonparser.GetInt(value)
				if err != nil {
					return nil
				}
				like.likee = uint32(i)
				likeeExist = true
			case bytes.Equal(key, []byte("ts")) && dataType == jsonparser.Number:
				i, err := jsonparser.GetInt(value)
				if err != nil {
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

			requestLikes = append(requestLikes, &like)
		}

	}, "likes")

	if errorExist {
		ctx.SetStatusCode(400)
		return
	} else {
		for _, like := range requestLikes {
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

	err := s.ListenAndServe("127.0.0.1:80")
	if err != nil {
		log.Fatalf("error in ListenAndServe: %s", err)
	}
}

func handler(ctx *fasthttp.RequestCtx) {
	path := ctx.Path()
	method := ctx.Method()
	//length := len(path)

	if bytes.Equal(method, []byte("POST")) {
		if bytes.Equal(path, []byte("/new/")) {
			addAccount(ctx)
		} else if bytes.Equal(path, []byte("/likes/")) {
			addLikes(ctx)
		} else {
			if _, err := strconv.Atoi(string(path[9 : len(path)-1])); err == nil {
				//editAccount(id, ctx)
			} else {
				ctx.SetStatusCode(404)
			}
		}
	} else if bytes.Equal(method, []byte("GET")) {

	} else {
		ctx.SetStatusCode(404)
	}
}
