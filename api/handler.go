package api

type FailReturn struct {
	Code    int64  `json:"code" example:"-1"`    //	code 1 正常  其他出现错误
	Message string `json:"message" example:"描述"` //	错误描述
}

type SuccessReturn struct {
	FailReturn
	Data string `json:"data"` //	返回json数据, 可为空
}

type PageData struct {
	Page      int64 `json:"page" example:"1"`        //	页码
	PageSize  int64 `json:"pageSize" example:"30"`   //	一页显示记录数
	TotalSize int64 `json:"totalSize" example:"100"` //	记录总数
	TotalPage int64 `json:"totalPage" example:"4"`   //	总页面数量
}
