class Xenosocket{
    constructor() {
        this.socket = null
        this.manualDisconnect = false

        this.chat = null
        this.setChat = null

        this.notifications = null
        this.setNotifications = null

        this.getFriends = () => {}

        this.setState = null

        this.userInfo = null
    }

    connect() {
        this.socket = new WebSocket("ws://localhost:7777/socket");
        this.manualDisconnect = false
        
        this.socket.onopen = () => {
            this.setState("DONE")
        };
        
        this.socket.onmessage = (event) => {
            var data = JSON.parse(event.data)
            switch(data.type) {
                case "DM":
                    if(this.chat.friendid === data.body.senderid || this.chat.friendid === data.body.receiverid) {
                        var newChat = Object.assign({}, this.chat)
                        newChat.history.push(data.body)
                        this.setChat(newChat)
                    } else if(!this.notifications.dms.includes(data.body.senderid)) {
                        var newNoti = Object.assign({}, this.notifications)
                        newNoti.dms.push(data.body.senderid)
                        this.setNotifications(newNoti)
                    }
                    break

                case "CHAT_HISTORY_RES":  
                    this.setChat(data.body)
                    break
                
                case "ALL_NOTIFICATIONS":
                    var newNotis = Object.assign({}, this.notifications)
                    newNotis.dms = data.body.senderids
                    newNotis.friendreq = data.body.friendreq
                    this.setNotifications(newNotis)
                    break
                
                case "FR_NOTI":
                    newNoti = Object.assign({}, this.notifications)
                    newNoti.friendreq = data.body.friendreq
                    this.setNotifications(newNoti)
                    break

                case "REFRESH_FRIENDS":
                    this.getFriends();
                    break

                default:
            }
            console.log(data)
        }

        this.socket.onclose = () => {
            if(this.manualDisconnect)
                return
            this.connect()
        }

        this.socket.onerror = (error) => {
            if(this.manualDisconnect)
                return
            alert(`Reconnecting...`);
        };
    }

    disconnect() {
        this.manualDisconnect = true
        this.socket.close()
    }

    sendDM(message, receiverId) {
        var body = JSON.stringify({
            type: "DM",
            body: {
                receiverid: parseInt(receiverId),
                message: message,
            }
        })
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