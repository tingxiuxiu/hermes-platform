// Package password 提供 Argon2id 密码哈希（ADR-0005）。
//
// Python 版用 pwdlib 混合 Argon2 + Bcrypt 存储历史哈希；
// Go 版不兼容旧哈希（Python 仍在开发期，评审已确认），统一使用 Argon2id。
package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Params 是 Argon2id 的计算参数。
type Params struct {
	Time    uint32 // 迭代次数
	Memory  uint32 // 内存开销，单位 KiB
	Threads uint8  // 并行度
	SaltLen uint32 // 盐长度，字节
	KeyLen  uint32 // 输出密钥长度，字节
}

// DefaultParams 是生产参数，对齐 OWASP 推荐下限：
// memory=64MiB, time=1, threads=4。
var DefaultParams = Params{
	Time:    1,
	Memory:  64 * 1024,
	Threads: 4,
	SaltLen: 16,
	KeyLen:  32,
}

// TestParams 是测试参数，成本低到可以在单测里跑几百次。
// 只能在测试中使用：64KiB 内存无法抵抗真实攻击。
var TestParams = Params{
	Time:    1,
	Memory:  64,
	Threads: 1,
	SaltLen: 16,
	KeyLen:  32,
}

// Hasher 按给定参数执行哈希与校验。
type Hasher struct {
	params Params
}

// New 用默认（生产）参数构造 Hasher。
func New() *Hasher { return &Hasher{params: DefaultParams} }

// NewWithParams 用自定义参数构造。测试通过它使用 TestParams。
func NewWithParams(p Params) *Hasher { return &Hasher{params: p} }

// Params 返回当前参数，供启动自检与日志使用。
func (h *Hasher) Params() Params { return h.params }

// Hash 生成 PHC 字符串格式的哈希：
//
//	$argon2id$v=19$m=65536,t=1,p=4$<base64 salt>$<base64 hash>
//
// 用 RawStdEncoding（无 padding），与主流实现（argon2 CLI、libsodium）一致。
func (h *Hasher) Hash(plain string) (string, error) {
	if plain == "" {
		return "", errors.New("password: plain password must not be empty")
	}

	salt := make([]byte, h.params.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("password: generate salt: %w", err)
	}

	key := argon2.IDKey(
		[]byte(plain), salt,
		h.params.Time, h.params.Memory, h.params.Threads,
		h.params.KeyLen,
	)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		h.params.Memory, h.params.Time, h.params.Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

// Verify 校验明文密码是否匹配已存储的哈希。
// 使用常量时间比较，避免通过响应时间侧信道泄漏信息。
func (h *Hasher) Verify(plain, encoded string) (bool, error) {
	if plain == "" || encoded == "" {
		return false, nil
	}

	params, salt, key, err := decode(encoded)
	if err != nil {
		return false, err
	}

	candidate := argon2.IDKey(
		[]byte(plain), salt,
		params.Time, params.Memory, params.Threads,
		uint32(len(key)),
	)
	return subtle.ConstantTimeCompare(key, candidate) == 1, nil
}

// NeedsRehash 判断哈希是否使用了过期参数，用于静默升级。
// 本轮不启用（Go 版不兼容旧哈希），保留接口以备后续调参。
func (h *Hasher) NeedsRehash(encoded string) bool {
	params, _, _, err := decode(encoded)
	if err != nil {
		return true
	}
	return params != h.params
}

// ErrInvalidHash 表示哈希字符串格式无法解析。
var ErrInvalidHash = errors.New("password: malformed hash")

func decode(encoded string) (Params, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	// ["", "argon2id", "v=19", "m=...,t=...,p=...", salt, hash]
	if len(parts) != 6 {
		return Params{}, nil, nil, ErrInvalidHash
	}
	if parts[1] != "argon2id" {
		return Params{}, nil, nil, fmt.Errorf("%w: unsupported algorithm %q", ErrInvalidHash, parts[1])
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return Params{}, nil, nil, fmt.Errorf("%w: bad version: %v", ErrInvalidHash, err)
	}
	if version != argon2.Version {
		return Params{}, nil, nil, fmt.Errorf("%w: unsupported version %d", ErrInvalidHash, version)
	}

	var p Params
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.Memory, &p.Time, &p.Threads); err != nil {
		return Params{}, nil, nil, fmt.Errorf("%w: bad params: %v", ErrInvalidHash, err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return Params{}, nil, nil, fmt.Errorf("%w: bad salt: %v", ErrInvalidHash, err)
	}
	key, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return Params{}, nil, nil, fmt.Errorf("%w: bad hash: %v", ErrInvalidHash, err)
	}

	p.SaltLen = uint32(len(salt))
	p.KeyLen = uint32(len(key))
	return p, salt, key, nil
}
