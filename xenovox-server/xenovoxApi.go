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
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	tokenCookie, _ := r.Cookie("xeno_token")
	token := tokenCookie.Value
	id := dbhandler.GetUserId(&token)
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

		// Get socket of the receiver of the message
		if receiverSocket, ok := sockets[messageObj.ReceiverId]; ok {
			err = receiverSocket.WriteMessage(mt, message)
			if err != nil {
				log.Println("err:", err)
				break
			}
		} else {
			log.Println("Key not found")
			continue
		}

		// Send message to receiver
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("err:", err)
			break
		}

		// Insert message to database
		messageObj.SenderId = id
		ret := dbhandler.InsertPrivateMessage(&messageObj)
		if ret == "FAILED" {
			log.Println(ret)
		}
		log.Printf("recv: %s", message)

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

	mainRouter.HandleFunc("/logout", logout).Methods("POST")
	mainRouter.HandleFunc("/users", getUsers).Methods("GET")
	mainRouter.HandleFunc("/send", sendMessage)

	log.Fatal(http.ListenAndServe(*addr, nil))
}

func addCorsHeaders(w http.ResponseWriter) http.ResponseWriter {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

	return w
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w = addCorsHeaders(w)
		next.ServeHTTP(w, r)
	})
}

func authorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w = addCorsHeaders(w)
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

func getUsers(w http.ResponseWriter, r *http.Request) {
	rows := dbhandler.GetUsers()
	//token, err := c.Cookie("xeno_token")

	jsonUsers := structs.Jsonify(rows)
	users := []structs.User{}

	for _, jsonUser := range jsonUsers {

		if jsonUser == "," {
			continue
		}
		user := &structs.User{}
		easyjson.Unmarshal([]byte(jsonUser), user)
		users = append(users, *user)
	}

	//c.IndentedJSON(http.StatusOK, users)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	jsonRes, _ := json.Marshal(users)
	w.Write(jsonRes)
}

/*package main

import (
	"net/http"

	dbhandler "github.com/Disidious/Xenovox/DbHandler"
	structs "github.com/Disidious/Xenovox/Structs"
	"github.com/gin-gonic/gin"
	"github.com/mailru/easyjson"
)

type apiResponse struct {
	Message string `json:"message"`
}

func main() {
	dbhandler.ConnectDb()
	defer dbhandler.CloseDb()

	router := gin.Default()
	router.Use(corsMiddleware())
	router.POST("/login", login)
	router.POST("/logout", logout)
	router.POST("/register", register)

	router.Use(authorize())
	router.GET("/users", getUsers)

	router.Run("localhost:7777")
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func authorize() gin.HandlerFunc {
	return func(c *gin.Context) {
		// middleware

		token, _ := c.Cookie("xeno_token")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, apiResponse{Message: "UNAUTHORIZED"})
			return
		}

		ret := dbhandler.Authorize(token)
		if ret == "UNAUTHORIZED" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, apiResponse{Message: ret})
			return
		}

		c.Next()
	}
}

func login(c *gin.Context) {
	var user structs.Users
	if err := c.BindJSON(&user); err != nil {
		c.IndentedJSON(http.StatusBadRequest, apiResponse{Message: "UEXPECTED_FAILURE"})
		return
	}

	ret, token := dbhandler.Login(&user)
	if ret == "AUTH_FAILED" {
		c.IndentedJSON(http.StatusUnauthorized, apiResponse{Message: ret})
	} else if ret == "SUCCESS" {
		c.SetCookie("xeno_token", token, 2592000, "/", "localhost", false, true)
		c.IndentedJSON(http.StatusOK, apiResponse{Message: "LOGGED_IN"})
	}
}

func logout(c *gin.Context) {
	_, err := c.Cookie("xeno_token")
	if err != nil {
		c.IndentedJSON(http.StatusForbidden, apiResponse{Message: "LOGOUT_FAILED"})
		return
	}

	c.SetCookie("xeno_token", "", -1, "/", "localhost", false, true)
	c.IndentedJSON(http.StatusOK, apiResponse{Message: "LOGGED_OUT"})
}

func register(c *gin.Context) {
	var newUser structs.Users
	if err := c.BindJSON(&newUser); err != nil {
		c.IndentedJSON(http.StatusBadRequest, apiResponse{Message: "UEXPECTED_FAILURE"})
		return
	}

	ret := dbhandler.Register(&newUser)
	if ret == "EMAIL_EXISTS" || ret == "USERNAME_EXISTS" {
		c.IndentedJSON(http.StatusForbidden, apiResponse{Message: ret})
	} else if ret == "SUCCESS" {
		c.IndentedJSON(http.StatusCreated, apiResponse{Message: ret})
	} else {
		c.IndentedJSON(http.StatusBadRequest, apiResponse{Message: ret})
	}
}

func getUsers(c *gin.Context) {
	rows := dbhandler.GetUsers()
	//token, err := c.Cookie("xeno_token")

	jsonUsers := structs.Jsonify(rows)
	users := []structs.Users{}

	for _, jsonUser := range jsonUsers {

		if jsonUser == "," {
			continue
		}
		user := &structs.Users{}
		easyjson.Unmarshal([]byte(jsonUser), user)
		users = append(users, *user)
	}

	c.IndentedJSON(http.StatusOK, users)
} */
