'use strict';

import React from 'react';

class Notices extends React.Component {
  constructor (props) {
    super(props);
    this.state = {
      notices: props.notices
    };
  }

  componentWillReceiveProps (nextProps) {
    this.setState({
      notices: nextProps.notices
    });
  }

  render () {
    return (
      <div className="details-panel-description subsection"><span style={{ fontWeight: 600 }}>Notices</span>&nbsp;
      {
        this.state.notices.length > 0 ? this.state.notices.map((notice) => {
          let noticeTitle = notice.title;
          if (notice.link) {
            noticeTitle = <span>{notice.title} <a href={notice.link} target="_blank"><span className="glyphicon glyphicon-new-window"></span></a></span>;
          }
          return <div key={notice.title}>{noticeTitle}</div>;
        }) : (<div>None</div>)
      }
    </div>
    );
  }
}

Notices.propTypes = {
  notices: React.PropTypes.array.isRequired
};

export default Notices;
