// Package ringbuffer 是定长环形缓冲：写满后覆盖最旧数据，按写入顺序读出。
package ringbuffer

// Buffer 是环形缓冲。head 指向最旧字节，size 是当前有效字节数。
type Buffer struct {
	data   []byte
	head   int
	size   int
	copied int
}

// New 建一个容量为 capacity 的缓冲，capacity 小于 1 时按 1 处理。
func New(capacity int) *Buffer {
	if capacity < 1 {
		capacity = 1
	}
	return &Buffer{data: make([]byte, capacity)}
}

// Len 返回当前长度，永远不超过 Cap。
func (b *Buffer) Len() int { return b.size }

// Cap 返回容量。
func (b *Buffer) Cap() int { return len(b.data) }

// Copied 返回累计复制字节数（写入与读出中批量 copy 的字节）。
func (b *Buffer) Copied() int { return b.copied }

// Write 写入 chunk，缓冲满时覆盖最旧字节，返回接收的全部字节数 len(chunk)。
// 最多两段批量 copy，不逐字节搬移。
func (b *Buffer) Write(chunk []byte) int {
	c := len(b.data)
	if len(chunk) >= c {
		// 新数据本身足以填满整个缓冲，旧数据全部被覆盖，只保留最后 c 个字节。
		tail := chunk[len(chunk)-c:]
		b.copied += copy(b.data, tail)
		b.head = 0
		b.size = c
		return len(chunk)
	}

	free := c - b.size
	if len(chunk) > free {
		// 需要覆盖最旧数据：head 前移被覆盖的字节数。
		b.head = (b.head + len(chunk) - free) % c
		b.size = c
	} else {
		b.size += len(chunk)
	}

	// 新数据从最新字节之后（即 head+size 的旧值）开始写入，最多绕一次。
	start := (b.head + b.size + c - len(chunk)) % c
	n := copy(b.data[start:], chunk)
	b.copied += n
	if n < len(chunk) {
		n2 := copy(b.data, chunk[n:])
		b.copied += n2
	}
	return len(chunk)
}

// Bytes 按写入顺序返回内容的副本（最旧在前、最新在后），最多两段批量 copy。
func (b *Buffer) Bytes() []byte {
	out := make([]byte, b.size)
	if b.size == 0 {
		return out
	}
	n := copy(out, b.data[b.head:])
	b.copied += n
	if n < b.size {
		n2 := copy(out[n:], b.data[:b.size-n])
		b.copied += n2
	}
	return out
}
