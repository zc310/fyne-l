package widget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// CheckGroup 支持动态生成一组带文字复选框，并实时回传 []bool
type CheckGroup struct {
	widget.BaseWidget
	title    string
	labels   []string
	values   []bool
	onChange func([]bool)

	box   *fyne.Container
	check []*widget.Check
}

// NewCheckGroup 构造组件
//
//	labels   : 每个复选框的文字，长度即个数
//	defaults : 初始选中状态，可传 nil（全 false）
//	onChange : 用户修改时回调，参数为最新 []bool
func NewCheckGroup(title string, labels []string, defaults []bool, onChange func([]bool)) *CheckGroup {
	c := &CheckGroup{
		title:    title,
		labels:   labels,
		onChange: onChange,
	}
	// 初始化 values
	n := len(labels)
	c.values = make([]bool, n)
	if defaults != nil && len(defaults) == n {
		copy(c.values, defaults)
	}
	c.ExtendBaseWidget(c)
	c.build()
	return c
}

// 创建内部控件
func (c *CheckGroup) build() {
	c.check = make([]*widget.Check, len(c.labels))
	objs := make([]fyne.CanvasObject, len(c.labels)) // 存 CanvasObject
	for i, txt := range c.labels {
		idx := i
		ch := widget.NewCheck(txt, func(v bool) {
			c.values[idx] = v
			if c.onChange != nil {
				c.onChange(c.cloneBool())
			}
		})
		ch.SetChecked(c.values[i])
		c.check[i] = ch
		objs[i] = ch // 接口类型
	}
	if c.title != "" {
		c.box = container.NewHBox(append([]fyne.CanvasObject{widget.NewLabel(c.title)}, objs...)...)
	} else {
		c.box = container.NewHBox(objs...)
	}
}

// 返回当前 []bool 副本，防止外部篡改
func (c *CheckGroup) cloneBool() []bool {
	ret := make([]bool, len(c.values))
	copy(ret, c.values)
	return ret
}

// CreateRenderer 实现 fyne.Widget
func (c *CheckGroup) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.box)
}

// Values 只读获取当前 []bool（可选）
func (c *CheckGroup) Values() []bool { return c.cloneBool() }

// SetValues 设置复选框的值
//
//	values: 要设置的布尔值切片，长度必须与标签数量一致
//	返回: 成功返回 true，失败返回 false
func (c *CheckGroup) SetValues(values []bool) bool {
	if len(values) != len(c.labels) {
		return false // 长度不匹配，设置失败
	}

	// 更新内部值
	copy(c.values, values)

	// 更新界面显示
	for i, ch := range c.check {
		if ch != nil {
			ch.SetChecked(c.values[i])
		}
	}

	return true
}

// SetValue 设置单个复选框的值
//
//	index: 复选框索引（从0开始）
//	value: 要设置的值
//	返回: 成功返回 true，失败返回 false
func (c *CheckGroup) SetValue(index int, value bool) bool {
	if index < 0 || index >= len(c.labels) {
		return false // 索引越界
	}

	c.values[index] = value
	if c.check[index] != nil {
		c.check[index].SetChecked(value)
	}

	return true
}
