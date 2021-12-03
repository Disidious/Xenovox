class Xenosocket {
    constructor() {
        this.socket = null
    }

    connect() {
        // Without this if condition the socket connects TWICE for some reason.
        if(this.socket != null) {
            return;
        }

        this.socket = new WebSocket("ws://localhost:7777/sendPM");

        this.socket.onopen = function(e) {
            //alert("[open] Connection established");
            //alert("Sending to server");
            //socket.send("My name is John");
        };
        
        this.socket.onmessage = function(event) {
            alert(`[message] Data received from server: ${event.data}`);
        };
        
        // this.socket.onclose = function(event) {
        //     if (event.wasClean) {
        //         alert(`[close] Connection closed cleanly, code=${event.code} reason=${event.reason}`);
        //     } else {
        //         // e.g. server process killed or network down
        //         // event.code is usually 1006 in this case
        //         alert('[close] Connection died');
        //     }
        // };

        this.socket.onerror = function(error) {
            alert(`[error] ${error.message}`);
        };
    }

    disconnect() {
        this.socket.close()
    }

    sendPM(message, receiverId) {
        var body = JSON.stringify({
            receiverId: parseInt(receiverId),
            message: message,
        })
        console.log(body)
        if(this.socket == null) {
            this.connect();
            setTimeout(() => {this.socket.send(body)}, 1000);
        } else {
            this.socket.send(body);
        }
    }
}

export default Xenosocket