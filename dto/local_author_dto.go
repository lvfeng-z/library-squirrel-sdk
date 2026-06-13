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

// RankedLocalAuthor 带排序的本地作者
type RankedLocalAuthor struct {
	Author    LocalAuthorDTO `json:"author"`
	RoleName  string         `json:"roleName"`
	SortOrder int            `json:"sortOrder"`
}

// RankedLocalAuthorWithWorkId 带作品ID的本地作者
type RankedLocalAuthorWithWorkId struct {
	Author    LocalAuthorDTO `json:"author"`
	RoleName  string         `json:"roleName"`
	SortOrder int            `json:"sortOrder"`
	WorkId    int64          `json:"workId"`
}
