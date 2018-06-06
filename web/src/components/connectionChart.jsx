'use strict';

import _ from 'lodash';
import React from 'react';

import ReactHighcharts from 'react-highcharts';

import trafficStore from './trafficStore';

const defaultConfig = {
  global: {
    timezoneOffset: (new Date()).getTimezoneOffset()
  },
  chart: {
    type: 'area',
    width: 300,
    height: 400,
    backgroundColor: '#466bb0'
  },
  title: '',
  xAxis: [{
    categories: [],
    type: 'datetime',
    crosshair: false,
    labels: {
      format: '{value:%m/%d %H:%M:%S}',
      step: 0,
      style: {
        color: '#ffffff'
      }
    },
    tickInterval: 60 * 1000 // 1 minute
  }],
  yAxis: [{
    title: {
      text: 'count',
      labels: {
        format: '{value:%.2f}',
        style: {
          color: '#ffffff'
        }
      }
    },
    labels: {
      style: {
        color: '#ffffff'
      }
    }
  }],
  tooltip: {
    shared: true,
    xDateFormat: '%Y/%m/%d %H:%M:%S',
    pointFormat: '<span style="color:{point.color}">\u25CF</span> {series.name}: <b>{point.y:%.2f}</b><br/>'
  },
  plotOptions: {
    area: {
      marker: {
        enabled: false,
        symbol: 'circle',
        radius: 1,
        states: {
          hover: {
            enabled: true
          }
        }
      }
    }
  },
  series: [{
    name: 'RPS',
    data: history.total,
    color: '#BAD5ED'
  },
  {
    name: 'errors',
    data: history.errors,
    color: '#B82424'
  }],
  legend: {
    itemStyle: {
      color: '#D6D6D6'
    }
  }
};

class ConnectionChart extends React.Component {
  constructor (props) {
    super(props);

    const config = props.config ? _.merge(defaultConfig, props.config) : defaultConfig;
    config.chart.width = props.width;
    config.chart.height = props.height;

    if (config.global) {
      ReactHighcharts.Highcharts.setOptions({
        global: config.global
      });
    }

    this.state = {
      connection: props.connection,
      config: config,
      initialized: false
    };
  }

  getHistory (region, source, target, until) {
    const connectionHistory = trafficStore.getConnectionHistoryRange(region, source, target, 0, until);

    const totalHistory = _.map(connectionHistory, connection => ({
      x: connection.updated,
      y: (connection.metrics.normal || 0) + (connection.metrics.warning || 0) + (connection.metrics.danger || 0),
    }));

    const errorsHistory = _.map(connectionHistory, connection => ({
      x: connection.updated,
      y: connection.metrics.danger || 0
    }));

    return { total: totalHistory, errors: errorsHistory };
  }

  offsetChanged = () => {
    this.setState({ initialized: false });
  };

  componentDidMount () {
    trafficStore.addOffsetChangeListener(this.offsetChanged);
  }

  componentWillReceiveProps (nextProps) {
    const connection = nextProps.connection;

    const chart = this.chart.getChart();
    const updated = trafficStore.getLastUpdatedServerTime();

    this.setState({
      connection: nextProps.connection
    });

    if (!this.state.initialized ||
        this.state.connection.source.name !== nextProps.connection.source.name ||
        this.state.connection.target.name !== nextProps.connection.target.name) {
      const history = this.getHistory(
        nextProps.region || this.state.region,
        nextProps.connection.source.name,
        nextProps.connection.target.name,
        updated
      );

      chart.series[0].setData(history.total);
      chart.series[1].setData(history.errors);

      if (!this.state.initialized) {
        this.setState({ initialized: true });
      }
      chart.redraw();
      return;
    }

    const shift = chart.series[0].data.length > trafficStore.getMaxHistoryLength();

    chart.series[0].addPoint({
      x: updated,
      y: connection.getVolumeTotal()
    }, false, shift, false);
    chart.series[1].addPoint({
      x: updated,
      y: connection.getVolume('danger') || 0
    }, false, shift, false);

    chart.redraw();
  }

  render () {
    return (
      <div className="connection-chart">
        <ReactHighcharts isPureConfig={true} config={this.state.config} ref={(chart) => { this.chart = chart; }}></ReactHighcharts>
      </div>
    );
  }
}

ConnectionChart.propTypes = {
  region: React.PropTypes.string.isRequired,
  connection: React.PropTypes.object.isRequired,
  config: React.PropTypes.object
};

export default ConnectionChart;
