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
                    if(this.chat.chatid === data.body.senderid || this.chat.chatid === data.body.receiverid) {
                        let newChat = Object.assign({}, this.chat)
                        newChat.history.push(data.body)
                        this.setChat(newChat)
                    } else if(!this.notifications.senderids.includes(data.body.senderid)) {
                        let newNoti = Object.assign({}, this.notifications)
                        newNoti.senderids.push(data.body.senderid)
                        newNoti.senderscores.push(1)
                        this.setNotifications(newNoti)
                    } else {
                        var idx = this.notifications.senderids.indexOf(data.body.senderid)
                        let newNoti = Object.assign({}, this.notifications)
                        newNoti.senderscores[idx]++
                        this.setNotifications(newNoti)
                    }
                    break

                case "CHAT_HISTORY_RES":  
                    this.setChat(data.body)
                    break
                
                case "ALL_NOTIFICATIONS":
                    let newNotis = Object.assign({}, this.notifications)
                    newNotis.senderids = data.body.senderids
                    newNotis.senderscores = data.body.senderscores
                    newNotis.groupids = data.body.groupids
                    newNotis.groupscores = data.body.groupscores
                    newNotis.friendreq = data.body.friendreq
                    this.setNotifications(newNotis)
                    break
                
                case "FR_NOTI":
                    let newNoti = Object.assign({}, this.notifications)
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
                
            setTimeout(() => {
                this.connect()
            }, 5000)
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

    getPrivateChat(friendId) {
        var body = JSON.stringify({
            type: "PRIVATE_HISTORY_REQ",
            body: {
                id: friendId
            }
        })
        this.socket.send(body);
    }

    getGroupChat(groupId) {
        var body = JSON.stringify({
            type: "GROUP_HISTORY_REQ",
            body: {
                id: groupId
            }
        })
        this.socket.send(body);
    }
}

export default Xenosocket