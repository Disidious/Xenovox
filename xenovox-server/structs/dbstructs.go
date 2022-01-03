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
	SenderId   int    `json:"senderId"`
	ReceiverId int    `json:"receiverId"`
}

type Relation struct {
	Id       int `json:"id"`
	User1Id  int `json:"user1id"`
	User2Id  int `json:"user2id"`
	Relation int `json:"relation"`
}
