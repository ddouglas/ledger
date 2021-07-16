import Base from './base';

export class Plaid extends Base {
  linkToken() {
    return this.request(
      {
        method: 'get',
        url: '/external/plaid/v1/link/token'
      },
      true
    );
  }
}
