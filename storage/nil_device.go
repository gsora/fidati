package storage

// NilDevice implements io.ReadWriter, returns successful data on both
// Read and Write, but will not do anything.
type NilDevice struct {
}

func (n NilDevice) Write(p []byte) (i int, err error) {
	return len(p), nil
}

func (n NilDevice) Read(p []byte) (i int, err error) {
	return len(p), nil
}
