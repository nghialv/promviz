'use strict';

import React from 'react';

import OptionsPanel from './optionsPanel';

import './replayClock.css';

class ReplayClock extends React.Component {
  static defaultProps = {
    offsetChanged: () => {}
  }

  constructor (props) {
    super(props);

    this.state = {
      time: props.time,
      inputTime: props.time,
      offset: 0,
      offsetChanged: props.offsetChanged,
      error: false
    };

    this.panel = undefined;
  }

  componentWillReceiveProps (nextProps) {
    if (nextProps.time === this.props.time) {
      return;
    }
    this.setState({ time: nextProps.time });
  }

  offsetChanged = (value) => {
    this.setState({ inputValue: value });
  };

  offsetEntered = () => {
    const inputDate = new Date(this.state.inputValue);

    // is valid date
    if (!inputDate.getTime || isNaN(inputDate.getTime())) {
      this.setState({
        error: true
      });
      return;
    }

    const offset = Date.now() - inputDate.getTime();
    if (offset < 0 || this.props.maxOffset - offset < 0) {
      this.setState({
        error: true
      });
      return;
    }
    this.setState({
      offset: offset,
      error: false
    });
    this.props.offsetChanged(offset);
    if (this.panel) {
      this.panel.optionsDropdownClicked();
    }
  };

  offsetCleared = () => {
    this.setState({
      offset: 0,
      inputValue: '',
      error: false
    });
    this.props.offsetChanged(0);
    this.panel.optionsDropdownClicked();
  };

  keyPressed = (value) => {
    if (value.key === 'Enter') {
      this.offsetEntered();
    }
  };

  render () {
    const currentDate = new Date(this.state.time);

    const dateToString = (date) => {
      const format = value => (`0${value}`).slice(-2);
      return `${date.getFullYear()}/${format(date.getMonth() + 1)}/${format(date.getDate())
            } ${format(date.getHours())}:${format(date.getMinutes())}:${format(date.getSeconds())}`;
    };

    const offsetToString = (offset) => {
      const abs = Math.abs(offset);
      const seconds = abs / 1000;
      const minutes = seconds / 60;
      const hours = minutes / 60;
      return `${Math.floor(hours)}h ${Math.floor(minutes % 60)}m ${Math.floor(seconds % 60)}s`;
    };

    const diffString = this.state.offset !== 0 ? `(${offsetToString(this.state.offset)} ago)` : '';

    return (
      <div className="clock">
        <OptionsPanel title={dateToString(currentDate) + ' ' + diffString} ref={(panel) => { this.panel = panel; }}>
          <div className="input-group offset-time">
            <input type="text" className={this.state.error ? 'form-control error' : 'form-control'} value={this.state.inputValue} onChange={event => this.offsetChanged(event.currentTarget.value)} defaultValue='' onKeyPress={event => this.keyPressed(event)} placeholder="2016/4/11 00:00:00" />
            <span className="input-group-btn">
              <button className="btn btn-default" onClick={this.offsetEntered}>Replay</button>
            </span>
            <span className="input-group-btn">
              <button className="btn btn-default" onClick={this.offsetCleared}><i className="glyphicon glyphicon-remove"></i></button>
            </span>
          </div>
        </OptionsPanel>
      </div>
    );
  }
}

ReplayClock.propTypes = {
  time: React.PropTypes.number.isRequired,
  maxOffset: React.PropTypes.number.isRequired,
  offsetChanged: React.PropTypes.func
};

export default ReplayClock;
