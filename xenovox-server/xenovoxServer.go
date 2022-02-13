package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"reflect"
	"time"

	dbhandler "github.com/Disidious/Xenovox/DbHandler"
	structs "github.com/Disidious/Xenovox/Structs"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type apiResponse struct {
	Message string `json:"message"`
}

type ChatType int

const (
	None    = -1
	Private = 0
	Group   = 1
)

type chat struct {
	chatType ChatType
	chatId   int
}

var addr = flag.String("addr", "localhost:7777", "http service address")

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	return origin == "http://localhost:3000"
}}

var sockets map[int]*websocket.Conn = make(map[int]*websocket.Conn)
var currChats map[int]chat = make(map[int]chat)

func resetSocket(id *int) {
	log.Println("Closing Socket ID", *id)
	sockets[*id].Close()
	delete(currChats, *id)
}

func socketIncomingHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Client connected")
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	tokenCookie, _ := r.Cookie("xeno_token")
	token := tokenCookie.Value
	id := dbhandler.GetUserId(&token)

	if clientSocket, ok := sockets[id]; ok {
		clientSocket.Close()
	}

	sockets[id] = c
	defer resetSocket(&id)

	sendAllNotifications(&id, c)

	currChats[id] = chat{
		chatType: None,
		chatId:   -1,
	}

	for {
		mt, message, err := c.ReadMessage()

		// If connection closed remove key from sockets map
		if err != nil {
			log.Println("err disconnect:", err)
			break
		}

		var request structs.ClientSocketMessage
		err = json.Unmarshal(message, &request)
		if err != nil {
			log.Println("Failed to unmarshal: ", err)
			break
		}

		switch request.Type {
		case "DM":
			bodyMap, ok := (request.Body).(map[string]interface{})
			if !ok {
				log.Println("Failed : couldn't get body from json")
				continue
			}

			// Convert message to Message object
			var messageObj structs.Message
			ok = structs.StructifyMap(&bodyMap, &messageObj)

			if !ok {
				log.Println("Failed : couldn't convert map to struct")
				continue
			}

			messageObj.SenderId = id
			if !sendDM(&messageObj, &mt) {
				continue
			}
			log.Printf("recv: %s", message)
		case "GM":
			bodyMap, ok := (request.Body).(map[string]interface{})
			if !ok {
				log.Println("Failed : couldn't get body from json")
				continue
			}

			// Convert message to Message object
			var messageObj structs.GroupMessage
			ok = structs.StructifyMap(&bodyMap, &messageObj)

			if !ok {
				log.Println("Failed : couldn't convert map to struct")
				continue
			}

			messageObj.SenderId = id
			if !sendGM(&messageObj, true, &mt) {
				continue
			}
			log.Printf("recv: %s", message)

		case "PRIVATE_HISTORY_REQ":
			bodyMap, ok := (request.Body).(map[string]interface{})
			if !ok {
				log.Println("Failed : couldn't get id from json")
				continue
			}

			friendId, ok := (bodyMap["id"].(float64))
			if !ok {
				log.Println("Failed : id of wrong type")
				continue
			}

			friendIdInt := int(friendId)
			if !getPrivateChat(&friendIdInt, &mt, &id, c) {
				currChats[id] = chat{
					chatType: Private,
					chatId:   -1,
				}
				continue
			}
			currChats[id] = chat{
				chatType: Private,
				chatId:   friendIdInt,
			}
		case "GROUP_HISTORY_REQ":
			bodyMap, ok := (request.Body).(map[string]interface{})
			if !ok {
				log.Println("Failed : couldn't get id from json")
				continue
			}

			groupId, ok := (bodyMap["id"].(float64))
			if !ok {
				log.Println("Failed : id of wrong type")
				continue
			}

			groupIdInt := int(groupId)
			if !getGroupChatAndMembers(&groupIdInt, &mt, &id, c) {
				currChats[id] = chat{
					chatType: Private,
					chatId:   -1,
				}
				continue
			}
			currChats[id] = chat{
				chatType: Group,
				chatId:   groupIdInt,
			}

		}
	}
}

func main() {
	dbhandler.ConnectDb()
	defer dbhandler.CloseDb()

	router := mux.NewRouter()
	mainRouter := router.PathPrefix("/").Subrouter()
	authRouter := router.PathPrefix("/auth").Subrouter()

	http.Handle("/", router)

	authRouter.Use(corsMiddleware)
	mainRouter.Use(authorizationMiddleware)

	authRouter.HandleFunc("/login", login).Methods("POST")
	authRouter.HandleFunc("/register", register).Methods("POST")

	mainRouter.HandleFunc("/checkauth", checkAuth).Methods("GET")
	mainRouter.HandleFunc("/logout", logout).Methods("POST")
	mainRouter.HandleFunc("/friends", getFriends).Methods("GET")
	mainRouter.HandleFunc("/groups", getGroups).Methods("GET")
	mainRouter.HandleFunc("/connections", getConnections).Methods("GET")
	mainRouter.HandleFunc("/info", getUserInfo).Methods("GET")
	mainRouter.HandleFunc("/socket", socketIncomingHandler)
	mainRouter.HandleFunc("/sendRelation", sendRelation).Methods("POST")
	mainRouter.HandleFunc("/friendRequests", getFriendRequests).Methods("GET")
	mainRouter.HandleFunc("/read", updateDMsToRead).Methods("POST")
	//mainRouter.HandleFunc("/createGroup", createGroup).Methods("POST")
	//mainRouter.HandleFunc("/leave", leaveGroup).Methods("POST")
	mainRouter.HandleFunc("/addToGroup", addToGroup).Methods("POST")

	log.Fatal(http.ListenAndServe(*addr, nil))
}

func addCorsHeaders(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	(*w).Header().Set("Access-Control-Allow-Credentials", "true")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		addCorsHeaders(&w)
		next.ServeHTTP(w, r)
	})
}

func authorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		addCorsHeaders(&w)
		w.Header().Set("Content-Type", "application/json")

		tokenCookie, err := r.Cookie("xeno_token")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			jsonRes, _ := json.Marshal(apiResponse{Message: "UNAUTHORIZED"})
			w.Write(jsonRes)
			return
		}

		token := tokenCookie.Value
		if token == "" {
			w.WriteHeader(http.StatusUnauthorized)
			jsonRes, _ := json.Marshal(apiResponse{Message: "UNAUTHORIZED"})
			w.Write(jsonRes)
			return
		}

		ret := dbhandler.Authorize(&token)
		if ret == "UNAUTHORIZED" {
			w.WriteHeader(http.StatusUnauthorized)
			jsonRes, _ := json.Marshal(apiResponse{Message: ret})
			w.Write(jsonRes)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func checkAuth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	jsonRes, _ := json.Marshal(apiResponse{Message: "AUTHORIZED"})
	w.Write(jsonRes)
}

func getId(w http.ResponseWriter, r *http.Request) (id int, ok bool) {
	tokenCookie, err := r.Cookie("xeno_token")
	if err != nil {
		ok = false
		w.WriteHeader(http.StatusBadRequest)
		jsonRes, _ := json.Marshal(apiResponse{Message: "UNEXPECTED_FAILURE"})
		w.Write(jsonRes)
		return
	}

	token := tokenCookie.Value
	id = dbhandler.GetUserId(&token)
	ok = true

	return
}

func failureRes(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	jsonRes, _ := json.Marshal(apiResponse{Message: "UNEXPECTED_FAILURE"})
	w.Write(jsonRes)
	return
}

func login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var user structs.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		jsonRes, _ := json.Marshal(apiResponse{Message: "UNEXPECTED_FAILURE"})
		w.Write(jsonRes)
		return
	}

	ret, token := dbhandler.Login(&user)
	if ret == "AUTH_FAILED" {
		w.WriteHeader(http.StatusUnauthorized)
		jsonRes, _ := json.Marshal(apiResponse{Message: ret})
		w.Write(jsonRes)
		return
	} else if ret == "SUCCESS" {
		tokenCookie := http.Cookie{Name: "xeno_token",
			Value:    token,
			Expires:  time.Now().Add(2592000 * time.Second),
			Path:     "/",
			Domain:   "localhost",
			HttpOnly: true}

		http.SetCookie(w, &tokenCookie)

		jsonRes, _ := json.Marshal(apiResponse{Message: "LOGGED_IN"})
		w.Write(jsonRes)

		return
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	_, err := r.Cookie("xeno_token")
	if err != nil {
		//c.IndentedJSON(http.StatusForbidden, apiResponse{Message: "LOGOUT_FAILED"})
		w.WriteHeader(http.StatusUnauthorized)
		jsonRes, _ := json.Marshal(apiResponse{Message: "LOGOUT_FAILED"})
		w.Write(jsonRes)

		return
	}
	tokenCookie := http.Cookie{
		Name:     "xeno_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Second),
		Path:     "/",
		Domain:   "localhost",
		HttpOnly: true}
	http.SetCookie(w, &tokenCookie)

	jsonRes, _ := json.Marshal(apiResponse{Message: "LOGGED_OUT"})
	w.Write(jsonRes)
}

func register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var newUser structs.User
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		failureRes(w, r)
		return
	}

	ret := dbhandler.Register(&newUser)
	if ret == "EMAIL_EXISTS" || ret == "USERNAME_EXISTS" {
		w.WriteHeader(http.StatusForbidden)
	} else if ret == "SUCCESS" {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
	jsonRes, _ := json.Marshal(apiResponse{Message: ret})
	w.Write(jsonRes)
}

func sendRelation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var newRelation structs.Relation
	err := json.NewDecoder(r.Body).Decode(&newRelation)
	if err != nil {
		failureRes(w, r)
		return
	}

	id, ok := getId(w, r)
	if !ok {
		return
	}

	newRelation.User1Id = id

	if id == newRelation.User2Id {
		failureRes(w, r)
		return
	}

	ret := dbhandler.UpsertRelation(&newRelation)
	if ret == "FAILED" {
		failureRes(w, r)
		return
	}

	if receiverSocket, ok := sockets[newRelation.User2Id]; ok {
		// Send notification or refresh friend to the receiver client based on the relation
		if newRelation.Relation == 0 {
			sendFRNotification(&id, receiverSocket)
		} else {
			sendRefresh(receiverSocket, false)
		}
	}
	jsonRes, _ := json.Marshal(apiResponse{Message: ret})
	w.Write(jsonRes)
}

func getUserInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, ok := getId(w, r)
	if !ok {
		return
	}

	row := dbhandler.GetUserInfo(&id)

	var userInfo structs.ClientUser
	row.Scan(&userInfo.Id, &userInfo.Name, &userInfo.Username, &userInfo.Email, &userInfo.Picture)
	jsonRes, _ := json.Marshal(userInfo)
	w.Write(jsonRes)
}

func getGroups(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, ok := getId(w, r)
	if !ok {
		return
	}

	rows, status := dbhandler.GetGroups(&id)
	if !status {
		failureRes(w, r)
		return
	}

	jsonString := structs.JsonifyRows(rows)

	w.Write([]byte(jsonString))
}

func getFriends(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, ok := getId(w, r)
	if !ok {
		return
	}

	rows, status := dbhandler.GetFriends(&id)
	if !status {
		failureRes(w, r)
		return
	}

	jsonString := structs.JsonifyRows(rows)

	w.Write([]byte(jsonString))
}

func getConnections(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, ok := getId(w, r)
	if !ok {
		return
	}

	frows, grows, status := dbhandler.GetAllConnections(&id)
	if !status {
		failureRes(w, r)
		return
	}

	fJsonString := structs.JsonifyRows(frows)
	gJsonString := structs.JsonifyRows(grows)

	jsonString := `{"friends":` + fJsonString + `,"groups":` + gJsonString + `}`

	w.Write([]byte(jsonString))
}

func getFriendRequests(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, ok := getId(w, r)
	if !ok {
		return
	}

	rows, status := dbhandler.GetFriendRequests(&id)
	if !status {
		failureRes(w, r)
		return
	}

	jsonString := structs.JsonifyRows(rows)

	w.Write([]byte(jsonString))
}

func updateDMsToRead(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, ok := getId(w, r)
	if !ok {
		return
	}

	var body map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		failureRes(w, r)
		return
	}

	chatId := int(body["id"].(float64))
	if !body["group"].(bool) {
		ok = dbhandler.UpdateDMsToRead(&chatId, &id)
	} else {
		ok = dbhandler.UpdateGMsToRead(&chatId, &id)
	}

	if !ok {
		failureRes(w, r)
		return
	}

	jsonRes, _ := json.Marshal(apiResponse{Message: "SUCCESS"})
	w.Write(jsonRes)
}

// func createGroup(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")

// 	id, ok := getId(w, r)
// 	if !ok {
// 		return
// 	}

// }

func addToGroup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, ok := getId(w, r)
	if !ok {
		return
	}

	var body map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		failureRes(w, r)
		return
	}

	friendIds := body["friendids"].([]interface{})
	groupId := int(body["groupid"].(float64))

	friendIdsCasted := make([]int, 0)
	for _, friendId := range friendIds {
		friendIdsCasted = append(friendIdsCasted, int(friendId.(float64)))
	}

	res := dbhandler.AddToGroup(&id, &friendIdsCasted, &groupId)
	if res != "SUCCESS" {
		failureRes(w, r)
		return
	}

	usernames := make(map[int]string, 0)
	var senderUsername string
	if rows, ok := dbhandler.GetUsernames(append(friendIdsCasted, id)); ok {
		users, ok := structs.StructifyRows(rows, reflect.TypeOf(structs.User{}))
		if !ok {
			failureRes(w, r)
			return
		}
		for _, user := range users {
			userCasted := user.(*structs.User)
			if userCasted.Id == id {
				senderUsername = userCasted.Username
				continue
			}

			usernames[userCasted.Id] = userCasted.Username
		}
	}

	// Send refresh groups to added user and send system message of adding a new member to a group
	sendToSender := true
	for _, friendId := range friendIdsCasted {
		if receiverSocket, ok := sockets[friendId]; ok {
			sendRefresh(receiverSocket, true)
		}

		sysMsg := senderUsername + " added " + usernames[friendId]
		mt := 1
		if !sendGM(&structs.GroupMessage{
			Message:  sysMsg,
			SenderId: id,
			GroupId:  groupId,
			IsSystem: true,
		}, sendToSender, &mt) {
			failureRes(w, r)
			return
		}
		sendToSender = false
	}

	jsonRes, _ := json.Marshal(apiResponse{Message: "SUCCESS"})
	w.Write(jsonRes)
}
