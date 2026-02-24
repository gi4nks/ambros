package utils

import "bytes"

type LimitedBuffer struct {
	buf       bytes.Buffer
	max       int64
	n         int64
	truncated bool
}

func NewLimitedBuffer(max int64) *LimitedBuffer {
	return &LimitedBuffer{max: max}
}

func (b *LimitedBuffer) Write(p []byte) (int, error) {
	if b.max <= 0 {
		_, _ = b.buf.Write(p)
		b.n += int64(len(p))
		return len(p), nil
	}
	if b.truncated {
		return len(p), nil
	}
	remaining := b.max - b.n
	if remaining <= 0 {
		b.truncated = true
		return len(p), nil
	}
	if int64(len(p)) > remaining {
		_, _ = b.buf.Write(p[:remaining])
		b.n += remaining
		b.truncated = true
		return len(p), nil
	}
	_, _ = b.buf.Write(p)
	b.n += int64(len(p))
	return len(p), nil
}

func (b *LimitedBuffer) String() string {
	return b.buf.String()
}

func (b *LimitedBuffer) Truncated() bool {
	return b.truncated
}
