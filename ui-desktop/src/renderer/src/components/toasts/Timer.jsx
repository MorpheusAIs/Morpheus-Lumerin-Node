const Timer = function(callback, delay) {
  let start
  let remaining = delay

  this.timerId = null

  this.pause = function() {
    window.clearTimeout(this.timerId)
    remaining -= new Date() - start
  }

  this.resume = function() {
    start = new Date()
    if (this.timerId) window.clearTimeout(this.timerId)
    this.timerId = window.setTimeout(callback, remaining)
  }

  this.stop = function() {
    if (this.timerId) window.clearTimeout(this.timerId)
    this.timerId = null
  }

  this.resume()
}

export default Timer
