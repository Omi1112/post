package service

import (
	"errors"
	"net/http"
	"os"
	"strconv"

	"github.com/SeijiOmi/posts-service/db"
	"github.com/SeijiOmi/posts-service/entity"
	"github.com/jmcvetta/napping"
)

// Behavior 投稿サービスを提供するメソッド群
type Behavior struct{}

// User オブジェクト構造
type User struct {
	id   int
	name string
}

// GetAll 投稿全件を取得
func (b Behavior) GetAll() ([]entity.Post, error) {
	db := db.GetDB()
	var post []entity.Post

	if err := db.Find(&post).Error; err != nil {
		return nil, err
	}

	return post, nil
}

// GetAllWithUserData 投稿情報にユーザ情報を紐づけて取得
func (b Behavior) GetAllWithUserData() ([]entity.PostWithUser, error) {
	posts, err := b.GetAll()
	if err != nil {
		return nil, err
	}

	return attachUserData(posts)
}

// GetByHelperUserID 投稿情報にユーザ情報を紐づけて取得
func (b Behavior) GetByHelperUserID(userID string) ([]entity.Post, error) {
	db := db.GetDB()
	var post []entity.Post

	if err := db.Where("helper_user_id = ?", userID).Find(&post).Error; err != nil {
		return post, err
	}

	return post, nil
}

// GetByHelperUserIDWithUserData 投稿情報にユーザ情報を紐づけて取得
func (b Behavior) GetByHelperUserIDWithUserData(userID string) ([]entity.PostWithUser, error) {
	posts, err := b.GetByHelperUserID(userID)
	if err != nil {
		return nil, err
	}

	return attachUserData(posts)
}

// CreateModel 投稿情報の生成
func (b Behavior) CreateModel(inputPost entity.Post, token string) (entity.Post, error) {
	userID, err := getUserIDByToken(token)
	if err != nil {
		return inputPost, err
	}

	createPost := inputPost
	createPost.UserID = uint(userID)
	db := db.GetDB()

	if err := db.Create(&createPost).Error; err != nil {
		return createPost, err
	}

	return createPost, nil
}

// SetHelpUserID 投稿情報のHlpUserIDにTokenから取得したユーザＩＤを格納する。
func (b Behavior) SetHelpUserID(id string, token string) (entity.Post, error) {
	findPost, userID, err := authAndGetPost(id, token)
	if err != nil {
		return entity.Post{}, err
	}

	findPost.HelperUserID = uint(userID)

	return updatePostExec(&findPost)
}

// TakeHelpUserID 投稿情報のHlpUserIDにTokenから取得したユーザＩＤを格納する。
func (b Behavior) TakeHelpUserID(id string, token string) (entity.Post, error) {
	findPost, _, err := authAndGetPost(id, token)
	if err != nil {
		return entity.Post{}, err
	}
	findPost.HelperUserID = 0

	return updatePostExec(&findPost)
}

// DonePayment 投稿情報を元に完了ステータスの登録とポイントの支払をする。
func (b Behavior) DonePayment(id string, token string) (entity.Post, error) {
	findPost, _, err := authAndGetPost(id, token)
	if err != nil {
		return entity.Post{}, err
	}

	if findPost.HelperUserID == 0 {
		return entity.Post{}, errors.New("PostID:" + strconv.Itoa(int(findPost.ID)) + " Don't have helper")
	}

	// ポイント支払いのため、マイナスポイントを登録する。
	createPoint(-int(findPost.Point), token)
	findPost.Status = entity.Payment

	return updatePostExec(&findPost)
}

// DoneAcceptance 投稿情報のHlpUserIDにTokenから取得したユーザＩＤを格納する。
func (b Behavior) DoneAcceptance(id string, token string) (entity.Post, error) {
	findPost, _, err := authAndGetPost(id, token)
	if err != nil {
		return entity.Post{}, err
	}

	if findPost.Status != entity.Payment {
		return entity.Post{}, errors.New("PostID:" + strconv.Itoa(int(findPost.ID)) + " Status not payment")
	}

	createPoint(int(findPost.Point), token)
	findPost.Status = entity.Acceptance

	return updatePostExec(&findPost)
}

// GetByID IDを元に投稿1件を取得
func (b Behavior) GetByID(id string) (entity.Post, error) {
	db := db.GetDB()
	var u entity.Post

	if err := db.Where("id = ?", id).First(&u).Error; err != nil {
		return u, err
	}

	return u, nil
}

// UpdateByID 指定されたidをinputPost通りに更新
func (b Behavior) UpdateByID(id string, inputPost entity.Post) (entity.Post, error) {
	findPost, err := b.GetByID(id)
	if err != nil {
		return findPost, err
	}

	updatePostData := inputPost
	updatePostData.ID = findPost.ID

	return updatePostExec(&updatePostData)
}

// DeleteByID 指定されたidを削除
func (b Behavior) DeleteByID(id string) error {
	db := db.GetDB()
	var u entity.Post

	if err := db.Where("id = ?", id).Delete(&u).Error; err != nil {
		return err
	}

	return nil
}

// GetAmountPaymentByUserID 現在の支払い可能ポイントを取得する。
func (b Behavior) GetAmountPaymentByUserID(id string) (int, error) {
	havePoint, err := getPointByUserID(id)
	if err != nil {
		return 0, err
	}

	paymentPoint, err := getScheduledPaymentPointByUserID(id)
	if err != nil {
		return 0, err
	}

	amountPayment := havePoint - paymentPoint
	return amountPayment, nil
}

// getScheduledPaymentPointByUserID 投稿情報ステータスがNoneのポイントの集計を行う。
func getScheduledPaymentPointByUserID(id string) (int, error) {
	db := db.GetDB()

	rows, err := db.Table("posts").
		Select("sum(point) as point").
		Where("user_id = ?", id).
		Where("status = ?", entity.None).
		Group("user_id").Rows()
	defer rows.Close()

	if err != nil {
		return 0, err
	}

	var point int
	for rows.Next() {
		rows.Scan(&point)
	}

	return point, nil
}

func getUsersData() map[int]*entity.User {
	var response []map[string]interface{}
	error := struct {
		Error string
	}{}
	baseURL := os.Getenv("USER_URL")
	_, err := napping.Get(baseURL+"/users", nil, &response, &error)
	if err != nil {
		panic(err)
	}

	users := map[int]*entity.User{}
	for _, val := range response {
		floatID, ok := val["id"].(float64)
		if !ok {
			panic("User id no exist")
		}
		id := int(floatID)
		name, _ := val["name"].(string)
		users[id] = &entity.User{ID: id, Name: name}
	}
	return users
}

func authAndGetPost(id string, token string) (entity.Post, int, error) {
	userID, err := getUserIDByToken(token)
	if err != nil {
		return entity.Post{}, 0, err
	}

	var b Behavior
	post, err := b.GetByID(id)
	return post, userID, err
}

func updatePostExec(post *entity.Post) (entity.Post, error) {
	db := db.GetDB()
	db.Save(post)

	return *post, nil
}

func getUserIDByToken(token string) (int, error) {
	response := struct {
		ID int
	}{}
	error := struct {
		Error string
	}{}

	baseURL := os.Getenv("USER_URL")
	resp, err := napping.Get(baseURL+"/auth/"+token, nil, &response, &error)

	if err != nil {
		return 0, err
	}

	if resp.Status() == http.StatusBadRequest {
		return 0, errors.New("token invalid")
	}

	return response.ID, nil
}

func createPoint(point int, token string) error {
	input := struct {
		Number int    `json:"number"`
		Token  string `json:"token"`
	}{
		point,
		token,
	}
	error := struct {
		Error string
	}{}

	baseURL := os.Getenv("POINT_URL")
	resp, err := napping.Post(baseURL+"/points", &input, nil, &error)

	if err != nil {
		return err
	}

	if resp.Status() == http.StatusBadRequest {
		return errors.New("token invalid")
	}

	return nil
}

func getPointByUserID(id string) (int, error) {
	response := struct {
		Total int `json:"total"`
	}{}
	error := struct {
		Error string
	}{}

	baseURL := os.Getenv("POINT_URL")
	resp, err := napping.Get(baseURL+"/sum/"+id, nil, &response, &error)

	if err != nil {
		return 0, err
	}

	if resp.Status() == http.StatusBadRequest {
		return 0, errors.New("token invalid")
	}

	return response.Total, nil
}

func attachUserData(posts []entity.Post) ([]entity.PostWithUser, error) {
	users := getUsersData()

	var returnData []entity.PostWithUser
	for _, post := range posts {
		if post.UserID == 0 {
			idStr := strconv.Itoa(int(post.ID))
			return []entity.PostWithUser{}, errors.New("postID:" + idStr + " don't have userID")
		}

		user, ok := users[int(post.UserID)]
		if !ok {
			user = &entity.User{}
		}

		helperUser, ok := users[int(post.HelperUserID)]
		if !ok {
			helperUser = &entity.User{}
		}

		returnData = append(returnData, entity.PostWithUser{Post: post, User: *user, HelperUser: *helperUser})
	}

	return returnData, nil
}
