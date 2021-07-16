import Base from './base';
import Items from './items';
import { Plaid } from './plaid';

export default class Ledger extends Base {
  #items;
  #plaid;
  constructor() {
    super();
    this.#items = new Items();
    this.#plaid = new Plaid();
  }

  plaid() {
    return this.#plaid;
  }

  items() {
    return this.#items;
  }
}
