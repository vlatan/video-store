package worker

type WorkerError struct {
	Err   error
	Stats WorkerStats
}

func (e *WorkerError) Error() string { return e.Err.Error() }
func (e *WorkerError) Unwrap() error { return e.Err }
