import { createSelector } from 'reselect';

export const getDevices = state => state.devices;

export const getDevicesList = createSelector(getDevices, devices => devices);
