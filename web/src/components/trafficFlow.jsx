'use strict';

import _ from 'lodash';
import { Alert } from 'react-bootstrap';
import React from 'react';
import TWEEN from 'tween.js'; // Start TWEEN updates for sparklines and loading screen fading out
import 'vizceral-react/dist/vizceral.css';
import keypress from 'keypress.js';
import queryString from 'query-string';
import request from 'superagent';

import './trafficFlow.css';
import Breadcrumbs from './breadcrumbs';
import DisplayOptions from './displayOptions';
import PhysicsOptions from './physicsOptions';
import FilterControls from './filterControls';
import DetailsPanelConnection from './detailsPanelConnection';
import DetailsPanelNode from './detailsPanelNode';
import LoadingCover from './loadingCover';
import Locator from './locator';
import OptionsPanel from './optionsPanel';
import CustomVizceral from './customVizceral';
import ReplayClock from './replayClock';
import ServerStatus from './serverStatus';

import filterActions from './filterActions';
import filterStore from './filterStore';

import trafficActions from './trafficActions';
import trafficStore from './trafficStore';

import AppConstants from '../appConstants';

const listener = new keypress.Listener();

const hasOwnPropFunc = Object.prototype.hasOwnProperty;

function animate (time) {
  requestAnimationFrame(animate);
  TWEEN.update(time);
}
requestAnimationFrame(animate);

const panelWidth = 400;

class TrafficFlow extends React.Component {
  static defaultProps = {
    interval: 10 * 1000,
    maxReplayOffset: 12 * 60 * 60
  }

  constructor (props) {
    super(props);

    this.state = {
      currentView: undefined,
      redirectedFrom: undefined,
      selectedChart: undefined,
      displayOptions: {
        allowDraggingOfNodes: false,
        showLabels: true
      },
      currentGraph_physicsOptions: {
        isEnabled: false,
        viscousDragCoefficient: 0.2,
        hooksSprings: {
          restLength: 50,
          springConstant: 0.2,
          dampingConstant: 0.1
        },
        particles: {
          mass: 1
        }
      },
      labelDimensions: {},
      appliedFilters: filterStore.getChangedFilters(),
      filters: filterStore.getFiltersArray(),
      searchTerm: '',
      matches: {
        total: -1,
        visible: -1
      },
      modes: {
        detailedNode: 'volume'
      },
      traffic: { nodes: [], connections: [] },
      styles: {},
      serverUpdatedTime: Date.now(),
      clientUpdatedTime: Date.now(),
      serverStatus: AppConstants.ServerStatus.DISCONNECTED
    };

    // Browser history support
    window.addEventListener('popstate', event => this.handlePopState(event.state));

    // Keyboard interactivity
    listener.simple_combo('esc', () => {
      if (this.state.detailedNode) {
        this.setState({ detailedNode: undefined });
      } else if (this.state.currentView.length > 0) {
        this.setState({ currentView: this.state.currentView.slice(0, -1) });
      }
    });
  }

  handlePopState () {
    const state = window.history.state || {};
    this.poppedState = true;
    this.setState({ currentView: state.selected, objectToHighlight: state.highlighted });
  }

  viewChanged = (data) => {
    const changedState = {
      currentView: data.view,
      searchTerm: '',
      matches: { total: -1, visible: -1 },
      redirectedFrom: data.redirectedFrom
    };
    if (hasOwnPropFunc.call(data, 'graph')) {
      let oldCurrentGraph = this.state.currentGraph;
      if (oldCurrentGraph == null) oldCurrentGraph = null;
      let newCurrentGraph = data.graph;
      if (newCurrentGraph == null) newCurrentGraph = null;
      if (oldCurrentGraph !== newCurrentGraph) {
        changedState.currentGraph = newCurrentGraph;
        const o = newCurrentGraph === null ? null : newCurrentGraph.getPhysicsOptions();
        changedState.currentGraph_physicsOptions = o;
      }
    }
    this.setState(changedState);
  }

  viewUpdated = () => {
    this.setState({});
  }

  objectHighlighted = (highlightedObject) => {
    // need to set objectToHighlight for diffing on the react component. since it was already highlighted here, it will be a noop
    this.setState({ highlightedObject: highlightedObject, objectToHighlight: highlightedObject ? highlightedObject.getName() : undefined, searchTerm: '', matches: { total: -1, visible: -1 }, redirectedFrom: undefined });
  }

  nodeContextSizeChanged = (dimensions) => {
    this.setState({ labelDimensions: dimensions });
  }

  checkInitialRoute () {
    // Check the location bar for any direct routing information
    const pathArray = window.location.pathname.split('/');
    const currentView = [];
    if (pathArray[1]) {
      currentView.push(pathArray[1]);
      if (pathArray[2]) {
        currentView.push(pathArray[2]);
      }
    }
    const parsedQuery = queryString.parse(window.location.search);

    this.setState({ currentView: currentView, objectToHighlight: parsedQuery.highlighted });
  }

  fetchData = () => {
    request.get(this.props.src)
      .set('Accept', 'application/json')
      .query({ offset: Math.floor(trafficStore.getTrafficOffset() / 1000) })
      .end((err, res) => {
        let serverStatus = AppConstants.ServerStatus.DISCONNECTED;
        if (res && res.status === 200) {
          trafficActions.updateTraffic(res.body);
          this.updateStyles(res.body.classes);
          serverStatus = AppConstants.ServerStatus.CONNECTED;
        }
        this.setState({ serverStatus: serverStatus, clientUpdatedTime: Date.now() });
      });
  };

  beginFetchingData () {
    this.setState({
      fetchInterval: setInterval(this.fetchData, this.props.interval)
    });
    this.fetchData();
  }

  updateStyles (classes) {
    const styles = {};
    _.map(classes, (clas) => {
      styles[clas.name] = clas.color;
    });
    this.setState({
      styles: styles
    });
  }

  componentDidMount () {
    this.checkInitialRoute();
    this.beginFetchingData();

    // Listen for changes to the stores
    filterStore.addChangeListener(this.filtersChanged);
    trafficStore.addChangeListener(this.trafficChanged);
  }

  componentWillUnmount () {
    filterStore.removeChangeListener(this.filtersChanged);
    trafficStore.removeChangeListener(this.trafficChanged);
    clearInterval(this.state.fetchInterval);
  }

  shouldComponentUpdate (nextProps, nextState) {
    if (!this.state.currentView ||
        this.state.currentView[0] !== nextState.currentView[0] ||
        this.state.currentView[1] !== nextState.currentView[1] ||
        this.state.highlightedObject !== nextState.highlightedObject) {
      const titleArray = (nextState.currentView || []).slice(0);
      titleArray.unshift('Vistio');
      document.title = titleArray.join(' / ');

      if (this.poppedState) {
        this.poppedState = false;
      } else if (nextState.currentView) {
        const highlightedObjectName = nextState.highlightedObject && nextState.highlightedObject.getName();
        const state = {
          title: document.title,
          url: nextState.currentView.join('/') + (highlightedObjectName ? `?highlighted=${highlightedObjectName}` : ''),
          selected: nextState.currentView,
          highlighted: highlightedObjectName
        };
        window.history.pushState(state, state.title, state.url);
      }
    }
    return true;
  }

  isSelectedNode () {
    return this.state.currentView && this.state.currentView[1] !== undefined;
  }

  zoomCallback = () => {
    const currentView = _.clone(this.state.currentView);
    if (currentView.length === 1 && this.state.focusedNode) {
      currentView.push(this.state.focusedNode.name);
    } else if (currentView.length === 2) {
      currentView.pop();
    }
    this.setState({ currentView: currentView });
  }

  displayOptionsChanged = (options) => {
    const displayOptions = _.merge({}, this.state.displayOptions, options);
    this.setState({ displayOptions: displayOptions });
  }

  physicsOptionsChanged = (physicsOptions) => {
    this.setState({ currentGraph_physicsOptions: physicsOptions });
    let currentGraph = this.state.currentGraph;
    if (currentGraph == null) currentGraph = null;
    if (currentGraph !== null) {
      currentGraph.setPhysicsOptions(physicsOptions);
    }
  }

  navigationCallback = (newNavigationState) => {
    this.setState({ currentView: newNavigationState });
  }

  detailsClosed = () => {
    // If there is a selected node, deselect the node
    if (this.isSelectedNode()) {
      this.setState({ currentView: [this.state.currentView[0]] });
    } else {
      // If there is just a detailed node, remove the detailed node.
      this.setState({ focusedNode: undefined, highlightedObject: undefined, objectToHighlight: undefined });
    }
  }

  filtersChanged = () => {
    this.setState({
      appliedFilters: filterStore.getChangedFilters(),
      filters: filterStore.getFiltersArray()
    });
  }

  filtersCleared = () => {
    if (!filterStore.isClear()) {
      if (!filterStore.isDefault()) {
        filterActions.resetFilters();
      } else {
        filterActions.clearFilters();
      }
    }
  }

  trafficChanged = () => {
    this.setState({
      traffic: trafficStore.getTraffic(),
      serverUpdatedTime: trafficStore.getLastUpdatedServerTime()
    });
  }

  locatorChanged = (value) => {
    this.setState({ searchTerm: value });
  }

  matchesFound = (matches) => {
    this.setState({ matches: matches });
  }

  nodeClicked = (node) => {
    if (this.state.currentView.length === 1) {
      // highlight node
      this.setState({ objectToHighlight: node.getName() });
    } else if (this.state.currentView.length === 2) {
      // detailed view of node
      this.setState({ currentView: [this.state.currentView[0], node.getName()] });
    }
  }

  offsetChanged = (offset) => {
    trafficActions.updateTrafficOffset(offset);
    this.fetchData();
  };

  resetLayoutButtonClicked = () => {
    const g = this.state.currentGraph;
    if (g != null) {
      g._relayout();
    }
  }

  dismissAlert = () => {
    this.setState({ redirectedFrom: undefined });
  }

  render () {
    const globalView = this.state.currentView && this.state.currentView.length === 0;
    const nodeView = !globalView && this.state.currentView && this.state.currentView[1] !== undefined;
    let nodeToShowDetails = this.state.currentGraph && this.state.currentGraph.focusedNode;
    nodeToShowDetails = nodeToShowDetails || (this.state.highlightedObject && this.state.highlightedObject.type === 'node' ? this.state.highlightedObject : undefined);
    const connectionToShowDetails = this.state.highlightedObject && this.state.highlightedObject.type === 'connection' ? this.state.highlightedObject : undefined;
    const showLoadingCover = !this.state.currentGraph;

    let matches;
    if (this.state.currentGraph) {
      matches = {
        totalMatches: this.state.matches.total,
        visibleMatches: this.state.matches.visible,
        total: this.state.currentGraph.nodeCounts.total,
        visible: this.state.currentGraph.nodeCounts.visible
      };
    }

    return (
      <div className="vizceral-container">
        { this.state.redirectedFrom ?
          <Alert onDismiss={this.dismissAlert}>
            <strong>{this.state.redirectedFrom.join('/') || '/'}</strong> does not exist, you were redirected to <strong>{this.state.currentView.join('/') || '/'}</strong> instead
          </Alert>
        : undefined }
        <div className="subheader">
          <Breadcrumbs rootTitle="global" navigationStack={this.state.currentView || []} navigationCallback={this.navigationCallback} />
          <ReplayClock time={this.state.serverUpdatedTime} maxOffset={this.props.maxReplayOffset * 1000} offsetChanged={offset => this.offsetChanged(offset) } />
          <ServerStatus endpoint={this.props.src} status={this.state.serverStatus} clientUpdatedTime={this.state.clientUpdatedTime} />
          <div style={{ float: 'right', paddingTop: '4px' }}>
            { (!globalView && matches) && <Locator changeCallback={this.locatorChanged} searchTerm={this.state.searchTerm} matches={matches} clearFilterCallback={this.filtersCleared} /> }
            <OptionsPanel title="Filters"><FilterControls /></OptionsPanel>
            <OptionsPanel title="Display"><DisplayOptions options={this.state.displayOptions} changedCallback={this.displayOptionsChanged} /></OptionsPanel>
            <OptionsPanel title="Physics"><PhysicsOptions options={this.state.currentGraph_physicsOptions} changedCallback={this.physicsOptionsChanged}/></OptionsPanel>
            <a role="button" className="reset-layout-link" onClick={this.resetLayoutButtonClicked}>Reload Layout</a><br />
          </div>
        </div>
        <div className="service-traffic-map">
          <div style={{ position: 'absolute', top: '0px', right: nodeToShowDetails || connectionToShowDetails ? '380px' : '0px', bottom: '100px', left: '0px' }}>
            <CustomVizceral traffic={this.state.traffic}
                      view={this.state.currentView}
                      showLabels={this.state.displayOptions.showLabels}
                      filters={this.state.filters}
                      viewChanged={this.viewChanged}
                      viewUpdated={this.viewUpdated}
                      objectHighlighted={this.objectHighlighted}
                      nodeContextSizeChanged={this.nodeContextSizeChanged}
                      objectToHighlight={this.state.objectToHighlight}
                      matchesFound={this.matchesFound}
                      match={this.state.searchTerm}
                      modes={this.state.modes}
                      styles={this.state.styles}
                      allowDraggingOfNodes={this.state.displayOptions.allowDraggingOfNodes}
            />
          </div>
          {
            !!nodeToShowDetails &&
            <DetailsPanelNode node={nodeToShowDetails}
                              nodeSelected={nodeView}
                              region={this.state.currentView[0]}
                              width={panelWidth}
                              zoomCallback={this.zoomCallback}
                              closeCallback={this.detailsClosed}
                              nodeClicked={node => this.nodeClicked(node)}
            />
          }
          {
            !!connectionToShowDetails &&
            <DetailsPanelConnection connection={connectionToShowDetails}
                                    region={this.state.currentView[0]}
                                    width={panelWidth}
                                    closeCallback={this.detailsClosed}
                                    nodeClicked={node => this.nodeClicked(node)}
            />
          }
          <LoadingCover show={showLoadingCover} />
        </div>
      </div>
    );
  }
}

TrafficFlow.propTypes = {
  src: React.PropTypes.string.isRequired,
  interval: React.PropTypes.number,
  maxReplayOffset: React.PropTypes.number
};

export default TrafficFlow;
