package main

import (
	"encoding/json"
	"log"
	"reflect"

	"github.com/gorilla/websocket"

	dbhandler "github.com/Disidious/Xenovox/DbHandler"
	structs "github.com/Disidious/Xenovox/Structs"
)

func getPrivateChat(friendId *int, mt *int, id *int, c *websocket.Conn) bool {
	rows, status := dbhandler.GetPrivateChat(id, friendId)
	if !status {
		return false
	}

	history, ok := structs.StructifyRows(rows, reflect.TypeOf(structs.ClientDM{}))

	if !ok {
		log.Println("Failed: could not convert rows to struct")
		return false
	}

	body := make(map[string]interface{})
	body["history"] = structs.ClientChatHistory{
		Group:   false,
		History: history,
		ChatId:  *friendId,
	}

	response := structs.ClientSocketMessage{
		Type: "CHAT_HISTORY_RES",
		Body: body,
	}
	jsonRes, _ := json.Marshal(response)

	c.WriteMessage(*mt, jsonRes)

	return true
}

func getGroupChatAndMembers(groupId *int, mt *int, id *int, c *websocket.Conn) bool {
	hrows, mrows, status := dbhandler.GetGroupChatAndMembers(id, groupId)
	if !status {
		return false
	}

	history, ok := structs.StructifyRows(hrows, reflect.TypeOf(structs.ClientGM{}))
	if !ok {
		log.Println("Failed: could not convert hrows to struct")
		return false
	}

	members, ok := structs.StructifyRows(mrows, reflect.TypeOf(structs.ClientUser{}))
	if !ok {
		log.Println("Failed: could not convert mrows to struct")
		return false
	}

	body := make(map[string]interface{})
	body["history"] = structs.ClientChatHistory{
		Group:   true,
		History: history,
		ChatId:  *groupId,
	}
	body["members"] = members

	response := structs.ClientSocketMessage{
		Type: "CHAT_HISTORY_RES",
		Body: body,
	}
	jsonRes, _ := json.Marshal(response)

	c.WriteMessage(*mt, jsonRes)

	return true
}

func sendDM(message *structs.Message, mt *int, id *int, c *websocket.Conn) bool {
	// Insert message to database
	ret := dbhandler.InsertDirectMessage(message)
	if ret == "FAILED" {
		log.Println(ret)
		return false
	} else if ret == "BLOCKED" {
		log.Println(ret)
		return false
	}
	response := structs.ClientSocketMessage{
		Type: "DM",
		Body: message.Convert(),
	}
	jsonRes, _ := json.Marshal(response)

	// Get socket of the receiver of the message
	if receiverSocket, ok := sockets[message.ReceiverId]; ok {
		// Send message to receiver
		receiverSocket.WriteMessage(*mt, jsonRes)
	} else {
		log.Println("Receiver Offline")
	}

	// Send message to sender
	err := c.WriteMessage(*mt, jsonRes)
	if err != nil {
		log.Println("err:", err)
	}

	return true
}

func sendGM(message *structs.GroupMessage, mt *int, id *int, c *websocket.Conn) bool {
	// Insert message to database
	ret := dbhandler.InsertGroupMessage(message)
	if ret == "FAILED" {
		log.Println(ret)
		return false
	}
	response := structs.ClientSocketMessage{
		Type: "GM",
		Body: message.Convert(),
	}
	jsonRes, _ := json.Marshal(response)

	// TODO: Send message to all members

	// Send message to sender
	err := c.WriteMessage(*mt, jsonRes)
	if err != nil {
		log.Println("err:", err)
	}

	return true
}

func sendAllNotifications(id *int, c *websocket.Conn) {
	set := dbhandler.GetUnreadMsgs(id)
	jsonString := structs.JsonifyRedisZ(set)

	var notifications structs.ClientNotifications
	json.Unmarshal([]byte(jsonString), &notifications)

	exists, status := dbhandler.FriendReqExists(id)
	if !status {
		log.Println("Failed : Couldn't send all notifications")
		return
	}

	notifications.FriendReq = exists

	response := structs.ClientSocketMessage{
		Type: "ALL_NOTIFICATIONS",
		Body: notifications,
	}
	jsonRes, _ := json.Marshal(response)

	err := c.WriteMessage(1, jsonRes)
	if err != nil {
		log.Println("err:", err)
	}
}

func sendFRNotification(id *int, c *websocket.Conn) {
	response := structs.ClientSocketMessage{
		Type: "FR_NOTI",
		Body: structs.ClientNotifications{
			FriendReq: true,
		},
	}

	jsonRes, _ := json.Marshal(response)

	err := c.WriteMessage(1, jsonRes)
	if err != nil {
		log.Println("err:", err)
	}
}

func sendRefresh(id *int, c *websocket.Conn, relation *int) {
	if *relation == 0 {
		return
	}

	response := structs.ClientSocketMessage{
		Type: "REFRESH_FRIENDS",
		Body: nil,
	}
	jsonRes, _ := json.Marshal(response)
	err := c.WriteMessage(1, jsonRes)
	if err != nil {
		log.Println("err:", err)
	}
}
