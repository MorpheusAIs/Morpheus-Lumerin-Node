'use strict';

const { app } = require('electron');

function restart (status = 0) {
  app.relaunch({ args: process.argv.slice(1).concat(['--relaunch']) });
  app.exit(status);
  return {};
}

module.exports = restart;
