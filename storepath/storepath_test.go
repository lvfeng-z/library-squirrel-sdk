package storepath

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"path"
	"strings"
	"testing"
)

// expectHex 测试内独立复算 s 的 sha256 十六进制摘要前 digits 位——锚定消歧哈希段
// 的口径：按原始输入（净化/截断之前）计算
func expectHex(t *testing.T, s string, digits int) string {
	t.Helper()
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])[:digits]
}

// TestWorkDirNamePlainIdForm 净化未变更的站点 ID 原文直用（形态锚定）
func TestWorkDirNamePlainIdForm(t *testing.T) {
	cases := []struct{ siteKey, siteWorkId, want string }{
		{"pixiv", "128937464", "pixiv_128937464"},
		{"bilibili", "BV1xx411c7mD_4538792", "bilibili_BV1xx411c7mD_4538792"},
		{"local", strings.Repeat("3f2a", 16), "local_" + strings.Repeat("3f2a", 16)}, // 64 位 hex id 直用
	}
	for _, c := range cases {
		got, err := WorkDirName(c.siteKey, c.siteWorkId)
		if err != nil {
			t.Errorf("WorkDirName(%q,%q) 意外出错：%v", c.siteKey, c.siteWorkId, err)
			continue
		}
		if got != c.want {
			t.Errorf("WorkDirName(%q,%q) = %q；want %q", c.siteKey, c.siteWorkId, got, c.want)
		}
	}
}

// TestComposedRelPathForm 完整 relPath 形态：主程序侧 path.Join 组合目录段与文件段
func TestComposedRelPathForm(t *testing.T) {
	dir, err := WorkDirName("bilibili", "BV1xx411c7mD_4538792")
	if err != nil {
		t.Fatalf("WorkDirName 意外出错：%v", err)
	}
	file, err := StoreFileName("videoTrack", 0, ".mp4")
	if err != nil {
		t.Fatalf("StoreFileName 意外出错：%v", err)
	}
	rel := path.Join("store/resource", dir, file)
	want := "store/resource/bilibili_BV1xx411c7mD_4538792/videoTrack_000.mp4"
	if rel != want {
		t.Errorf("组合 relPath = %q；want %q", rel, want)
	}
}

// TestWorkDirNameTruncation 派生段超长截断消歧：净化结果保留前 88 字符 + 原始 ID 的
// sha256 前 8 位 hex；共享 88 前缀的不同长 ID 由哈希段区分；长度按字符计（多字节
// 字符不按字节折算）；恰好 96 字符不截断
func TestWorkDirNameTruncation(t *testing.T) {
	long1 := strings.Repeat("a", 120)
	long2 := strings.Repeat("a", 88) + strings.Repeat("b", 40) // 与 long1 共享 88 前缀

	d1, err := WorkDirName("pixiv", long1)
	if err != nil {
		t.Fatalf("WorkDirName(long1) 意外出错：%v", err)
	}
	want1 := "pixiv_" + strings.Repeat("a", segmentLengthPrefix) + "_" + expectHex(t, long1, 8)
	if d1 != want1 {
		t.Errorf("超长截断形态 = %q；want %q", d1, want1)
	}
	if n := len([]rune(d1)); n != len("pixiv_")+segmentLengthPrefix+1+8 {
		t.Errorf("截断后段长 %d，与 88前缀+分隔符+8hex 形态不符：%q", n, d1)
	}

	d2, _ := WorkDirName("pixiv", long2)
	if d1 == d2 {
		t.Error("共享 88 前缀的两个不同 ID 截断后必须由哈希段区分")
	}

	// 恰好 96 字符（原文直用上限）不截断、不附哈希；97 字符进入截断
	atLimit := strings.Repeat("c", segmentLengthLimit)
	got, err := WorkDirName("pixiv", atLimit)
	if err != nil {
		t.Fatalf("WorkDirName(96字符) 意外出错：%v", err)
	}
	if got != "pixiv_"+atLimit {
		t.Errorf("96 字符应原文直用，got %q", got)
	}
	overLimit := strings.Repeat("c", segmentLengthLimit+1)
	got, _ = WorkDirName("pixiv", overLimit)
	if want := "pixiv_" + strings.Repeat("c", segmentLengthPrefix) + "_" + expectHex(t, overLimit, 8); got != want {
		t.Errorf("97 字符应截断消歧，got %q；want %q", got, want)
	}

	// 多字节字符按字符计：50 个全角字符 150 字节，仍在 96 字符内原文直用
	cjk := strings.Repeat("图", 50)
	got, err = WorkDirName("pixiv", cjk)
	if err != nil {
		t.Fatalf("WorkDirName(多字节) 意外出错：%v", err)
	}
	if got != "pixiv_"+cjk {
		t.Errorf("多字节 ID 按字符计未超限应直用，got %q", got)
	}
}

// TestWorkDirNameSanitizeCollision 净化单射（碰撞对）：半角非法字符与其全角等价
// 字符净化后同形，含半角者发生净化变更、追加按各自原始 ID 计算的消歧段，两个
// 不同 ID 派生互异目录段；原生全角者净化未变更、原文直用
func TestWorkDirNameSanitizeCollision(t *testing.T) {
	half := "a/b" // 含半角斜杠，净化为全角 ／
	full := "a／b" // 原生全角斜杠，净化无变更

	gotHalf, err := WorkDirName("pixiv", half)
	if err != nil {
		t.Fatalf("WorkDirName(half) 意外出错：%v", err)
	}
	wantHalf := "pixiv_a／b_" + expectHex(t, half, 8)
	if gotHalf != wantHalf {
		t.Errorf("净化变更形态 = %q；want %q", gotHalf, wantHalf)
	}

	gotFull, err := WorkDirName("pixiv", full)
	if err != nil {
		t.Fatalf("WorkDirName(full) 意外出错：%v", err)
	}
	if gotFull != "pixiv_"+full {
		t.Errorf("净化未变更应原文直用，got %q", gotFull)
	}

	if gotHalf == gotFull {
		t.Error("净化同形的两个不同 ID 必须派生互异目录段")
	}
}

// TestWorkDirNameSanitizedEmpty 净化后为空：派生段整体退为原始 ID 的 sha256 前
// 16 位 hex；不同原始 ID 互异
func TestWorkDirNameSanitizedEmpty(t *testing.T) {
	// 控制字符全删为空；纯空格句点尾部剥离为空
	for _, id := range []string{"\x01", "\x02\x7f", " ... ", strings.Repeat(" ", 4)} {
		got, err := WorkDirName("local", id)
		if err != nil {
			t.Errorf("WorkDirName(%q) 意外出错：%v", id, err)
			continue
		}
		if want := "local_" + expectHex(t, id, 16); got != want {
			t.Errorf("净化为空回退形态(%q) = %q；want %q", id, got, want)
		}
	}
	a, _ := WorkDirName("local", "\x01")
	b, _ := WorkDirName("local", "\x02")
	if a == b {
		t.Error("净化为空的不同 ID 必须由 16 位 hex 区分")
	}
}

// TestWorkDirNameEmptyRejected siteKey/siteWorkId 空值显式拒绝
func TestWorkDirNameEmptyRejected(t *testing.T) {
	if _, err := WorkDirName("", "123"); !errors.Is(err, ErrEmptySiteKey) {
		t.Errorf("空 siteKey 应报 ErrEmptySiteKey，got %v", err)
	}
	if _, err := WorkDirName("pixiv", ""); !errors.Is(err, ErrEmptySiteWorkId) {
		t.Errorf("空 siteWorkId 应报 ErrEmptySiteWorkId，got %v", err)
	}
}

// TestStoreFileNameForm 文件名形态：role_seq 三位零填充 + ext（含点）直拼；
// 序号超三位自然扩位；空 ext = 无扩展名
func TestStoreFileNameForm(t *testing.T) {
	cases := []struct {
		role string
		seq  int
		ext  string
		want string
	}{
		{"image", 0, ".jpg", "image_000.jpg"},
		{"videoTrack", 12, ".mp4", "videoTrack_012.mp4"},
		{"videoMain", 0, ".mp4", "videoMain_000.mp4"},
		{"thumbnail", 7, ".webp", "thumbnail_007.webp"},
		{"document", 3, ".md", "document_003.md"},
		{"audioMain", 125, ".flac", "audioMain_125.flac"},
		{"image", 1000, ".jpg", "image_1000.jpg"}, // 序号超三位自然扩位
		{"image", 0, "", "image_000"},             // 空 ext = 无扩展名
	}
	for _, c := range cases {
		got, err := StoreFileName(c.role, c.seq, c.ext)
		if err != nil {
			t.Errorf("StoreFileName(%q,%d,%q) 意外出错：%v", c.role, c.seq, c.ext, err)
			continue
		}
		if got != c.want {
			t.Errorf("StoreFileName(%q,%d,%q) = %q；want %q", c.role, c.seq, c.ext, got, c.want)
		}
	}
}

// TestStoreFileNameRoleRejected role 非法拒绝：空、路径分隔符、Windows 非法字符、
// 控制字符、句点/空格收尾
func TestStoreFileNameRoleRejected(t *testing.T) {
	for _, role := range []string{
		"", "im/age", `im\age`, "im:age", "im*age", "im?age", `im"age`,
		"im<age", "im>age", "im|age", "im\x01age", "image.", "image ", ".",
	} {
		if _, err := StoreFileName(role, 0, ".jpg"); !errors.Is(err, ErrInvalidRole) {
			t.Errorf("role %q 应报 ErrInvalidRole，got %v", role, err)
		}
	}
}

// TestStoreFileNameSeqExtRejected seq 负数与 ext 非法拒绝：无前导点、仅点号、
// 非法字符、句点/空格收尾
func TestStoreFileNameSeqExtRejected(t *testing.T) {
	if _, err := StoreFileName("image", -1, ".jpg"); !errors.Is(err, ErrNegativeSeq) {
		t.Errorf("负 seq 应报 ErrNegativeSeq，got %v", err)
	}
	for _, ext := range []string{"jpg", ".", ".jpg.", ".jpg ", ".j/g", ".jp\x01g"} {
		if _, err := StoreFileName("image", 0, ext); !errors.Is(err, ErrInvalidExt) {
			t.Errorf("ext %q 应报 ErrInvalidExt，got %v", ext, err)
		}
	}
}
