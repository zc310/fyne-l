package widget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// MarkdownHelpConfig 配置结构体
type MarkdownHelpConfig struct {
	Title        string            // 对话框标题
	MarkdownText string            // Markdown 文本内容
	ButtonText   string            // 按钮文字
	ButtonIcon   fyne.Resource     // 按钮图标
	Importance   widget.Importance // 按钮重要性
	WindowSize   fyne.Size         // 对话框大小（可选，为零值时自动计算）
}

// DefaultConfig 返回默认配置
func DefaultConfig() *MarkdownHelpConfig {
	return &MarkdownHelpConfig{
		Title:      "帮助",
		ButtonText: "",
		ButtonIcon: nil,
		Importance: widget.LowImportance,
		WindowSize: fyne.Size{Width: 0, Height: 0}, // 零值，表示自动计算
	}
}

// MarkdownHelpWidget Markdown 帮助组件
type MarkdownHelpWidget struct {
	widget.BaseWidget
	parent fyne.Window
	config *MarkdownHelpConfig
	button *widget.Button
}

// NewMarkdownHelpWidget 创建新的帮助组件
func NewMarkdownHelpWidget(parent fyne.Window, config *MarkdownHelpConfig) *MarkdownHelpWidget {
	if config == nil {
		config = DefaultConfig()
	}

	// 设置默认图标
	if config.ButtonIcon == nil {
		config.ButtonIcon = theme.HelpIcon()
	}

	m := &MarkdownHelpWidget{
		parent: parent,
		config: config,
	}

	m.createButton()
	m.ExtendBaseWidget(m)
	return m
}

// NewSimpleMarkdownHelp 快速创建简单帮助按钮
func NewSimpleMarkdownHelp(parent fyne.Window, title, markdown string) *MarkdownHelpWidget {
	config := DefaultConfig()
	config.Title = title
	config.MarkdownText = markdown

	return NewMarkdownHelpWidget(parent, config)
}

// NewMarkdownHelpWithText 创建带文字的帮助按钮
func NewMarkdownHelpWithText(parent fyne.Window, title, markdown, buttonText string) *MarkdownHelpWidget {
	config := DefaultConfig()
	config.Title = title
	config.MarkdownText = markdown
	config.ButtonText = buttonText

	return NewMarkdownHelpWidget(parent, config)
}

// createButton 创建内部按钮
func (m *MarkdownHelpWidget) createButton() {
	m.button = widget.NewButtonWithIcon(m.config.ButtonText, m.config.ButtonIcon, m.showHelp)
	m.button.Importance = m.config.Importance
}

// showHelp 显示帮助对话框
func (m *MarkdownHelpWidget) showHelp() {
	if m.config.MarkdownText == "" {
		// 如果没有内容，显示默认提示
		m.config.MarkdownText = "# 帮助\n暂无帮助内容"
	}

	helpContent := widget.NewRichTextFromMarkdown(m.config.MarkdownText)
	helpContent.Wrapping = fyne.TextWrapWord

	scrollContainer := container.NewScroll(helpContent)

	// 自动计算合适的大小
	dialogSize := m.calculateDialogSize()

	customDialog := dialog.NewCustom(m.config.Title, "关闭", scrollContainer, m.parent)

	// 只有明确设置了窗口大小时才调整大小
	if m.config.WindowSize.Width > 0 && m.config.WindowSize.Height > 0 {
		customDialog.Resize(m.config.WindowSize)
	} else {
		// 使用自动计算的大小
		customDialog.Resize(dialogSize)
	}

	customDialog.Show()
}

// calculateDialogSize 自动计算合适的对话框大小
func (m *MarkdownHelpWidget) calculateDialogSize() fyne.Size {
	// 获取父窗口大小作为参考
	parentSize := m.parent.Canvas().Size()

	// 计算合适的对话框大小（父窗口的 70%-80%）
	width := parentSize.Width * 0.75
	height := parentSize.Height * 0.7

	// 设置最小和最大限制
	var (
		minWidth  float32 = 400.0
		maxWidth  float32 = 800.0
		minHeight float32 = 300.0
		maxHeight float32 = 600.0
	)

	if width < minWidth {
		width = minWidth
	} else if width > maxWidth {
		width = maxWidth
	}

	if height < minHeight {
		height = minHeight
	} else if height > maxHeight {
		height = maxHeight
	}

	return fyne.NewSize(float32(width), float32(height))
}

// CreateRenderer 创建组件渲染器
func (m *MarkdownHelpWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(m.button)
}

// SetMarkdownText 动态设置 Markdown 文本
func (m *MarkdownHelpWidget) SetMarkdownText(text string) {
	m.config.MarkdownText = text
}

// SetTitle 动态设置对话框标题
func (m *MarkdownHelpWidget) SetTitle(title string) {
	m.config.Title = title
}

// SetButtonText 动态设置按钮文字
func (m *MarkdownHelpWidget) SetButtonText(text string) {
	m.config.ButtonText = text
	m.refreshButton()
}

// SetButtonIcon 动态设置按钮图标
func (m *MarkdownHelpWidget) SetButtonIcon(icon fyne.Resource) {
	m.config.ButtonIcon = icon
	m.refreshButton()
}

// SetImportance 设置按钮重要性
func (m *MarkdownHelpWidget) SetImportance(importance widget.Importance) {
	m.config.Importance = importance
	m.refreshButton()
}

// SetWindowSize 设置对话框大小（可选）
func (m *MarkdownHelpWidget) SetWindowSize(size fyne.Size) {
	m.config.WindowSize = size
}

// UpdateConfig 更新配置
func (m *MarkdownHelpWidget) UpdateConfig(updater func(*MarkdownHelpConfig)) {
	updater(m.config)
	m.refreshButton()
}

// refreshButton 刷新按钮
func (m *MarkdownHelpWidget) refreshButton() {
	m.createButton()
	m.Refresh()
}

// GetButton 获取内部按钮（用于外部布局）
func (m *MarkdownHelpWidget) GetButton() *widget.Button {
	return m.button
}
