import React from 'react';
import {
  BrowserRouter as Router,
  Routes,
  Route
} from 'react-router-dom';
import './App.css'

import LoginPage from './components/pages/login';
import TempForm from './components/tempComponent';
import SocketPage from './components/pages/socket'

function App() {
  return (
    <Router>
      <div className="App">
        <Routes>
          <Route exact path='/' element={<TempForm/>}/>
          <Route exact path='/send' element={<SocketPage/>}/>
        </Routes>
      </div>
    </Router>
  );
}

export default App;
