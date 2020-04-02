package entity

// JoinPost 投稿情報に付随情報がついた状態のデータ。
type JoinPost struct {
	Post       Post  `json:"post"`
	Tags       []Tag `json:"tags"`
	User       User  `json:"user"`
	HelperUser User  `json:"helperUser"`
}
