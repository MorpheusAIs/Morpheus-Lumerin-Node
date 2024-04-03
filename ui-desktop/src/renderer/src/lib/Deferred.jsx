export default class Defered {
  constructor() {
    const obj = this
    obj.promise = new Promise(function (resolve, reject) {
      Object.assign(obj, { resolve, reject })
    })
  }
}
