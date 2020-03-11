package entity

// Post オブジェクト構造
type Post struct {
	ID           uint   `json:"id"`
	UserID       uint   `json:"userId"`
	HelperUserID uint   `json:"helpserUserId"`
	Body         string `json:"body"`
}
