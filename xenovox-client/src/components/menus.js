import React from 'react';

class UserMenu extends React.Component {
    
    // Unfinished component
    // Should be a menu for users containing (Unfriend - Block - Invite to group)

    componentDidMount() {
        console.log("did mount")
        const menu = document.getElementById('userMenu');
        if(menu === null) {
            console.log("null here")
            return
        }
        menu.style.setProperty('--mouse-x', this.props.left + 'px')
        menu.style.setProperty('--mouse-y', this.props.top + 'px')
    }

    render() {
        return (
            <div 
            id="userMenu"
            className="user-menu" 
            style={{display: this.props.display}}>
                <ul>
                    <li><button>Element-1</button></li>
                    <li><button>Element-2</button></li>
                    <li><button>Element-3</button></li>
                    <li><button>Element-4</button></li>
                </ul>
            </div>  
        );
    }
}

export default UserMenu;