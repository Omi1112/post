package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/SeijiOmi/posts-service/db"
	"github.com/SeijiOmi/posts-service/entity"
	"github.com/jmcvetta/napping"
	"github.com/stretchr/testify/assert"
)

var client = new(http.Client)
var testServer *httptest.Server
var postDefault = entity.Post{Body: "test", Point: 100}
var tmpBaseUserURL string
var tmpBasePointURL string

func TestMain(m *testing.M) {
	setup()
	ret := m.Run()
	teardown()
	os.Exit(ret)
}

func setup() {
	tmpBaseUserURL = os.Getenv("USER_URL")
	tmpBasePointURL = os.Getenv("POINT_URL")
	os.Setenv("USER_URL", "http://post-mock-user:3000")
	os.Setenv("POINT_URL", "http://post-mock-point:3000")
	db.Init()
	router := router()
	testServer = httptest.NewServer(router)
}

func teardown() {
	testServer.Close()
	initPostTable()
	db.Close()
	os.Setenv("USER_URL", tmpBaseUserURL)
	os.Setenv("POINT_URL", tmpBasePointURL)
}

func TestPostCreate(t *testing.T) {
	inputPost := struct {
		Body  string `json:"body"`
		Point uint   `json:"point"`
		Token string `json:"token"`
	}{
		"tests",
		100,
		"tests",
	}
	input, _ := json.Marshal(inputPost)
	resp, _ := http.Post(testServer.URL+"/posts", "application/json", bytes.NewBuffer(input))
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestPostCreateNumericErrValid(t *testing.T) {
	inputPost := struct {
		Body  string `json:"body"`
		Point string `json:"point"`
	}{
		"tests",
		"tests",
	}
	input, _ := json.Marshal(inputPost)
	resp, _ := http.Post(testServer.URL+"/posts", "application/json", bytes.NewBuffer(input))
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestPostCreateMinusErrValid(t *testing.T) {
	inputPost := struct {
		Body  string `json:"body"`
		Point int    `json:"point"`
	}{
		"tests",
		-1,
	}
	input, _ := json.Marshal(inputPost)
	resp, _ := http.Post(testServer.URL+"/posts", "application/json", bytes.NewBuffer(input))
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestPostDone(t *testing.T) {
	initPostTable()
	createDefaultPost(1, 1, 2)

	inputPost := struct {
		ID    int    `json:"id"`
		Token string `json:"token"`
	}{
		1,
		"testToken",
	}
	input, _ := json.Marshal(inputPost)

	resp, _ := http.Post(testServer.URL+"/done", "application/json", bytes.NewBuffer(input))
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestPutDone(t *testing.T) {
	initPostTable()
	createDefaultPost(1, 1, 2)

	inputPost := struct {
		ID    int    `json:"id"`
		Token string `json:"token"`
	}{
		1,
		"testToken",
	}
	input, _ := json.Marshal(inputPost)

	resp, _ := http.Post(testServer.URL+"/done", "application/json", bytes.NewBuffer(input))
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestAmountGetByUserID(t *testing.T) {
	response := struct {
		AmountPayment int
	}{}
	error := struct {
		Error string
	}{}

	initPostTable()
	createDefaultPost(0, 1, 2)

	resp, err := napping.Get(testServer.URL+"/amount/1", nil, &response, &error)
	assert.Equal(t, nil, err)
	assert.Equal(t, http.StatusOK, resp.Status())
	assert.NotEqual(t, 0, response.AmountPayment)
}

func TestGetUserByID(t *testing.T) {
	response := []entity.PostWithUser{}
	error := struct {
		Error string
	}{}

	initPostTable()
	createDefaultPost(0, 1, 3)
	createDefaultPost(0, 1, 3)
	createDefaultPost(0, 2, 3)

	resp, err := napping.Get(testServer.URL+"/user/1", nil, &response, &error)
	assert.Equal(t, nil, err)
	assert.Equal(t, http.StatusOK, resp.Status())
	assert.Equal(t, 2, len(response))
}

func createDefaultPost(id uint, userID uint, helpserUserID uint) entity.Post {
	db := db.GetDB()
	post := postDefault
	post.ID = id
	post.UserID = userID
	post.HelperUserID = helpserUserID
	db.Create(&post)
	return post
}

func initPostTable() {
	db := db.GetDB()
	var u entity.Post
	db.Delete(&u)
}
