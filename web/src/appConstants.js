'use strict';

import keymirror from 'keymirror';

const AppConstants = {
  ActionTypes: keymirror({
    SERVER_ACTION: null,
    VIEW_ACTION: null,

    UPDATE_FILTER: null,
    UPDATE_DEFAULT_FILTERS: null,
    RESET_FILTERS: null,
    CLEAR_FILTERS: null,

    UPDATE_TRAFFIC: null,
    CLEAR_TRAFFIC: null,
    UPDATE_TRAFFIC_OFFSET: null
  }),
  ServerStatus: keymirror({
    DISCONNECTED: null,
    CONNECTED: null
  })
};

export default AppConstants;
