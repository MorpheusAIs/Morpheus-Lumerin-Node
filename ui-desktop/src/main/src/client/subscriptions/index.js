'use strict'

import { subscribeSingleCore, unsubscribeSingleCore } from './single-core'
import { subscribeWithoutCore, unsubscribeWithoutCore } from './no-core'

function subscribe(core) {
  console.log("SUBSCRIBING");
  subscribeSingleCore(core)
  subscribeWithoutCore()
}

function unsubscribe(core) {
  unsubscribeSingleCore(core)
  unsubscribeWithoutCore()
}

export default { subscribe, unsubscribe }
