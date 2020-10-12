package pipeline

type Options struct {
	Caps       int32
	BufferSize int
	Handler    func(int32, interface{}) error
}

func NewOptions() *Options {
	return &Options{
		Caps:       256,
		BufferSize: 1024,
		Handler: func(int32, interface{}) error {
			return nil
		},
	}
}
