package pipeline

type Options struct {
	Caps       int32
	BufferSize int
	Handler    func(interface{}) error
}

func NewOptions() *Options {
	return &Options{
		Caps:       64,
		BufferSize: 1024,
		Handler: func(interface{}) error {
			return nil
		},
	}
}
