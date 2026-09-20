// Package storepath 落盘路径派生——桶段、作品目录段与 store 文件名的唯一权威形态。
//
// 落盘布局（relPath，workDir 相对、正斜杠）：
//
//	store/work/{bucket 2hex}/{siteKey}_{siteWorkId 派生段}/{role}_{seq 三位零填充}{ext}
//	示例：store/work/73/pixiv_128937464/image_000.jpg
//	      store/work/16/bilibili_BV1xx411c7mD_4538792/videoTrack_000.mp4
//
// 桶段 = 复合键（siteKey + "_" + siteWorkId）SHA256 前 2 位小写 hex，共 256 桶：
// 几万作品级库的 store/work 根目录按桶扇出，Explorer/外部工具打开根目录与各
// 桶目录保持流畅；同键恒同桶，路径锚定不变量（重下同复合键恒同路径）保持。
//
// "store/work/" 前缀属主程序库内布局，不在本包——本包只产出桶段、目录段与
// 文件段，完整 relPath 由主程序侧 path.Join 组合；插件可引用本包推导兄弟文件名（如
// document 关联的 image）。作品目录段 = siteKey（站点身份注册表 slug，原文直用）
// + "_" + siteWorkId 派生段（单射：不同 siteWorkId 恒得不同段）。store 文件段
// 恒带 role_seq——单 store 资源不省略段，role 内序号是落盘文件与续传配对的身份键。
package storepath

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

var (
	// ErrEmptySiteKey siteKey 为空——作品目录名要求站点键非空（注册表 slug 恒有值）
	ErrEmptySiteKey = errors.New("siteKey 为空")
	// ErrEmptySiteWorkId siteWorkId 为空——作品目录名要求站点作品 ID 非空
	ErrEmptySiteWorkId = errors.New("siteWorkId 为空")
	// ErrInvalidRole role 非法——须非空、不含控制字符与 Windows 非法字符（含两域路径分隔符）、不以句点或空格收尾
	ErrInvalidRole = errors.New("role 非法")
	// ErrNegativeSeq seq 为负——store 文件名序号要求 >= 0
	ErrNegativeSeq = errors.New("seq 为负")
	// ErrInvalidExt ext 非法——空串（无扩展名）或以点开头（如 .jpg）、不含控制字符与 Windows 非法字符、不以句点或空格收尾
	ErrInvalidExt = errors.New("ext 非法")
)

// 目录段长度预算：派生段按字符计超过 96 进入截断消歧，截断保留净化结果前 88 字符，
// 其后拼接消歧哈希段（"_" + 8 位 hex，共 9 字符）。
const (
	segmentLengthLimit  = 96
	segmentLengthPrefix = 88
)

// windowsIllegalRunes Windows 文件名非法字符（含正斜杠与反斜杠两域路径分隔符）
const windowsIllegalRunes = `\/:*?"<>|`

// WorkDirName 返回作品目录段：{siteKey}_{siteWorkId 派生段}。
//
// siteKey 原文直用——站点身份注册表 slug（^[a-z][a-z0-9-]{1,30}$）由注册表保证，
// 此处仅校验非空。siteWorkId 为站点侧任意 ID，派生规则见 deriveSegment。
func WorkDirName(siteKey, siteWorkId string) (string, error) {
	if siteKey == "" {
		return "", ErrEmptySiteKey
	}
	if siteWorkId == "" {
		return "", ErrEmptySiteWorkId
	}
	return siteKey + "_" + deriveSegment(siteWorkId), nil
}

// StoreFileName 返回 store 文件名：{role}_{seq 三位零填充}{ext}，如 image_000.jpg、
// videoTrack_012.mp4。ext 含前导点（如 ".jpg"），空串 = 无扩展名；seq 超三位自然
// 扩位（如 1000 → "1000"）。role 与 ext 是契约输入（非站点任意数据），不做净化
// 变形——非法值显式报错。
func StoreFileName(role string, seq int, ext string) (string, error) {
	if role == "" || !isFileNameSafe(role) || endsWithDotOrSpace(role) {
		return "", fmt.Errorf("%w：%q", ErrInvalidRole, role)
	}
	if seq < 0 {
		return "", fmt.Errorf("%w：%d", ErrNegativeSeq, seq)
	}
	if ext != "" && (!strings.HasPrefix(ext, ".") || !isFileNameSafe(ext) || endsWithDotOrSpace(ext)) {
		return "", fmt.Errorf("%w：%q", ErrInvalidExt, ext)
	}
	return fmt.Sprintf("%s_%03d%s", role, seq, ext), nil
}

// BucketSegment 返回库内布局的桶段目录名：复合键（siteKey + "_" + siteWorkId）的
// SHA256 前 2 位小写 hex，共 256 桶。siteKey 由注册表 regex（^[a-z][a-z0-9-]{1,30}$，
// 不含下划线）保证，复合键的下划线连接边界无歧义；输入合法性校验由同链路的
// WorkDirName 承担（桶段与作品目录段取同一对 siteKey/siteWorkId），本函数恒返
// 字符串、无 error。
//
// 桶段分摊几万作品级库 store/work 根目录的扇出，Explorer/外部工具打开根目录
// 与各桶目录保持流畅；同键恒同桶，路径锚定不变量（同复合键恒得同库内路径）保持。
func BucketSegment(siteKey, siteWorkId string) string {
	return hashHex(siteKey+"_"+siteWorkId, 2)
}

// deriveSegment 生成 siteWorkId 的目录段，保证单射（不同 siteWorkId 恒得不同段）：
//   - 净化后为空：整体退为原始 ID 的 sha256 前 16 位 hex；
//   - 净化未变更且不超长：原文直用；
//   - 其余（净化发生变更，或派生段超 96 字符）：净化结果截断至前 88 字符，其后
//     拼接 "_" 与原始 ID 的 sha256 前 8 位 hex。净化与截断都可能令不同 ID 得同段
//     （如半角非法字符与其全角等价字符净化后同形、长 ID 共享前缀），消歧哈希段
//     一律按原始 ID 计算，由它承担区分。
func deriveSegment(siteWorkId string) string {
	base := sanitizeSegment(siteWorkId)
	if base == "" {
		return hashHex(siteWorkId, 16)
	}
	if base == siteWorkId && utf8.RuneCountInString(base) <= segmentLengthLimit {
		return base
	}
	return truncateRunes(base, segmentLengthPrefix) + "_" + hashHex(siteWorkId, 8)
}

// segmentSanitizer Windows 非法字符 → 全角等价字符（保形替换）
var segmentSanitizer = strings.NewReplacer(
	`\`, "＼",
	"/", "／",
	":", "：",
	"*", "＊",
	"?", "？",
	`"`, "＂",
	"<", "＜",
	">", "＞",
	"|", "｜",
)

// sanitizeSegment 净化站点侧 ID 为文件系统安全的路径段：Windows 非法字符替换为
// 全角等价字符（保形）；控制字符删除（不可见）；尾部空格与句点剥离——Windows
// 文件系统层对二者静默剥离，目录实际名不含它们。
func sanitizeSegment(id string) string {
	replaced := segmentSanitizer.Replace(id)
	var b strings.Builder
	b.Grow(len(replaced))
	for _, r := range replaced {
		if r <= 0x1F || r == 0x7F {
			continue
		}
		b.WriteRune(r)
	}
	return strings.TrimRight(b.String(), " .")
}

// hashHex 返回 s 的 SHA256 十六进制摘要前 digits 位
func hashHex(s string, digits int) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])[:digits]
}

// truncateRunes 截取 s 前 n 个字符（按字符不按字节，避免在多字节字符中间截断）
func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

// isFileNameSafe 判断 s 可作为文件名组成段：不含控制字符与 Windows 非法字符。
// 首尾空格/句点由调用方按其位置另行校验（目录段净化剥离、契约段整体拒绝）。
func isFileNameSafe(s string) bool {
	for _, r := range s {
		if r <= 0x1F || r == 0x7F || strings.ContainsRune(windowsIllegalRunes, r) {
			return false
		}
	}
	return true
}

// endsWithDotOrSpace 判断 s 是否以句点或空格收尾（Windows 文件系统层会剥离尾部的
// 二者，含它们的文件名与实际落盘名不一致）
func endsWithDotOrSpace(s string) bool {
	return strings.HasSuffix(s, ".") || strings.HasSuffix(s, " ")
}
