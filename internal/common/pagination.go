package common

type PageQuery struct {
	Page  int    `form:"page,default=1" binding:"min=1"`
	Limit int    `form:"limit,default=20" binding:"min=1,max=200"`
	Sort  string `form:"sort,default=created_at.desc"`
}

type PageMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalItems int `json:"total_items,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}
