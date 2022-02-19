package main

import (
	"encoding/json"
	"log"
	"reflect"

	"github.com/gorilla/websocket"

	dbhandler "github.com/Disidious/Xenovox/dbHandler"
	structs "github.com/Disidious/Xenovox/structs"
)

func getPrivateChat(friendId *int, id *int) bool {
	rows, status := dbhandler.GetPrivateChat(id, friendId)
	if !status {
		return false
	}

	history, ok := structs.StructifyRows(rows, reflect.TypeOf(structs.ClientDM{}))

	if !ok {
		log.Println("Failed: could not convert rows to struct")
		return false
	}

	response := structs.ClientSocketMessage{
		Type: "CHAT_HISTORY_RES",
		Body: structs.ClientChatHistory{
			Group:   false,
			History: history,
			ChatId:  *friendId,
		},
	}
	jsonRes, _ := json.Marshal(response)

	if c, ok := sockets[*id]; ok {
		err := c.WriteMessage(websocket.TextMessage, jsonRes)
		if err != nil {
			log.Println(err)
		}
	} else {
		log.Println("Socket not found")
	}

	return true
}

func getGroupChat(groupId *int, id *int) bool {
	rows, status := dbhandler.GetGroupChat(id, groupId)
	if !status {
		return false
	}

	history, ok := structs.StructifyRows(rows, reflect.TypeOf(structs.ClientGM{}))
	if !ok {
		log.Println("Failed: could not convert hrows to struct")
		return false
	}

	response := structs.ClientSocketMessage{
		Type: "CHAT_HISTORY_RES",
		Body: structs.ClientChatHistory{
			Group:   true,
			History: history,
			ChatId:  *groupId,
		},
	}
	jsonRes, _ := json.Marshal(response)

	if c, ok := sockets[*id]; ok {
		err := c.WriteMessage(websocket.TextMessage, jsonRes)
		if err != nil {
			log.Println(err)
		}
	} else {
		log.Println("Socket not found")
	}

	return true
}

func sendDM(message *structs.Message) bool {
	clientMsg := message.Convert()
	rows, status := dbhandler.GetUsernameAndPicture(&message.SenderId, true)
	if !status {
		return false
	}
	rows.Next()
	rows.Scan(&clientMsg.Username, &clientMsg.Picture)

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
	}

	if c, ok := sockets[message.SenderId]; ok {
		err := c.WriteMessage(websocket.TextMessage, jsonRes)
		if err != nil {
			log.Println(err)
		}
	} else {
		log.Println("Socket not found")
	}

	return true
}

func sendGM(message *structs.GroupMessage, sendToSender bool, sendMembers bool, sendGroupInfo bool) bool {
	body := make(map[string]interface{})

	clientMsg := message.Convert()
	rows, status := dbhandler.GetUsernameAndPicture(&message.SenderId, true)
	if !status {
		return false
	}
	rows.Next()
	rows.Scan(&clientMsg.Username, &clientMsg.Picture)
	body["message"] = clientMsg

	// Store notifications for all group members that are offline or not on chat
	mrows, ok := dbhandler.GetGroupMembers(&message.GroupId)
	if ok {
		members, ok := structs.StructifyRows(mrows, reflect.TypeOf(structs.ClientGroupMember{}))
		if ok {
			if sendGroupInfo {
				rowInfo, ok := dbhandler.GetGroupInfo(&message.GroupId)
				if !ok {
					return false
				}

				ret, ok := structs.StructifyRows(rowInfo, reflect.TypeOf(structs.ClientGroup{}))
				body["groupinfo"] = ret[0]
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
				castedMember := member.(*structs.ClientGroupMember)
				if castedMember.UserId == message.SenderId && !sendToSender {
					continue
				}

				if receiverSocket, ok := sockets[castedMember.UserId]; ok {
					// Send message to receiver
					receiverSocket.WriteMessage(websocket.TextMessage, jsonRes)
				}

				if currChat, ok := currChats[castedMember.UserId]; !ok || currChat.chatType != Group || currChat.chatId != message.GroupId {
					membersToNotify = append(membersToNotify, structs.User{
						Id: castedMember.UserId,
					})
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

func sendAllNotifications(id *int) bool {
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

	if c, ok := sockets[*id]; ok {
		err := c.WriteMessage(websocket.TextMessage, jsonRes)
		if err != nil {
			log.Println(err)
			return false
		}
	} else {
		log.Println("Socket not found")
	}

	return true
}

func sendFRNotification(id *int) bool {
	response := structs.ClientSocketMessage{
		Type: "FR_NOTI",
		Body: structs.ClientNotifications{
			FriendReq: true,
		},
	}

	jsonRes, _ := json.Marshal(response)

	if c, ok := sockets[*id]; ok {
		err := c.WriteMessage(websocket.TextMessage, jsonRes)
		if err != nil {
			log.Println(err)
			return false
		}
	} else {
		log.Println("Socket not found")
	}

	return true
}

func sendRefresh(id *int, entityId *int, remFlag bool, isGroup bool) bool {
	var resType string
	body := make(map[string]interface{})

	body["remove"] = remFlag
	if remFlag {
		body["entity"] = entityId
	}

	if isGroup {
		resType = "REFRESH_GROUPS"
		if rows, ok := dbhandler.GetGroupInfo(entityId); ok && !remFlag {
			if objs, ok := structs.StructifyRows(rows, reflect.TypeOf(structs.ClientGroup{})); ok {
				body["entity"] = objs[0]
			} else {
				return false
			}
		} else if !remFlag {
			return false
		}
	} else {
		resType = "REFRESH_FRIENDS"
		if rows, ok := dbhandler.GetUserInfoMin(entityId); ok && !remFlag {
			if objs, ok := structs.StructifyRows(rows, reflect.TypeOf(structs.ClientUser{})); ok {
				body["entity"] = objs[0]
			} else {
				return false
			}
		} else if !remFlag {
			return false
		}
	}

	response := structs.ClientSocketMessage{
		Type: resType,
		Body: body,
	}
	jsonRes, _ := json.Marshal(response)

	if c, ok := sockets[*id]; ok {
		err := c.WriteMessage(websocket.TextMessage, jsonRes)
		if err != nil {
			log.Println(err)
			return false
		}
	} else {
		log.Println("Socket not found")
	}

	return true
}
