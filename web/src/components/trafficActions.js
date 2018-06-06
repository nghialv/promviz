'use strict';

import AppDispatcher from '../appDispatcher';
import AppConstants from '../appConstants';

export default {
  updateTraffic: (traffic) => {
    AppDispatcher.handleAction({
      actionType: AppConstants.ActionTypes.UPDATE_TRAFFIC,
      data: traffic
    });
  },
  clearFilters: () => {
    AppDispatcher.handleAction({
      actionType: AppConstants.ActionTypes.CLEAR_TRAFFIC
    });
  },
  updateTrafficOffset: (offset) => {
    AppDispatcher.handleAction({
      actionType: AppConstants.ActionTypes.UPDATE_TRAFFIC_OFFSET,
      data: offset
    });
  }
};
