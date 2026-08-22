package state

// Effect runs a reactive side-effect function that tracks state dependencies.
// If the effect function returns a cleanup func(), it will be executed before the next run or upon disposal.
func Effect(fn func() func()) func() {
	var cleanup func()
	var unsubs []func()
	var disposed bool

	var run func()
	run = func() {
		if disposed {
			return
		}
		if cleanup != nil {
			cleanup()
			cleanup = nil
		}
		for _, unsub := range unsubs {
			unsub()
		}
		unsubs = nil

		ctx := &trackerContext{}
		trackerMutex.Lock()
		ctx.parent = activeTracker
		activeTracker = ctx
		trackerMutex.Unlock()

		cleanup = fn()

		trackerMutex.Lock()
		activeTracker = ctx.parent
		deps := ctx.dependencies
		trackerMutex.Unlock()

		for _, dep := range deps {
			switch d := dep.(type) {
			case *Value[int]:
				unsub := d.Subscribe(func(int) { run() })
				unsubs = append(unsubs, unsub)
			case *Value[string]:
				unsub := d.Subscribe(func(string) { run() })
				unsubs = append(unsubs, unsub)
			case *Value[bool]:
				unsub := d.Subscribe(func(bool) { run() })
				unsubs = append(unsubs, unsub)
			case *Value[float64]:
				unsub := d.Subscribe(func(float64) { run() })
				unsubs = append(unsubs, unsub)
			case subscribableWithCallback:
				unsub := d.subscribeGeneric(func() { run() })
				unsubs = append(unsubs, unsub)
			}
		}
	}

	run()

	return func() {
		disposed = true
		if cleanup != nil {
			cleanup()
			cleanup = nil
		}
		for _, unsub := range unsubs {
			unsub()
		}
		unsubs = nil
	}
}

// Batch groups multiple state updates and dispatches notifications all at once at the end.
func Batch(fn func()) {
	trackerMutex.Lock()
	batchDepth++
	trackerMutex.Unlock()

	defer func() {
		trackerMutex.Lock()
		batchDepth--
		var callbacks []func()
		if batchDepth == 0 {
			callbacks = pendingNotify
			pendingNotify = nil
		}
		trackerMutex.Unlock()

		for _, notify := range callbacks {
			notify()
		}
	}()

	fn()
}
