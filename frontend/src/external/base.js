import axios from 'axios';
import { getInstance } from '@/auth';

export default class Base {
  #baseURL;
  constructor() {
    this.#baseURL = 'https://ledger.onetwentyseven.dev/api';
  }

  async request(axConfig, useAuth = false) {
    axConfig.baseURL = this.#baseURL;
    let headers = {};
    if (useAuth) {
      const authService = getInstance();
      const token = await authService.getTokenSilently();
      headers['Authorization'] = `Bearer ${token}`;
    }

    axConfig.headers = headers;

    console.log(axConfig);

    return axios.request(axConfig).catch(async err => {
      // Handle expired sessions
      const { status, data } = err.response;
      let message = data.message;
      const promises = [];
      if (status === 401) {
        promises.push(
          store.dispatch('storeToken', null),
          store.dispatch('storeUser', null),
          store.dispatch('setRedirectURL', window.location.pathname),
          router.push('/')
        );
        if (data.message == 'token is expired') {
          message = 'Session has expired. Please log back into to continue';
          promises.push(
            store.dispatch('storeAlertProps', {
              message: message,
              variant: 'error'
            })
          );
        }
        return Promise.all(promises);
      }
      return Promise.reject(err);
    });
  }
}
