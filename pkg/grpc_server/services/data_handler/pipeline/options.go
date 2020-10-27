package pipeline

type Options struct {
	Caps        int32
	WorkerCount int32
	BufferSize  int
	Handler     func(int32, interface{}) error
}

func NewOptions() *Options {
	return &Options{
		Caps:        256,
		WorkerCount: 32,
		BufferSize:  8192,
		Handler: func(int32, interface{}) error {
			return nil
		},
	}
}
