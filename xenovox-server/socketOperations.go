package main

import (
	"encoding/json"
	"log"
	"reflect"

	"github.com/gorilla/websocket"

	dbhandler "github.com/Disidious/Xenovox/DbHandler"
	structs "github.com/Disidious/Xenovox/Structs"
)

func getPrivateChat(friendId *int, id *int, c *websocket.Conn) bool {
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

	c.WriteMessage(websocket.TextMessage, jsonRes)

	return true
}

func getGroupChatAndMembers(groupId *int, id *int, c *websocket.Conn) bool {
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

	c.WriteMessage(websocket.TextMessage, jsonRes)

	return true
}

func sendDM(message *structs.Message) bool {
	// Insert message to database
	dmId, ret := dbhandler.InsertDirectMessage(message)
	if ret == "FAILED" {
		log.Println(ret)
		return false
	} else if ret == "BLOCKED" {
		log.Println(ret)
		return false
	}

	clientMsg := message.Convert()
	clientMsg.Id = dmId
	dbhandler.GetUsernameAndPicture(&message.SenderId, true).Scan(&clientMsg.Username, &clientMsg.Picture)

	response := structs.ClientSocketMessage{
		Type: "DM",
		Body: clientMsg,
	}
	jsonRes, _ := json.Marshal(response)

	if currChat, ok := currChats[message.ReceiverId]; !ok || currChat.chatType != Private || currChat.chatId != message.SenderId {
		dbhandler.AppendUnreadDM(&message.SenderId, &message.ReceiverId)
	}

	// Get socket of the receiver of the message
	if receiverSocket, ok := sockets[message.ReceiverId]; ok {
		// Send message to receiver
		receiverSocket.WriteMessage(websocket.TextMessage, jsonRes)
	} else {
		log.Println("Receiver Offline")
	}

	// Send message to sender
	err := sockets[message.SenderId].WriteMessage(websocket.TextMessage, jsonRes)
	if err != nil {
		log.Println("err:", err)
	}

	return true
}

func sendGM(message *structs.GroupMessage, sendToSender bool, sendMembers bool, sendGroupInfo bool) bool {
	// Insert message to database
	gmId, ret := dbhandler.InsertGroupMessage(message)
	if ret == "FAILED" {
		return false
	}

	body := make(map[string]interface{})

	clientMsg := message.Convert()
	clientMsg.Id = gmId
	dbhandler.GetUsernameAndPicture(&message.SenderId, true).Scan(&clientMsg.Username, &clientMsg.Picture)
	body["message"] = clientMsg

	// Store notifications for all group members that are offline or not on chat
	mrows, ok := dbhandler.GetGroupMembers(&message.GroupId)
	if ok {
		members, ok := structs.StructifyRows(mrows, reflect.TypeOf(structs.User{}))
		if ok {
			if sendGroupInfo {
				rowInfo := dbhandler.GetGroupInfo(&message.GroupId)
				body["groupinfo"], ok = structs.StructifyRows(rowInfo, reflect.TypeOf(structs.ClientGroup{}))
				if !ok {
					return false
				}
			}
			if sendMembers {
				body["members"] = members
			}
			response := structs.ClientSocketMessage{
				Type: "GM",
				Body: body,
			}
			jsonRes, _ := json.Marshal(response)

			var membersToNotify []structs.User
			for _, member := range members {
				castedMember := member.(*structs.User)
				if castedMember.Id == message.SenderId && !sendToSender {
					continue
				}

				if receiverSocket, ok := sockets[castedMember.Id]; ok {
					// Send message to receiver
					receiverSocket.WriteMessage(websocket.TextMessage, jsonRes)
				}

				if currChat, ok := currChats[castedMember.Id]; !ok || currChat.chatType != Group || currChat.chatId != message.GroupId {
					membersToNotify = append(membersToNotify, *castedMember)
				}
			}
			dbhandler.AppendUnreadGM(&message.GroupId, &membersToNotify)
		} else {
			return false
		}
	} else {
		return false
	}

	return true
}

func sendAllNotifications(id *int, c *websocket.Conn) {
	set := dbhandler.GetUnreadMsgs(id)
	jsonString := structs.JsonifyRedisZ(set)

	var notifications structs.ClientNotifications
	json.Unmarshal([]byte(jsonString), &notifications)

	exists := dbhandler.FriendReqExists(id)

	notifications.FriendReq = exists

	response := structs.ClientSocketMessage{
		Type: "ALL_NOTIFICATIONS",
		Body: notifications,
	}
	jsonRes, _ := json.Marshal(response)

	err := c.WriteMessage(websocket.TextMessage, jsonRes)
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

	err := c.WriteMessage(websocket.TextMessage, jsonRes)
	if err != nil {
		log.Println("err:", err)
	}
}

func sendRefresh(c *websocket.Conn, groups bool) {
	resType := "REFRESH_FRIENDS"
	if groups {
		resType = "REFRESH_GROUPS"
	}

	response := structs.ClientSocketMessage{
		Type: resType,
		Body: nil,
	}
	jsonRes, _ := json.Marshal(response)
	err := c.WriteMessage(1, jsonRes)
	if err != nil {
		log.Println("err:", err)
	}
}
