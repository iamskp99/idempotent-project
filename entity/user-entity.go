package userEntity

type User struct {
	Id      uint64 `json:"id"`
	Name    string `json:"name"`
	Balance int32  `json:"balance"`
}
