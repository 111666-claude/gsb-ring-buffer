package ringbuffer

import "testing"

func TestNewHasMinCapacity(t *testing.T) {
if New(0).Cap() < 1 {
t.Fatal("容量至少要 1")
}
}

func TestWriteAndBytes(t *testing.T) {
buffer := New(8)
written := buffer.Write([]byte("abc"))
if written != 3 || buffer.Len() != 3 {
t.Fatalf("写入不对：%d %d", written, buffer.Len())
}
if string(buffer.Bytes()) != "abc" {
t.Fatalf("内容不对：%q", string(buffer.Bytes()))
}
}

func TestLenNeverExceedsCap(t *testing.T) {
buffer := New(4)
buffer.Write([]byte("abcdefgh"))
if buffer.Len() != 4 {
t.Fatalf("长度应被容量限制：%d", buffer.Len())
}
}

func TestEmptyBuffer(t *testing.T) {
buffer := New(4)
if buffer.Len() != 0 || len(buffer.Bytes()) != 0 {
t.Fatal("空缓冲应没有内容")
}
}

func TestCapIsReported(t *testing.T) {
if New(16).Cap() != 16 {
t.Fatalf("容量不对：%d", New(16).Cap())
}
}

func TestCopiedIsCounted(t *testing.T) {
buffer := New(4)
buffer.Write([]byte("ab"))
buffer.Bytes()
if buffer.Copied() == 0 {
t.Fatal("复制量应该被统计")
}
}
