/* eslint-disable jsx-a11y/anchor-is-valid */
import React, { useEffect, useState } from 'react';
import { Col, Row, Button, Modal } from 'react-bootstrap';
import { Navigate } from 'react-router-dom'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faLocationArrow, faUserPlus } from '@fortawesome/free-solid-svg-icons'

import AddFriendModal from './../components/modals'
import Spinner from "./../components/loadingspinner";

function sendDM(socket, friendId) {
    //socket.sendDM("message", 11)
    var messageBox = document.getElementById("message");

    if(messageBox.value === "") {
        return;
    }

    socket.sendDM(messageBox.value, friendId);
    document.getElementById("message").value = "";
}

function logout(setLoggedOut, url, socket) {
    fetch(url + '/logout', {
        credentials: 'include',
        method: 'POST'
    }).then(() => {
        socket.disconnect()
        setLoggedOut(true)
    })
    .catch((error) => {
        console.log(error)
    })
}

function getUserInfo(url, socket, setState, setInfo) {
    fetch(url + '/info', {
        credentials: 'include',
        method: 'GET'
    }).then(response => response.json())
    .then(data => {
        socket.userInfo = data
        setInfo(data)
        setState("DONE")
    })
    .catch((error) => {
        setState("FAILED")
        console.log(error)
    })
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
    .catch((error) => {
        console.log(error)
    })
}

function getChat(socket, friendId) {
    socket.getChat(friendId)
}

function Home(props) {
    const[state, setState] = useState("LOADING")
    const[socketState, setSocState] = useState("CONNECTING")

    const[userInfo, setInfo] = useState({id: -1, username: "", name: "", email: "", picture: ""})
    const[friends, setFriends] = useState([])
    const[chat, setChat] = useState({friendId: -1, history:[]})

    const[loggedOut, setLoggedOut] = useState(false)

    const [friendModalShow, setFriendModalShow] = React.useState(false);

    const handleEnter = (e) => {
        if(e.key !== 'Enter') {
            return
        }
        sendDM(props.socket, chat.friendId)
    }
    
    props.socket.chat = chat
    
    const calledOnce = React.useRef(false);
    useEffect(()=>{
        if(calledOnce.current){
            return
        }
        props.socket.setChat = setChat
        props.socket.setState = setSocState
        props.socket.connect()

        getUserInfo(props.url, props.socket, setState, setInfo)
        getFriends(props.url,setFriends)

        calledOnce.current = true
    }, [props.socket, props.url, chat])

    useEffect(()=>{
        var history = document.getElementById("history");
        if(history !== null) {
            history.scrollTop = history.scrollHeight;
        }
    },[chat])
    
    if(loggedOut){
        return (<Navigate to='/'/>);
    }

    if(state === 'LOADING' || socketState === 'CONNECTING') {
        return (
            <Spinner mode={"SCREEN"}/>
        );
    }

    return (
        <div className="home-container">
            <Row className="width-fix">
                <Col>
                    <Button type="button" className="btn-main" style={{float: "right"}} 
                    onClick={() => logout(setLoggedOut, props.url, props.socket)}>
                        Logout
                    </Button>
                    <Button type="button" className="btn-main" style={{float: "right"}} 
                    onClick={() => getChat(props.socket)}>
                        Temp
                    </Button>
                </Col>
            </Row>
            <Row className="width-fix">
                <Col className="col-10">
                    <div className="chat-container">
                        <div id="history" className="history-container">
                            {
                                chat.history.map((el, key) => (
                                    <p key={key} className="chat-msg">
                                        <b>
                                            {
                                                el.receiverId === userInfo.id ?
                                                friends.find(friend => friend.id === chat.friendId).username
                                                :
                                                userInfo.username
                                            }
                                        </b>
                                        <br/>
                                        {el.message}
                                        <br/>
                                    </p>
                                ))
                            }
                        </div>
                        <div className="message-container">
                            <input type="text" id="message" className="message-box textbox-main" onKeyPress={(e) => handleEnter(e)} 
                            disabled={chat.friendId === -1} autocomplete="off"/>

                            <Button type="button" className="btn-main btn-send " 
                            onClick={() => sendDM(props.socket, chat.friendId)}
                            disabled={chat.friendId === -1}>
                                <FontAwesomeIcon icon={faLocationArrow} />
                            </Button>
                        </div>
                    </div>
                </Col>
                <Col className="col-2">
                    <div className="friends-container">
                        <div style={{margin: "1em"}}>
                            <h1>Friends</h1>
                            <Button className="btn-small"
                            onClick={() => setFriendModalShow(true)}>
                                <FontAwesomeIcon icon={faUserPlus} size={'xs'} />
                            </Button>
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
                                    <Button className="friends-btn" key={key} onClick={()=>{getChat(props.socket, el.id)}}>
                                        {el.username}
                                    </Button>
                                ))
                            }
                        </div>
                    </div>
                </Col>
            </Row>

            <AddFriendModal 
            show={friendModalShow} 
            onHide={() => setFriendModalShow(false)} 
            url={props.url}
            getFriends={()=>{getFriends(props.url, setFriends)}}/>
        </div>
    );
}
export default Home;