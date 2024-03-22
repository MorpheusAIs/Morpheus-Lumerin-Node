import { handleActions } from 'redux-actions';
import {
  EVENT_DEVICES_DEVICE_UPDATED,
  EVENT_DEVICES_RESET,
  EVENT_DEVICES_STATE_UPDATED
} from '../events/devices';

const initialState = {
  isDiscovering: false,
  devices: {},
  error: null
};

const reducer = handleActions(
  {
    [EVENT_DEVICES_DEVICE_UPDATED]: (state, action) => ({
      ...state,
      devices: {
        ...state.devices,
        [action.payload.host]: {
          ...(state.devices[action.payload.host] || {}),
          ...action.payload
        }
      }
    }),
    [EVENT_DEVICES_STATE_UPDATED]: (state, { payload }) => ({
      ...state,
      isDiscovering: payload.isDiscovering
    }),
    [EVENT_DEVICES_RESET]: state => ({
      ...state,
      devices: []
    })
  },
  initialState
);

export default reducer;
