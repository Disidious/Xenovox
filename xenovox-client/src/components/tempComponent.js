import React from 'react';
import { Col, Row } from 'react-bootstrap';

class TempForm extends React.Component {
    constructor(props) {
        super(props);
        this.state = {value: ''};

        this.login = this.login.bind(this);
        this.getUsers = this.getUsers.bind(this);
        this.logout = this.logout.bind(this);
    }

    login() {
        var username = document.getElementById("username");
        var pass = document.getElementById("pass");
        //console.log(username.value, pass.value, " here");
        fetch('http://localhost:7777/auth/login', {
            credentials: 'include',
            method: 'POST',
            body: JSON.stringify({
                username: username.value,
                password: pass.value,
            })
        }).then(response => response.text())
        .then(data => this.setState({value: data}))
    }

    getUsers() {
        fetch('http://localhost:7777/users', {
            credentials: 'include',
            method: 'GET'
        }).then(response => response.text())
        .then(data => this.setState({value: data}))
    }

    logout() {
        fetch('http://localhost:7777/logout', {
            credentials: 'include',
            method: 'POST'
        }).then(response => response.text())
        .then(data => this.setState({value: data}))
    }
  
    render() {
        return (
            <div>
                <div>
                    <label>Username:</label><br/>
                    <input type="text" id="username" name="username"/><br/>
                    <label>Password:</label><br/>
                    <input type="text" id="pass" name="pass"/><br/><br/>
                    <button onClick={this.login}>Login</button>
                    <button onClick={this.getUsers}>Get Users</button>
                    <button onClick={this.logout}>Logout</button>

                    <a href="/send">
                        <button>Chat</button>
                    </a>
                </div>
                <div>
                    <pre id="output">{this.state.value}</pre>
                </div>
    
            </div>
        );
    }
}

export default TempForm;