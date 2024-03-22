'use strict';

import { app } from 'electron';
import ua from 'universal-analytics';
import { noop } from 'lodash';
import settings from 'electron-settings';
import isDev from 'electron-is-dev';

let visitor;
const analytics = {};

analytics.screenview = function (...args) {
  if (!visitor) return;
  visitor.screenview(...args).send();
}

analytics.event = function (...args) {
  if (!visitor) return;
  visitor.event(...args).send();
}

analytics.init = function (userAgent) {
  visitor = ua(settings.getSync('app.trackingId'));
  visitor.set('ds', 'app');
  visitor.set('an', app.getName());
  visitor.set('av', app.getVersion());
  visitor.set('ua', userAgent);
  analytics.event('App', 'App initiated');
}

const analyticsDev = {
  init: noop,
  event: noop,
  screenview: noop
};

export default isDev ? analyticsDev : analytics;
