'use strict';

import _ from 'lodash';

import Vizceral from 'vizceral-react';
import React from 'react';

class CustomVizceral extends Vizceral {
  constructor (props) {
    super(props);

    this.state = {
      styles: {}
    };
  }

  shouldComponentUpdate (nextProps) {
    if (nextProps.styles) {
      if (this.shouldStylesUpdate(nextProps.styles)) {
        this.setState({ styles: nextProps.styles });
        this.vizceral.updateStyles({ colorTraffic: nextProps.styles });
        this.refreshNodes();
      }
    }
    return true;
  }

  refreshNodes () {
    if (!(this.vizceral.currentGraph && this.vizceral.currentGraph.nodes)) {
      return;
    }
    _.map(this.vizceral.currentGraph.nodes, (value) => {
      if (value.view) {
        value.view.refresh(true);
      }
    });
  }

  shouldStylesUpdate (styles) {
    const currentKeys = Object.keys(this.state.styles).sort();
    const newKeys = Object.keys(styles).sort();

    if (currentKeys.toString() === newKeys.toString()) {
      return false;
    }

    return !_.every(newKeys, key => styles[key] === this.state.styles[key]);
  }
}

CustomVizceral.propTypes = {
  styles: React.PropTypes.object
};

export default CustomVizceral;
