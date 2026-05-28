package dto

// PoiDTO POI
type PoiDTO struct {
	ID         int64   `json:"id"`
	PoiName    *string `json:"poiName"`
	CreateTime int64   `json:"createTime"`
	UpdateTime int64   `json:"updateTime"`
}
