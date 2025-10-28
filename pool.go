package zeroslog

import (
	"bytes"
	"sync"
)

var (
	bufPool = sync.Pool{
		New: func() any {
			return new(bytes.Buffer)
		},
	}

	tsPool = sync.Pool{
		New: func() any {
			b := make([]byte, 0, 32)

			return &b
		},
	}
)
