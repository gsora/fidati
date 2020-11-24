package u2fhid

import "testing"

type test struct {
	name string
	f    func(*testing.T)
}

type fakeToken struct {
	shouldReturnData bool
	data             []byte
}

func (f *fakeToken) HandleMessage(b []byte) []byte {
	if f.shouldReturnData {
		return f.data
	}

	return nil
}
