package ringbuffer

import (
	"math/rand"
	"testing"
)

// refBuffer 是「切片 + 头下标」的朴素参考实现，用 append 与切片拼接。
type refBuffer struct {
	data []byte
	head int
	cap  int
}

func newRefBuffer(capacity int) *refBuffer {
	if capacity <= 0 {
		capacity = 1
	}
	return &refBuffer{cap: capacity}
}

func (r *refBuffer) Write(p []byte) int {
	r.data = append(r.data, p...)
	if overflow := len(r.data) - r.head - r.cap; overflow > 0 {
		r.head += overflow
	}
	r.data = append([]byte(nil), r.data[r.head:]...)
	r.head = 0
	return len(p)
}

func (r *refBuffer) Len() int { return len(r.data) - r.head }

func (r *refBuffer) Bytes() []byte {
	return append([]byte(nil), r.data[r.head:]...)
}

func TestNewHasMinCapacity(t *testing.T) {
	if New(0).Cap() != 1 {
		t.Fatalf("容量 0 应按 1 处理：%d", New(0).Cap())
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

func TestOverwriteOldestWhenFull(t *testing.T) {
	buffer := New(8)
	written := buffer.Write([]byte("abcdefghij"))
	if written != 10 {
		t.Fatalf("Write 应返回全部接收的字节数：%d", written)
	}
	if string(buffer.Bytes()) != "cdefghij" {
		t.Fatalf("满后应覆盖最旧数据：%q", string(buffer.Bytes()))
	}
}

func TestWrapAroundAcrossWrites(t *testing.T) {
	buffer := New(8)
	for index := 0; index < 100000; index++ {
		buffer.Write([]byte{byte('0' + index%10)})
	}
	if string(buffer.Bytes()) != "23456789" {
		t.Fatalf("回绕后应保留最近 8 个：%q", string(buffer.Bytes()))
	}
	if buffer.Len() != 8 {
		t.Fatalf("长度应为容量：%d", buffer.Len())
	}
}

func TestLenNeverExceedsCap(t *testing.T) {
	buffer := New(4)
	buffer.Write([]byte("abcdefgh"))
	if buffer.Len() != 4 {
		t.Fatalf("长度应被容量限制：%d", buffer.Len())
	}
	if string(buffer.Bytes()) != "efgh" {
		t.Fatalf("应保留最后 4 个：%q", string(buffer.Bytes()))
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
	before := buffer.Copied()
	buffer.Bytes()
	if buffer.Copied() <= before {
		t.Fatal("读出的批量复制应计入 Copied")
	}
	if before != 2 {
		t.Fatalf("写入 2 字节应复制 2 字节：%d", before)
	}
}

func TestCopiedReflectsBulkComplexity(t *testing.T) {
	const ops = 100000
	buffer := New(8)
	for index := 0; index < ops; index++ {
		buffer.Write([]byte{byte('0' + index%10)})
	}
	buffer.Bytes()
	if limit := ops + buffer.Cap(); buffer.Copied() > limit {
		t.Fatalf("总复制量应为 O(容量+写入量)：copied=%d limit=%d", buffer.Copied(), limit)
	}
}

func TestBytesLenInvariant(t *testing.T) {
	buffer := New(8)
	for _, chunk := range []string{"", "a", "abcdefghij", "xy", "123456789"} {
		buffer.Write([]byte(chunk))
		if len(buffer.Bytes()) != buffer.Len() {
			t.Fatalf("len(Bytes()) != Len()：%d != %d", len(buffer.Bytes()), buffer.Len())
		}
	}
}

func TestDeterministic(t *testing.T) {
	run := func() string {
		buffer := New(8)
		for _, chunk := range []string{"abc", "defghij", "klmno"} {
			buffer.Write([]byte(chunk))
		}
		return string(buffer.Bytes())
	}
	if run() != run() {
		t.Fatal("同一操作序列重复执行结果应相同")
	}
}

func TestAgainstReference(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for trial := 0; trial < 200; trial++ {
		capacity := 1 + rng.Intn(16)
		buffer := New(capacity)
		reference := newRefBuffer(capacity)
		for op := 0; op < 200; op++ {
			chunk := make([]byte, rng.Intn(2*capacity+2))
			if _, err := rng.Read(chunk); err != nil {
				t.Fatal(err)
			}
			gotWritten := buffer.Write(chunk)
			wantWritten := reference.Write(chunk)
			if gotWritten != wantWritten {
				t.Fatalf("cap=%d 写入返回值不一致：%d != %d", capacity, gotWritten, wantWritten)
			}
			if buffer.Len() != reference.Len() {
				t.Fatalf("cap=%d Len 不一致：%d != %d", capacity, buffer.Len(), reference.Len())
			}
			got, want := string(buffer.Bytes()), string(reference.Bytes())
			if got != want {
				t.Fatalf("cap=%d Bytes 不一致：%q != %q", capacity, got, want)
			}
		}
	}
}
