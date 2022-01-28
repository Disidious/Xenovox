package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"time"

	dbhandler "github.com/Disidious/Xenovox/DbHandler"
	structs "github.com/Disidious/Xenovox/Structs"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type apiResponse struct {
	Message string `json:"message"`
}

var addr = flag.String("addr", "localhost:7777", "http service address")

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	return origin == "http://localhost:3000"
}}

var sockets map[int]*websocket.Conn = make(map[int]*websocket.Conn)
var currFriendIds map[int]int = make(map[int]int)

func resetSocket(id *int) {
	sockets[*id].Close()
	delete(sockets, *id)
	delete(currFriendIds, *id)
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

	currFriendIds[id] = -1

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
				log.Println("Failed : couldn't get id from json")
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
			if currFriendId, ok := currFriendIds[messageObj.ReceiverId]; ok && currFriendId == id {
				messageObj.Read = true
			} else {
				messageObj.Read = false
			}

			if !sendDM(&messageObj, &mt, &id, c) {
				continue
			}
			log.Printf("recv: %s", message)

		case "CHAT_HISTORY_REQ":
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
			if !getChat(&friendIdInt, &mt, &id, c) {
				currFriendIds[id] = -1
				continue
			}
			currFriendIds[id] = friendIdInt
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
	mainRouter.HandleFunc("/info", getUserInfo).Methods("GET")
	mainRouter.HandleFunc("/socket", socketIncomingHandler)
	mainRouter.HandleFunc("/sendRelation", sendRelation).Methods("POST")
	mainRouter.HandleFunc("/friendRequests", getFriendRequests).Methods("GET")
	mainRouter.HandleFunc("/read", updateDMsToRead).Methods("POST")

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

func login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var user structs.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		jsonRes, _ := json.Marshal(apiResponse{Message: "UEXPECTED_FAILURE"})
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
		w.WriteHeader(http.StatusOK)
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
	w.WriteHeader(http.StatusOK)
	jsonRes, _ := json.Marshal(apiResponse{Message: "LOGGED_OUT"})
	w.Write(jsonRes)
}

func register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var newUser structs.User
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		jsonRes, _ := json.Marshal(apiResponse{Message: "UEXPECTED_FAILURE"})
		w.Write(jsonRes)
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
		w.WriteHeader(http.StatusBadRequest)
		jsonRes, _ := json.Marshal(apiResponse{Message: "UEXPECTED_FAILURE"})
		w.Write(jsonRes)
		return
	}

	tokenCookie, err := r.Cookie("xeno_token")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		jsonRes, _ := json.Marshal(apiResponse{Message: "UEXPECTED_FAILURE"})
		w.Write(jsonRes)
		return
	}

	token := tokenCookie.Value
	id := dbhandler.GetUserId(&token)
	newRelation.User1Id = id

	if id == newRelation.User2Id {
		w.WriteHeader(http.StatusBadRequest)
		jsonRes, _ := json.Marshal(apiResponse{Message: "UNEXPECTED_FAILURE"})
		w.Write(jsonRes)
		return
	}

	ret := dbhandler.UpsertRelation(&newRelation)
	if ret == "FAILED" {
		w.WriteHeader(http.StatusBadRequest)
		jsonRes, _ := json.Marshal(apiResponse{Message: "UEXPECTED_FAILURE"})
		w.Write(jsonRes)
		return
	}

	if receiverSocket, ok := sockets[newRelation.User2Id]; ok {
		// Send refresh friends request and a notification to receiver client
		sendRefresh(&newRelation.User2Id, receiverSocket, &newRelation.Relation)
		if newRelation.Relation == 0 {
			sendFRNotification(&id, receiverSocket)
		}
	}
	jsonRes, _ := json.Marshal(apiResponse{Message: ret})
	w.Write(jsonRes)
}

func getUserInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tokenCookie, err := r.Cookie("xeno_token")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		jsonRes, _ := json.Marshal(apiResponse{Message: "UEXPECTED_FAILURE"})
		w.Write(jsonRes)
		return
	}

	token := tokenCookie.Value
	id := dbhandler.GetUserId(&token)
	row := dbhandler.GetUserInfo(&id)

	var userInfo structs.ClientUser
	row.Scan(&userInfo.Id, &userInfo.Name, &userInfo.Username, &userInfo.Email, &userInfo.Picture)
	jsonRes, _ := json.Marshal(userInfo)
	w.Write(jsonRes)
}

func getFriends(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tokenCookie, err := r.Cookie("xeno_token")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		jsonRes, _ := json.Marshal(apiResponse{Message: "UEXPECTED_FAILURE"})
		w.Write(jsonRes)
		return
	}

	token := tokenCookie.Value
	id := dbhandler.GetUserId(&token)

	rows, status := dbhandler.GetFriends(&id)
	if !status {
		w.WriteHeader(http.StatusBadRequest)
		jsonRes, _ := json.Marshal(apiResponse{Message: "UEXPECTED_FAILURE"})
		w.Write(jsonRes)
		return
	}

	jsonString := structs.JsonifyRows(rows)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(jsonString))
}

func getFriendRequests(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tokenCookie, err := r.Cookie("xeno_token")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		jsonRes, _ := json.Marshal(apiResponse{Message: "UEXPECTED_FAILURE"})
		w.Write(jsonRes)
		return
	}

	token := tokenCookie.Value
	id := dbhandler.GetUserId(&token)

	rows, status := dbhandler.GetFriendRequests(&id)
	if !status {
		w.WriteHeader(http.StatusBadRequest)
		jsonRes, _ := json.Marshal(apiResponse{Message: "UEXPECTED_FAILURE"})
		w.Write(jsonRes)
		return
	}

	jsonString := structs.JsonifyRows(rows)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(jsonString))
}

func updateDMsToRead(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var friend structs.ClientFriend
	err := json.NewDecoder(r.Body).Decode(&friend)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		jsonRes, _ := json.Marshal(apiResponse{Message: "UEXPECTED_FAILURE"})
		w.Write(jsonRes)
		return
	}

	tokenCookie, err := r.Cookie("xeno_token")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		jsonRes, _ := json.Marshal(apiResponse{Message: "UEXPECTED_FAILURE"})
		w.Write(jsonRes)
		return
	}

	token := tokenCookie.Value
	id := dbhandler.GetUserId(&token)

	status := dbhandler.UpdateDMsToRead(&id, &friend.Id)
	if status == "FAILED" {
		w.WriteHeader(http.StatusBadRequest)
		jsonRes, _ := json.Marshal(apiResponse{Message: "UEXPECTED_FAILURE"})
		w.Write(jsonRes)
		return
	}

	jsonRes, _ := json.Marshal(apiResponse{Message: status})
	w.Write(jsonRes)
}
