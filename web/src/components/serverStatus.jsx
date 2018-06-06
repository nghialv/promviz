'use strict';

import React from 'react';

import { Tooltip, OverlayTrigger } from 'react-bootstrap';

import AppConstants from '../appConstants';

import './serverStatus.css';

class ServerStatus extends React.Component {
  constructor (props) {
    super(props);

    this.state = {
      endpoint: props.endpoint,
      status: AppConstants.ServerStatus.DISCONNECTED,
      clientUpdatedTime: undefined
    };
  }

  componentWillReceiveProps (nextProps) {
    this.setState(nextProps);
  }

  render () {
    const msToTimeAgo = (ms) => {
      if (!ms) { return 'Unknown'; }
      const secs = Math.round(ms / 1000);
      const hours = Math.floor(secs / (60 * 60));

      const dM = secs % (60 * 60);
      const minutes = Math.floor(dM / 60);

      const dS = dM % 60;
      const seconds = Math.ceil(dS);

      const timeAgo = [];
      if (hours > 0) {
        timeAgo.push(`${hours}h`);
      }
      if (hours > 0 || minutes > 0) {
        timeAgo.push(`${minutes}m`);
      }
      timeAgo.push(`${seconds}s`);
      // return { h: hours, m: minutes, s: seconds };
      return timeAgo.join(':');
    };

    const tooltip = (
      <Tooltip id="update-status">
        <td>Fetched: {msToTimeAgo(Date.now() - this.state.clientUpdatedTime)} ago</td>
      </Tooltip>
    );

    return (
      <div className="server-status">
        <OverlayTrigger placement="bottom" overlay={tooltip}>
          <div className={this.state.status === AppConstants.ServerStatus.DISCONNECTED ? 'indicator disconnected' : 'indicator connected'}></div>
        </OverlayTrigger>
      </div>
    );
  }
}

ServerStatus.propTypes = {
  endpoint: React.PropTypes.string.isRequired,
  status: React.PropTypes.string.isRequired,
  clientUpdatedTime: React.PropTypes.number
};

export default ServerStatus;
