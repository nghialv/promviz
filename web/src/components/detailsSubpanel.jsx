'use strict';

import React from 'react';

class DetailsSubpanel extends React.Component {
  constructor (props) {
    super(props);
    this.state = {
      expanded: props.expanded ? props.expanded : false
    };
  }

  componentWillReceiveProps (nextProps) {
    if (nextProps.expanded !== undefined) {
      this.setState({ expanded: nextProps.expanded });
    }
  }

  render () {
    const badge = this.props.badge;
    const title = this.props.title.replace(/\s/g, '_');
    const headingId = `${title}Heading`;
    const collapseId = `collapse${title}`;

    const expanded = this.state.expanded;
    const iconClass = `glyphicon ${expanded ? 'glyphicon-chevron-down' : 'glyphicon-chevron-right'}`;
    const iconStyle = {
      fontSize: '12px',
      paddingRight: expanded ? '5px' : undefined
    };
    return (
      <div className="panel-group node-details-subpanel" role="tablist">
        <div className="panel panel-default">
          <div className="panel-heading" role="tab" id={headingId}>
            <h4 className="panel-title">
              <a role="button" data-toggle="collapse" href={`#${collapseId}`} aria-controls={collapseId} className={`accordion-toggle${expanded ? '' : ' collapsed'}`} onClick={() => this.setState({ expanded: !expanded })}>
                <span className={iconClass} style={iconStyle}></span> {this.props.title.toUpperCase()} {badge ? <span className="badge">{badge}</span> : undefined}
              </a>
            </h4>
          </div>
          <div id={collapseId} className={`panel-collapse collapse${expanded ? ' in' : ''}`} role="tabpanel" aria-labelledby={headingId} aria-expanded={expanded ? 'true' : 'false'}>
            <div className="panel-body">
              <div className="subsection">
                {this.props.children}
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }
}

DetailsSubpanel.propTypes = {
  title: React.PropTypes.string.isRequired,
  expanded: React.PropTypes.bool,
  badge: React.PropTypes.number
};

export default DetailsSubpanel;
