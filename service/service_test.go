package service

import (
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/SeijiOmi/posts-service/db"
	"github.com/SeijiOmi/posts-service/entity"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

/*
	テストの前準備
*/

var client = new(http.Client)

var postDefault = entity.Post{Name: "test", Email: "test@co.jp", Password: "password"}

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
	db.Init()
	initPostTable()
}

// テスト実施後共通処理
func teardown() {
	db.Close()
}

/*
	ここからが個別のテスト実装
*/

func TestGetAll(t *testing.T) {
	initPostTable()
	createDefaultPost()
	createDefaultPost()

	var b Behavior
	posts, err := b.GetAll()
	assert.Equal(t, err, nil)
	assert.Equal(t, len(posts), 2)
}

func TestCreateModel(t *testing.T) {
	var b Behavior
	post, err := b.CreateModel(postDefault)

	assert.Equal(t, nil, err)
	assert.Equal(t, postDefault.Name, post.Name)
	assert.Equal(t, postDefault.Email, post.Email)
	assert.NotEqual(t, postDefault.Password, post.Password)
	err = bcrypt.CompareHashAndPassword([]byte(post.Password), []byte(postDefault.Password))
	assert.Equal(t, nil, err)
}

func TestGetByIDExists(t *testing.T) {
	post := createDefaultPost()
	var b Behavior
	post, err := b.GetByID(strconv.Itoa(int(post.ID)))

	assert.Equal(t, nil, err)
	assert.Equal(t, postDefault.Name, post.Name)
	assert.Equal(t, postDefault.Email, post.Email)
}

func TestGetByIDNotExists(t *testing.T) {
	var b Behavior
	post, err := b.GetByID(string(postDefault.ID))

	assert.NotEqual(t, nil, err)
	var nilPost entity.Post
	assert.Equal(t, nilPost, post)
}

func TestUpdateByIDExists(t *testing.T) {
	post := createDefaultPost()

	updatePost := entity.Post{Name: "not", Email: "not@co.jp", Password: "notpassword"}

	var b Behavior
	post, err := b.UpdateByID(strconv.Itoa(int(post.ID)), updatePost)

	assert.Equal(t, nil, err)
	assert.Equal(t, updatePost.Name, post.Name)
	assert.Equal(t, updatePost.Email, post.Email)
	assert.NotEqual(t, updatePost.Password, post.Password)
	err = bcrypt.CompareHashAndPassword([]byte(post.Password), []byte(updatePost.Password))
	assert.Equal(t, nil, err)
}

func TestUpdateByIDNotExists(t *testing.T) {
	post := createDefaultPost()

	updatePost := entity.Post{Name: "not", Email: "not@co.jp", Password: "notpassword"}

	var b Behavior
	post, err := b.UpdateByID("0", updatePost)

	assert.NotEqual(t, nil, err)
	var nilPost entity.Post
	assert.Equal(t, nilPost, post)
}

func TestDeleteByIDExists(t *testing.T) {
	post := createDefaultPost()

	db := db.GetDB()
	var beforeCount int
	db.Table("posts").Count(&beforeCount)

	var b Behavior
	err := b.DeleteByID(strconv.Itoa(int(post.ID)))

	var afterCount int
	db.Table("posts").Count(&afterCount)

	assert.Equal(t, nil, err)
	assert.Equal(t, beforeCount-1, afterCount)
}

func TestDeleteByIDNotExists(t *testing.T) {
	initPostTable()
	createDefaultPost()

	db := db.GetDB()
	var beforeCount int
	db.Table("posts").Count(&beforeCount)

	var b Behavior
	err := b.DeleteByID("0")

	var afterCount int
	db.Table("posts").Count(&afterCount)

	assert.Equal(t, nil, err)
	assert.Equal(t, beforeCount, afterCount)
}

func createDefaultPost() entity.Post {
	db := db.GetDB()
	post := postDefault
	db.Create(&post)
	return post
}

func initPostTable() {
	db := db.GetDB()
	var u entity.Post
	db.Delete(&u)
}
