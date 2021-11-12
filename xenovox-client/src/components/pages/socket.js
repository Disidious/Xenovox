import React from 'react';
import { Col, Row } from 'react-bootstrap';

let socket = null

function connect() {
    socket = new WebSocket("ws://localhost:7777/send");

    socket.onopen = function(e) {
        alert("[open] Connection established");
        //alert("Sending to server");
        //socket.send("My name is John");
    };
    
    socket.onmessage = function(event) {
        alert(`[message] Data received from server: ${event.data}`);
    };
    
    socket.onclose = function(event) {
    if (event.wasClean) {
        alert(`[close] Connection closed cleanly, code=${event.code} reason=${event.reason}`);
    } else {
        // e.g. server process killed or network down
        // event.code is usually 1006 in this case
        alert('[close] Connection died');
    }
    };

    socket.onerror = function(error) {
        alert(`[error] ${error.message}`);
    };
}

function send() {
    var body = JSON.stringify({
        receiverId: parseInt(document.getElementById("rid").value),
        message: document.getElementById("message").value,
    })
    if(socket == null) {
        connect();
        setTimeout(function(){socket.send(body)}, 1000);
    } else {
        socket.send(body);
    }
}

function SocketPage() {
    return (
        <div>
            <div className="temp-div">
                <label>Id:</label><br/>
                <input type="text" id="rid" name="rid"/><br/>
                <label>Message:</label><br/>
                <input type="text" id="message" name="message"/><br/><br/>
                <button onClick={send}>Send</button>
                <a href="/">
                    <button>Back</button>
                </a>
            </div>
        </div>
    );
}
export default SocketPage;