class Xenosocket{
    constructor() {
        this.socket = null

        this.chat = null
        this.setChat = null

        this.setState = null

        this.userInfo = null
    }

    connect() {
        this.socket = new WebSocket("ws://localhost:7777/socket");
        
        this.socket.onopen = () => {
            this.setState("DONE")
        };
        
        this.socket.onmessage = (event) => {
            var data = JSON.parse(event.data)
            switch(data.type) {
                case "DM":
                    if(this.chat.friendId === data.body.senderId || this.chat.friendId === data.body.receiverId) {
                        var newChat = Object.assign({}, this.chat)
                        newChat.history.push(data.body)
                        this.setChat(newChat)
                    }
                    break
                case "CHAT_HISTORY_RES":
                    //console.log(this.chat)  
                    this.setChat(data.body)
                    break
                default:
            }
            console.log(data)
        }

        this.socket.onerror = function(error) {
            alert(`[error] ${error.message}`);
        };
    }

    disconnect() {
        this.socket.close()
    }

    sendDM(message, receiverId) {
        var body = JSON.stringify({
            type: "DM",
            body: {
                receiverId: parseInt(receiverId),
                message: message,
            }
        })
        //console.log(body)
        this.socket.send(body);
    }

    getChat(friendId) {
        var body = JSON.stringify({
            type: "CHAT_HISTORY_REQ",
            body: {
                id: friendId
            }
        })
        this.socket.send(body);
    }
}

export default Xenosocket