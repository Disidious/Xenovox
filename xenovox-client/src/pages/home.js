/* eslint-disable jsx-a11y/anchor-is-valid */
import React, { useEffect, useState } from 'react';
import { Col, Row, Container, Button } from 'react-bootstrap';
import { Navigate } from 'react-router-dom'

import Spinner from "./../components/loadingspinner";

function sendPM(socket) {
    var rid = document.getElementById("rid").value
    var message = document.getElementById("message").value
    socket.sendPM(message, rid)
}

function logout(setLoggedOut, url, socket) {
    fetch(url + '/logout', {
        credentials: 'include',
        method: 'POST'
    }).then(() => {
        socket.disconnect()
        setLoggedOut(true)
    }
    )
}

function getFriends(url, setFriends) {
    fetch(url + '/friends', {
        credentials: 'include',
        method: 'GET'
    }).then(response => response.json())
    .then(data => {
        if(data.length === 0) {
            setFriends([{id:-1}])
            return
        }
        setFriends(data)
    })
}

function Home(props) {
    const[loggedOut, setLoggedOut] = useState(false)
    const[friends, setFriends] = useState([])
    const[chat, setChat] = useState([])

    useEffect(()=>{
        props.soc.connect()
        getFriends(props.url,setFriends)
    }, [props.soc])
    
    if(loggedOut){
        return (<Navigate to='/'/>);
    }
    return (
        <div className="home-container">
            <Row className="width-fix">
                <Col>
                    <Button type="button" className="btn-main" style={{float: "right"}} 
                    onClick={() => logout(setLoggedOut, props.url, props.soc)}>
                        Logout
                    </Button>
                </Col>
            </Row>
            <Row className="width-fix">
                <Col>
                    <div className="chat-container">

                    </div>
                </Col>
                <Col className="col-2">
                    <div className="friends-container">
                        <h1>Friends</h1>
                        {
                            friends.length === 0 ?
                            <center style={{paddingTop: "50%"}}>
                                <br/>
                                <Spinner/>
                            </center>
                            :
                            friends[0].id === -1 ?
                            <center style={{paddingTop: "50%"}}>
                                <br/>
                                <p className="info-message">
                                    You don't have any friends
                                </p>
                            </center>
                            :
                            friends.map((el, key) => (
                                <Button className="friends-btn" key={key} onClick={()=>{console.log(friends[key].username)}}>{el.username}
                                    <br/>
                                </Button>
                            ))
                        }
                    </div>
                </Col>
            </Row>
        </div>
        // <div>
        //     <div className="temp-div">
        //         <label>Id:</label><br/>
        //         <input type="text" id="rid" name="rid"/><br/>
        //         <label>Message:</label><br/>
        //         <input type="text" id="message" name="message"/><br/>
        //         <button onClick={() => sendPM(props.soc)}>Send</button><br/><br/>
        //         <button onClick={() => logout(setLoggedOut, props.url, props.soc)}>Logout</button><br/><br/>
        //     </div>
        // </div>
    );
}
export default Home;