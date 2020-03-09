package service

import (
	"github.com/SeijiOmi/posts-service/db"
	"github.com/SeijiOmi/posts-service/entity"
)

// Behavior 投稿サービスを提供するメソッド群
type Behavior struct{}

// GetAll 投稿全件を取得
func (b Behavior) GetAll() ([]entity.Post, error) {
	db := db.GetDB()
	var u []entity.Post

	if err := db.Find(&u).Error; err != nil {
		return nil, err
	}

	return u, nil
}

// CreateModel 投稿情報の生成
func (b Behavior) CreateModel(inputPost entity.Post) (entity.Post, error) {
	createPost := inputPost
	db := db.GetDB()

	if err := db.Create(&createPost).Error; err != nil {
		return createPost, err
	}

	return createPost, nil
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
	db := db.GetDB()
	var findPost entity.Post
	if err := db.Where("id = ?", id).First(&findPost).Error; err != nil {
		return findPost, err
	}

	updatePost := inputPost
	updatePost.ID = findPost.ID
	db.Save(&updatePost)

	return updatePost, nil
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
