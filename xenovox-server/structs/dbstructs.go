package structs

type User struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Picture  string `json:"picture"`
	Token    string `json:"token"`
}

type Message struct {
	Id         int    `json:"id"`
	Message    string `json:"message"`
	SenderId   int    `json:"sendId"`
	ReceiverId int    `json:"receiverId"`
}
