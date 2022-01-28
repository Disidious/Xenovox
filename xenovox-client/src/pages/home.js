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

    if(messageBox.value.replaceAll(' ','') === "") {
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

function markAsRead(url, friendId, notifications, setNotifications) {
    if(!notifications.dms.includes(friendId))
        return

    fetch(url + '/read', {
        credentials: 'include',
        method: 'POST',
        body: JSON.stringify({
            id: friendId
        })
    }).then(() => {
        var notiIdx = notifications.dms.indexOf(friendId)
        if(notiIdx !== -1) {
            var newNotifications = Object.assign({}, notifications)
            newNotifications.dms.splice(notiIdx, 1)
            setNotifications(newNotifications)
        }
    })
}

function Home(props) {
    const[state, setState] = useState("LOADING")
    const[socketState, setSocState] = useState("CONNECTING")

    const[userInfo, setInfo] = useState({id: -1, username: "", name: "", email: "", picture: ""})
    const[friends, setFriends] = useState([])
    const[chat, setChat] = useState({friendid: -1, history:[]})
    const[notifications, setNotifications] = useState({dms: [], groups: [], friendreq: false})

    const[loggedOut, setLoggedOut] = useState(false)

    const[friendModalShow, setFriendModalShow] = React.useState(false);

    const handleEnter = (e) => {
        if(e.key !== 'Enter') {
            return
        }
        sendDM(props.socket, chat.friendid)
    }
    
    props.socket.chat = chat
    props.socket.notifications = notifications
    
    const calledOnce = React.useRef(false);
    useEffect(()=>{
        if(calledOnce.current){
            return
        }
        props.socket.setChat = setChat
        props.socket.setState = setSocState
        props.socket.setNotifications = setNotifications
        props.socket.getFriends = () => {getFriends(props.url, setFriends)}
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
                <Col style={{float: "left"}} >
                    <div className='user-info'>
                        <div>
                            <h1 style={{float: "left"}}>
                                {userInfo.username}
                            </h1>
                            <h1 style={{color: "rgb(51, 153, 255)"}}>
                                #{userInfo.id} 
                            </h1>
                        </div>
                    </div>
                </Col>
                <Col>
                    <Button type="button" className="btn-main" style={{float: "right"}} 
                    onClick={() => logout(setLoggedOut, props.url, props.socket)}>
                        Logout
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
                                                el.receiverid === userInfo.id ?
                                                friends.find(friend => friend.id === chat.friendid).username
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
                            disabled={chat.friendid === -1} autoComplete="off"/>

                            <Button type="button" className="btn-main btn-send " 
                            onClick={() => sendDM(props.socket, chat.friendid)}
                            disabled={chat.friendid === -1}>
                                <FontAwesomeIcon icon={faLocationArrow} />
                            </Button>
                        </div>
                    </div>
                </Col>
                <Col className="col-2">
                    <div className="friends-container">
                        <div style={{margin: "1em"}}>
                            <h1>Friends</h1>
                            <Button className={
                                notifications.friendreq ?
                                "glowing-btn btn-small"
                                :
                                "btn-small"
                            }
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
                                    <Button className={
                                        notifications.dms.includes(el.id) ? 
                                        "friends-btn glowing-btn" 
                                        : 
                                        "friends-btn"
                                    } key={key} onClick={()=>{
                                        if(chat.friendid !== el.id) {
                                            getChat(props.socket, el.id)
                                            if(notifications.dms.includes(el.id)) {
                                                markAsRead(props.url, el.id, notifications, setNotifications)
                                            }
                                        }
                                    }}>
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
            getFriends={()=>{getFriends(props.url, setFriends)}}
            removeNoti={()=>{
                var newNotifications = Object.assign({}, notifications)
                newNotifications.friendreq = false
                setNotifications(newNotifications)
            }}/>
        </div>
    );
}
export default Home;