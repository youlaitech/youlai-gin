package model

type MenuQuery struct {
	Keywords string `form:"keywords"`
	Status   *int   `form:"status"`
}
