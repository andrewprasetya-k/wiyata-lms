package service

// runAsync executes fn in a new goroutine and recovers from any panic inside it,
// so that best-effort background work (e.g. notification fan-out) never crashes
// the process or affects the request that triggered it.
func runAsync(fn func()) {
	go func() {
		defer func() {
			_ = recover()
		}()
		fn()
	}()
}
