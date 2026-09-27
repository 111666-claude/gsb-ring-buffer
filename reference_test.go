package ringbuffer

import (
	"math/rand"
	"testing"
)

// refBuffer 是「切片 + 头下标」的朴素参考实现：
// 用 append 与切片拼接保留最后 min(n, cap) 个字节，不追求复制效率，只用于对照行为。
type refBuffer struct {
	data []byte
	head int
	cap  int
}

func newRefBuffer(capacity int) *refBuffer {
	if capacity < 1 {
		capacity = 1
	}
	return &refBuffer{data: make([]byte, 0, capacity), cap: capacity}
}

func (b *refBuffer) Len() int { return len(b.data) }

func (b *refBuffer) Cap() int { return b.cap }

// Write 用 append 拼接新旧数据，再切掉超出容量的最旧部分。
func (b *refBuffer) Write(chunk []byte) int {
	b.data = append(b.data, chunk...)
	if len(b.data) > b.cap {
		b.data = b.data[len(b.data)-b.cap:]
	}
	b.head = 0
	return len(chunk)
}

// Bytes 用切片拼接按写入顺序（头下标之后、再绕回开头）返回内容。
func (b *refBuffer) Bytes() []byte {
	if len(b.data) == 0 {
		return []byte{}
	}
	out := append([]byte{}, b.data[b.head:]...)
	if b.head > 0 {
		out = append(out, b.data[:b.head]...)
	}
	return out
}

func TestMatchesReferenceSamples(t *testing.T) {
	cases := []struct {
		capacity int
		writes   []string
	}{
		{8, []string{"abcdefghij"}},
		{8, nil},
		{4, []string{"ab", "cd", "e", "fg"}},
		{1, []string{"a", "b", "cd", "e"}},
		{3, []string{"abc", "defgh", "i"}},
		{0, []string{"x", "yz"}},
	}
	for _, tc := range cases {
		got := New(tc.capacity)
		want := newRefBuffer(tc.capacity)
		for _, chunk := range tc.writes {
			g := got.Write([]byte(chunk))
			w := want.Write([]byte(chunk))
			if g != w {
				t.Fatalf("容量 %d 写入 %q 返回值不一致：%d != %d", tc.capacity, chunk, g, w)
			}
		}
		assertMatchesReference(t, got, want)
	}
}

func TestMatchesReferenceRandom(t *testing.T) {
	for seed := int64(0); seed < 50; seed++ {
		rng := rand.New(rand.NewSource(seed))
		capacity := 1 + rng.Intn(16)
		got := New(capacity)
		want := newRefBuffer(capacity)
		for op := 0; op < 200; op++ {
			chunk := make([]byte, rng.Intn(capacity*2+1))
			for i := range chunk {
				chunk[i] = byte('0' + rng.Intn(10))
			}
			g := got.Write(chunk)
			w := want.Write(chunk)
			if g != w {
				t.Fatalf("seed=%d op=%d 返回值不一致：%d != %d", seed, op, g, w)
			}
			assertMatchesReference(t, got, want)
		}
	}
}

func assertMatchesReference(t *testing.T, got *Buffer, want *refBuffer) {
	t.Helper()
	if got.Cap() != want.Cap() {
		t.Fatalf("Cap 不一致：%d != %d", got.Cap(), want.Cap())
	}
	if got.Len() != want.Len() {
		t.Fatalf("Len 不一致：%d != %d", got.Len(), want.Len())
	}
	if got.Len() > got.Cap() {
		t.Fatalf("Len %d 超过 Cap %d", got.Len(), got.Cap())
	}
	if string(got.Bytes()) != string(want.Bytes()) {
		t.Fatalf("Bytes 不一致：%q != %q", got.Bytes(), want.Bytes())
	}
	if len(got.Bytes()) != got.Len() {
		t.Fatal("len(Bytes()) != Len()")
	}
}
