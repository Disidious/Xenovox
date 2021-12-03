import React from 'react';

class Spinner extends React.Component {
    render() {
      return (
        <div className="loading">
            <div className="arc"></div>
            <div className="arc"></div>
            <div className="arc"></div>
        </div>
      );
    }
}

export default Spinner;