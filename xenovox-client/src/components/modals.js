import React from 'react';
import {Modal, Button} from 'react-bootstrap';
import Spinner from "./../components/loadingspinner";

class CreateGroupModal extends React.Component {
    constructor(props) {
        super(props)
        this.state = {
            value: "STANDBY",
            memberstoadd: []
        }
    }

    addMember(id) {
        this.state.memberstoadd.push(id)
        this.setState({value: "STANDBY", memberstoadd: this.state.memberstoadd})
        console.log(this.state.memberstoadd)
    }

    remMember(id) {
        let idx = this.state.memberstoadd.indexOf(id)
        this.state.memberstoadd.splice(idx, 1)
        this.setState({value: "STANDBY", memberstoadd: this.state.memberstoadd})
    }

    isAFriend(id) {
        return this.props.friends.findIndex(friend => friend.id === id) !== -1
    }

    createGroup(group) {
        if(group.name === '') {
            return
        }

        this.state.memberstoadd.forEach(member => !this.isAFriend(member) ? this.remMember(member) : null)

        fetch(this.props.url + '/createGroup', {
            credentials: 'include',
            method: 'POST',
            body: JSON.stringify({
                group: group,
                members: this.state.memberstoadd
            })
        }).then(response => {
            if(response.status !== 200) {
                this.setState({value: "FAILED", memberstoadd: this.state.memberstoadd})
                return
            }
            this.props.hide()
        }).catch((error) => {
            console.log(error)
        })
    }

    render() {
        return (
            <Modal
            show={true}
            size="md"
            aria-labelledby="contained-modal-title-vcenter"
            centered>
                <Modal.Header>
                    <Modal.Title id="contained-modal-title-vcenter">
                        Create Group
                    </Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <p className="field-title"><b>Group Name</b></p>
                    <input id="groupName" className='textbox-main'/>
                    {
                        this.state.value === "FAILED" ?
                        <p className="error-message">
                            Couldn't create group
                        </p>
                        :
                        <p className="error-message">
                            &nbsp;
                        </p>
                    }
                    <div className="scrollable table-container">
                        <table style={{width: "100%"}}>
                            <tbody className="table">
                                <tr>
                                    <th>Username#Id</th>
                                </tr>
                                {
                                    this.props.friends.length === 0 ?
                                    <tr>
                                        <td>
                                            <p className="info-message">
                                                <br/>
                                                You don't have any friends
                                            </p>
                                        </td>
                                    </tr>
                                    :
                                    this.props.friends.map((el, key) => (
                                        <tr key={key}>
                                            <td>
                                                {el.username}#{el.id}
                                            </td>
                                            <td style={{float: 'right'}}>
                                                {
                                                    this.state.memberstoadd.includes(el.id) ?
                                                    <Button className="btn-main btn-cancel table-btn"
                                                    onClick={()=>{ this.remMember(el.id) }}>Remove</Button>
                                                    :
                                                    <Button className="btn-main btn-confirm table-btn"
                                                    onClick={()=>{ this.addMember(el.id) }}>Add</Button>
                                                }
                                            </td>
                                        </tr>
                                    ))
                                }
                            </tbody>
                        </table>
                    </div>
                </Modal.Body>
                <Modal.Footer>
                    <Button className="btn-main btn-confirm" onClick={() => 
                        this.createGroup({name: document.getElementById("groupName").value})
                        }>Create</Button>
                    <Button className="btn-main btn-cancel" onClick={() => {
                        this.props.hide()
                        }}>Close</Button>
                </Modal.Footer>
            </Modal>
        )
    }
}

class GroupInviteModal extends React.Component {
    constructor(props) {
        super(props)
        this.state = {value: "STANDBY", groupname: ""}
    }

    addToGroup(groupId, groupName) {
        fetch(this.props.url + '/addToGroup', {
            credentials: 'include',
            method: 'POST',
            body: JSON.stringify({
                friendids: [parseInt(this.props.info.chatid)],
                groupid: groupId
            })
        }).then( response => response.json() )
        .then(data => {
            this.setState({value: data.message, groupname: groupName})
        }).catch((error) => {
            console.log(error)
        })
    }

    render() {
        return (
            <Modal
            show={true}
            size="md"
            aria-labelledby="contained-modal-title-vcenter"
            centered>
                <Modal.Header>
                    <Modal.Title id="contained-modal-title-vcenter">
                        Groups
                    </Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <div className="scrollable table-container">
                        <table style={{width: "100%"}}>
                            <tbody className="table">
                                <tr>
                                    <th>Name</th>
                                </tr>
                                {
                                    this.props.groups.length === 0 ?
                                    <tr>
                                        <td>
                                            <p className="info-message">
                                                <br/>
                                                You are not in any group
                                            </p>
                                        </td>
                                    </tr>
                                    :
                                    this.props.groups.map((el, key) => (
                                        <tr key={key}>
                                            <td>
                                                {el.name}
                                            </td>
                                            <td style={{float: 'right'}}>
                                                <Button className="btn-main btn-confirm table-btn"
                                                onClick={()=>{ this.addToGroup(el.id, el.name) }}>Add</Button>
                                            </td>
                                        </tr>
                                    ))
                                }
                            </tbody>
                        </table>
                    </div>
                    {
                        this.state.value === "UNEXPECTED_FAILURE" ?
                        <p className="error-message">
                            Couldn't add your friend to {this.state.groupname}
                        </p>
                        :
                        this.state.value === "SUCCESS" ?
                        <p className="success-message">
                            Friend added to {this.state.groupname}
                        </p>
                        :
                        <p className="success-message">&nbsp;</p>
                    }
                </Modal.Body>
                <Modal.Footer>
                    <Button className="btn-main btn-cancel" onClick={() => {
                        this.props.hide()
                        }}>Close</Button>
                </Modal.Footer>
            </Modal>
        );
    }
}

class AddFriendModal extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            frState: "STANDBY",
            requestsState: "STANDBY",
            requests: []
        };
    }

    addFriend() {
        var friendIdBox = document.getElementById('addUserId')
        if(friendIdBox.value === '')
            return

        fetch(this.props.url + '/sendRelation', {
            credentials: 'include',
            method: 'POST',
            body: JSON.stringify({
                user2Id: parseInt(friendIdBox.value),
                relation: 0
            })
        }).then(response => response.json())
        .then(data => {
            this.setState({frState: data.message})
        })
    }

    getFriendRequests() {
        this.setState({frState: this.state.frState, requestsState: "LOADING"})
        fetch(this.props.url + '/friendRequests', {
            credentials: 'include',
            method: 'GET'
        }).then(response => response.json())
        .then(data => {
            if(data.length === 0)
                this.props.removeNoti()
            this.setState({frState: this.state.frState, requestsState: "SUCCESS", requests: data})
        })
    }

    acceptRequest(friendId) {
        fetch(this.props.url + '/sendRelation', {
            credentials: 'include',
            method: 'POST',
            body: JSON.stringify({
                user2Id: friendId,
                relation: 1
            })
        }).then(() => {
            this.getFriendRequests()
            this.props.getFriends()
        })
    }

    rejectRequest(friendId) {
        fetch(this.props.url + '/sendRelation', {
            credentials: 'include',
            method: 'POST',
            body: JSON.stringify({
                user2Id: friendId,
                relation: -1
            })
        }).then(() => {
            this.getFriendRequests()
        })
    }

    componentDidMount() {
        this.getFriendRequests()
    }

    render() {
        return (
            <Modal
            show={true}
            size="md"
            aria-labelledby="contained-modal-title-vcenter"
            centered>
                <Modal.Header>
                    <Modal.Title id="contained-modal-title-vcenter">
                        Add Friend
                    </Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <p className="field-title"><b>User ID</b></p>
                    <input type="number" id="addUserId" className='textbox-main' style={{width: "50%"}}/>
                    {
                        this.state.frState === 'STANDBY'?
                        <p className="success-message">&nbsp;</p>
                        :
                        this.state.frState === 'SUCCESS' ?
                        <p className="success-message">
                            Request Pending
                        </p>
                        :
                        <p className="error-message">
                            Couldn't send request
                        </p>
                    }
                </Modal.Body>
                <Modal.Footer>
                    <Button className="btn-main btn-confirm" onClick={() => this.addFriend(this.props.url)}>Send Friend Request</Button>
                </Modal.Footer>
                <Modal.Header>
                    <Modal.Title id="contained-modal-title-vcenter">
                        Friend Requests
                    </Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <table style={{width: "100%"}}>
                        <tbody className="table">
                            <tr>
                                <th>Username#Id</th>
                            </tr>
                            {
                                this.state.requestsState === 'LOADING' ?
                                <tr>
                                    <td>
                                        <Spinner/>
                                    </td>
                                </tr>
                                :
                                this.state.requests.length === 0 ?
                                <tr>
                                    <td>
                                        <p className="info-message">
                                            <br/>
                                            You don't have any friend requests
                                        </p>
                                    </td>
                                </tr>
                                :
                                this.state.requests.map((el, key) => (
                                    <tr key={key}>
                                        <td>
                                            {el.username}#{el.userid}
                                        </td>
                                        <td style={{float: 'right'}}>
                                            <Button className="btn-main btn-confirm" style={{marginRight: '0.5em'}}
                                            onClick={()=>{ this.acceptRequest(el.userid) }}>Accept</Button>

                                            <Button className="btn-main btn-cancel"
                                            onClick={()=>{ this.rejectRequest(el.userid) }}>Reject</Button>
                                        </td>
                                    </tr>
                                ))
                            }
                        </tbody>
                    </table>
                </Modal.Body>
                <Modal.Footer>
                    <Button className="btn-main btn-cancel" onClick={() => {
                        this.props.hide()
                        }}>Close</Button>
                </Modal.Footer>
            </Modal>
        );
    }
}

class ConfirmationModal extends React.Component {
    render() {
        return(
            <Modal
            size="md"
            show={true}
            aria-labelledby="contained-modal-title-vcenter"
            centered>
                <Modal.Header>
                    <Modal.Title id="contained-modal-title-vcenter">
                        {this.props.info.title}
                    </Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <p><b>{this.props.info.content}</b></p>
                </Modal.Body>
                <Modal.Footer>
                    <Button className="btn-main btn-confirm-neg" onClick={() => {
                        this.props.info.action()
                        this.props.hide()
                        }}>
                        {this.props.info.actionName}
                    </Button>
                    <Button className="btn-main btn-cancel" onClick={() => {
                        this.props.hide()
                        }}>Close</Button>
                </Modal.Footer>
            </Modal>
        )
    }
}

export {AddFriendModal, GroupInviteModal, CreateGroupModal, ConfirmationModal};