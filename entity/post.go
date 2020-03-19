package entity

// Status 投稿情報の状態を示す。
type Status int

const (
	// None 投稿完了
	None Status = iota
	// Payment 投稿者支払完了
	Payment
	// Acceptance ヘルパー受け取り完了
	Acceptance
)

// Post オブジェクト構造
type Post struct {
	ID           uint   `json:"id"`
	UserID       uint   `json:"userId"`
	HelperUserID uint   `json:"helperUserId"`
	Body         string `json:"body"`
	Point        uint   `json:"point" binding:"numeric,min=0"`
	Status       Status `json:"status"`
}
