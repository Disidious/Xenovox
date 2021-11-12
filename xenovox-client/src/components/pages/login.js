import React from 'react';
import { Col, Row } from 'react-bootstrap';

function login() {
    var username = document.getElementById("username");
    var pass = document.getElementById("pass");
    //console.log(username.value, pass.value, " here");
    fetch('http://localhost:7777/login', {
        credentials: 'include',
        method: 'POST',
        body: JSON.stringify({
            username: username.value,
            password: pass.value,
        })
    }).then(response => response.json())
    .then(data => console.log(data))
}

function getUsers() {
    fetch('http://localhost:7777/users', {
        credentials: 'include',
        method: 'GET'
    }).then(response => response.json())
    .then(data => console.log(data))

    serverOutput = 'data';
}

function logout() {
    fetch('http://localhost:7777/logout', {
        credentials: 'include',
        method: 'POST'
    }).then(response => response.json())
    .then(data => console.log(data))
}

var serverOutput = '';

function LoginPage() {
    return (
        <div>
            <div className="temp-div">
                <label>Username:</label><br/>
                <input type="text" id="username" name="username"/><br/>
                <label>Password:</label><br/>
                <input type="text" id="pass" name="pass"/><br/><br/>
                <button onClick={login}>Login</button><br/>
                <button onClick={getUsers}>Get Users</button><br/>
                <button onClick={logout}>Logout</button><br/>
            </div>
            <div className="temp-div" style={{marginLeft:"100px"}}>
                <p id="output">{serverOutput}</p>
            </div>

        </div>
    );
}
export default LoginPage;