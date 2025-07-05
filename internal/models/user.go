package models

type User struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	UserNum  int    `json:"user_num"`
}
