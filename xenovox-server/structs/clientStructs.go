package structs

type ClientUser struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Picture  string `json:"picture"`
}

type ClientDM struct {
	Id         int    `json:"id"`
	Message    string `json:"message"`
	SenderId   int    `json:"senderid"`
	Username   string `json:"username"`
	Picture    string `json:"string"`
	ReceiverId int    `json:"receiverid"`
}

type ClientGM struct {
	Id       int    `json:"id"`
	Message  string `json:"message"`
	SenderId int    `json:"senderid"`
	Username string `json:"username"`
	Picture  string `json:"string"`
	GroupId  int    `json:"groupid"`
	IsSystem bool   `json:"issystem"`
}

type ClientGroup struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	OwnerId int    `json:"ownerid"`
	Picture string `json:"picture"`
}

type ClientGroupMember struct {
	Id       int    `json:"id"`
	UserId   int    `json:"userid"`
	Username string `json:"username"`
	Picture  string `json:"picture"`
}

type ClientChatHistory struct {
	Group   bool        `json:"group"`
	ChatId  int         `json:"chatid"`
	History interface{} `json:"history"`
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
