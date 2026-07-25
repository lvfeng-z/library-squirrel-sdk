package transport

import (
	"fmt"
	"strconv"
	"strings"
)

// EncodePullRole 编码 PullRequest.role 字段:同 role 多 spec 时附加 specIndex 后缀(role#idx)。
// serveSpecsPull 据 specIndex 在 specs 顺序中定位 reader,替代旧 map[role] 同 role 覆盖
// (旧版 article 29 image 同 role 时只留 1 个 reader,29 个消费者分享 1 张图数据)。
func EncodePullRole(role string, specIndex int) string {
	return fmt.Sprintf("%s#%d", role, specIndex)
}

// DecodePullRole 解析 PullRequest.role:拆 "role#idx";无 # 或解析失败回退 (role, 0)。
func DecodePullRole(encoded string) (role string, specIndex int) {
	if idx := strings.LastIndex(encoded, "#"); idx >= 0 {
		if n, err := strconv.Atoi(encoded[idx+1:]); err == nil && n >= 0 {
			return encoded[:idx], n
		}
	}
	return encoded, 0
}
