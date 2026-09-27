package ringbuffer

import "testing"

func TestNewHasMinCapacity(t *testing.T) {
	if New(0).Cap() < 1 {
		t.Fatal("容量至少要 1")
	}
	if New(-3).Cap() != 1 {
		t.Fatal("非正容量应按 1 处理")
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
	if buffer.Len() > buffer.Cap() {
		t.Fatal("Len 不能超过 Cap")
	}
}

func TestEmptyBuffer(t *testing.T) {
	buffer := New(4)
	if buffer.Len() != 0 || len(buffer.Bytes()) != 0 {
		t.Fatal("空缓冲应没有内容")
	}
	if written := buffer.Write(nil); written != 0 || buffer.Len() != 0 {
		t.Fatalf("空写入不应改变缓冲：%d %d", written, buffer.Len())
	}
	if buffer.Copied() != 0 {
		t.Fatal("空缓冲不应产生复制")
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

func TestOverwriteKeepsNewestAndReturnsAll(t *testing.T) {
	buffer := New(8)
	written := buffer.Write([]byte("abcdefghij"))
	if written != 10 {
		t.Fatalf("Write 应返回全部接收字节数，得到 %d", written)
	}
	if got := string(buffer.Bytes()); got != "cdefghij" {
		t.Fatalf("覆盖写后应保留最后 8 个字节，得到 %q", got)
	}
	if buffer.Len() != 8 {
		t.Fatalf("长度应为容量 8，得到 %d", buffer.Len())
	}
}

func TestWraparoundOrder(t *testing.T) {
	buffer := New(4)
	buffer.Write([]byte("ab"))
	buffer.Write([]byte("cd"))
	if got := string(buffer.Bytes()); got != "abcd" {
		t.Fatalf("写满内容不对：%q", got)
	}
	buffer.Write([]byte("e"))
	if got := string(buffer.Bytes()); got != "bcde" {
		t.Fatalf("覆盖一个最旧字节后不对：%q", got)
	}
	buffer.Write([]byte("fg"))
	if got := string(buffer.Bytes()); got != "defg" {
		t.Fatalf("跨回绕点覆盖后不对：%q", got)
	}
}

func TestExactCapWrites(t *testing.T) {
	for _, c := range []int{1, 2, 5, 8} {
		buffer := New(c)
		chunk := []byte("abcdefghijklmnop")[:c]
		if written := buffer.Write(chunk); written != c {
			t.Fatalf("容量 %d：返回值应为 %d，得到 %d", c, c, written)
		}
		if buffer.Len() != c || string(buffer.Bytes()) != string(chunk) {
			t.Fatalf("容量 %d：恰好写满不对 len=%d bytes=%q", c, buffer.Len(), buffer.Bytes())
		}
		// 再写 c 个字节应整块覆盖。
		next := []byte("STUVWXYZ01234567")[:c]
		buffer.Write(next)
		if string(buffer.Bytes()) != string(next) {
			t.Fatalf("容量 %d：整块覆盖不对：%q", c, buffer.Bytes())
		}
	}
}

func TestCopiedReflectsBatchCopies(t *testing.T) {
	buffer := New(4)
	buffer.Write([]byte("ab")) // 写入 2 字节
	if buffer.Copied() != 2 {
		t.Fatalf("写入后复制量应为 2，得到 %d", buffer.Copied())
	}
	buffer.Write([]byte("cdef")) // m >= cap，整块放 4 字节
	if buffer.Copied() != 6 {
		t.Fatalf("整块覆盖后复制量应为 6，得到 %d", buffer.Copied())
	}
	if got := string(buffer.Bytes()); got != "cdef" {
		t.Fatalf("内容应为 cdef，得到 %q", got)
	}
	if buffer.Copied() != 10 {
		t.Fatalf("读出后复制量应为 10，得到 %d", buffer.Copied())
	}
}

func TestHundredThousandWritesCopiedIsLinear(t *testing.T) {
	buffer := New(8)
	const ops = 100000
	for index := 0; index < ops; index++ {
		buffer.Write([]byte{byte('0' + index%10)})
	}
	content := buffer.Bytes()
	if string(content) != "23456789" {
		t.Fatalf("10 万次写入后应保留最近 8 个，得到 %q", content)
	}
	if buffer.Len() != 8 || len(content) != buffer.Len() {
		t.Fatalf("长度不变量不成立：%d %d", buffer.Len(), len(content))
	}
	// 每次写入批量复制 1 字节，读出最多两段共 8 字节：O(写入量 + 容量)。
	if buffer.Copied() > ops+2*buffer.Cap() {
		t.Fatalf("复制量 %d 超过 O(写入量+容量) 上界", buffer.Copied())
	}
	if buffer.Copied() != ops+8 {
		t.Fatalf("复制量应为 %d，得到 %d", ops+8, buffer.Copied())
	}
}

func TestRepeatableSequence(t *testing.T) {
	sequence := [][]int{{4, 0}, {4, 3}, {4, 9}, {4, 2}, {4, 7}}
	run := func() (string, int, int) {
		buffer := New(4)
		for _, op := range sequence {
			buffer.Write([]byte{byte('0' + op[1])})
		}
		return string(buffer.Bytes()), buffer.Len(), buffer.Copied()
	}
	firstContent, firstLen, firstCopied := run()
	secondContent, secondLen, secondCopied := run()
	if firstContent != secondContent || firstLen != secondLen || firstCopied != secondCopied {
		t.Fatal("同一操作序列重复执行结果不同")
	}
}
