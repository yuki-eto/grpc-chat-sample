package entity

type Player struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Token       string `json:"-"`
	AccessToken string `json:"-"`
}
