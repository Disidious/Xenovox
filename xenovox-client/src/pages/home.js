/* eslint-disable jsx-a11y/anchor-is-valid */
import React, { useEffect, useState } from 'react'
import { Col, Row, Button } from 'react-bootstrap'
import { Navigate } from 'react-router-dom'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faLocationArrow, faUserPlus, faPlus } from '@fortawesome/free-solid-svg-icons'

import {ConfirmationModal, AddFriendModal, GroupInviteModal, CreateGroupModal} from './../components/modals'
import UserMenu from './../components/menus'
import Spinner from "./../components/loadingspinner"

function sendDM(socket, friendId) {
    var messageBox = document.getElementById("message")

    if(messageBox.value.replaceAll(' ','') === "") {
        return
    }

    socket.sendDM(messageBox.value, friendId)
    document.getElementById("message").value = ""
}

function sendGM(socket, groupId) {
    var messageBox = document.getElementById("message")

    if(messageBox.value.replaceAll(' ','') === "") {
        return
    }

    socket.sendGM(messageBox.value, groupId)
    document.getElementById("message").value = ""
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

function getGroups(url, setGroups) {
    fetch(url + '/groups', {
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
            setGroups([])
        } else {
            setGroups(data)
        }
    })
    .catch((error) => {
        console.log(error)
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
            setFriends([])
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
            setFriends([])
        } else {
            setFriends(data.friends)
        }

        if(data.groups.length === 0) {
            setGroups([])
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
        setRelationsTab(true)
    else if(!isGroup && relationsTab)
        setRelationsTab(false)
}

function Home(props) {
    const[state, setState] = useState("LOADING")
    const[socketState, setSocState] = useState("CONNECTING")

    const[userInfo, setInfo] = useState({id: -1, username: "", name: "", email: "", picture: ""})
    const[friends, setFriends] = useState([{id:-1}])
    const[groups, setGroups] = useState([{id:-1}])
    const[groupMembers, setGroupMembers] = useState([])
    const[chat, setChat] = useState({group: false, chatid: -1, history: []})
    const[notifications, setNotifications] = useState({senderids: [], senderscores: [], groupids: [], groupscores: [], friendreq: false})
    const newDividerIdx = React.useRef(-1)
    const prevUnreadScore = React.useRef(-1)

    const[loggedOut, setLoggedOut] = useState(false)

    const[userMenuProps, setUserMenuProps] = useState({display: 'none', top: -1, left: -1, group: false, chatid: -1, chatname: ''})

    const[confirmationModalInfo, setConfirmationModalInfo] = useState({show: false, title: '', content: '', actionName: '', action: ()=>{}})
    const[friendModalShow, setFriendModalShow] = useState(false)
    const[groupInviteModalInfo, setGroupInviteModalInfo] = useState({show: false, chatid: -1})
    const[createGroupModalShow, setCreateGroupModalShow] = useState(false)
    const[relationsTab, setRelationsTab] = useState(false)

    const handleContextMenu = (event, chatId, isGroup, chatName) => {
        event.preventDefault()
        let left = event.clientX
        let top = event.clientY
        let display = 'block'
        setUserMenuProps({display: display, top: top, left: left, group: isGroup, chatid: chatId, chatname: chatName})
    }

    const handleHistory = (id, isGroup) => {
        if(chat.chatid !== id || chat.group !== isGroup) {
            newDividerIdx.current = -1
            prevUnreadScore.current = -1
            
            if(!isGroup && notifications.senderids.includes(id)) {
                let idx = notifications.senderids.indexOf(id)
                prevUnreadScore.current = notifications.senderscores[idx]
                markAsRead(props.url, id, false, notifications, setNotifications)
            } else if(isGroup && notifications.groupids.includes(id)) {
                let idx = notifications.groupids.indexOf(id)
                prevUnreadScore.current = notifications.groupscores[idx]
                markAsRead(props.url, id, true, notifications, setNotifications)
            }
            getChat(props.socket, id, isGroup)
        }
    }

    const setUnreadDivider = (len) => {
        if(newDividerIdx.current === -1 && prevUnreadScore.current !== -1) {
            newDividerIdx.current = len - prevUnreadScore.current
        } else {
            newDividerIdx.current = -1
        }
    }

    const handleMsg = () => {
        newDividerIdx.current = -1
        prevUnreadScore.current = -1

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

    const setFriendInfo = (friendInfo, rem) => {
        let newFriends = [...friends]
        if(rem === true) {
            let idx = newFriends.findIndex(friend => friend.id === friendInfo)
            newFriends.splice(idx, 1)

            let notiIdx = notifications.senderids.indexOf(friendInfo)
            if(notiIdx !== -1) {
                let newNotifications = Object.assign({}, notifications)
                newNotifications.senderids.splice(notiIdx, 1)
                newNotifications.senderscores.splice(notiIdx, 1)
                setNotifications(newNotifications)
            }
            if(chat.chatid === friendInfo){
                setChat({group: false, chatid: -1, history: []})
            }
        } else if(rem === false) {
            if(newFriends.findIndex(friend => friend.id === friendInfo.id) === -1) {
                newFriends.push(friendInfo)
            }
        } else {
            let idx = newFriends.findIndex(friend => friend.id === friendInfo.id)
            newFriends[idx] = friendInfo
        }

        setFriends(newFriends)
    }

    const setGroupInfo = (groupInfo, rem) => {
        let newGroups = [...groups]
        if(rem === true) {
            let idx = newGroups.findIndex(group => group.id === groupInfo)
            console.log(idx)
            console.log(groupInfo)
            newGroups.splice(idx, 1)

            let notiIdx = notifications.groupids.indexOf(groupInfo)
            if(notiIdx !== -1) {
                let newNotifications = Object.assign({}, notifications)
                newNotifications.groupids.splice(notiIdx, 1)
                newNotifications.groupscores.splice(notiIdx, 1)
                setNotifications(newNotifications)
            }
            if(chat.chatid === groupInfo){
                setChat({group: false, chatid: -1, history: []})
            }
        } else if(rem === false) {
            if(newGroups.findIndex(group => group.id === groupInfo.id) === -1) {
                newGroups.push(groupInfo)
            }
        } else {
            let idx = newGroups.findIndex(group => group.id === groupInfo.id)
            newGroups[idx] = groupInfo
        }

        setGroups(newGroups)
    }

    props.socket.chat = chat
    props.socket.notifications = notifications
    
    const calledOnce = React.useRef(false)
    useEffect(()=>{
        if(calledOnce.current){
            return
        }
        props.socket.setChat = setChat
        props.socket.setGroupMembers = setGroupMembers
        props.socket.setState = setSocState
        props.socket.setNotifications = setNotifications
        props.socket.setUnreadDivider = setUnreadDivider
        props.socket.refreshed = calledOnce
        props.socket.connect()

        getUserInfo(props.url, props.socket, setState, setInfo)
        getConnections(props.url, setFriends, setGroups)

        document.addEventListener('click', (event) => {
            let userMenu = document.getElementById("userMenu")
            if(userMenu === null || userMenu.style.display === 'none') {
                return
            }

            setUserMenuProps({display: 'none', top: -1, left: -1, group: false, chatid: -1, chatname: ''})
        })

        calledOnce.current = true
    }, [props.socket, props.url])

    useEffect(()=>{
        var history = document.getElementById("history")
        if(history !== null) {
            history.scrollTop = history.scrollHeight
        }
    },[chat])
    
    useEffect(()=>{
        props.socket.setGroupInfo = setGroupInfo
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [groups, chat])

    useEffect(()=>{
        props.socket.setFriendInfo = setFriendInfo
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [friends, chat])

    if(loggedOut){
        return (<Navigate to='/'/>)
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
                    null
                }
            </div>
        )
    }

    console.log("Refreshed")
    return (
        <div>
            <UserMenu
            url={props.url}
            info={userMenuProps}
            setModalState={setConfirmationModalInfo}
            refreshFriends={(chatId) => {
                setFriendInfo(chatId, true)
            }}
            refreshGroups={(chatId) => {
                setGroupInfo(chatId, true)
            }}
            resetChat={(chatId) => {
                if(chat.chatid === chatId){
                    setChat({group: false, chatid: -1, history: []})
                }
            }}
            showGroupInviteModal={setGroupInviteModalInfo}
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
                                                null
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
                                null
                            }
                            <div id="history" className="history-container scrollable">
                                {
                                    chat.history.map((el, key) => (
                                        <div key={key}>
                                            {
                                                key === newDividerIdx.current ?
                                                <div className="messages-divider">
                                                    <p><span>NEW</span></p>
                                                </div>
                                                :
                                                null
                                            }
                                            <p className={
                                                el.issystem ?
                                                "chat-msg info-message"
                                                :
                                                "chat-msg"}>
                                                {
                                                    el.issystem ?
                                                    null
                                                    :
                                                    <b>
                                                        {el.username}
                                                        <br/>
                                                    </b>
                                                }
                                                {el.message}
                                                <br/>
                                            </p>
                                        </div>
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
                                {
                                    relationsTab ?
                                    <Button className="btn-small"
                                    onClick={() => setCreateGroupModalShow(true)}>
                                        <FontAwesomeIcon icon={faPlus} size={'xs'} />
                                    </Button>
                                    :
                                    <Button className={
                                        notifications.friendreq ?
                                        "glowing-btn btn-small"
                                        :
                                        "btn-small"
                                    }
                                    onClick={() => setFriendModalShow(true)}>
                                        <FontAwesomeIcon icon={faUserPlus} size={'xs'} />
                                    </Button>
                                }

                                <div className="tab-list scrollable">
                                {
                                    !relationsTab ?
                                    // Friends here
                                    friends.length === 0 ?
                                    <center style={{paddingTop: "50%"}}>
                                        <br/>
                                        <p className="info-message">
                                            You don't have any friends
                                        </p>
                                    </center>
                                    :
                                    friends[0].id === -1 ?
                                    <div style={{paddingTop: "50%"}}>
                                        <br/>
                                        <Spinner/>
                                    </div>
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
                                        onContextMenu={event=>handleContextMenu(event, el.id, false, el.username)}
                                        disabled={el.id === chat.chatid && !chat.group}>
                                            {el.username}
                                            {
                                                notifications.senderids.includes(el.id) ?
                                                <div className="unread-noti">
                                                    {notifications.senderscores[notifications.senderids.indexOf(el.id)]}
                                                </div>
                                                :
                                                null
                                            }
                                        </button>
                                    ))
                                    :
                                    // Groups here
                                    groups.length === 0 ?
                                    <center style={{paddingTop: "50%"}}>
                                        <br/>
                                        <p className="info-message">
                                            You are not in any group
                                        </p>
                                    </center>
                                    :
                                    groups[0].id === -1 ?
                                    <div style={{paddingTop: "50%"}}>
                                        <br/>
                                        <Spinner/>
                                    </div>
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
                                        onContextMenu={event=>handleContextMenu(event, el.id, true, el.name)}
                                        disabled={el.id === chat.chatid && chat.group}>
                                            {el.name}
                                            {
                                                notifications.groupids.includes(el.id) ?
                                                <div className="unread-noti">
                                                    {notifications.groupscores[notifications.groupids.indexOf(el.id)]}
                                                </div>
                                                :
                                                null
                                            }
                                        </button>
                                    ))
                                }
                                </div>
                            </div>
                            <div className="relations-tabs">
                                <button className={
                                    notifications.senderids.length !== 0 && relationsTab ?
                                    "friends-tab glowing-btn"
                                    :
                                    "friends-tab"
                                } disabled={!relationsTab}
                                onClick={()=>switchTabs(relationsTab, setRelationsTab, false)}>Friends</button>
                                
                                <button className={
                                    notifications.groupids.length !== 0 && !relationsTab ?
                                    "groups-tab glowing-btn"
                                    :
                                    "groups-tab"
                                } disabled={relationsTab}
                                onClick={()=>switchTabs(relationsTab, setRelationsTab, true)}>Groups</button>
                            </div>
                        </div>
                    </Col>
                </Row>
                {
                    confirmationModalInfo.show ?
                    <ConfirmationModal
                    info={confirmationModalInfo}
                    hide={()=>setConfirmationModalInfo({show: false, title: '', content: '', actionName: '', action: ()=>{}})}
                    />
                    :
                    null
                }
                {
                    friendModalShow ?
                    <AddFriendModal 
                    hide={() => setFriendModalShow(false)} 
                    url={props.url}
                    getFriends={()=>{getFriends(props.url, setFriends)}}
                    removeNoti={()=>{
                        var newNotifications = Object.assign({}, notifications)
                        newNotifications.friendreq = false
                        setNotifications(newNotifications)
                    }}/>
                    :
                    null
                }

                {
                    groupInviteModalInfo.show ?
                    <GroupInviteModal
                    hide={() => setGroupInviteModalInfo({show: false, chatid: -1})}
                    url={props.url}
                    info={groupInviteModalInfo}
                    groups={groups}/>
                    :
                    null
                }
                {
                    createGroupModalShow ?
                    <CreateGroupModal
                    hide={() => setCreateGroupModalShow(false)}
                    url={props.url}
                    friends={friends}/>
                    :
                    null
                }

            </div>
        </div>
    )
}
export default Home