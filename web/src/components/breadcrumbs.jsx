'use strict';

import React from 'react';

import './breadcrumbs.css';

class Breadcrumbs extends React.Component {
  constructor (props) {
    super(props);
    this.state = {
    };
  }

  handleClick (index) {
    const newState = this.props.navigationStack.slice(0, index + 1);
    this.props.navigationCallback(newState);
  }

  shouldComponentUpdate (nextProps) {
    if (nextProps.navigationStack) {
      if (nextProps.navigationStack.length !== this.props.navigationStack) {
        return true;
      }

      for (let i = 0; i < this.props.navigationStack.length; i++) {
        if (nextProps.navigationStack[i] !== this.props.navigationStack[i]) {
          return true;
        }
      }
    }
    return false;
  }

  render () {
    const navStack = this.props.navigationStack.slice() || [];
    navStack.unshift(this.props.rootTitle);

    return (
      <div className="breadcrumbs">
        <ol>
            {
              navStack.map((state, index) =>
                ((index !== navStack.length - 1) ?
                <li key={index + state}><a className="clickable" onClick={() => { this.handleClick(index - 1); }}>{ state }</a></li> :
                <li key={index + state}>{ state }</li>)
              )
            }
        </ol>
      </div>
    );
  }
}

Breadcrumbs.propTypes = {
  rootTitle: React.PropTypes.string.isRequired,
  navigationStack: React.PropTypes.array.isRequired,
  navigationCallback: React.PropTypes.func.isRequired
};

export default Breadcrumbs;
