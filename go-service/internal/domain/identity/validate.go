package identity

import "unicode/utf8"

// textLen 返回字符串的**字符数**。
//
// 必须用字符数而不是字节数：Python 侧 Pydantic 的 min_length / max_length
// 统计的是字符。若这里用 len()（字节数），一个汉字（UTF-8 占 3 字节）会让
// "最小 2 个字符" 的校验被绕过，Go 侧放行而 Python 侧拒绝，契约不一致。
func textLen(s string) int {
	return utf8.RuneCountInString(s)
}
