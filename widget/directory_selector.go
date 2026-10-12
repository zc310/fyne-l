package widget

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/ncruces/zenity"
	"github.com/pkg/browser"
)

// DirectorySelector 目录选择器组件
type DirectorySelector struct {
	widget.BaseWidget

	button           *widget.Button
	multiButton      *widget.Button // 多选按钮
	pathLabel        *widget.Label
	pathButton       *widget.Button
	selectedPath     string
	selectedPaths    []string // 多选路径
	OnPathSelected   func(string)
	OnPathsSelected  func([]string) // 多选回调
	window           fyne.Window
	maxDisplayLength int
	showHidden       bool
	allowMultiple    bool // 是否允许多选

	LastFile *LastFile
}

// DirectorySelectorConfig 配置选项
type DirectorySelectorConfig struct {
	Window           fyne.Window
	OnPathSelected   func(string)
	OnPathsSelected  func([]string)
	MaxDisplayLength int
	ShowHidden       bool
	ButtonText       string
	AllowMultiple    bool // 是否允许多选
}

// NewDirectorySelector 创建新的目录选择器
func NewDirectorySelector(window fyne.Window, onPathSelected func(string)) *DirectorySelector {
	return NewDirectorySelectorWithConfig(DirectorySelectorConfig{
		Window:         window,
		OnPathSelected: onPathSelected,
	})
}

// NewDirectorySelectorWithConfig 使用配置创建目录选择器
func NewDirectorySelectorWithConfig(config DirectorySelectorConfig) *DirectorySelector {
	if config.MaxDisplayLength <= 0 {
		config.MaxDisplayLength = 40
	}
	if config.ButtonText == "" {
		config.ButtonText = "选择目录"
	}

	ds := &DirectorySelector{
		window:           config.Window,
		OnPathSelected:   config.OnPathSelected,
		OnPathsSelected:  config.OnPathsSelected,
		selectedPath:     "",
		selectedPaths:    []string{},
		maxDisplayLength: config.MaxDisplayLength,
		showHidden:       config.ShowHidden,
		allowMultiple:    config.AllowMultiple,
	}

	// 创建选择目录按钮
	ds.button = widget.NewButtonWithIcon(config.ButtonText, theme.FolderOpenIcon(), ds.selectDirectory)

	// 创建多选目录按钮，并根据当前配置显示或隐藏
	ds.multiButton = widget.NewButtonWithIcon("选择多个目录", theme.FolderNewIcon(), ds.selectMultipleDirectories)
	if !ds.allowMultiple {
		ds.multiButton.Hide()
	}

	// 创建显示路径的标签
	ds.pathLabel = widget.NewLabel("未选择目录")
	ds.pathLabel.TextStyle = fyne.TextStyle{Bold: true}

	// 创建可点击的路径按钮
	ds.pathButton = widget.NewButton("", ds.showPathDetails)
	ds.pathButton.Importance = widget.LowImportance

	ds.ExtendBaseWidget(ds)
	return ds
}

// SetMaxDisplayLength 设置路径显示的最大长度
func (ds *DirectorySelector) SetMaxDisplayLength(length int) {
	if length > 10 { // 最小长度限制
		ds.maxDisplayLength = length
		ds.updatePathDisplay()
	}
}

// SetShowHidden 设置是否显示隐藏文件
func (ds *DirectorySelector) SetShowHidden(show bool) {
	ds.showHidden = show
}

// SetAllowMultiple 设置是否允许多选
func (ds *DirectorySelector) SetAllowMultiple(allow bool) {
	ds.allowMultiple = allow
	if allow {
		ds.multiButton.Show()
	} else {
		ds.multiButton.Hide()
	}
	ds.updatePathDisplay()
	ds.Refresh()
}

// CreateRenderer 创建组件渲染器
func (ds *DirectorySelector) CreateRenderer() fyne.WidgetRenderer {
	// 使用容器来组合路径按钮和标签
	pathContainer := container.NewStack(
		ds.pathButton,
		container.NewCenter(ds.pathLabel),
	)

	// 使用水平布局，按钮在左边，路径显示在右边
	buttons := []fyne.CanvasObject{ds.button}
	if ds.multiButton != nil {
		buttons = append(buttons, ds.multiButton)
	}

	content := container.NewHBox(
		container.NewHBox(buttons...),
		layout.NewSpacer(),
		pathContainer,
	)

	return widget.NewSimpleRenderer(content)
}

// selectDirectory 选择单个目录
func (ds *DirectorySelector) selectDirectory() {
	var defFilename string
	if ds.LastFile != nil {
		defer ds.LastFile.Close()
		defFilename = ds.LastFile.Value
	}
	options := []zenity.Option{
		zenity.Filename(defFilename),
		zenity.Title("选择目录"),
		zenity.Directory(),
	}

	if ds.showHidden {
		options = append(options, zenity.ShowHidden())
	}

	selectedDir, err := zenity.SelectFile(options...)

	if err != nil {
		if !errors.Is(err, zenity.ErrCanceled) {
			dialog.ShowError(fmt.Errorf("选择目录失败: %v", err), ds.window)
		}
		return
	}

	if selectedDir != "" {
		if ds.LastFile != nil {
			ds.LastFile.Value = selectedDir
		}
		ds.SetSelectedPath(selectedDir)
		ds.selectedPaths = []string{selectedDir} // 更新多选路径
		if ds.OnPathsSelected != nil {
			ds.OnPathsSelected(cloneStrings(ds.selectedPaths))
		}
	}
}

// selectMultipleDirectories 选择多个目录 - 使用 zenity.SelectFileMultiple
func (ds *DirectorySelector) selectMultipleDirectories() {
	var defFilename string
	if ds.LastFile != nil {
		defer ds.LastFile.Close()
		defFilename = ds.LastFile.Value
	}
	options := []zenity.Option{
		zenity.Filename(defFilename),
		zenity.Title("选择多个目录"),
		zenity.Directory(),
	}

	if ds.showHidden {
		options = append(options, zenity.ShowHidden())
	}

	paths, err := zenity.SelectFileMultiple(options...)
	if err != nil {
		if !errors.Is(err, zenity.ErrCanceled) {
			dialog.ShowError(fmt.Errorf("选择目录失败: %v", err), ds.window)
		}
		return
	}

	if len(paths) > 0 {
		if ds.LastFile != nil {
			ds.LastFile.Value = paths[0]
		}
		// 合并新选择的目录到现有列表，去重
		ds.addDirectories(paths)
	}
}

// addDirectories 添加目录到现有列表，自动去重
func (ds *DirectorySelector) addDirectories(newPaths []string) {
	if len(newPaths) == 0 {
		return
	}

	// 创建路径映射用于去重
	existingPaths := make(map[string]bool)
	for _, path := range ds.selectedPaths {
		// 标准化路径格式
		cleanPath := filepath.Clean(path)
		existingPaths[cleanPath] = true
	}

	// 添加新路径，跳过已存在的
	addedCount := 0
	for _, path := range newPaths {
		cleanPath := filepath.Clean(path)
		if !existingPaths[cleanPath] {
			ds.selectedPaths = append(ds.selectedPaths, path)
			existingPaths[cleanPath] = true
			addedCount++
		}
	}

	if addedCount > 0 {
		// 更新显示和回调
		if len(ds.selectedPaths) == 1 {
			ds.selectedPath = ds.selectedPaths[0]
		} else {
			ds.selectedPath = "" // 多选时清空单一路径
		}
		ds.updatePathDisplay()

		if ds.OnPathsSelected != nil {
			ds.OnPathsSelected(cloneStrings(ds.selectedPaths))
		}

		// 显示添加成功的提示
		if addedCount == 1 {
			dialog.ShowInformation("添加成功", fmt.Sprintf("已添加 1 个目录"), ds.window)
		} else {
			dialog.ShowInformation("添加成功", fmt.Sprintf("已添加 %d 个目录", addedCount), ds.window)
		}
	} else {
		dialog.ShowInformation("提示", "所有选择的目录都已存在列表中", ds.window)
	}
}

// showPathDetails 显示路径详细信息
func (ds *DirectorySelector) showPathDetails() {
	if ds.selectedPath == "" && len(ds.selectedPaths) == 0 {
		dialog.ShowInformation("路径信息", "未选择任何目录", ds.window)
		return
	}

	pathDialog := dialog.NewCustom(
		"目录管理",
		"关闭",
		ds.createPathDetailContent(),
		ds.window,
	)

	pathDialog.Resize(fyne.NewSize(700, 500))
	pathDialog.Show()
}

// createPathDetailContent 创建路径详细信息内容
func (ds *DirectorySelector) createPathDetailContent() fyne.CanvasObject {
	if ds.allowMultiple && len(ds.selectedPaths) > 0 {
		return ds.createMultiplePathDetailContent()
	}
	return ds.createSinglePathDetailContent()
}

// createSimplePathList 创建简洁版路径列表
func (ds *DirectorySelector) createSimplePathList() *widget.List {
	var list *widget.List
	list = widget.NewList(
		func() int {
			return len(ds.selectedPaths)
		},
		func() fyne.CanvasObject {
			// 创建可点击的目录名按钮
			dirNameBtn := widget.NewButton("", nil)
			dirNameBtn.Importance = widget.LowImportance
			dirNameBtn.Alignment = widget.ButtonAlignLeading

			// 简洁的列表项：可点击目录名 + 操作按钮
			return container.NewHBox(
				dirNameBtn,
				layout.NewSpacer(),
				widget.NewButtonWithIcon("", theme.FolderOpenIcon(), nil),
				widget.NewButtonWithIcon("", theme.DeleteIcon(), nil),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			path := ds.selectedPaths[id]
			dirName := filepath.Base(path)

			hbox := item.(*fyne.Container)
			if len(hbox.Objects) >= 4 {
				// 更新目录名按钮
				dirNameBtn := hbox.Objects[0].(*widget.Button)
				displayText := fmt.Sprintf("%s - %s", dirName, ds.truncateText(path, 60))
				dirNameBtn.SetText(displayText)

				// 设置目录名点击事件 - 查看完整信息
				dirNameBtn.OnTapped = func() {
					ds.showDirectoryInfo(path)
				}

				// 更新打开按钮事件
				openBtn := hbox.Objects[2].(*widget.Button)
				openBtn.OnTapped = func() {
					if err := browser.OpenFile(path); err != nil {
						dialog.ShowError(fmt.Errorf("无法打开目录: %v", err), ds.window)
					}
				}

				// 更新删除按钮事件
				deleteBtn := hbox.Objects[3].(*widget.Button)
				deleteBtn.OnTapped = func() {
					confirmDialog := dialog.NewConfirm(
						"确认删除",
						fmt.Sprintf("确定要删除 '%s' 吗？", dirName),
						func(confirmed bool) {
							if confirmed {
								newPaths := make([]string, 0, len(ds.selectedPaths)-1)
								for i, p := range ds.selectedPaths {
									if i != id {
										newPaths = append(newPaths, p)
									}
								}
								ds.SetSelectedPaths(newPaths)
								list.Refresh() // 刷新列表
							}
						},
						ds.window,
					)
					confirmDialog.Show()
				}
			}
		},
	)

	return list
}

// showDirectoryInfo 显示单个目录的完整信息
func (ds *DirectorySelector) showDirectoryInfo(path string) {
	dirName := filepath.Base(path)
	parentDir := filepath.Dir(path)

	// 创建信息内容
	infoContent := container.NewVBox(
		widget.NewLabelWithStyle("目录完整信息", fyne.TextAlignCenter, fyne.TextStyle{Bold: true, Italic: true}),
		widget.NewSeparator(),

		// 基本信息网格
		container.NewGridWithColumns(2,
			widget.NewLabelWithStyle("目录名称:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel(dirName),
			widget.NewLabelWithStyle("完整路径:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			container.NewHScroll(widget.NewLabel(path)),
			widget.NewLabelWithStyle("父级目录:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel(parentDir),
			widget.NewLabelWithStyle("路径长度:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel(fmt.Sprintf("%d 字符", len(path))),
		),

		layout.NewSpacer(),

		// 操作按钮
		container.NewHBox(
			widget.NewButtonWithIcon("复制路径", theme.ContentCopyIcon(), func() {
				ds.window.Clipboard().SetContent(path)
				dialog.ShowInformation("成功", "路径已复制到剪贴板", ds.window)
			}),
			widget.NewButtonWithIcon("打开目录", theme.FolderOpenIcon(), func() {
				if err := browser.OpenFile(path); err != nil {
					dialog.ShowError(fmt.Errorf("无法打开目录: %v", err), ds.window)
				}
			}),
		),
	)

	// 创建对话框
	infoDialog := dialog.NewCustom(
		fmt.Sprintf("目录信息 - %s", dirName),
		"关闭",
		infoContent,
		ds.window,
	)

	infoDialog.Resize(fyne.NewSize(600, 300))
	infoDialog.Show()
}

// createMultiplePathDetailContent 创建多路径详细信息内容
func (ds *DirectorySelector) createMultiplePathDetailContent() fyne.CanvasObject {
	// 顶部统计信息和操作说明
	header := container.NewVBox(
		widget.NewLabelWithStyle(
			fmt.Sprintf("已选择 %d 个目录", len(ds.selectedPaths)),
			fyne.TextAlignCenter,
			fyne.TextStyle{Bold: true},
		),
		widget.NewLabelWithStyle(
			"点击目录名查看完整信息 • 点击图标打开目录或删除",
			fyne.TextAlignCenter,
			fyne.TextStyle{Italic: true},
		),
	)

	// 路径列表
	pathList := ds.createSimplePathList()

	// 批量操作按钮
	batchActions := container.NewHBox(
		widget.NewButtonWithIcon("添加目录", theme.ContentAddIcon(), func() {
			ds.selectMultipleDirectories()
		}),
		widget.NewButtonWithIcon("复制全部", theme.ContentCopyIcon(), func() {
			allPaths := strings.Join(ds.selectedPaths, "\n")
			ds.window.Clipboard().SetContent(allPaths)
			dialog.ShowInformation("成功", "所有路径已复制到剪贴板", ds.window)
		}),
		widget.NewButtonWithIcon("清空全部", theme.DeleteIcon(), func() {
			confirmDialog := dialog.NewConfirm(
				"确认清空",
				fmt.Sprintf("确定要清空所有 %d 个目录吗？", len(ds.selectedPaths)),
				func(confirmed bool) {
					if confirmed {
						ds.ClearSelection()
						// 清空后刷新列表
						pathList.Refresh()
					}
				},
				ds.window,
			)
			confirmDialog.Show()
		}),
	)

	// 创建主内容
	content := container.NewBorder(
		header,
		batchActions,
		nil, nil,
		container.NewScroll(pathList),
	)

	return content
}

// truncateText 智能截断文本
func (ds *DirectorySelector) truncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}

	// 如果包含路径分隔符，尝试智能截断
	if strings.Contains(text, string(filepath.Separator)) {
		parts := strings.Split(text, string(filepath.Separator))
		if len(parts) >= 3 {
			// 保留开头和结尾部分
			head := parts[0]
			tail := parts[len(parts)-1]

			availableLength := maxLength - 5 // 为 "..." 留出空间
			if len(head)+len(tail)+3 <= availableLength {
				// 如果开头和结尾能放下，就显示它们
				return head + string(filepath.Separator) + "..." + string(filepath.Separator) + tail
			}
		}
	}

	// 简单截断
	return text[:maxLength-3] + "..."
}

// 在 showPathDetails 中为单个目录也添加完整路径查看功能
func (ds *DirectorySelector) createSinglePathDetailContent() fyne.CanvasObject {
	// 基本信息卡片
	infoCard := ds.createInfoCard(ds.selectedPath)

	// 操作按钮
	actionButtons := container.NewHBox(
		widget.NewButtonWithIcon("复制路径", theme.ContentCopyIcon(), func() {
			ds.window.Clipboard().SetContent(ds.selectedPath)
			dialog.ShowInformation("成功", "路径已复制到剪贴板", ds.window)
		}),
		widget.NewButtonWithIcon("打开目录", theme.FolderOpenIcon(), func() {
			if err := browser.OpenFile(ds.selectedPath); err != nil {
				dialog.ShowError(fmt.Errorf("无法打开目录: %v", err), ds.window)
			}
		}),
		widget.NewButtonWithIcon("重新选择", theme.ViewRefreshIcon(), func() {
			ds.selectDirectory()
		}),
	)

	return container.NewVBox(
		widget.NewLabelWithStyle("目录信息", fyne.TextAlignCenter, fyne.TextStyle{Bold: true, Italic: true}),
		infoCard,
		layout.NewSpacer(),
		actionButtons,
	)
}

// createInfoCard 创建目录信息卡片
func (ds *DirectorySelector) createInfoCard(path string) *widget.Card {
	dirName := filepath.Base(path)
	parentDir := filepath.Dir(path)

	infoContent := container.NewVBox(
		widget.NewLabelWithStyle("完整路径:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewScroll(widget.NewLabel(path)),
		widget.NewSeparator(),
		container.NewGridWithColumns(2,
			widget.NewLabel("目录名称:"),
			widget.NewLabel(dirName),
			widget.NewLabel("父级目录:"),
			widget.NewLabel(parentDir),
			widget.NewLabel("路径长度:"),
			widget.NewLabel(fmt.Sprintf("%d 字符", len(path))),
		),
	)

	return widget.NewCard("详细信息", "", infoContent)
}

// updatePathDisplay 更新路径显示
func (ds *DirectorySelector) updatePathDisplay() {
	if ds.allowMultiple && len(ds.selectedPaths) > 0 {
		displayText := ""
		if len(ds.selectedPaths) == 1 {
			displayText = ds.truncatePathForDisplay(ds.selectedPaths[0])
		} else {
			displayText = fmt.Sprintf("已选择 %d 个目录", len(ds.selectedPaths))
		}
		ds.pathLabel.SetText(displayText)
		ds.pathButton.SetText(displayText)
	} else if ds.selectedPath != "" {
		displayPath := ds.truncatePathForDisplay(ds.selectedPath)
		ds.pathLabel.SetText(displayPath)
		ds.pathButton.SetText(displayPath)
	} else {
		ds.pathLabel.SetText("未选择目录")
		ds.pathButton.SetText("未选择目录")
	}
	ds.Refresh()
}

// truncatePathForDisplay 智能截断路径显示
func (ds *DirectorySelector) truncatePathForDisplay(fullPath string) string {
	if len(fullPath) <= ds.maxDisplayLength {
		return fullPath
	}

	// 尝试基于路径分隔符智能截断
	parts := strings.Split(fullPath, string(filepath.Separator))
	if len(parts) <= 2 {
		// 简单路径，直接截断
		return ds.simpleTruncate(fullPath)
	}

	// 保留开头和结尾部分
	headParts := []string{parts[0]}            // 根目录
	tailParts := []string{parts[len(parts)-1]} // 最后一级目录

	availableLength := ds.maxDisplayLength - 3 // 减去 "..."
	currentLength := len(headParts[0]) + len(tailParts[0])

	// 从开头逐步添加部分，直到达到长度限制
	for i := 1; i < len(parts)-1 && currentLength < availableLength; i++ {
		part := parts[i]
		if currentLength+len(part)+1 <= availableLength {
			headParts = append(headParts, part)
			currentLength += len(part) + 1
		} else {
			break
		}
	}

	// 构建显示路径
	head := strings.Join(headParts, string(filepath.Separator))
	tail := strings.Join(tailParts, string(filepath.Separator))

	return head + string(filepath.Separator) + "..." + string(filepath.Separator) + tail
}

// simpleTruncate 简单截断
func (ds *DirectorySelector) simpleTruncate(fullPath string) string {
	if len(fullPath) <= ds.maxDisplayLength {
		return fullPath
	}
	headLength := ds.maxDisplayLength / 3
	tailLength := ds.maxDisplayLength - headLength - 3
	return fullPath[:headLength] + "..." + fullPath[len(fullPath)-tailLength:]
}

// GetSelectedPath 获取当前选择的路径（单选的路径）
func (ds *DirectorySelector) GetSelectedPath() string {
	return ds.selectedPath
}

// GetSelectedPaths 获取所有选择的路径（多选）
func (ds *DirectorySelector) GetSelectedPaths() []string {
	return cloneStrings(ds.selectedPaths)
}

// SetSelectedPath 设置单个路径并触发回调
func (ds *DirectorySelector) SetSelectedPath(path string) {
	ds.selectedPath = path
	ds.selectedPaths = []string{path} // 同步更新多选路径
	ds.updatePathDisplay()

	if ds.OnPathSelected != nil {
		ds.OnPathSelected(ds.selectedPath)
	}
	if ds.OnPathsSelected != nil {
		ds.OnPathsSelected(cloneStrings(ds.selectedPaths))
	}
}

// SetSelectedPaths 设置多个路径并触发回调
func (ds *DirectorySelector) SetSelectedPaths(paths []string) {
	if len(paths) == 0 {
		ds.ClearSelection()
		return
	}

	ds.selectedPaths = cloneStrings(paths)
	if len(paths) == 1 {
		ds.selectedPath = paths[0]
	} else {
		ds.selectedPath = "" // 多选时清空单一路径
	}
	ds.updatePathDisplay()

	if ds.OnPathsSelected != nil {
		ds.OnPathsSelected(cloneStrings(ds.selectedPaths))
	}
	// 如果是单选，也触发单一路径回调
	if len(paths) == 1 && ds.OnPathSelected != nil {
		ds.OnPathSelected(paths[0])
	}
}

// ClearSelection 清空选择
func (ds *DirectorySelector) ClearSelection() {
	ds.selectedPath = ""
	ds.selectedPaths = []string{}
	ds.updatePathDisplay()

	// 触发回调
	if ds.OnPathsSelected != nil {
		ds.OnPathsSelected(cloneStrings(ds.selectedPaths))
	}
	if ds.OnPathSelected != nil {
		ds.OnPathSelected("")
	}
}

// IsEmpty 检查是否已选择目录
func (ds *DirectorySelector) IsEmpty() bool {
	return ds.selectedPath == "" && len(ds.selectedPaths) == 0
}

// GetSelectedCount 获取选择的目录数量
func (ds *DirectorySelector) GetSelectedCount() int {
	if ds.allowMultiple {
		return len(ds.selectedPaths)
	}
	if ds.selectedPath != "" {
		return 1
	}
	return 0
}

// SimpleDirectorySelector 简化版目录选择器（基于主选择器的轻量封装）
type SimpleDirectorySelector struct {
	*DirectorySelector
}

// NewSimpleDirectorySelector 创建简化版目录选择器
func NewSimpleDirectorySelector(window fyne.Window, onSelected func(string)) *SimpleDirectorySelector {
	config := DirectorySelectorConfig{
		Window:           window,
		OnPathSelected:   onSelected,
		MaxDisplayLength: 30,
		ButtonText:       "选择目录",
		AllowMultiple:    false,
	}

	baseSelector := NewDirectorySelectorWithConfig(config)
	return &SimpleDirectorySelector{baseSelector}
}

// NewMultiDirectorySelector 创建支持多选的目录选择器
func NewMultiDirectorySelector(window fyne.Window, onPathsSelected func([]string)) *DirectorySelector {
	config := DirectorySelectorConfig{
		Window:           window,
		OnPathsSelected:  onPathsSelected,
		MaxDisplayLength: 40,
		ButtonText:       "选择目录",
		AllowMultiple:    true,
	}

	return NewDirectorySelectorWithConfig(config)
}

// GetPath 为简化版提供兼容性方法
func (sds *SimpleDirectorySelector) GetPath() string {
	return sds.GetSelectedPath()
}

func (sds *SimpleDirectorySelector) SetPath(path string) {
	sds.SetSelectedPath(path)
}

func (sds *SimpleDirectorySelector) Clear() {
	sds.ClearSelection()
}
