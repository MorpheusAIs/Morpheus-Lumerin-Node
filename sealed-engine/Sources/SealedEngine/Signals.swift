import Darwin

enum Signals {
    /// Every catchable fatal signal exits at once, before the crash reporter can
    /// snapshot thread registers. SIGKILL cannot be caught and leaves no report.
    static func installExitHandlers() {
        let fatal: [Int32] = [SIGSEGV, SIGBUS, SIGILL, SIGFPE, SIGTRAP, SIGABRT, SIGSYS, SIGQUIT, SIGEMT]
        for sig in fatal {
            var action = sigaction()
            action.__sigaction_u.__sa_handler = { _ in _exit(113) }
            sigemptyset(&action.sa_mask)
            action.sa_flags = 0
            sigaction(sig, &action, nil)
        }
    }
}
