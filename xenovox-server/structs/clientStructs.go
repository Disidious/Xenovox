package structs

type ClientUser struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
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
	Message    string `json:"message"`
	SenderId   int    `json:"senderId"`
	ReceiverId int    `json:"receiverId"`
}

type ClientChatHistory struct {
	FriendId int             `json:"friendId"`
	History  []ClientMessage `json:"history"`
}

type ClientFriendRequest struct {
	RelationId int    `json:"relationId"`
	Username   string `json:"username"`
	UserId     int    `json:"userId"`
}

type ClientSocketMessage struct {
	Type string      `json:"type"`
	Body interface{} `json:"body"`
}
