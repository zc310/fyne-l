package widget

import (
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ShowCopyableDialog 弹出可复制文本的对话框
// parent: 父窗口
// title : 对话框标题
// text  : 要显示的文本
func ShowCopyableDialog(parent fyne.Window, title, text string) {
	// 多行 Entry，只读，自动换行，方便复制
	entry := widget.NewMultiLineEntry()
	entry.SetText(text)
	//entry.Disable() // 只读（仍可选中复制）

	// 自定义 content
	content := container.NewBorder(nil, nil, nil, nil, entry)
	// 用 dialog.NewCustom 创建对话框
	d := dialog.NewCustom(title, "关闭", content, parent)
	d.Resize(fyne.NewSize(450, 300)) // 可调整大小
	d.Show()
}

func RunPopUp(s string, w fyne.Window) *widget.PopUp {
	m := widget.NewModalPopUp(
		container.NewVBox(widget.NewRichTextFromMarkdown(s)),
		w.Canvas(),
	)
	m.Show()
	return m
}

type LastFile struct {
	key   string
	Value string
}

// Cache 按字节键值存储缓存数据。
//
// 调用方可以通过 SetCache 提供持久化实现。
// 默认实现是内存缓存，适用于不需要持久化的场景。
type Cache interface {
	Get(key []byte) ([]byte, bool)
	Set(key, value []byte) error
}

type memoryCache struct {
	mu     sync.RWMutex
	values map[string][]byte
}

func newMemoryCache() *memoryCache {
	return &memoryCache{values: make(map[string][]byte)}
}

func (c *memoryCache) Get(key []byte) ([]byte, bool) {
	c.mu.RLock()
	value, ok := c.values[string(key)]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	return append([]byte(nil), value...), true
}

func (c *memoryCache) Set(key, value []byte) error {
	c.mu.Lock()
	c.values[string(key)] = append([]byte(nil), value...)
	c.mu.Unlock()
	return nil
}

var (
	cacheMu sync.RWMutex
	cache   Cache = newMemoryCache()
)

// SetCache 替换 NewLastFile 和字符串缓存辅助函数使用的缓存。
// 传入 nil 时重置为新的内存缓存。
func SetCache(c Cache) {
	if c == nil {
		c = newMemoryCache()
	}
	cacheMu.Lock()
	cache = c
	cacheMu.Unlock()
}

// GetCache 返回 NewLastFile 和字符串缓存辅助函数当前使用的缓存。
func GetCache() Cache {
	cacheMu.RLock()
	c := cache
	cacheMu.RUnlock()
	return c
}

// GetCachedString 返回缓存字符串；键不存在时返回 fallback。
func GetCachedString(key, fallback string) string {
	value, ok := GetCache().Get([]byte(key))
	if !ok {
		return fallback
	}
	return string(value)
}

// SetCachedString 将字符串存入当前配置的缓存。
func SetCachedString(key, value string) error {
	return GetCache().Set([]byte(key), []byte(value))
}

func (p *LastFile) Close() {
	_ = SetCachedString(p.key, p.Value)
}

func NewLastFile(key, value string) *LastFile {
	lf := &LastFile{key: key, Value: value}
	lf.Value = GetCachedString(lf.key, lf.Value)
	return lf
}
