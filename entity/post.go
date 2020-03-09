package entity

// Post オブジェクト構造
type Post struct {
	ID     uint   `json:"id"`
	UserID uint   `json:"userId"`
	Body   string `json:"body"`
}
