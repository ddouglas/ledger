import Base from './base';

export default class Items extends Base {
  items() {
    return this.request(
      {
        method: 'get',
        url: '/items',
      },
      true
    );
  }
  createItem(data) {
    return this.request(
      {
        method: 'post',
        url: '/items',
        data: data,
      },
      true
    );
  }

  deleteItem(itemID) {
    return this.request(
      {
        method: 'delete',
        url: `/items/${itemID}`,
      },
      true
    );
  }
}
