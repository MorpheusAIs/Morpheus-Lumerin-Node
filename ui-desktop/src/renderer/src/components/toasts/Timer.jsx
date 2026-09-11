const Timer = function(callback, delay) {
  let start
  let remaining = delay

  this.timerId = null
  // `timerId` alone cannot tell "paused under the cursor" apart from "never
  // started", and callers were using it for exactly that. Track it explicitly.
  this.paused = false

  this.pause = function() {
    if (this.paused) return
    this.paused = true
    window.clearTimeout(this.timerId)
    this.timerId = null
    remaining -= new Date() - start
  }

  this.resume = function() {
    this.paused = false
    start = new Date()
    if (this.timerId) window.clearTimeout(this.timerId)
    // A toast held open past its deadline should still close promptly on
    // mouse-out rather than vanish the instant the cursor leaves.
    this.timerId = window.setTimeout(callback, Math.max(remaining, 0))
  }

  this.stop = function() {
    this.paused = false
    if (this.timerId) window.clearTimeout(this.timerId)
    this.timerId = null
  }

  this.resume()
}

export default Timer
