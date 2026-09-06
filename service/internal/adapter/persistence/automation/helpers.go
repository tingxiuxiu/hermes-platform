package automation

import (
	"hash/fnv"
)

// hashString 把字符串哈希成 int64，供 pg_advisory_xact_lock 作为锁键。
// 用 FNV-1a 保证确定性与跨进程一致性（同一输入永远得到同一锁键）。
func hashString(s string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	// 转成 int64 时去掉符号位，避免负锁键（advisory lock 键通常用正数）
	return int64(h.Sum64() & 0x7fffffffffffffff)
}

// nullIfEmpty 把空字符串编码为 NULL。
func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// nullLabelArray 把空标签切片编码为 NULL。
//
// labels 列是 VARCHAR(64)[]，必须传 []string 让 pgx 走原生数组编码，
// 不能传 JSON 字符串——后者会被 PG 当作单个文本而非数组。
func nullLabelArray(labels []string) any {
	if len(labels) == 0 {
		return nil
	}
	return labels
}
