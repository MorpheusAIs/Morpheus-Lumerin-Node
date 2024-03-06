'use strict';

function createEventsRegistry () {
  const registeredEvents = [];

  return {
    getAll: () => registeredEvents,
    register: registration => registeredEvents.push(registration)
  };
}

module.exports = createEventsRegistry;
