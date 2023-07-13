package outbound

type Blackhole struct{}

func NewBlackhole() Blackhole {
	return Blackhole{}
}

func (Blackhole) Read(data []byte) (int, error) {
	return 0, nil
}

func (Blackhole) Write(data []byte) (int, error) {
	return len(data), nil
}

func (Blackhole) Close() error {
	return nil
}
