# ring-buffer

定长环形缓冲（覆盖写、按序读、容量与复制量统计），只用 Go 标准库。

```
go test ./...
go run ./cmd/ring-buffer --cap 8 --write abcdefghij
go run ./cmd/ring-buffer --cap 8 --ops 100000
```

## 口径（README 为准）

- **写满之后覆盖最旧的数据**：`--cap 8 --write abcdefghij` 按口径应该保留 `cdefghij`（丢掉最旧的 `ab`），
  现在给 `abcdefgh`（新数据被直接丢弃）。
- 连续写入的顺序同样成立：`--cap 8 --ops 100000`（写 0..9 循环）按口径应该是最近 8 个 `23456789`，
  现在给 `01234567`。
- `Write` 返回真正写入的字节数：容量为 c、已有 n 个字节时，一次写 m 个字节的结果是
  `min(m, c-n) + 覆盖掉的数量`（即除「被覆盖后仍留在缓冲区」之外都算写入），口径要求返回 `m`（全部接收）。
- `Bytes()` 按**写入顺序**返回：最旧的在前、最新的在后；`Len()` 不超过 `Cap()`。
- 不变量：`len(Bytes()) == Len()`；写 n 个字节后缓冲区里正好是最后 min(n, cap) 个字节；
  同一操作序列重复执行结果相同；容量为 0 时按 1 处理。
- 规模：10 万次写入加一次读出的总复制量是 O(容量 + 写入量)，
  不许每次 `Bytes()` 都扫全量、也不许写入时整块搬移。

## 实现要点

- `Write` 用最多两段批量 `copy` 写入，满时推进 `head` 覆盖最旧数据，返回全部接收的字节数。
- `Bytes()` 用最多两段批量 `copy` 按写入顺序读出；`Copied()` 计入写入与读出的全部批量复制。
- `head` 随读写真实回绕；`Len()` 永远不超过 `Cap()`，容量 0 按 1 处理。

## 计划新增的能力

```
--cap N      容量
--write 文本 写入一段文本
--ops N      写入 N 个递增数字
```

## 输出

```
content=…
len=… cap=… copied=…
```

## 目录

```
ring.go              环形缓冲
cmd/ring-buffer      命令行入口
ring_test.go         go test 用例
```
