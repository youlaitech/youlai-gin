package common

// Option 下拉选项（泛型）
type Option[T any] struct {
	Value T      `json:"value"`
	Label string `json:"label"`
	Children []Option[T] `json:"children,omitempty"`
}
