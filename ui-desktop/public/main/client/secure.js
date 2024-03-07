const secureStore = require('keytar');

const APP_NAME = 'lumerin-wallet';
const SECURE_PASSWORD_KEY = 'password';

module.exports = {
  APP_NAME,
  SECURE_PASSWORD_KEY,
  setSessionPassword(password) {
    return secureStore
      .setPassword(APP_NAME, SECURE_PASSWORD_KEY, password)
      .catch(e => {
        console.log('Failed to set password in keystore: ', e);
      });
  },
  getSessionPassword() {
    return secureStore.getPassword(APP_NAME, SECURE_PASSWORD_KEY);
  }
};
