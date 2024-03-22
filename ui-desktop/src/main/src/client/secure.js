import secureStore from 'keytar';
const APP_NAME = 'lumerin-wallet';
const SECURE_PASSWORD_KEY = 'password';

export function setSessionPassword(password) {
  return secureStore
    .setPassword(APP_NAME, SECURE_PASSWORD_KEY, password)
    .catch(e => {
      console.log('Failed to set password in keystore: ', e);
    });
}

export function getSessionPassword() {
  return secureStore.getPassword(APP_NAME, SECURE_PASSWORD_KEY);
}

export default {
  APP_NAME,
  SECURE_PASSWORD_KEY,
  setSessionPassword,
  getSessionPassword
};
