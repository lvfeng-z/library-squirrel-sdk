package dto

// LocalAuthorDTO 本地作者
type LocalAuthorDTO struct {
	ID         int64   `json:"id"`
	AuthorName *string `json:"authorName"`
	Introduce  *string `json:"introduce"`
	LastUse    *int64  `json:"lastUse"`
	CreateTime int64   `json:"createTime"`
	UpdateTime int64   `json:"updateTime"`
}
