package main

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"

	dbhandler "github.com/Disidious/Xenovox/DbHandler"
	structs "github.com/Disidious/Xenovox/Structs"
)

func getChat(body *interface{}, message *[]byte, mt *int, id *int, c *websocket.Conn) bool {
	// {"id": 11}
	// log.Println(body.id)
	// friendId := (*body).(int)
	// log.Println(friendId)

	bodyMap, ok := (*body).(map[string]interface{})
	if !ok {
		log.Println("Failed : couldn't get id from json")
		return false
	}

	friendId, ok := (bodyMap["id"].(float64))
	if !ok {
		log.Println("Failed : id of wrong type")
		return false
	}

	friendIdInt := int(friendId)
	rows, status := dbhandler.GetChat(id, &friendIdInt)
	if !status {
		return false
	}

	jsonChat := structs.Jsonify(rows)
	chatMessages := []structs.ClientMessage{}

	for _, jsonMessage := range jsonChat {
		if jsonMessage == "," {
			continue
		}

		chatM := &structs.ClientMessage{}
		//dbChatM := &structs.Message{}
		json.Unmarshal([]byte(jsonMessage), chatM)

		chatMessages = append(chatMessages, *chatM)
	}

	response := structs.ClientSocketMessage{
		Type: "CHAT_HISTORY_RES",
		Body: structs.ClientChatHistory{
			History:  chatMessages,
			FriendId: friendIdInt,
		},
	}
	jsonRes, _ := json.Marshal(response)

	c.WriteMessage(*mt, jsonRes)

	return true
}

func sendDM(body *interface{}, message *[]byte, mt *int, id *int, c *websocket.Conn) bool {
	// Convert message to Message object
	var messageObj structs.Message
	bodyBytes, _ := json.Marshal(body)
	err := json.Unmarshal(bodyBytes, &messageObj)

	if err != nil {
		log.Println("Failed to unmarshal: ", err)
		return false
	}

	// Insert message to database
	messageObj.SenderId = *id
	ret := dbhandler.InsertPrivateMessage(&messageObj)
	if ret == "FAILED" {
		log.Println(ret)
		return false
	} else if ret == "BLOCKED" {
		log.Println(ret)
		return false
	}
	response := structs.ClientSocketMessage{
		Type: "DM",
		Body: messageObj.Convert(),
	}
	jsonRes, _ := json.Marshal(response)
	log.Printf("recv: %s", message)

	// Get socket of the receiver of the message
	if receiverSocket, ok := sockets[messageObj.ReceiverId]; ok {
		// Send message to receiver
		err = receiverSocket.WriteMessage(*mt, jsonRes)
		if err != nil {
			log.Println("err:", err)
		}
	} else {
		log.Println("Receiver Offline")
	}

	// Send message to sender
	err = c.WriteMessage(*mt, jsonRes)
	if err != nil {
		log.Println("err:", err)
	}

	return true
}
