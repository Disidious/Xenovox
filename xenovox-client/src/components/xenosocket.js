class Xenosocket{
    constructor() {
        this.socket = null
        this.manualDisconnect = false
        this.reconnected = false

        this.chat = null
        this.setChat = null
        this.setGroupMembers = null

        this.notifications = null
        this.setNotifications = null

        this.refreshed = null

        this.getGroups = () => {}
        this.getFriends = () => {}
        this.handleBeforeHistory = () => {}

        this.setState = null
    }

    connect() {
        if(this.socket !== null && this.socket.readyState === WebSocket.OPEN)
            return

        console.log('opened')
        this.socket = new WebSocket("ws://localhost:7777/socket");
        this.manualDisconnect = false
        
        this.socket.onopen = () => {
            this.setState("DONE")
            if(this.reconnected) {
                this.reconnected = false
                if(this.chat.chatid === -1)
                    return

                if(this.chat.group) {
                    this.getGroupChat(this.chat.chatid)
                } else {
                    this.getPrivateChat(this.chat.chatid)
                }
            }
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
                        let idx = this.notifications.senderids.indexOf(data.body.senderid)
                        let newNoti = Object.assign({}, this.notifications)
                        newNoti.senderscores[idx]++
                        this.setNotifications(newNoti)
                    }
                    break
                case "GM":
                    if(this.chat.chatid === data.body.message.groupid) {
                        if(data.body.members !== undefined) {
                            this.setGroupMembers(data.body.members)
                            console.log("came")
                        }

                        let newChat = Object.assign({}, this.chat)
                        newChat.history.push(data.body.message)
                        this.setChat(newChat)
                    } else if(!this.notifications.groupids.includes(data.body.message.groupid)) {
                        let newNoti = Object.assign({}, this.notifications)
                        newNoti.groupids.push(data.body.message.groupid)
                        newNoti.groupscores.push(1)
                        this.setNotifications(newNoti)
                    } else {
                        let idx = this.notifications.groupids.indexOf(data.body.groupid)
                        let newNoti = Object.assign({}, this.notifications)
                        newNoti.groupscores[idx]++
                        this.setNotifications(newNoti)
                    }
                    break
                case "CHAT_HISTORY_RES":  
                    console.log(data)
                    if(data.body.history.group) {
                        this.setGroupMembers(data.body.members)
                    }
                    this.setUnreadDivider(data.body.history.history.length)
                    this.setChat(data.body.history)

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
                
                case "REFRESH_GROUPS":
                    this.getGroups();
                    break

                default:
            }
            console.log(data)
        }

        this.socket.onclose = () => {
            if(this.manualDisconnect)
                return
            console.log('closed')
            setTimeout(() => {
                this.refreshed.current = false
                this.reconnected = true
                this.connect()
            }, 2000)
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

    sendGM(message, groupId) {
        var body = JSON.stringify({
            type: "GM",
            body: {
                groupid: parseInt(groupId),
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