/* eslint-disable jsx-a11y/anchor-is-valid */
import React, { useEffect, useState } from 'react';
import { Col, Row, Button } from 'react-bootstrap';
import { Navigate } from 'react-router-dom'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faLocationArrow, faUserPlus } from '@fortawesome/free-solid-svg-icons'

import AddFriendModal from './../components/modals'
import Usermenu from './../components/menus'
import Spinner from "./../components/loadingspinner";

function sendDM(socket, friendId) {
    var messageBox = document.getElementById("message");

    if(messageBox.value.replaceAll(' ','') === "") {
        return;
    }

    socket.sendDM(messageBox.value, friendId);
    document.getElementById("message").value = "";
}

function sendGM(socket, groupId) {
    var messageBox = document.getElementById("message");

    if(messageBox.value.replaceAll(' ','') === "") {
        return;
    }

    socket.sendGM(messageBox.value, groupId);
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
    }).then(response => {
        if(response.status !== 200)
            return null
        return response.json()
    })
    .then(data => {
        socket.userInfo = data
        setInfo(data)
        setState("DONE")
    })
}

function getFriends(url, setFriends) {
    fetch(url + '/friends', {
        credentials: 'include',
        method: 'GET'
    }).then(response => {
        if(response.status !== 200)
            return null
        return response.json()
    })
    .then(data => {
        if(data === null)
            return

        if(data.length === 0) {
            setFriends([{id:-1}])
        } else {
            setFriends(data)
        }
    })
    .catch((error) => {
        console.log(error)
    })
}

function getConnections(url, setFriends, setGroups) {
    fetch(url + '/connections', {
        credentials: 'include',
        method: 'GET'
    }).then(response => {
        if(response.status !== 200)
            return null
        return response.json()
    })
    .then(data => {
        if(data === null)
            return

        if(data.friends.length === 0) {
            setFriends([{id:-1}])
        } else {
            setFriends(data.friends)
        }

        if(data.groups.length === 0) {
            setGroups([{id:-1}])
        } else {
            setGroups(data.groups)
        }
    })
    .catch((error) => {
        console.log(error)
    })
}

function getChat(socket, chatId, group) {
    if(!group) {
        socket.getPrivateChat(chatId)
        return
    }
    
    socket.getGroupChat(chatId)
}

function markAsRead(url, friendId, notifications, setNotifications) {
    if(!notifications.senderids.includes(friendId))
        return

    fetch(url + '/read', {
        credentials: 'include',
        method: 'POST',
        body: JSON.stringify({
            id: friendId
        })
    }).then(response => {
        if(response.status !== 200)
            return

        var notiIdx = notifications.senderids.indexOf(friendId)
        if(notiIdx !== -1) {
            var newNotifications = Object.assign({}, notifications)
            newNotifications.senderids.splice(notiIdx, 1)
            newNotifications.senderscores.splice(notiIdx, 1)
            setNotifications(newNotifications)
        }
    })
}

function Home(props) {
    const[state, setState] = useState("LOADING")
    const[socketState, setSocState] = useState("CONNECTING")

    const[userInfo, setInfo] = useState({id: -1, username: "", name: "", email: "", picture: ""})
    const[friends, setFriends] = useState([])
    const[groups, setGroups] = useState([])
    const[groupMembers, setGroupMembers] = useState([])
    const[chat, setChat] = useState({group: false, chatid: -1, history: []})
    const[notifications, setNotifications] = useState({senderids: [], senderscores: [], groupids: [], groupscores: [], friendreq: false})

    const[loggedOut, setLoggedOut] = useState(false)

    const[userMenuProps, setUserMenuProps] = useState({display: 'none', top: '0', left: '0', userid: -1})

    const[friendModalShow, setFriendModalShow] = useState(false);

    const handleHistory = (id, isGroup) => {
        if(chat.chatid !== id) {
            getChat(props.socket, id, isGroup)
            if(notifications.senderids.includes(id) && !isGroup) {
                markAsRead(props.url, id, notifications, setNotifications)
            }
        }
    }
    const handleMsg = () => {
        if(!chat.group)
            sendDM(props.socket, chat.chatid)
        else
            sendGM(props.socket, chat.chatid)
    }
    const handleEnter = (e) => {
        if(e.key !== 'Enter') {
            return
        }
        handleMsg()
    }
    
    props.socket.chat = chat
    props.socket.notifications = notifications
    
    const calledOnce = React.useRef(false);
    useEffect(()=>{
        if(calledOnce.current){
            return
        }
        props.socket.setChat = setChat
        props.socket.setGroupMembers = setGroupMembers
        props.socket.setState = setSocState
        props.socket.setNotifications = setNotifications
        props.socket.getFriends = () => {getFriends(props.url, setFriends)}
        props.socket.connect()

        getUserInfo(props.url, props.socket, setState, setInfo)
        //getFriends(props.url,setConnections)
        getConnections(props.url, setFriends, setGroups)

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
            <div>
                <Spinner mode={"SCREEN"}/>
                {
                    state === 'UNEXPECTED_FAILURE' ?
                    <center>
                        <p className="error-message">
                            Wrong username or password
                            <br/>
                            Please try again
                        </p>
                    </center>
                    :
                    ""
                }
            </div>
        );
    }

    console.log("Refreshed")
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
                        onClick={() => handleHistory(1, true)}>
                        Temp
                    </Button>
                    <Button type="button" className="btn-main" style={{float: "right"}} 
                    onClick={() => logout(setLoggedOut, props.url, props.socket)}>
                        Logout
                    </Button>
                </Col>
            </Row>
            <Row className="width-fix">
                <Col className="col-10">
                    <div className="chat-container">
                        <div id="history" className="history-container scrollable">
                            {
                                chat.history.map((el, key) => (
                                    <p key={key} className="chat-msg">
                                        {
                                            !chat.group ?
                                            <b>
                                                {
                                                    el.senderid === userInfo.id ?
                                                    userInfo.username
                                                    :
                                                    friends.find(friend => friend.id === el.senderid).username
                                                }
                                            </b>
                                            :
                                            <b>
                                                {
                                                    el.senderid === userInfo.id ?
                                                    userInfo.username
                                                    :
                                                    groupMembers.find(member => member.id === el.senderid).username
                                                }
                                            </b>
                                        }
                                        <br/>
                                        {el.message}
                                        <br/>
                                    </p>
                                ))
                            }
                        </div>
                        <div className="message-container">
                            <input type="text" id="message" className="message-box textbox-main" onKeyPress={(e) => handleEnter(e)} 
                            disabled={chat.chatid === -1} autoComplete="off"/>

                            <Button type="button" className="btn-main btn-send " 
                            onClick={() => {
                                handleMsg()
                            }}
                            disabled={chat.chatid === -1}>
                                <FontAwesomeIcon icon={faLocationArrow} />
                            </Button>
                        </div>
                    </div>
                </Col>
                <Col className="col-2">
                    <div className="relations-container">
                        <div className="friends-container">
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
                            <div className="friend-list scrollable">
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
                                    <button className={
                                        notifications.senderids.includes(el.id) ? 
                                        "friends-btn glowing-btn" 
                                        : 
                                        "friends-btn"
                                    } key={key} onClick={()=>{
                                        handleHistory(el.id, false)
                                    }}
                                    onContextMenu={()=>{}}>
                                        {el.username}
                                        {
                                            notifications.senderids.includes(el.id) ?
                                            <div className="unread-noti">
                                                {notifications.senderscores[notifications.senderids.indexOf(el.id)]}
                                            </div>
                                            :
                                            ""
                                        }
                                    </button>
                                ))
                            }
                            </div>
                        </div>
                    </div>
                </Col>
            </Row>

            <AddFriendModal 
            show={friendModalShow} 
            onHide={() => setFriendModalShow(false)} 
            url={props.url}
            getFriends={()=>{getFriends(props.url, setFriends, setGroups)}}
            removeNoti={()=>{
                var newNotifications = Object.assign({}, notifications)
                newNotifications.friendreq = false
                setNotifications(newNotifications)
            }}/>
        </div>
    );
}
export default Home;