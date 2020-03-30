package entity

// PostTag 投稿情報とタグを紐づけ
type PostTag struct {
	PostID uint `json:"postId"`
	TagID  uint `json:"tagId"`
}
