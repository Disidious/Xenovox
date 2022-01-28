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
	SenderId   int    `json:"senderid"`
	ReceiverId int    `json:"receiverid"`
}

type ClientChatHistory struct {
	FriendId int         `json:"friendid"`
	History  interface{} `json:"history"`
}

type ClientFriendRequest struct {
	RelationId int    `json:"relationid"`
	Username   string `json:"username"`
	UserId     int    `json:"userid"`
}

type ClientNotifications struct {
	SenderIds    []int `json:"senderids"`
	SenderScores []int `json:"senderscores"`
	GroupIds     []int `json:"groupids"`
	GroupScores  []int `json:"groupscores"`
	FriendReq    bool  `json:"friendreq"`
}

type ClientSocketMessage struct {
	Type string      `json:"type"`
	Body interface{} `json:"body"`
}
