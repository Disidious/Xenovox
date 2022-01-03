/* eslint-disable jsx-a11y/anchor-is-valid */
import { Col, Row, Form, Button } from 'react-bootstrap';
import { Navigate } from 'react-router-dom'
import React, { useState, useEffect } from 'react';

import logo from "./../images/logo.png"

import Spinner from "./../components/loadingspinner";

function checkAuth(url, setState) {
    fetch(url + '/checkauth', {
        credentials: 'include',
        method: 'GET'
    }).then(response => response.json())
    .then(data => {
        if(data.message === "AUTHORIZED") {
            setState('REDIRECT')
            return
        }
        setState('CHECKED')
    })
    .catch((error) => {
        console.log(error)
        setState('CHECKED')
    })
}

function login(url, setState) {
    var username = document.getElementById("username").value;
    var pass = document.getElementById("pass").value;
    
    if(username === "" || pass === "") {
        setState("MISSING")
        return;
    }

    fetch(url + '/auth/login', {
        credentials: 'include',
        method: 'POST',
        body: JSON.stringify({
            username: username,
            password: pass,
        })
    }).then(response => response.text())
    .then(data => {
        var message = JSON.parse(data).message
        if(message === "LOGGED_IN") {
            setState('REDIRECT')
            return
        }
        setState('UNAUTHORIZED')
    })
    .catch((error) => {
        console.log(error)
        setState('UNEXPECTED_FAILURE')
    })

    setState('LOADING')
}

function register(url, setState, setMode) {
    var username = document.getElementById("username").value;
    var pass = document.getElementById("pass").value;
    var name = document.getElementById("name").value;
    var email = document.getElementById("email").value;

    if(username === "" || pass === "" || name === "" || email === "") {
        setState("MISSING")
        return;
    }

    fetch(url + '/auth/register', {
        credentials: 'include',
        method: 'POST',
        body: JSON.stringify({
            username: username,
            password: pass,
            email: email,
            name: name
        })
    }).then(response => response.text())
    .then(data => {
        var message = JSON.parse(data).message
        setState(message)
        if(message === "SUCCESS")
            switchToLogin(setMode)
    })

    setState('LOADING')
}

function switchToRegister(setMode) {
    setMode('REGISTER')
    document.getElementById("username").value = ""
    document.getElementById("pass").value = ""
}

function switchToLogin(setMode) {
    setMode('LOGIN')
    document.getElementById("username").value = ""
    document.getElementById("pass").value = ""
}

// function handleEnter(e) {
//     if(e.key !== 'Enter') {
//         return
//     }
    
//     if(mode === 'LOGIN') {
//         login(url, setState)
//     } else {
//         register(url, setState)
//     }
// }

function Login(props) {
    const[state, setState] = useState('STANDBY')
    const[mode, setMode] = useState('LOGIN')

    const handleEnter = (e) => {
        if(e.key !== 'Enter') {
            return
        }
        
        if(mode === 'LOGIN') {
            login(props.url, setState)
        } else {
            register(props.url, setState, setMode)
        }
    }

    useEffect(()=>{
        checkAuth(props.url, setState)
    }, [props.url])

    if(state === 'REDIRECT') {
        return <Navigate to='/home'/>
    }
    else if(state === 'STANDBY') {
        return (
            <Spinner mode={"SCREEN"}/>
        );
    }
    else {
        return (
            <Row className="justify-content-center h-100 width-fix"> 
                <Col className="center">
                    <div className="creds-container">
                        <Form>
                            <center>
                                <img src={logo} className="logo" alt="logo"/>
                                <h1 className="neonText">
                                    XENOVOX
                                </h1>
                            </center>
                            {
                                mode === 'REGISTER' ?
                                <div>
                                    
                                    <Form.Group>
                                        <Form.Label className="form-label">Name</Form.Label>
                                        <Form.Control id="name" type="name" className="form-input" placeholder="Name" 
                                         onKeyPress={(e) => handleEnter(e)}/>
                                    </Form.Group>
                                    <br/>
                                    <Form.Group>
                                        <Form.Label className="form-label">Email</Form.Label>
                                        <Form.Control id="email" type="email" className="form-input" placeholder="Email" 
                                        onKeyPress={(e) => handleEnter(e)}/>
                                    </Form.Group>
                                    <br/>
                                </div>
                                :
                                <br/>
                            }
                            <Form.Group>
                                <Form.Label className="form-label">Username</Form.Label>
                                <Form.Control id="username" type="username" className="form-input" placeholder="Username" 
                                onKeyPress={(e) => handleEnter(e)}/>
                            </Form.Group>
                            <br/>
                            <Form.Group>
                                <Form.Label className="form-label">Password</Form.Label>
                                <Form.Control id="pass" type="password" className="form-input" placeholder="Password" 
                                onKeyPress={(e) => handleEnter(e)}/>
                            </Form.Group>
                            <br/>
                            {
                                mode === 'LOGIN' ?
                                <center>
                                    <Button type="button" className="btn-main" disabled={state === "LOADING"} 
                                    onClick={() => {login(props.url, setState)}}>
                                        Sign in
                                    </Button>
                                    <p>
                                        Don't have an account yet? 
                                        <a href="#" onClick={()=>{
                                            switchToRegister(setMode)
                                        }}> Sign Up</a>
                                    </p>
                                </center>
                                :
                                <center>
                                    <Button type="button" className="btn-main" disabled={state === "LOADING"} 
                                    onClick={() => {register(props.url, setState, setMode)}}>
                                        Sign Up
                                    </Button>
                                    <p>
                                        Already have an account? 
                                        <a href="#" onClick={()=>{
                                            switchToLogin(setMode)
                                        }}> Sign In</a>
                                    </p>
                                </center>
                            }
                            
                            <center style={{display: state === "LOADING" ? 'block':'none'}}>
                                <Spinner/>
                            </center>
                            <center>
                                {
                                    state === "UNAUTHORIZED" ?
                                    <p className="error-message">
                                        Wrong username or password
                                        <br/>
                                        Please try again
                                    </p>
                                    :
                                    state === "EMAIL_EXISTS" ?
                                    <p className="error-message">
                                        Email already exists
                                    </p>
                                    :
                                    state === "USERNAME_EXISTS" ?
                                    <p className="error-message">
                                        Username already exists
                                    </p>
                                    :
                                    state === "MISSING" ?
                                    <p className="error-message">
                                        Please fill all fields
                                    </p>
                                    :
                                    state === "FAILED" ?
                                    <p className="error-message">
                                        Registration failed
                                    </p>
                                    :
                                    state === "UNEXPECTED_FAILURE" ?
                                    <p className="error-message">
                                        Server is down<br/>Please try again later
                                    </p>
                                    :
                                    <div/>
                                }
                            </center>
                            <center>
                                {
                                    state === "SUCCESS" ?
                                    <p className="success-message">
                                        Registered!
                                    </p>
                                    :
                                    <div/>
                                }
                            </center>
                        </Form>
                    </div>
                </Col>
                <Col>
                    <div className="login-image"/>
                </Col>
            </Row>
        );
    }
}


export default Login;