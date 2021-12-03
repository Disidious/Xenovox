package structs

type ClientUser struct {
	Id       int    `json:"id"`
	Username string `json:"username`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Picture  string `json:"picture"`
}

type ClientFriend struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Picture  string `json:"picture"`
}

type ClientMessage struct {
	Id           int    `json:"id"`
	Message      string `json:"message"`
	SenderId     int    `json:"senderId"`
	ReceiverId   int    `json:"receiverId"`
	GroupMessage bool   `json:"groupMessage"`
}
