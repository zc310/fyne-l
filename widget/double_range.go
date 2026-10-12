package widget

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// DoubleRange 范围设置组件：文字  最小值 ~ 最大值（float64）
type DoubleRange struct {
	widget.BaseWidget

	label    string
	min, max float64
	onChange func(min, max float64)

	minEntry *MinWidthEntry
	maxEntry *MinWidthEntry
	box      *fyne.Container // 根容器
}

// NewDoubleRange 构造函数
func NewDoubleRange(label string, min, max float64, onChange func(min, max float64)) *DoubleRange {
	r := &DoubleRange{
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
func (r *DoubleRange) build() {
	r.minEntry = r.makeEntry(r.min, func(v float64) {
		r.min = v
		if r.onChange != nil {
			r.onChange(r.min, r.max)
		}
	})
	r.maxEntry = r.makeEntry(r.max, func(v float64) {
		r.max = v
		if r.onChange != nil {
			r.onChange(r.min, r.max)
		}
	})

	// 根布局
	if r.label == "" {
		r.box = container.NewHBox(r.minEntry, widget.NewLabel("~"), r.maxEntry)
	} else {
		r.box = container.NewHBox(widget.NewLabel(r.label), r.minEntry, widget.NewLabel("~"), r.maxEntry)
	}
}

// 生成带校验的浮点输入框
func (r *DoubleRange) makeEntry(init float64, setter func(float64)) *MinWidthEntry {
	e := NewMinWidthEntry(60)
	e.SetText(strconv.FormatFloat(init, 'f', -1, 64))
	e.OnChanged = func(s string) {
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			setter(v)
		}
	}
	return e
}

// CreateRenderer 实现 fyne.Widget 接口
func (r *DoubleRange) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(r.box)
}

// Range 返回当前最小值、最大值
func (r *DoubleRange) Range() (float64, float64) { return r.min, r.max }

// SetRange 外部设置范围并刷新界面
func (r *DoubleRange) SetRange(min, max float64) {
	r.min = min
	r.max = max
	setEntryTextWithoutChange(r.minEntry, &r.minEntry.OnChanged, strconv.FormatFloat(min, 'f', -1, 64))
	setEntryTextWithoutChange(r.maxEntry, &r.maxEntry.OnChanged, strconv.FormatFloat(max, 'f', -1, 64))
	if r.onChange != nil {
		r.onChange(min, max)
	}
}

// SetValue 设置范围值（与 SetRange 功能相同，提供一致性接口）
func (r *DoubleRange) SetValue(min, max float64) {
	r.SetRange(min, max)
}

// SetMin 单独设置最小值
func (r *DoubleRange) SetMin(min float64) {
	r.min = min
	setEntryTextWithoutChange(r.minEntry, &r.minEntry.OnChanged, strconv.FormatFloat(min, 'f', -1, 64))
	if r.onChange != nil {
		r.onChange(r.min, r.max)
	}
}

// SetMax 单独设置最大值
func (r *DoubleRange) SetMax(max float64) {
	r.max = max
	setEntryTextWithoutChange(r.maxEntry, &r.maxEntry.OnChanged, strconv.FormatFloat(max, 'f', -1, 64))
	if r.onChange != nil {
		r.onChange(r.min, r.max)
	}
}

// GetValue 获取当前范围值（返回 min, max）
func (r *DoubleRange) GetValue() (float64, float64) {
	return r.min, r.max
}

// Validate 验证范围是否有效（min <= max）
func (r *DoubleRange) Validate() bool {
	return r.min <= r.max
}

// Reset 重置为指定值
func (r *DoubleRange) Reset(min, max float64) {
	r.SetValue(min, max)
}

// Disable 禁用组件
func (r *DoubleRange) Disable() {
	r.minEntry.Disable()
	r.maxEntry.Disable()
}

// Enable 启用组件
func (r *DoubleRange) Enable() {
	r.minEntry.Enable()
	r.maxEntry.Enable()
}

// SetOnChange 设置值变化回调函数
func (r *DoubleRange) SetOnChange(callback func(min, max float64)) {
	r.onChange = callback
}

// SetPrecision 设置显示精度
func (r *DoubleRange) SetPrecision(precision int) {
	// 更新当前值的显示精度
	setEntryTextWithoutChange(r.minEntry, &r.minEntry.OnChanged, strconv.FormatFloat(r.min, 'f', precision, 64))
	setEntryTextWithoutChange(r.maxEntry, &r.maxEntry.OnChanged, strconv.FormatFloat(r.max, 'f', precision, 64))

	// 更新后续输入的处理
	r.minEntry.OnChanged = func(s string) {
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			r.min = v
			if r.onChange != nil {
				r.onChange(r.min, r.max)
			}
		}
	}
	r.maxEntry.OnChanged = func(s string) {
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			r.max = v
			if r.onChange != nil {
				r.onChange(r.min, r.max)
			}
		}
	}
}

// SetEntryWidth 设置输入框宽度
func (r *DoubleRange) SetEntryWidth(width float32) {
	r.minEntry.SetMinWidth(width)
	r.maxEntry.SetMinWidth(width)
}

// GetLabel 获取标签文本
func (r *DoubleRange) GetLabel() string {
	return r.label
}

// SetLabel 设置标签文本
func (r *DoubleRange) SetLabel(label string) {
	r.label = label
	// 重新构建界面以更新标签
	r.build()
	r.Refresh()
}
