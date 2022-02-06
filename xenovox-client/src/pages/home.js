/* eslint-disable jsx-a11y/anchor-is-valid */
import React, { useEffect, useState } from 'react';
import { Col, Row, Button } from 'react-bootstrap';
import { Navigate } from 'react-router-dom'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faLocationArrow, faUserPlus } from '@fortawesome/free-solid-svg-icons'

import AddFriendModal from './../components/modals'
import UserMenu from './../components/menus'
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

function markAsRead(url, chatId, isGroup, notifications, setNotifications) {
    fetch(url + '/read', {
        credentials: 'include',
        method: 'POST',
        body: JSON.stringify({
            group: isGroup,
            id: chatId
        })
    }).then(response => {
        if(response.status !== 200)
            return
        if(!isGroup) {
            let notiIdx = notifications.senderids.indexOf(chatId)
            if(notiIdx !== -1) {
                let newNotifications = Object.assign({}, notifications)
                newNotifications.senderids.splice(notiIdx, 1)
                newNotifications.senderscores.splice(notiIdx, 1)
                setNotifications(newNotifications)
            }
        } else {
            let notiIdx = notifications.groupids.indexOf(chatId)
            if(notiIdx !== -1) {
                let newNotifications = Object.assign({}, notifications)
                newNotifications.groupids.splice(notiIdx, 1)
                newNotifications.groupscores.splice(notiIdx, 1)
                setNotifications(newNotifications)
            }
        }
    })
}

function switchTabs(relationsTab, setRelationsTab, isGroup) {
    if(isGroup && !relationsTab)
        setRelationsTab(true);
    else if(!isGroup && relationsTab)
        setRelationsTab(false);
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

    const[userMenuProps, setUserMenuProps] = useState({display: 'none', top: -1, left: -1, group: false, chatid: -1})

    const[friendModalShow, setFriendModalShow] = useState(false);
    const[relationsTab, setRelationsTab] = useState(false);

    const handleContextMenu = (event, chatId, isGroup) => {
        event.preventDefault()
        let left = event.clientX
        let top = event.clientY
        let display = 'block'
        setUserMenuProps({display: display, top: top, left: left, group: isGroup, chatid: chatId})
    }

    const handleHistory = (id, isGroup) => {
        if(chat.chatid !== id) {
            getChat(props.socket, id, isGroup)
            if(!isGroup && notifications.senderids.includes(id)) {
                markAsRead(props.url, id, false, notifications, setNotifications)
            } else if(isGroup && notifications.groupids.includes(id)) {
                markAsRead(props.url, id, true, notifications, setNotifications)
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
        props.socket.refreshed = calledOnce
        props.socket.getFriends = () => {getFriends(props.url, setFriends)}
        props.socket.connect()

        getUserInfo(props.url, props.socket, setState, setInfo)
        getConnections(props.url, setFriends, setGroups)

        document.addEventListener('click', (event) => {
            let userMenu = document.getElementById("userMenu")
            if(userMenu === null || userMenu.style.display === 'none') {
                return
            }

            setUserMenuProps({display: 'none', top: -1, left: -1, userid: -1})
        })

        calledOnce.current = true
    }, [props.socket, props.url, chat, userMenuProps])

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
        <div>
            <UserMenu
            url={props.url}
            info={userMenuProps}
            refreshFriends={() => {getFriends(props.url, setFriends)}}
            />
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
                            {
                                chat.chatid !== -1 ?
                                <div className='user-info-chat'>
                                    <div>
                                        <h1 style={{float: "left", fontSize: "large!important"}}>
                                            {
                                                !chat.group ?
                                                "@" + friends.find(friend => friend.id === chat.chatid).username
                                                :
                                                chat.group ?
                                                "@" + groups.find(group => group.id === chat.chatid).name
                                                :
                                                ""
                                            }
                                        </h1>
                                        <h1 style={{color: "rgb(51, 153, 255)"}}>
                                            {
                                                "#" + chat.chatid
                                            } 
                                        </h1>
                                    </div>
                                </div>
                                :
                                ""
                            }
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
                            <div className="tab-container">
                                {
                                    !relationsTab ?
                                    <h1>Friends</h1>
                                    :
                                    <h1>Groups</h1>
                                }
                                <Button className={
                                    notifications.friendreq ?
                                    "glowing-btn btn-small"
                                    :
                                    "btn-small"
                                }
                                onClick={() => setFriendModalShow(true)}>
                                    <FontAwesomeIcon icon={faUserPlus} size={'xs'} />
                                </Button>
                                <div className="tab-list scrollable">
                                {
                                    !relationsTab ?
                                    // Friends here
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
                                            "relation-btn glowing-btn" 
                                            : 
                                            "relation-btn"
                                        } key={key} onClick={()=>{
                                            handleHistory(el.id, false)
                                        }}
                                        onContextMenu={event=>handleContextMenu(event, el.id, false)}
                                        disabled={el.id === chat.chatid}>
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
                                    :
                                    // Groups here
                                    groups.length === 0 ?
                                    <center style={{paddingTop: "50%"}}>
                                        <br/>
                                        <Spinner/>
                                    </center>
                                    :
                                    groups[0].id === -1 ?
                                    <center style={{paddingTop: "50%"}}>
                                        <br/>
                                        <p className="info-message">
                                            You are not in any group
                                        </p>
                                    </center>
                                    :
                                    groups.map((el, key) => (
                                        <button className={
                                            notifications.groupids.includes(el.id) ? 
                                            "relation-btn glowing-btn" 
                                            : 
                                            "relation-btn"
                                        } key={key} onClick={()=>{
                                            handleHistory(el.id, true)
                                        }}
                                        onContextMenu={event=>handleContextMenu(event, el.id, true)}
                                        disabled={el.id === chat.chatid}>
                                            {el.name}
                                            {
                                                notifications.groupids.includes(el.id) ?
                                                <div className="unread-noti">
                                                    {notifications.groupscores[notifications.groupids.indexOf(el.id)]}
                                                </div>
                                                :
                                                ""
                                            }
                                        </button>
                                    ))
                                }
                                </div>
                            </div>
                            <div className="relations-tabs">
                                <button className='friends-tab' disabled={!relationsTab}
                                onClick={()=>switchTabs(relationsTab, setRelationsTab, false)}>Friends</button>
                                
                                <button className='groups-tab' disabled={relationsTab}
                                onClick={()=>switchTabs(relationsTab, setRelationsTab, true)}>Groups</button>
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
        </div>
    );
}
export default Home;