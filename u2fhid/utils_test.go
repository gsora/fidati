package u2fhid

type fakeToken struct {
	shouldReturnData bool
}

func (f *fakeToken) HandleMessage(b []byte) []byte {
	if f.shouldReturnData {
		return []byte("data")
	}

	return nil
}
