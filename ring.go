// Package ringbuffer 是定长环形缓冲：覆盖写、按序读、容量与复制量统计。
package ringbuffer

// Buffer 是定长环形缓冲。写满后继续写会覆盖最旧的数据。
type Buffer struct {
	data   []byte
	head   int
	size   int
	copied int
}

// New 建一个容量为 capacity 的缓冲；capacity 小于 1 时按 1 处理。
func New(capacity int) *Buffer {
	if capacity <= 0 {
		capacity = 1
	}
	return &Buffer{data: make([]byte, capacity)}
}

// Len 返回当前长度，永远不超过 Cap()。
func (b *Buffer) Len() int { return b.size }

// Cap 返回容量。
func (b *Buffer) Cap() int { return len(b.data) }

// Copied 是累计复制字节数（写入与读出的批量复制都计入）。
func (b *Buffer) Copied() int { return b.copied }

// Write 写入 chunk，返回全部接收的字节数；缓冲满时覆盖最旧的数据。
func (b *Buffer) Write(chunk []byte) int {
	n := len(chunk)
	if n == 0 {
		return 0
	}
	capacity := len(b.data)
	if n >= capacity {
		// 只保留最后 capacity 个字节，整段覆盖。
		chunk = chunk[n-capacity:]
		b.copied += copy(b.data, chunk)
		b.head = 0
		b.size = capacity
		return n
	}
	tail := (b.head + b.size) % capacity
	first := capacity - tail
	if first > n {
		first = n
	}
	b.copied += copy(b.data[tail:tail+first], chunk[:first])
	b.copied += copy(b.data[:n-first], chunk[first:])
	b.size += n
	if b.size > capacity {
		b.head = (b.head + b.size - capacity) % capacity
		b.size = capacity
	}
	return n
}

// Bytes 按写入顺序返回内容：最旧的在前、最新的在后。最多两段 copy。
func (b *Buffer) Bytes() []byte {
	out := make([]byte, b.size)
	if b.size == 0 {
		return out
	}
	capacity := len(b.data)
	first := capacity - b.head
	if first > b.size {
		first = b.size
	}
	b.copied += copy(out, b.data[b.head:b.head+first])
	b.copied += copy(out[first:], b.data[:b.size-first])
	return out
}
