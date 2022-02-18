import React from 'react';
import {ConfirmationModal} from './../components/modals'

class UserMenu extends React.Component {
    unfriendUser(chatId) {
        fetch(this.props.url + '/sendRelation', {
            credentials: 'include',
            method: 'POST',
            body: JSON.stringify({
                user2id: parseInt(chatId),
                relation: -1
            })
        }).then((res)=>{
            if(res.status !== 200)
                return
            //this.props.resetChat(chatId)
            this.props.refreshFriends(chatId)
        })
    }
    
    blockUser(chatId) {
        fetch(this.props.url + '/sendRelation', {
            credentials: 'include',
            method: 'POST',
            body: JSON.stringify({
                user2id: parseInt(chatId),
                relation: 2
            })
        }).then((res)=>{
            if(res.status !== 200)
                return
            //this.props.resetChat(chatId)
            this.props.refreshFriends(chatId)
        })
    }

    leaveGroup(chatId) {
        fetch(this.props.url + '/leave', {
            credentials: 'include',
            method: 'POST',
            body: JSON.stringify({
                groupid: parseInt(chatId)
            })
        }).then((res)=>{
            if(res.status !== 200)
                return
            //this.props.resetChat(chatId)
            this.props.refreshGroups(chatId)
        })
    }

    componentDidMount() {
        const menu = document.getElementById('userMenu');
        if(menu === null) {
            console.log("null here")
            return
        }
        menu.style.setProperty('--mouse-x', this.props.info.left + 'px')
        menu.style.setProperty('--mouse-y', this.props.info.top + 'px')
    }

    componentDidUpdate() {
        const menu = document.getElementById('userMenu');
        if(menu === null) {
            console.log("null here")
            return
        }
        menu.style.setProperty('--mouse-x', this.props.info.left + 'px')
        menu.style.setProperty('--mouse-y', this.props.info.top + 'px')
    }

    render() {
        return (
            <div 
            id="userMenu"
            className="user-menu" 
            style={{display: this.props.info.display}}>
                {
                    !this.props.info.group ?
                    <ul>
                        <li><button onClick={()=>{
                            this.props.showGroupInviteModal({
                                show: true, 
                                chatid: this.props.info.chatid})
                            }}>Add to Group</button></li>
                        <li><button onClick={()=>{
                            let chatId = this.props.info.chatid
                            this.props.setModalState({
                                show: true, 
                                title: 'Unfriend ' + this.props.info.chatname, 
                                content: 'Are you sure you want to unfriend ' + this.props.info.chatname + '?', 
                                actionName: 'Unfriend', 
                                action: ()=>{this.unfriendUser(chatId)}})
                            }}>Unfriend</button></li>
                        <li><button onClick={()=>{
                            let chatId = this.props.info.chatid
                            this.props.setModalState({
                                show: true, 
                                title: 'Block ' + this.props.info.chatname, 
                                content: 'Are you sure you want to block ' + this.props.info.chatname + '?', 
                                actionName: 'Block', 
                                action: ()=>{this.blockUser(chatId)}})
                            }}>Block</button></li>
                    </ul>
                    :
                    <ul>
                        <li><button>Show Members</button></li>
                        <li><button onClick={()=>{
                            let chatId = this.props.info.chatid
                            this.props.setModalState({
                                show: true, 
                                title: 'Leave ' + this.props.info.chatname, 
                                content: 'Are you sure you want to leave ' + this.props.info.chatname + '?', 
                                actionName: 'Leave', 
                                action: ()=>{this.leaveGroup(chatId)}})
                        }}>Leave Group</button></li>
                    </ul>
                    
                }
            </div>  
        );
    }
}

export default UserMenu;