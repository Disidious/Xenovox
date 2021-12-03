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
	"github.com/mailru/easyjson"
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

func sendMessage(w http.ResponseWriter, r *http.Request) {
	log.Println("Client connected")
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	tokenCookie, _ := r.Cookie("xeno_token")
	token := tokenCookie.Value
	id := dbhandler.GetUserId(&token)

	if clientSocket, ok := sockets[id]; ok {
		clientSocket.Close()
	}

	sockets[id] = c

	for {
		mt, message, err := c.ReadMessage()

		// If connection closed remove key from sockets map
		if err != nil {
			log.Println("err:", err)
			delete(sockets, id)
			break
		}

		// Convert message to Message object
		var messageObj structs.Message
		err = json.Unmarshal(message, &messageObj)
		if err != nil {
			log.Println("Failed to unmarshal: ", err)
			break
		}

		// Insert message to database
		messageObj.SenderId = id
		ret := dbhandler.InsertPrivateMessage(&messageObj)
		if ret == "FAILED" {
			log.Println(ret)
			break
		} else if ret == "BLOCKED" {
			log.Println(ret)
			continue
		}
		log.Printf("recv: %s", message)

		// Get socket of the receiver of the message
		if receiverSocket, ok := sockets[messageObj.ReceiverId]; ok {
			// Send message to receiver
			err = receiverSocket.WriteMessage(mt, message)
			if err != nil {
				log.Println("err:", err)
			}
		} else {
			log.Println("Receiver Offline")
		}

		// Send message to sender
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("err:", err)
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

	//flag.Parse()
	//log.SetFlags(0)
	authRouter.HandleFunc("/login", login).Methods("POST")
	authRouter.HandleFunc("/register", register).Methods("POST")

	mainRouter.HandleFunc("/checkauth", checkAuth).Methods("GET")
	mainRouter.HandleFunc("/logout", logout).Methods("POST")
	mainRouter.HandleFunc("/friends", getFriends).Methods("GET")
	mainRouter.HandleFunc("/info", getUserInfo).Methods("GET")
	mainRouter.HandleFunc("/sendPM", sendMessage)
	mainRouter.HandleFunc("/sendRelation", sendRelation)

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
	tokenCookie := http.Cookie{Name: "xeno_token",
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

	ret := dbhandler.UpsertRelation(&newRelation)
	if ret == "FAILED" {
		w.WriteHeader(http.StatusBadRequest)
		jsonRes, _ := json.Marshal(apiResponse{Message: "UEXPECTED_FAILURE"})
		w.Write(jsonRes)
		return
	}

	jsonRes, _ := json.Marshal(apiResponse{Message: ret})
	w.Write(jsonRes)
}

func getUserInfo(w http.ResponseWriter, r *http.Request) {
	tokenCookie, err := r.Cookie("xeno_token")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		jsonRes, _ := json.Marshal(apiResponse{Message: "UEXPECTED_FAILURE"})
		w.Write(jsonRes)
		return
	}

	token := tokenCookie.Value
	id := dbhandler.GetUserId(&token)
	row := dbhandler.GetUserInfo(id)

	var userInfo structs.ClientUser
	row.Scan(&userInfo.Id, &userInfo.Name, &userInfo.Username, &userInfo.Email, &userInfo.Picture)
	jsonRes, _ := json.Marshal(userInfo)
	w.Write(jsonRes)
}

func getFriends(w http.ResponseWriter, r *http.Request) {
	tokenCookie, err := r.Cookie("xeno_token")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		jsonRes, _ := json.Marshal(apiResponse{Message: "UEXPECTED_FAILURE"})
		w.Write(jsonRes)
		return
	}

	token := tokenCookie.Value
	id := dbhandler.GetUserId(&token)

	rows, status := dbhandler.GetFriends(id)
	if !status {
		w.WriteHeader(http.StatusBadRequest)
		jsonRes, _ := json.Marshal(apiResponse{Message: "UEXPECTED_FAILURE"})
		w.Write(jsonRes)
		return
	}

	jsonUsers := structs.Jsonify(rows)
	friends := []structs.ClientFriend{}

	for _, jsonUser := range jsonUsers {

		if jsonUser == "," {
			continue
		}
		user := &structs.ClientFriend{}
		easyjson.Unmarshal([]byte(jsonUser), user)
		friends = append(friends, *user)
	}

	//c.IndentedJSON(http.StatusOK, users)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	jsonRes, _ := json.Marshal(friends)
	w.Write(jsonRes)
}
