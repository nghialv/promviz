'use strict';

import React from 'react';
import './physicsOptions.css';

class PhysicsOptions extends React.Component {
  constructor (props) {
    super(props);
    this.state = {
      isEnabled: true,
      jaspersReplusionBetweenParticles: true,
      viscousDragCoefficient: 0.2,
      hooksSprings: {
        restLength: 50,
        springConstant: 0.2,
        dampingConstant: 0.1
      },
      particles: {
        mass: 1
      }
    };
  }

  componentWillReceiveProps (nextProps) {
    this.setState(nextProps.options);
  }

  changeState (newState) {
    this.setState(newState);
    this.props.changedCallback(newState);
  }

  render () {
    const isEnabled = this.state.isEnabled;
    const jaspersReplusionBetweenParticles = this.state.jaspersReplusionBetweenParticles;
    return (
      <div className="physics-options">
        <div>
          <input type="checkbox" name="isEnabled" value="isEnabled" checked={isEnabled} onChange={() => this.changeState({ isEnabled: !isEnabled })}/>
          <label style={{ cursor: 'default' }} onClick={() => this.changeState({ isEnabled: !isEnabled })}>Enabled</label>
        </div>
        <div>
          <label htmlFor="hooksSprings_restLength">Spring Rest Length</label>
          <input id="hooksSprings_restLength" type="number" step="any" name="hooksSprings_restLength" value={this.state.hooksSprings.restLength} onChange={event => this.changeState({ hooksSprings: { restLength: event.currentTarget.valueAsNumber } })}/>
        </div>
        <div>
          <label htmlFor="hooksSprings_springConstant">Spring Constant</label>
          <input id="hooksSprings_springConstant" type="number" min="0" step="any" max="1" name="hooksSprings_springConstant" value={this.state.hooksSprings.springConstant} onChange={event => this.changeState({ hooksSprings: { springConstant: event.currentTarget.valueAsNumber } })}/>
        </div>
        <div>
          <label htmlFor="hooksSprings_dampingConstant">Spring Damping Constant</label>
          <input id="hooksSprings_dampingConstant" type="number" min="0" step="any" max="1" name="hooksSprings_dampingConstant" value={this.state.hooksSprings.dampingConstant} onChange={event => this.changeState({ hooksSprings: { dampingConstant: event.currentTarget.valueAsNumber } })}/>
        </div>
        <div>
          <label htmlFor="viscousDragCoefficient">Viscous Drag Coefficient</label>
          <input id="viscousDragCoefficient" type="number" min="0" step="any" max="1" name="viscousDragCoefficient" value={this.state.viscousDragCoefficient} onChange={event => this.changeState({ viscousDragCoefficient: event.currentTarget.valueAsNumber })}/>
        </div>
        <div>
          <label htmlFor="particles_mass">Particle Mass</label>
          <input id="particles_mass" type="number" min="0.1" step="any" name="particles_mass" value={this.state.particles.mass} onChange={event => this.changeState({ particles: { mass: event.currentTarget.valueAsNumber } })}/>
        </div>
        <div>
          <input type="checkbox" name="jaspersReplusionBetweenParticles" value="jaspersReplusionBetweenParticles" checked={jaspersReplusionBetweenParticles} onChange={() => this.changeState({ jaspersReplusionBetweenParticles: !jaspersReplusionBetweenParticles })}/>
          <label style={{ cursor: 'default' }} onClick={() => this.changeState({ jaspersReplusionBetweenParticles: !jaspersReplusionBetweenParticles })}>Jasper's repulsion between particles</label>
        </div>
      </div>
    );
  }
}

PhysicsOptions.propTypes = {
  options: React.PropTypes.object,
  changedCallback: React.PropTypes.func.isRequired
};

export default PhysicsOptions;
