package widget

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// IntRange 范围设置组件：文字  最小值 ~ 最大值
type IntRange struct {
	widget.BaseWidget
	center   bool
	label    string
	min, max int
	onChange func(min, max int)

	minEntry *IntEntry
	maxEntry *IntEntry
	box      *fyne.Container // 根容器
}

// NewIntRange 构造函数
func NewIntRange(label string, min, max int, onChange func(min, max int), center bool) *IntRange {
	r := &IntRange{
		center:   center,
		label:    label,
		min:      min,
		max:      max,
		onChange: onChange,
	}
	r.ExtendBaseWidget(r)
	r.build()
	return r
}

// 创建内部控件
func (r *IntRange) build() {
	r.minEntry = r.makeEntry(r.min, func(v int) {
		r.min = v
		if r.onChange != nil {
			r.onChange(r.min, r.max)
		}
	})
	r.maxEntry = r.makeEntry(r.max, func(v int) {
		r.max = v
		if r.onChange != nil {
			r.onChange(r.min, r.max)
		}
	})

	// 根布局：文字  [最小值 ~ 最大值]（每个输入框带对号）
	if r.label == "" {
		r.box = container.NewHBox(
			r.minEntry,
			widget.NewLabel("~"),
			r.maxEntry,
		)
	} else {
		if r.center {
			r.box = container.NewHBox(
				r.minEntry,
				widget.NewLabel("≤"+r.label+"≤"),
				r.maxEntry,
			)
		} else {
			r.box = container.NewHBox(
				widget.NewLabel(r.label),
				r.minEntry,
				widget.NewLabel("~"),
				r.maxEntry,
			)
		}
	}
}

// 生成带校验的数字输入框
func (r *IntRange) makeEntry(init int, setter func(int)) *IntEntry {
	e := NewIntEntryWithNoValidator(r.min, r.max, 0)
	e.SetMinWidth(45)
	e.SetText(strconv.Itoa(init))
	e.OnChanged = func(s string) {
		if v, err := strconv.Atoi(s); err == nil {
			setter(v)
		}
	}
	return e
}

// CreateRenderer 实现 fyne.Widget 接口
func (r *IntRange) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(r.box)
}

// Range 返回当前最小值、最大值
func (r *IntRange) Range() (int, int) { return r.min, r.max }

// SetEntryMinWidth 设置最小值、最大值输入框的最小宽度，不改变范围值。
func (r *IntRange) SetEntryMinWidth(width float32) {
	r.minEntry.SetMinWidth(width)
	r.maxEntry.SetMinWidth(width)
	r.box.Refresh()
	r.Refresh()
}

// SetRange 外部设置范围并刷新界面
func (r *IntRange) SetRange(min, max int) {
	r.min = min
	r.max = max
	setEntryTextWithoutChange(r.minEntry, &r.minEntry.OnChanged, strconv.Itoa(min))
	setEntryTextWithoutChange(r.maxEntry, &r.maxEntry.OnChanged, strconv.Itoa(max))
	if r.onChange != nil {
		r.onChange(min, max)
	}
}

// SetValue 设置范围值（与 SetRange 功能相同，提供一致性接口）
func (r *IntRange) SetValue(min, max int) {
	r.SetRange(min, max)
}

// SetMin 单独设置最小值
func (r *IntRange) SetMin(min int) {
	r.min = min
	setEntryTextWithoutChange(r.minEntry, &r.minEntry.OnChanged, strconv.Itoa(min))
	if r.onChange != nil {
		r.onChange(r.min, r.max)
	}
}

// SetMax 单独设置最大值
func (r *IntRange) SetMax(max int) {
	r.max = max
	setEntryTextWithoutChange(r.maxEntry, &r.maxEntry.OnChanged, strconv.Itoa(max))
	if r.onChange != nil {
		r.onChange(r.min, r.max)
	}
}

// GetValue 获取当前范围值（返回 min, max）
func (r *IntRange) GetValue() (int, int) {
	return r.min, r.max
}

// Validate 验证范围是否有效（min <= max）
func (r *IntRange) Validate() bool {
	return r.min <= r.max
}

// Reset 重置为指定值
func (r *IntRange) Reset(min, max int) {
	r.SetValue(min, max)
}

// Disable 禁用组件
func (r *IntRange) Disable() {
	r.minEntry.Disable()
	r.maxEntry.Disable()
}

// Enable 启用组件
func (r *IntRange) Enable() {
	r.minEntry.Enable()
	r.maxEntry.Enable()
}

// SetOnChange 设置值变化回调函数
func (r *IntRange) SetOnChange(callback func(min, max int)) {
	r.onChange = callback
}

// setEntryTextWithoutChange 修改输入框文本但不触发输入回调。
func setEntryTextWithoutChange(entry interface{ SetText(string) }, onChanged *func(string), text string) {
	callback := *onChanged
	*onChanged = nil
	entry.SetText(text)
	*onChanged = callback
}
