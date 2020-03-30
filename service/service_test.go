package service

import (
	"net/http"
	"os"
	"testing"

	"github.com/SeijiOmi/posts-service/db"
	"github.com/SeijiOmi/posts-service/entity"
	"github.com/stretchr/testify/assert"
)

var client = new(http.Client)
var postDefault = entity.Post{Body: "test", Point: 100}
var tagDefault = entity.Tag{Body: "test"}
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
	initPostTable()
}
func teardown() {
	initPostTable()
	db.Close()
	os.Setenv("USER_URL", tmpBaseUserURL)
	os.Setenv("POINT_URL", tmpBasePointURL)
}

func TestGetAll(t *testing.T) {
	initPostTable()
	createDefaultPost(0, 1, 0)
	post := createDefaultPost(0, 1, 2)

	var b Behavior
	posts, err := b.GetAll()
	assert.Equal(t, err, nil)
	assert.Equal(t, len(posts), 2)
	// 最後に作成した投稿情報が先頭であることを確認
	assert.Equal(t, post.ID, posts[0].ID)
}

func TestFindByColumn(t *testing.T) {
	initPostTable()
	createDefaultPost(0, 1, 1)
	post := createDefaultPost(0, 1, 2)

	var b Behavior
	posts, err := b.FindByColumn("user_id", "1")
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(posts))
	// 最後に作成した投稿情報が先頭であることを確認
	assert.Equal(t, post.ID, posts[0].ID)
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

func TestAttachJoinData(t *testing.T) {
	initTable()
	post := createDefaultPost(0, 1, 2)
	tag := createDefaultTag()
	createTestPostTag(post.ID, tag.ID)
	tag = createDefaultTag()
	createTestPostTag(post.ID, tag.ID)

	posts := []entity.Post{post}

	postsJoinData, err := attachJoinData(posts)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, "", postsJoinData[0].User.Name)
	assert.NotEqual(t, "", postsJoinData[0].HelperUser.Name)
	assert.NotEqual(t, "", postsJoinData[0].Tags[0].Body)
	assert.NotEqual(t, "", postsJoinData[0].Tags[1].Body)
}

func TestGetTagByPostID(t *testing.T) {
	initTable()
	post := createDefaultPost(0, 1, 2)
	tag := createDefaultTag()
	createTestPostTag(post.ID, tag.ID)
	tag = createDefaultTag()
	createTestPostTag(post.ID, tag.ID)

	tags, err := getTagByPostID(post.ID)

	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(tags))
}

func TestGetTagByPostIDNoData(t *testing.T) {
	initTable()

	_, err := getTagByPostID(1)

	assert.Equal(t, nil, err)
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

func TestCreateModel(t *testing.T) {
	initTable()

	var b Behavior
	createPost := entity.JoinPost{
		Post: postDefault,
		Tags: []entity.Tag{
			entity.Tag{ID: 0, Body: "TEST2"},
			entity.Tag{ID: 0, Body: "TEST3"},
			entity.Tag{ID: 0, Body: "TEST1"},
			entity.Tag{ID: 0, Body: "TEST1"},
		},
	}
	post, err := b.CreateModel(createPost, "testToken")

	assert.Equal(t, nil, err)
	assert.Equal(t, postDefault.Body, post.Post.Body)
	assert.Equal(t, postDefault.Point, post.Post.Point)
	assert.NotEqual(t, uint(0), post.Post.UserID)
}

func TestCreateTagModel(t *testing.T) {
	initTable()
	tag := entity.Tag{ID: 0, Body: "TEST1"}
	tagFirst, errFirst := createTagModel(tag)
	tagSecond, errSecond := createTagModel(tag)
	assert.Equal(t, nil, errFirst)
	assert.Equal(t, nil, errSecond)
	assert.Equal(t, tagFirst, tagSecond)
}

func TestDone(t *testing.T) {
	initPostTable()
	createDefaultPost(1, 2, 1)
	var b Behavior
	post, err := b.DonePayment("1", "testToken")

	assert.Equal(t, nil, err)
	assert.Equal(t, entity.Payment, post.Post.Status)

	post, err = b.DoneAcceptance("1", "testToken")
	assert.Equal(t, nil, err)
	assert.Equal(t, entity.Acceptance, post.Post.Status)
}

func TestDoneAcceptanceErr(t *testing.T) {
	initPostTable()
	createDefaultPost(1, 2, 1)
	var b Behavior
	post, err := b.DoneAcceptance("1", "testToken")

	assert.NotEqual(t, nil, err)
	assert.NotEqual(t, entity.Payment, post.Post.Status)
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

func TestGetByTagIDWithUserData(t *testing.T) {
	post := createDefaultPost(0, 1, 2)
	tag := createDefaultTag()
	createPostTagModel(post.ID, tag.ID)
	tag = createDefaultTag()
	createPostTagModel(post.ID, tag.ID)

	var b Behavior
	posts, err := b.GetByTagIDWithUserData("1")
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(posts))
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

func createDefaultTag() entity.Tag {
	db := db.GetDB()
	tag := tagDefault
	db.Create(&tag)
	return tag
}

func createTestPostTag(postID uint, tagID uint) entity.PostTag {
	db := db.GetDB()
	createPostTag := entity.PostTag{
		PostID: postID,
		TagID: tagID,
	}
	db.Create(&createPostTag)
	return createPostTag
}

func initTable() {
	initPostTable()
	initTagTable()
	initPostTagsTable()
}

func initPostTable() {
	db := db.GetDB()
	var u entity.Post
	db.Delete(&u)
}
func initTagTable() {
	db := db.GetDB()
	var t entity.Tag
	db.Delete(&t)
}
func initPostTagsTable() {
	db := db.GetDB()
	var pt entity.PostTag
	db.Delete(&pt)
}
