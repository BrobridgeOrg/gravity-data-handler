package pipeline

type Options struct {
	Caps           int32
	WorkerCount    int32
	BufferSize     int
	PrepareHandler func(int32, interface{}) (interface{}, error)
	Handler        func(int32, interface{}) error
}

func NewOptions() *Options {
	return &Options{
		Caps:        256,
		WorkerCount: 32,
		BufferSize:  8192,
		PrepareHandler: func(int32, interface{}) (interface{}, error) {
			return nil, nil
		},
		Handler: func(int32, interface{}) error {
			return nil
		},
	}
}
