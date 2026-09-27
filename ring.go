// Package ringbuffer 是定长环形缓冲。
package ringbuffer

// Buffer 是环形缓冲。缺陷：满了直接丢弃新数据（应该覆盖最旧的），扩容时也不保持顺序。
type Buffer struct {
	data   []byte
	head   int
	size   int
	copied int
}

// New 建一个容量为 capacity 的缓冲。
func New(capacity int) *Buffer {
	if capacity <= 0 {
		capacity = 1
	}
	return &Buffer{data: make([]byte, capacity)}
}

// Len 返回当前长度。
func (b *Buffer) Len() int { return b.size }

// Cap 返回容量。
func (b *Buffer) Cap() int { return len(b.data) }

// Copied 是累计复制字节数（统计用）。
func (b *Buffer) Copied() int { return b.copied }

// Write 写入。缺陷：满了直接丢，不覆盖最旧的。
func (b *Buffer) Write(chunk []byte) int {
	written := 0
	for _, byte := range chunk {
		if b.size == len(b.data) {
			break
		}
		b.data[(b.head+b.size)%len(b.data)] = byte
		b.size++
		written++
	}
	return written
}

// Bytes 按写入顺序返回内容。缺陷：从 head 开始复制，扩容/回绕时会错位。
func (b *Buffer) Bytes() []byte {
	out := make([]byte, 0, b.size)
	for index := 0; index < b.size; index++ {
		b.copied++
		out = append(out, b.data[(b.head+index)%len(b.data)])
	}
	return out
}
