'use strict';

import React from 'react';

class Stepper extends React.Component {
  constructor (props) {
    super(props);
    this.state = {
    };
  }

  stepChanged (index) {
    this.props.changeCallback(index);
  }

  render () {
    return (
      <ol className="stepper">
        {
          this.props.steps.map((step, index) => {
            let className = this.props.selectedStep === index ? 'is-current' : undefined;
            className = className || this.props.selectedStep > index ? 'is-lower' : undefined;
            if (className === 'is-lower' && this.props.selectedStep > index) { className += ' show-bar'; }
            let stepName = step.name ? step.name.trim() : undefined;
            stepName = stepName || '&nbsp;';
            return (
              <li key={index} className={className} data-step=" " onClick={() => this.stepChanged(index)} dangerouslySetInnerHTML={{ __html: stepName }}>
              </li>
            );
          })
        }
      </ol>
    );
  }
}

Stepper.propTypes = {
  steps: React.PropTypes.array.isRequired,
  selectedStep: React.PropTypes.number.isRequired,
  changeCallback: React.PropTypes.func.isRequired
};

export default Stepper;
