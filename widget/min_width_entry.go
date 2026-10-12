package widget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// MinWidthEntry 支持设置最小宽度的输入框
type MinWidthEntry struct {
	widget.Entry
	minWidth float32
}

// NewMinWidthEntry 创建组件
func NewMinWidthEntry(minWidth float32) *MinWidthEntry {
	e := &MinWidthEntry{minWidth: minWidth}
	e.ExtendBaseWidget(e)
	return e
}

// MinSize 覆盖返回最小尺寸
func (e *MinWidthEntry) MinSize() fyne.Size {
	base := e.Entry.MinSize()
	if base.Width < e.minWidth {
		base.Width = e.minWidth
	}
	return base
}

// SetMinWidth 动态设置最小宽度
func (e *MinWidthEntry) SetMinWidth(width float32) {
	e.minWidth = width
	e.Refresh() // 触发重新布局
}

// GetMinWidth 获取当前最小宽度
func (e *MinWidthEntry) GetMinWidth() float32 {
	return e.minWidth
}

// SetText 设置文本并保持最小宽度
func (e *MinWidthEntry) SetText(text string) {
	e.Entry.SetText(text)
	e.Refresh()
}

// SetPlaceHolder 设置占位符文本
func (e *MinWidthEntry) SetPlaceHolder(placeholder string) {
	e.Entry.SetPlaceHolder(placeholder)
	e.Refresh()
}

// Disable 禁用输入框
func (e *MinWidthEntry) Disable() {
	e.Entry.Disable()
	e.Refresh()
}

// Enable 启用输入框
func (e *MinWidthEntry) Enable() {
	e.Entry.Enable()
	e.Refresh()
}

// SetReadOnly 设置只读状态
func (e *MinWidthEntry) SetReadOnly(readOnly bool) {
	if readOnly {
		e.Entry.Disable()
	} else {
		e.Entry.Enable()
	}
	e.Refresh()
}

// Clear 清空文本内容
func (e *MinWidthEntry) Clear() {
	e.Entry.SetText("")
	e.Refresh()
}

// SetOnChanged 设置文本变化回调函数
func (e *MinWidthEntry) SetOnChanged(callback func(string)) {
	e.Entry.OnChanged = callback
}

// SetValidation 设置验证函数
func (e *MinWidthEntry) SetValidation(validation func(string) error) {
	e.Entry.Validator = validation
}

// IsValid 验证当前文本是否有效
func (e *MinWidthEntry) IsValid() bool {
	return e.Entry.Validate() == nil
}

// Focus 获取焦点
func (e *MinWidthEntry) Focus() {
	e.Entry.FocusGained()
}

// Unfocus 失去焦点
func (e *MinWidthEntry) Unfocus() {
	e.Entry.FocusLost()
}
