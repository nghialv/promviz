/* eslint no-restricted-syntax: 0 */

'use strict';

import React from 'react';
import numeral from 'numeral';

import Notices from './notices';
import ConnectionChart from './connectionChart';

import './detailsPanel.css';

class DetailsPanelConnection extends React.Component {
  constructor (props) {
    super(props);
    this.state = {
      region: props.region,
      connection: props.connection
    };
  }

  componentWillReceiveProps (nextProps) {
    this.setState({
      region: nextProps.region,
      connection: nextProps.connection
    });
  }

  render () {
    const connection = this.state.connection;
    const notices = (connection && connection.notices) || [];

    const total = connection.getVolumeTotal();
    const errors = connection.getVolume('danger') || 0;
    const errorRate = errors / total || 0;

    return (
      <div className="details-panel">
        <div className="subsection">
          <div className="details-panel-title">{connection.getName()}
          </div>
          <div className="details-panel-close" onClick={this.props.closeCallback}>
            <span className="glyphicon glyphicon-remove" aria-hidden="true"></span>
          </div>
        </div>
        <div className="details-panel-description subsection">
          <span>RPS: {`${numeral(total).format('0.[00]')}`}</span><br />
          <span>Error: {`${numeral(errorRate * 100).format('0.[00]')}`}%</span>
        </div>
        <Notices notices={notices} />
        <ConnectionChart region={this.props.region} connection={connection} />
      </div>
    );
  }
}

DetailsPanelConnection.propTypes = {
  closeCallback: React.PropTypes.func.isRequired,
  connection: React.PropTypes.object.isRequired,
  nodeClicked: React.PropTypes.func,
  region: React.PropTypes.string
};

DetailsPanelConnection.defaultProps = {
  nodeClicked: () => {},
  region: ''
};

export default DetailsPanelConnection;
