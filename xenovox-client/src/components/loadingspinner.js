import React from 'react';

class Spinner extends React.Component {
    render() {
      return (
        <div>
          {
            this.props.mode === "SCREEN" ?
            <center style={{marginTop: "20%"}}>
              <div className="loading">
                <div className="arc"></div>
                <div className="arc"></div>
                <div className="arc"></div>
              </div>
            </center>
            :
            <center>
              <div className="loading">
                <div className="arc"></div>
                <div className="arc"></div>
                <div className="arc"></div>
              </div>
            </center>
          }
        </div>
      );
    }
}

export default Spinner;