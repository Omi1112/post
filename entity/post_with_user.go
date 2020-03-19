package entity

// PostWithUser 投稿データにユーザ情報を併せた構造体
type PostWithUser struct {
	Post       Post `json:"post"`
	User       User `json:"user"`
	HelperUser User `json:"helperUser"`
}
