package widget

import (
	"fmt"

	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// CPUSettingWidget CPU使用率设置组件
type CPUSettingWidget struct {
	widget.BaseWidget

	edtCpu    *IntEntry
	label     *widget.Label
	noteLabel *widget.RichText
	container *fyne.Container
	sysInfo   *SysInfoWidget
	lblTitle  *widget.Label
}

// NewCPUSettingWidget 创建CPU使用率设置组件
// defaultValue: 默认CPU使用率(1-100)
// showNote: 是否显示注意事项
func NewCPUSettingWidget(defaultValue int, showSysInfo bool) *CPUSettingWidget {
	c := &CPUSettingWidget{
		edtCpu: NewIntEntry(1, 100),
		label:  widget.NewLabel("使用："),
	}

	c.ExtendBaseWidget(c)
	c.setupUI(defaultValue, showSysInfo)

	return c
}

// setupUI 设置界面
func (c *CPUSettingWidget) setupUI(defaultValue int, showSysInfo bool) {
	// 设置输入框属性
	c.edtCpu.SetText(fmt.Sprintf("%d", defaultValue))
	c.edtCpu.SetMinWidth(75)
	c.edtCpu.SetPlaceHolder("请输入 1-100 的整数")
	c.edtCpu.Validator = c.validateCPUValue
	c.lblTitle = widget.NewLabel("CPU设置")
	c.noteLabel = widget.NewRichText(
		&widget.TextSegment{
			Text: "设置过大会内存不足无法计算",
			Style: widget.RichTextStyle{
				ColorName: theme.ColorNamePlaceHolder,
			},
		})
	// 创建基础水平布局
	hbox := container.NewHBox(
		c.label,
		c.edtCpu,
		widget.NewLabel("%"), c.noteLabel,
	)
	if showSysInfo {
		c.sysInfo = NewSysInfoWidget()
		c.container = container.NewVBox(
			c.lblTitle,
			hbox,
			widget.NewSeparator(), c.sysInfo)
	} else {
		c.container = container.NewVBox(
			c.lblTitle,
			hbox)
	}

}

// Stop 在启用系统信息显示时停止后台更新。
func (c *CPUSettingWidget) Stop() {
	if c.sysInfo != nil {
		c.sysInfo.Stop()
	}
}

// validateCPUValue 验证CPU值是否合法
func (c *CPUSettingWidget) validateCPUValue(text string) error {
	if text == "" {
		return nil // 允许为空，由具体业务逻辑处理
	}

	value, err := strconv.Atoi(text)
	if err != nil {
		return fmt.Errorf("请输入有效的整数")
	}

	if value < 1 || value > 100 {
		return fmt.Errorf("请输入 1-100 的整数")
	}

	return nil
}

// CreateRenderer 实现widget接口
func (c *CPUSettingWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.container)
}

// GetValue 获取当前设置的CPU使用率
func (c *CPUSettingWidget) GetValue() (int, error) {
	text := c.edtCpu.Text
	if text == "" {
		return 0, fmt.Errorf("CPU使用率不能为空")
	}

	value, err := strconv.Atoi(text)
	if err != nil {
		return 0, err
	}

	if value < 1 || value > 100 {
		return 0, fmt.Errorf("CPU使用率必须在 1-100 范围内")
	}

	return value, nil
}

// SetValue 设置CPU使用率
func (c *CPUSettingWidget) SetValue(value int) error {
	if value < 1 || value > 100 {
		return fmt.Errorf("CPU使用率必须在 1-100 范围内")
	}

	c.edtCpu.SetText(fmt.Sprintf("%d", value))
	return nil
}

// SetLabelText 设置标签文本
func (c *CPUSettingWidget) SetLabelText(text string) {
	c.label.SetText(text)
}

// SetEntryWidth 设置输入框宽度
func (c *CPUSettingWidget) SetEntryWidth(width float32) {
	c.edtCpu.SetMinWidth(width)
}

// OnChanged 设置值改变时的回调函数
func (c *CPUSettingWidget) OnChanged(f func(int)) {
	c.edtCpu.OnChanged = func(text string) {
		if text == "" {
			return
		}

		if value, err := strconv.Atoi(text); err == nil && value >= 1 && value <= 100 {
			f(value)
		}
	}
}

// Enable 启用组件
func (c *CPUSettingWidget) Enable() {
	c.edtCpu.Enable()
}

// Disable 禁用组件
func (c *CPUSettingWidget) Disable() {
	c.edtCpu.Disable()
}
