package service

import (
	"net/http"
	"os"
	"testing"

	"github.com/SeijiOmi/posts-service/db"
	"github.com/SeijiOmi/posts-service/entity"
	"github.com/stretchr/testify/assert"
)

/*
	テストの前準備
*/

var client = new(http.Client)
var postDefault = entity.Post{Body: "test", Point: 100}
var tmpBaseUserURL string
var tmpBasePointURL string

// テストを統括するテスト時には、これが実行されるイメージでいる。
func TestMain(m *testing.M) {
	// テスト実施前の共通処理（自作関数）
	setup()
	ret := m.Run()
	// テスト実施後の共通処理（自作関数）
	teardown()
	os.Exit(ret)
}

// テスト実施前共通処理
func setup() {
	tmpBaseUserURL = os.Getenv("USER_URL")
	tmpBasePointURL = os.Getenv("POINT_URL")
	os.Setenv("USER_URL", "http://post-mock-user:3000")
	os.Setenv("POINT_URL", "http://post-mock-point:3000")
	db.Init()
	initPostTable()
}

// テスト実施後共通処理
func teardown() {
	initPostTable()
	db.Close()
	os.Setenv("USER_URL", tmpBaseUserURL)
	os.Setenv("POINT_URL", tmpBasePointURL)
}

/*
	ここからが個別のテスト実装
*/

func TestGetAll(t *testing.T) {
	initPostTable()
	createDefaultPost(0, 1, 0)
	createDefaultPost(0, 1, 0)

	var b Behavior
	posts, err := b.GetAll()
	assert.Equal(t, err, nil)
	assert.Equal(t, len(posts), 2)
}

func TestFindByColumn(t *testing.T) {
	initPostTable()
	createDefaultPost(0, 1, 1)
	createDefaultPost(0, 1, 2)

	var b Behavior
	posts, err := b.FindByColumn("user_id", "1")
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(posts))
}

func TestAttachUserData(t *testing.T) {
	initPostTable()
	post := createDefaultPost(0, 1, 0)
	var posts []entity.Post
	posts = append(posts, post)

	postsWithUser, err := attachUserData(posts)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, "", postsWithUser[0].User.Name)
}

func TestGetByHelperUserIDWithUserData(t *testing.T) {
	initPostTable()
	createDefaultPost(0, 1, 1)
	createDefaultPost(0, 1, 2)

	var b Behavior
	postsWithUser, err := b.GetByHelperUserIDWithUserData("1")
	assert.Equal(t, nil, err)
	assert.Equal(t, 1, len(postsWithUser))
	assert.NotEqual(t, "", postsWithUser[0].User.Name)
	assert.NotEqual(t, "", postsWithUser[0].HelperUser.Name)
}

func TestGetUserIDWithUserData(t *testing.T) {
	initPostTable()
	createDefaultPost(0, 1, 1)
	createDefaultPost(0, 2, 1)

	var b Behavior
	postsWithUser, err := b.GetByUserIDWithUserData("1")
	assert.Equal(t, nil, err)
	assert.Equal(t, 1, len(postsWithUser))
	assert.NotEqual(t, "", postsWithUser[0].User.Name)
	assert.NotEqual(t, "", postsWithUser[0].HelperUser.Name)
}

func TestValidCreateModel(t *testing.T) {
	var b Behavior
	post, err := b.CreateModel(postDefault, "testToken")

	assert.Equal(t, nil, err)
	assert.Equal(t, postDefault.Body, post.Body)
	assert.Equal(t, postDefault.Point, post.Point)
	assert.NotEqual(t, uint(0), post.UserID)
}

func TestDone(t *testing.T) {
	initPostTable()
	createDefaultPost(1, 2, 1)
	var b Behavior
	post, err := b.DonePayment("1", "testToken")

	assert.Equal(t, nil, err)
	assert.Equal(t, entity.Payment, post.Status)

	post, err = b.DoneAcceptance("1", "testToken")
	assert.Equal(t, nil, err)
	assert.Equal(t, entity.Acceptance, post.Status)
}

func TestDoneAcceptanceErr(t *testing.T) {
	initPostTable()
	createDefaultPost(1, 2, 1)
	var b Behavior
	post, err := b.DoneAcceptance("1", "testToken")

	assert.NotEqual(t, nil, err)
	assert.NotEqual(t, entity.Payment, post.Status)
}

func TestGetAmountPaymentByUserID(t *testing.T) {
	initPostTable()
	createDefaultPost(0, 1, 2)
	var b Behavior
	amountPayment, err := b.GetAmountPaymentByUserID("1")

	assert.Equal(t, nil, err)
	assert.NotEqual(t, 0, amountPayment)
}

func TestGetScheduledPaymentPointByUserID(t *testing.T) {
	initPostTable()
	createDefaultPost(0, 1, 2)
	payment, err := getScheduledPaymentPointByUserID("1")

	assert.Equal(t, nil, err)
	assert.Equal(t, int(postDefault.Point), payment)
}

func TestGetPointByUserID(t *testing.T) {
	total, err := getPointByUserID("1")

	assert.Equal(t, nil, err)
	assert.NotEqual(t, 0, total)
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
