import React from 'react';

class UserMenu extends React.Component {
    blockUser() {
        fetch(this.props.url + '/sendRelation', {
            credentials: 'include',
            method: 'POST',
            body: JSON.stringify({
                user2Id: parseInt(this.props.info.userid),
                relation: 2
            })
        }).then(()=>{
            this.props.refreshFriends()
        })
    }

    unfriendUser() {
        fetch(this.props.url + '/sendRelation', {
            credentials: 'include',
            method: 'POST',
            body: JSON.stringify({
                user2Id: parseInt(this.props.info.userid),
                relation: -1
            })
        }).then(()=>{
            this.props.refreshFriends()
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
                <ul>
                    <li><button>Invite to Group</button></li>
                    <li><button onClick={()=>this.unfriendUser()}>Unfriend</button></li>
                    <li><button onClick={()=>this.blockUser()}>Block</button></li>
                </ul>
            </div>  
        );
    }
}

export default UserMenu;