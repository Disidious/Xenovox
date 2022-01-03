import React from 'react';
import {
  BrowserRouter as Router,
  Routes,
  Route
} from 'react-router-dom';
import './App.css';
import '../node_modules/bootstrap/dist/css/bootstrap.min.css';

import Login from './pages/login';
import Home from './pages/home'
import Xenosocket from './components/xenosocket'

let socket = new Xenosocket()
let url = 'http://localhost:7777'

function App() {
  return (
    <Router>
      <div className="App">
        <Routes>
          <Route exact path='/' element={<Login url={url}/>}/>
          <Route exact path='/home/:id' element={<Home socket={socket} url={url}/>}/>
          <Route exact path='/home' element={<Home socket={socket} url={url}/>}/>
        </Routes>
      </div>
    </Router>
  );
}

export default App;
