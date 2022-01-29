import React from 'react';
import {Modal, Button} from 'react-bootstrap';
import Spinner from "./../components/loadingspinner";

class AddFriendModal extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            frState: "STANDBY",
            requestsState: "STANDBY",
            requests: []
        };

        this.refreshRequests = true;
    }

    addFriend() {
        var friendIdBox = document.getElementById('addUserId')
        fetch(this.props.url + '/sendRelation', {
            credentials: 'include',
            method: 'POST',
            body: JSON.stringify({
                user2Id: parseInt(friendIdBox.value),
                relation: 0
            })
        }).then(response => response.text())
        .then(data => {
            var message = JSON.parse(data).message
            this.setState({frState: message})
        })
    }

    getFriendRequests() {
        this.setState({frState: this.state.frState, requestsState: "LOADING"})
        this.refreshRequests = false;
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

    componentDidUpdate() {
        if(this.props.show && this.refreshRequests)
            this.getFriendRequests()
    }

    render() {
        return (
            <Modal
            show={this.props.show}
            size="lg"
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
                    <br/>
                    <table style={{width: "100%"}}>
                        <tbody className="table">
                            <tr>
                                <th>Username#Id</th>
                            </tr>
                            {
                                this.state.requests.map((el, key) => (
                                    <tr key={key}>
                                        <td>
                                            {el.username}#{el.userid}
                                        </td>
                                        <td style={{float: 'right'}}>
                                            <Button className="btn-main btn-confirm" style={{marginRight: '0.5em'}}
                                            onClick={()=>{ this.acceptRequest(el.userid) }}>&#10004;</Button>

                                            <Button className="btn-main btn-cancel"
                                            onClick={()=>{ this.rejectRequest(el.userid) }}>&#10060;</Button>
                                        </td>
                                    </tr>
                                ))
                            }
                        </tbody>
                    </table>
                    {
                        this.state.requestsState === 'LOADING' ?
                        <center>
                            <Spinner/>
                        </center>
                        :
                        <div/>
                    }
                </Modal.Body>
                <Modal.Footer>
                    <Button className="btn-main btn-cancel" onClick={() => {
                        this.props.onHide()
                        this.refreshRequests = true;
                        this.setState({frState: "STANDBY", requestsState: "STANDBY"})
                        }}>Close</Button>
                </Modal.Footer>
            </Modal>
        );
    }
}

export default AddFriendModal;