package model

type DeptQuery struct {
	Keywords string `form:"keywords"`
	Status   *int   `form:"status"`
}
