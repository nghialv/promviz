'use strict';

import _ from 'lodash';
import EventEmitter from 'events';

import AppDispatcher from '../appDispatcher';
import AppConstants from '../appConstants';

const CHANGE_EVENT = 'change';

const defaultFilters = {
  rps: { value: -1 },
  error: { value: -1 },
  clas: { value: [] },
  notice: { value: -1 },
};

const noFilters = {
  rps: { value: -1 },
  error: { value: -1 },
  clas: { value: [] },
  notice: { value: -1 },
};

const store = {
  filters: {
    rps: {
      name: 'rps',
      type: 'connection',
      passes: (object, value) => object.volumeTotal >= value,
      value: -1
    },
    error: {
      name: 'error',
      type: 'connection',
      passes: (object, value) => (value === -1 && !object.volumePercent.danger) || object.volumePercent.danger >= value,
      value: -1
    },
    clas: {
      name: 'clas',
      type: 'node',
      passes: (object, value) => value.length <= 0 || value.indexOf(object.class || '') >= 0,
      value: []
    },
    notice: {
      name: 'notice',
      type: 'connection',
      passes: (object, value) => {
        if (!object.notices || object.notices.length === 0) {
          return value === -1;
        }
        return _.some(object.notices, notice => notice.severity >= value);
      },
      value: -1
    },

  },
  states: {
    rps: [
      {
        name: 'high(>1000)',
        value: 1000
      },
      {
        name: '(>300)',
        value: 300
      },
      {
        name: '(>5)',
        value: 5
      },
      {
        name: 'all',
        value: -1
      }
    ],
    error: [
      {
        name: 'high(>10)',
        value: 0.10
      },
      {
        name: '(>5)',
        value: 0.05
      },
      {
        name: '(>1)',
        value: 0.01
      },
      {
        name: 'all',
        value: -1
      }
    ],
    notice: [
      {
        name: 'danger',
        value: 2
      },
      {
        name: 'warning',
        value: 1
      },
      {
        name: 'info',
        value: 0
      },
      {
        name: 'all',
        value: -1
      }
    ]
  }
};

const resetDefaults = function () {
  _.mergeWith(store.filters, defaultFilters, (obj, src) => {
    if (_.isArray(obj)) {
      return src;
    }
    return undefined;
  });
};

const clearFilters = function () {
  _.mergeWith(store.filters, noFilters, (obj, src) => {
    if (_.isArray(obj)) {
      return src;
    }
    return undefined;
  });
};

resetDefaults();

class FilterStore extends EventEmitter {
  constructor () {
    super();
    this.requests = {};

    AppDispatcher.register((payload) => {
      const action = payload.action;
      switch (action.actionType) {
      case AppConstants.ActionTypes.UPDATE_FILTER:
        this.updateFilters(action.data);
        this.emit(CHANGE_EVENT);
        break;
      case AppConstants.ActionTypes.UPDATE_DEFAULT_FILTERS:
        this.updateDefaultFilters(action.data);
        this.emit(CHANGE_EVENT);
        break;
      case AppConstants.ActionTypes.RESET_FILTERS:
        resetDefaults();
        this.emit(CHANGE_EVENT);
        break;
      case AppConstants.ActionTypes.CLEAR_FILTERS:
        clearFilters();
        this.emit(CHANGE_EVENT);
        break;
      default:
        return true;
      }
      return true;
    });
  }

  addChangeListener (cb) {
    this.on(CHANGE_EVENT, cb);
  }

  removeChangeListener (cb) {
    this.removeListener(CHANGE_EVENT, cb);
  }

  getDefaultFilters () {
    return defaultFilters;
  }

  getFilters () {
    return store.filters;
  }

  getFiltersArray () {
    return _.map(store.filters, filter => _.clone(filter));
  }

  getStates () {
    return store.states;
  }

  getChangedFilters () {
    return _.filter(store.filters, filter => filter.value !== defaultFilters[filter.name].value);
  }

  getStepFromValue (name) {
    const index = _.findIndex(store.states[name], step => step.value === store.filters[name].value);
    if (index === -1) {
      return _.findIndex(store.states[name], step => step.value === defaultFilters[name].value);
    }
    return index;
  }

  updateFilters (filters) {
    Object.keys(filters).forEach((filter) => {
      store.filters[filter].value = filters[filter];
    });
  }

  updateDefaultFilters (defaults) {
    _.merge(defaultFilters, defaults);
  }

  isLastClass (clas) {
    return store.filters.clas.value.length === 1 && store.filters.clas.value.indexOf(clas) === 0;
  }

  isDefault () {
    return _.every(store.filters, filter => filter.value === defaultFilters[filter.name].value);
  }

  isClear () {
    return _.every(store.filters, filter => filter.value === noFilters[filter.name].value);
  }
}

const filterStore = new FilterStore();

export default filterStore;
