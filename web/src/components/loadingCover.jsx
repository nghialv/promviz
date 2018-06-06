'use strict';

import React from 'react';
import TWEEN from 'tween.js';

import './loadingCover.css';

const logo = require('url!./istio.png'); // eslint-disable-line import/no-extraneous-dependencies

const helperStyles = {
  display: 'inline-block',
  height: '100%',
  verticalAlign: 'middle'
};

const loaderStyles = {
  display: 'inline-block',
  verticalAlign: 'middle',
  fontSize: '1.5em',
  color: '#555'
};


class LoadingCover extends React.Component {
  constructor (props) {
    super(props);
    this.state = {
      show: props.show,
      showing: props.show
    };
  }

  componentWillReceiveProps (nextProps) {
    if (nextProps.show !== this.props.show) {
      if (!nextProps.show) {
        // If transitioning to not show...
        // TODO: Figure out how to do this with pure CSS instead of javascript
        const data = { opacity: 1 };
        this.tween = new TWEEN.Tween(data)
          .to({ opacity: 0 }, 1000)
          .easing(TWEEN.Easing.Linear.None)
          .onUpdate(() => {
            this.setState({ opacity: data.opacity });
          })
          .onComplete(() => {
            this.setState({ showing: false });
          })
          .start();
      } else {
        // If transitioning to show...
        this.setState({ showing: true });
      }
    }
  }

  render () {
    const wrapperStyles = {
      display: this.state.showing ? 'initial' : 'none'
    };

    const coverStyles = {
      opacity: this.state.showing ? this.state.opacity : undefined
    };

    return (
      <div className="loading-cover-wrapper" style={wrapperStyles}>
        { this.state.showing
          ? <div className="loading-cover" style={coverStyles}>
              <span style={helperStyles}></span>
              <div style={loaderStyles}>
                <img className="loading-image" src={logo} />
                Loading...
              </div>
            </div>
          : undefined
        }
      </div>
    );
  }
}

LoadingCover.propTypes = {
  show: React.PropTypes.bool.isRequired
};

export default LoadingCover;
