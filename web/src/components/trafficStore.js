'use strict';

import _ from 'lodash';
import EventEmitter from 'events';

import AppDispatcher from '../appDispatcher';
import AppConstants from '../appConstants';

const CHANGE_EVENT = 'change';
const OFFSET_CHANGE_EVENT = 'offset_changed';

const store = {
  traffic: { nodes: [], connections: [] },
  regions: {},
  lastUpdatedServerTime: 0,
  lastUpdatedClientTime: 0,
  offset: 0
};

const clearTraffic = function () {
  store.traffic = { nodes: [], connections: [] };
};

class TrafficStore extends EventEmitter {
  constructor () {
    super();
    this.requests = {};

    AppDispatcher.register((payload) => {
      const action = payload.action;
      switch (action.actionType) {
      case AppConstants.ActionTypes.UPDATE_TRAFFIC:
        if (this.updateTraffic(action.data)) {
          this.emit(CHANGE_EVENT);
        }
        break;
      case AppConstants.ActionTypes.CLEAR_TRAFFIC:
        clearTraffic();
        this.emit(CHANGE_EVENT);
        break;
      case AppConstants.ActionTypes.UPDATE_TRAFFIC_OFFSET:
        this.updateOffset(action.data);
        this.emit(OFFSET_CHANGE_EVENT);
        break;
      default:
        return true;
      }
      return true;
    });
  }

  getMaxHistoryLength () { return 30; }

  addChangeListener (cb) {
    this.on(CHANGE_EVENT, cb);
  }

  addOffsetChangeListener (cb) {
    this.on(OFFSET_CHANGE_EVENT, cb);
  }

  removeChangeListener (cb) {
    this.removeListener(CHANGE_EVENT, cb);
  }

  removeOffsetChangeListener (cb) {
    this.removeListener(OFFSET_CHANGE_EVENT, cb);
  }

  getTraffic () {
    return store.traffic;
  }

  getHistoryLength () {
    return store.traffic.length;
  }

  getLastUpdatedServerTime () {
    return store.lastUpdatedServerTime;
  }

  getLastUpdatedClientTime () {
    return store.lastUpdatedClientTime;
  }

  findRegionNode (node, region) {
    if (!node.nodes || node.nodes.length <= 0) {
      return undefined;
    }
    if (node.renderer === 'region' && node.name === region) {
      return node;
    }
    const result = _.reject(_.map(node.nodes, n => this.findRegionNode(n, region)), r => r === undefined);
    return result.length > 0 ? result[0] : undefined;
  }

  generateConnectionName (source, target) {
    const sourceName = typeof source === 'string' ? source : source.name;
    const targetName = typeof target === 'string' ? target : target.name;
    return `${sourceName}-${targetName}`;
  }

  getNodeHistory (region, name) {
    return store.regions[region].nodes[name] || [];
  }

  getConnectionHistory (region, source, target) {
    return store.regions[region].connections[this.generateConnectionName(source, target)] || [];
  }

  getNodeHistoryRange (region, name, since, until) {
    return _.filter(this.getNodeHistory(region, name), n => since <= n.updated && n.updated <= until);
  }

  getConnectionHistoryRange (region, source, target, since, until) {
    return _.filter(this.getConnectionHistory(region, source, target), c => since <= c.updated && c.updated <= until);
  }

  getTrafficOffset () {
    return store.offset;
  }

  updateOffset (offset) {
    store.offset = offset;
    store.regions = {};
  }

  updateTraffic (traffic) {
    store.traffic = traffic;
    store.lastUpdatedClientTime = Date.now();
    let serverUpdateTime = Date.now();

    if (traffic.serverUpdateTime) {
      serverUpdateTime = traffic.serverUpdateTime * 1000;

      if (store.lastUpdatedServerTime === serverUpdateTime) {
        return false;
      }

      store.lastUpdatedServerTime = serverUpdateTime;
    }

    const listupRegion = (node, self) => {
      if (!node.nodes || node.nodes.length <= 0) {
        return [];
      }
      return _.flatten(_.filter(node.nodes, n => n.renderer === 'region').concat(node.nodes, n => self(n, self)));
    };

    const regionNodes = listupRegion(traffic, listupRegion);
    const $this = this;
    _.map(regionNodes, (regionNode) => {
      if (!store.regions[regionNode.name]) {
        store.regions[regionNode.name] = {
          nodes: {},
          connections: {}
        };
      }
      _.map(regionNode.nodes, (node) => {
        node.updated = serverUpdateTime;
        const region = store.regions[regionNode.name];
        if (!region.nodes[node.name]) {
          region.nodes[node.name] = [];
        }
        const nodes = region.nodes[node.name];
        nodes.push(node);
        if (nodes.length > $this.getMaxHistoryLength()) {
          nodes.shift();
        }
      });
      _.map(regionNode.connections, (connection) => {
        connection.updated = serverUpdateTime;
        const connectionName = $this.generateConnectionName(connection.source, connection.target);
        const region = store.regions[regionNode.name];
        if (!region.connections[connectionName]) {
          region.connections[connectionName] = [];
        }
        const connections = region.connections[connectionName];
        connections.push(connection);
        if (connections.length > $this.getMaxHistoryLength()) {
          connections.shift();
        }
      });
    });

    return true;
  }
}

const trafficStore = new TrafficStore();

export default trafficStore;
