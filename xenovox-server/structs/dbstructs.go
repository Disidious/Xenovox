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
	SenderId   int    `json:"senderid"`
	ReceiverId int    `json:"receiverid"`
	Read       bool   `json:"read"`
}

type GroupMessage struct {
	Id       int    `json:"id"`
	Message  string `json:"message"`
	SenderId int    `json:"senderid"`
	GroupId  int    `json:"receiverid"`
}

type Relation struct {
	Id       int `json:"id"`
	User1Id  int `json:"user1id"`
	User2Id  int `json:"user2id"`
	Relation int `json:"relation"`
}
