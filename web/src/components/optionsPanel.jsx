'use strict';

import React from 'react';

import './optionsPanel.css';

class OptionsPanel extends React.Component {
  constructor (props) {
    super(props);
    this.state = {
      showOptions: false,
      alignRight: false
    };
  }

  componentDidMount () {
    this.setState({ alignRight: this.shouldAlignRight() });
  }

  shouldAlignRight () {
    const elm = this.refs.optionsPanel;
    const boundingRect = elm.getBoundingClientRect();
    const panelBoundingRect = elm.children[1].getBoundingClientRect();

    const isEntirelyVisible = (boundingRect.left + panelBoundingRect.width <= window.innerWidth);
    return !isEntirelyVisible;
  }

  clearDocumentClick () {
    if (this.documentClickHandler) {
      document.removeEventListener('click', this.documentClickHandler, false);
      this.documentClickHandler = undefined;
    }
  }

  setDocumentClick () {
    this.clearDocumentClick();
    this.documentClickHandler = this.handleDocumentClick.bind(this);
    document.addEventListener('click', this.documentClickHandler, false);
  }

  optionsDropdownClicked () {
    this.setState({
      showOptions: !this.state.showOptions,
      alignRight: this.shouldAlignRight()
    });
  }

  handleDocumentClick () {
    this.setState({ showOptions: false });
  }

  componentWillUpdate (nextProps, nextState) {
    if (nextState.showOptions !== this.state.showOptions) {
      if (nextState.showOptions) {
        this.setDocumentClick();
      } else {
        this.clearDocumentClick();
      }
    }
  }

  handleClick (event) {
    event.stopPropagation();
    event.nativeEvent.stopImmediatePropagation();
  }

  render () {
    const panelStyles = {
      visibility: !this.state.showOptions ? 'hidden' : undefined,
      border: '1px solid grey'
    };

    if (this.state.alignRight) {
      panelStyles.right = 0;
    }

    return (
      <div ref="optionsPanel" className="options-panel" onClick={this.optionsDropdownClicked.bind(this)}>
        <div className="options-panel-title">
          <a role="button" className="options-link">
            {this.props.title}
            <span className="caret"></span>
          </a>
        </div>
        <div className="options-panel-content" style={panelStyles} onClick={this.handleClick}>
          { this.props.children }
        </div>
      </div>
    );
  }
}

OptionsPanel.propTypes = {
  title: React.PropTypes.string.isRequired
};

export default OptionsPanel;
