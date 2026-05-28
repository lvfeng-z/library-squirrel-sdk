package dto

// SearchType 搜索类型
type SearchType int

const (
	SearchTypeLocalTag        SearchType = 1
	SearchTypeSiteTag         SearchType = 2
	SearchTypeLocalAuthor     SearchType = 3
	SearchTypeSiteAuthor      SearchType = 4
	SearchTypeWorksSiteName   SearchType = 5
	SearchTypeWorksNickname   SearchType = 6
	SearchTypeWorksUploadTime SearchType = 7
	SearchTypeWorksLastView   SearchType = 8
	SearchTypeMediaType       SearchType = 9
	SearchTypeSite            SearchType = 10
	SearchTypeWorkSet         SearchType = 11
)

// SearchCondition 搜索条件
type SearchCondition struct {
	Type     SearchType  `json:"type"`
	Value    interface{} `json:"value"`
	Operator string      `json:"operator,omitempty"`
}

// 操作符常量
const (
	OperatorEqual          = "="
	OperatorNotEqual       = "!="
	OperatorGreaterThan    = ">"
	OperatorLessThan       = "<"
	OperatorGreaterOrEqual = ">="
	OperatorLessOrEqual    = "<="
	OperatorLike           = "LIKE"
)

// MediaType 媒体类型
type MediaType int

const (
	MediaTypePicture  MediaType = 1
	MediaTypeVideo    MediaType = 2
	MediaTypeDocument MediaType = 3
	MediaTypeAudio    MediaType = 4
)

// MediaExtMapping 媒体类型对应的扩展名映射
var MediaExtMapping = map[MediaType][]string{
	MediaTypePicture:  {".jpg", ".png", ".jpeg", ".gif"},
	MediaTypeVideo:    {".mp4", ".avi", ".mkv"},
	MediaTypeDocument: {".pdf", ".docx", ".doc", ".xlsx", ".txt"},
	MediaTypeAudio:    {".mp3", ".wav", ".aac"},
}

// SearchConditionQuery 搜索条件分页查询请求
type SearchConditionQuery struct {
	Types   []SearchType `json:"types,omitempty"`
	Keyword string       `json:"keyword,omitempty"`
}

// NewSearchCondition 创建搜索条件
func NewSearchCondition(searchType SearchType, value interface{}) *SearchCondition {
	return &SearchCondition{
		Type:  searchType,
		Value: value,
	}
}
