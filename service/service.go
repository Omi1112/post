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

// GetAll 投稿全件を取得
func (b Behavior) GetAll() ([]entity.Post, error) {
	db := db.GetDB()
	var post []entity.Post

	if err := db.Order("id desc").Find(&post).Error; err != nil {
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

// FindByColumn 指定されたカラムで検索を行う。
func (b Behavior) FindByColumn(column string, id string) ([]entity.Post, error) {
	db := db.GetDB()
	var post []entity.Post

	if err := db.Where(column+" = ?", id).Order("id desc").Find(&post).Error; err != nil {
		return post, err
	}

	return post, nil
}

// GetByHelperUserIDWithUserData 投稿情報にユーザ情報を紐づけて取得（ヘルパーユーザーIDで検索）
func (b Behavior) GetByHelperUserIDWithUserData(userID string) ([]entity.PostWithUser, error) {
	posts, err := b.FindByColumn("helper_user_id", userID)
	if err != nil {
		return nil, err
	}

	return attachUserData(posts)
}

// GetByUserIDWithUserData 投稿情報にユーザ情報を紐づけて取得（ユーザーIDで検索）
func (b Behavior) GetByUserIDWithUserData(userID string) ([]entity.PostWithUser, error) {
	posts, err := b.FindByColumn("user_id", userID)
	if err != nil {
		return nil, err
	}

	return attachUserData(posts)
}

// GetByTagIDWithUserData タグＩＤで投稿情報を検索する。（ヘルパーユーザーIDで検索）
func (b Behavior) GetByTagIDWithUserData(tagID string) ([]entity.PostWithUser, error) {
	db := db.GetDB()
	rows, err := db.Table("posts").
		Select("posts.*").
		Joins("inner join post_tags on posts.id = post_tags.post_id").
		Where("post_tags.tag_id = ?", tagID).
		Rows()
	defer rows.Close()
	if err != nil {
		return []entity.PostWithUser{}, err
	}

	var posts []entity.Post
	for rows.Next() {
		var post entity.Post
		db.ScanRows(rows, &post)
		posts = append(posts, post)
	}

	return attachUserData(posts)
}

// CreateModel 投稿情報の生成
func (b Behavior) CreateModel(inputPost entity.JoinPost, token string) (entity.JoinPost, error) {
	userID, err := getUserIDByToken(token)
	if err != nil {
		return entity.JoinPost{}, err
	}
	createPost := inputPost.Post
	createPost.UserID = uint(userID)

	tx := db.StartBegin()

	if err := tx.Create(&createPost).Error; err != nil {
		db.EndRollback()
		return entity.JoinPost{}, err
	}

	for _, inputTag := range inputPost.Tags {
		tag, err := createTagModel(inputTag)
		if err != nil {
			db.EndRollback()
			return entity.JoinPost{}, err
		}

		err = createPostTagModel(createPost.ID, tag.ID)
		if err != nil {
			db.EndRollback()
			return entity.JoinPost{}, err
		}
	}

	db.EndCommit()
	return attachJoinDataSingle(createPost)
}

// SetHelpUserID 投稿情報のHlpUserIDにTokenから取得したユーザＩＤを格納する。
func (b Behavior) SetHelpUserID(id string, token string) (entity.PostWithUser, error) {
	findPost, userID, err := authAndGetPost(id, token)
	if err != nil {
		return entity.PostWithUser{}, err
	}

	findPost.HelperUserID = uint(userID)

	post, err := updatePostExec(&findPost)
	if err != nil {
		return entity.PostWithUser{}, err
	}

	return attachUserDataSingle(post)
}

// TakeHelpUserID 投稿情報のHlpUserIDにTokenから取得したユーザＩＤを格納する。
func (b Behavior) TakeHelpUserID(id string, token string) (entity.PostWithUser, error) {
	findPost, _, err := authAndGetPost(id, token)
	if err != nil {
		return entity.PostWithUser{}, err
	}
	findPost.HelperUserID = 0

	post, err := updatePostExec(&findPost)
	if err != nil {
		return entity.PostWithUser{}, err
	}

	return attachUserDataSingle(post)
}

// DonePayment 投稿情報を元に完了ステータスの登録とポイントの支払をする。
func (b Behavior) DonePayment(id string, token string) (entity.PostWithUser, error) {
	findPost, _, err := authAndGetPost(id, token)
	if err != nil {
		return entity.PostWithUser{}, err
	}

	if findPost.HelperUserID == 0 {
		return entity.PostWithUser{}, errors.New("PostID:" + strconv.Itoa(int(findPost.ID)) + " Don't have helper")
	}

	findPost.Status = entity.Payment

	post, err := updatePostExec(&findPost)
	if err != nil {
		return entity.PostWithUser{}, err
	}

	postWithUser, err := attachUserDataSingle(post)
	if err != nil {
		return entity.PostWithUser{}, err
	}

	// ポイント支払いのため、マイナスポイントを登録する。
	comment := postWithUser.HelperUser.Name + "さんが助けてくれました！"
	createPoint(-int(findPost.Point), comment, token)

	return postWithUser, nil
}

// DoneAcceptance 投稿情報のHlpUserIDにTokenから取得したユーザＩＤを格納する。
func (b Behavior) DoneAcceptance(id string, token string) (entity.PostWithUser, error) {
	findPost, _, err := authAndGetPost(id, token)
	if err != nil {
		return entity.PostWithUser{}, err
	}

	if findPost.Status != entity.Payment {
		return entity.PostWithUser{}, errors.New("PostID:" + strconv.Itoa(int(findPost.ID)) + " Status not payment")
	}

	findPost.Status = entity.Acceptance

	post, err := updatePostExec(&findPost)
	if err != nil {
		return entity.PostWithUser{}, err
	}

	postWithUser, err := attachUserDataSingle(post)
	if err != nil {
		return entity.PostWithUser{}, err
	}

	comment := postWithUser.User.Name + "さんを助けました！"
	createPoint(int(findPost.Point), comment, token)

	return postWithUser, nil
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

func createTagModel(inputTag entity.Tag) (entity.Tag, error) {
	db := db.GetDB()
	createTag := inputTag

	// 既に存在するタグの場合そのデータを返却する。
	tag, _ := getTagByBody(createTag.Body)
	empty := entity.Tag{}
	if tag != empty {
		return tag, nil
	}

	if err := db.Create(&createTag).Error; err != nil {
		return createTag, err
	}

	return createTag, nil
}

// FindTagLikeBody Tag.BodyをLike検索する。
func (b Behavior) FindTagLikeBody(body string) ([]entity.Tag, error) {
	db := db.GetDB()
	var tags []entity.Tag

	if err := db.Where("body LIKE ?", "%"+body+"%").Find(&tags).Error; err != nil {
		return []entity.Tag{}, err
	}

	return tags, nil
}

func getTagByBody(body string) (entity.Tag, error) {
	db := db.GetDB()
	var tag entity.Tag

	if err := db.Where("body = ?", body).First(&tag).Error; err != nil {
		return entity.Tag{}, err
	}

	return tag, nil
}

func createPostTagModel(postID uint, tagID uint) error {
	db := db.GetDB()
	createPostTag := entity.PostTag{
		PostID: postID,
		TagID:  tagID,
	}
	if err := db.Create(&createPostTag).Error; err != nil {
		return err
	}

	return nil
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

func createPoint(point int, comment string, token string) error {
	input := struct {
		Number  int    `json:"number"`
		Comment string `json:"comment"`
		Token   string `json:"token"`
	}{
		point,
		comment,
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

func attachUserDataSingle(post entity.Post) (entity.PostWithUser, error) {
	posts := []entity.Post{post}
	postsWithUser, err := attachUserData(posts)
	if err != nil {
		return entity.PostWithUser{}, err
	}

	return postsWithUser[0], nil
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

func attachJoinDataSingle(post entity.Post) (entity.JoinPost, error) {
	posts := []entity.Post{post}
	postJoinPost, err := attachJoinData(posts)
	if err != nil {
		return entity.JoinPost{}, err
	}

	return postJoinPost[0], nil
}

func attachJoinData(posts []entity.Post) ([]entity.JoinPost, error) {
	users := getUsersData()

	var returnData []entity.JoinPost
	for _, post := range posts {
		if post.UserID == 0 {
			idStr := strconv.Itoa(int(post.ID))
			return []entity.JoinPost{}, errors.New("postID:" + idStr + " don't have userID")
		}

		user, ok := users[int(post.UserID)]
		if !ok {
			user = &entity.User{}
		}

		helperUser, ok := users[int(post.HelperUserID)]
		if !ok {
			helperUser = &entity.User{}
		}

		tags, err := getTagByPostID(post.ID)
		if err != nil {
			return []entity.JoinPost{}, err
		}

		returnData = append(returnData, entity.JoinPost{Post: post, User: *user, HelperUser: *helperUser, Tags: tags})
	}

	return returnData, nil
}

func getTagByPostID(postID uint) ([]entity.Tag, error) {
	db := db.GetDB()
	rows, err := db.
		Table("tags").
		Select("tags.*").
		Joins("inner join post_tags on tags.id = post_tags.tag_id").
		Where("post_tags.post_id = ?", postID).
		Rows()

	if err != nil {
		return []entity.Tag{}, err
	}

	var tags []entity.Tag
	for rows.Next() {
		var tag entity.Tag
		db.ScanRows(rows, &tag)
		tags = append(tags, tag)
	}

	return tags, nil
}
