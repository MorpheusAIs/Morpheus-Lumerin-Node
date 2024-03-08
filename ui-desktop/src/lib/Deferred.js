const Deferred = function() {
  const obj = this;
  obj.promise = new Promise(function(resolve, reject) {
    Object.assign(obj, { resolve, reject });
  });
};

module.exports = Deferred;
