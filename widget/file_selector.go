package widget

import (
	"errors"
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/ncruces/zenity"
	"github.com/pkg/browser"
)

// FileSelectorMode 文件选择器模式
type FileSelectorMode int

const (
	ModeMultiple  FileSelectorMode = iota // 多文件选择模式（默认）
	ModeSingle                            // 单文件选择模式
	ModeDirectory                         // 目录选择模式
)

// FileFilter 文件过滤器
type FileFilter struct {
	Name     string   // 过滤器名称
	Patterns []string // 文件模式
}

func cloneStrings(values []string) []string {
	return append([]string(nil), values...)
}

// FileSelector 文件选择器组件
type FileSelector struct {
	widget.BaseWidget

	button         *widget.Button
	countLabel     *widget.Label
	countButton    *widget.Button
	files          []string
	OnFileSelected func([]string)
	window         fyne.Window
	currentDialog  dialog.Dialog // 跟踪当前打开的对话框
	mode           FileSelectorMode
	filters        []FileFilter // 文件过滤器
	defaultFilter  *FileFilter  // 默认过滤器
	LastFile       *LastFile
}

// NewFileSelector 创建新的文件选择器
func NewFileSelector(window fyne.Window, onFileSelected func([]string)) *FileSelector {
	return NewFileSelectorWithMode(window, onFileSelected, ModeMultiple)
}

// NewFileSelectorWithMode 创建新的文件选择器并指定模式
func NewFileSelectorWithMode(window fyne.Window, onFileSelected func([]string), mode FileSelectorMode) *FileSelector {
	// 默认过滤器：文本文件和所有文件
	defaultFilters := []FileFilter{
		{Name: "文本文件", Patterns: []string{"*.txt"}},
		{Name: "所有文件", Patterns: []string{"*"}},
	}

	fs := &FileSelector{
		window:         window,
		OnFileSelected: onFileSelected,
		files:          []string{},
		mode:           mode,
		filters:        defaultFilters,
		defaultFilter:  &defaultFilters[0], // 默认使用第一个过滤器
	}

	// 根据模式设置按钮文本和图标
	var buttonText string
	var buttonIcon fyne.Resource

	switch mode {
	case ModeSingle:
		buttonText = "选择文件"
		buttonIcon = theme.FileIcon()
	case ModeDirectory:
		buttonText = "选择目录"
		buttonIcon = theme.FolderIcon()
	default: // 多文件模式
		buttonText = "选择文件"
		buttonIcon = theme.FolderOpenIcon()
	}

	// 创建选择文件按钮
	fs.button = widget.NewButtonWithIcon(buttonText, buttonIcon, fs.selectFiles)

	// 创建显示文件数量的标签
	fs.countLabel = widget.NewLabel("0 个文件")
	fs.countLabel.TextStyle = fyne.TextStyle{Bold: true}

	// 创建可点击的文件计数按钮
	fs.countButton = widget.NewButton("", fs.showFileList)
	fs.countButton.Importance = widget.LowImportance

	fs.ExtendBaseWidget(fs)
	return fs
}

// NewFileSelectorWithFilters 创建新的文件选择器并指定过滤器
func NewFileSelectorWithFilters(window fyne.Window, onFileSelected func([]string), mode FileSelectorMode, filters []FileFilter) *FileSelector {
	fs := &FileSelector{
		window:         window,
		OnFileSelected: onFileSelected,
		files:          []string{},
		mode:           mode,
		filters:        filters,
	}

	// 设置默认过滤器
	if len(filters) > 0 {
		fs.defaultFilter = &filters[0]
	} else {
		// 如果没有提供过滤器，使用默认的
		defaultFilters := []FileFilter{
			{Name: "所有文件", Patterns: []string{"*"}},
		}
		fs.filters = defaultFilters
		fs.defaultFilter = &defaultFilters[0]
	}

	// 根据模式设置按钮文本和图标
	var buttonText string
	var buttonIcon fyne.Resource

	switch mode {
	case ModeSingle:
		buttonText = "选择文件"
		buttonIcon = theme.FileIcon()
	case ModeDirectory:
		buttonText = "选择目录"
		buttonIcon = theme.FolderIcon()
	default: // 多文件模式
		buttonText = "选择文件"
		buttonIcon = theme.FolderOpenIcon()
	}

	// 创建选择文件按钮
	fs.button = widget.NewButtonWithIcon(buttonText, buttonIcon, fs.selectFiles)

	// 创建显示文件数量的标签
	fs.countLabel = widget.NewLabel("0 个文件")
	fs.countLabel.TextStyle = fyne.TextStyle{Bold: true}

	// 创建可点击的文件计数按钮
	fs.countButton = widget.NewButton("", fs.showFileList)
	fs.countButton.Importance = widget.LowImportance

	fs.ExtendBaseWidget(fs)
	return fs
}

// CreateRenderer 创建组件渲染器
func (fs *FileSelector) CreateRenderer() fyne.WidgetRenderer {
	// 使用容器来组合计数按钮和标签
	countContainer := container.NewStack(
		fs.countButton,
		container.NewCenter(fs.countLabel),
	)

	// 使用水平布局，按钮在左边，文件计数在右边
	content := container.NewHBox(
		fs.button,
		layout.NewSpacer(),
		countContainer,
	)

	return widget.NewSimpleRenderer(content)
}

// selectFiles 选择文件 - 根据模式选择不同的选择方式
func (fs *FileSelector) selectFiles() {
	switch fs.mode {
	case ModeSingle:
		fs.selectSingleFile()
	case ModeDirectory:
		fs.selectDirectory()
	default: // 多文件模式
		fs.selectMultipleFiles(false)
	}
}

// selectSingleFile 选择单个文件
func (fs *FileSelector) selectSingleFile() {
	var defFilename string
	if fs.LastFile != nil {
		defer fs.LastFile.Close()
		defFilename = fs.LastFile.Value
	}

	// 转换过滤器格式
	zenityFilters := fs.convertFilters()

	selectedFile, err := zenity.SelectFile(
		zenity.Filename(defFilename),
		zenity.Title("选择文件"),
		zenity.FileFilters(zenityFilters),
	)

	if err != nil {
		if errors.Is(err, zenity.ErrCanceled) {
			// 用户取消了选择
			return
		}
		dialog.ShowError(err, fs.window)
		return
	}

	if selectedFile != "" {
		if fs.LastFile != nil {
			fs.LastFile.Value = selectedFile
		}
		// 单文件模式下，替换而不是追加
		fs.files = []string{selectedFile}
		fs.updateFileCount()

		// 调用回调函数
		if fs.OnFileSelected != nil {
			fs.OnFileSelected(cloneStrings(fs.files))
		}
	}
}

// selectMultipleFiles 选择多个文件（追加模式）
func (fs *FileSelector) selectMultipleFiles(replace bool) {
	var defFilename string
	if fs.LastFile != nil {
		defer fs.LastFile.Close()
		defFilename = fs.LastFile.Value
	}
	// 转换过滤器格式
	zenityFilters := fs.convertFilters()

	selectedFiles, err := zenity.SelectFileMultiple(zenity.Filename(defFilename),
		zenity.Title("选择多个文件"),
		zenity.FileFilters(zenityFilters),
	)

	if err != nil {
		if errors.Is(err, zenity.ErrCanceled) {
			// 用户取消了选择
			return
		}
		dialog.ShowError(err, fs.window)
		return
	}

	if len(selectedFiles) > 0 {
		if fs.LastFile != nil {
			fs.LastFile.Value = selectedFiles[0]
		}
		if replace {
			fs.files = selectedFiles
		} else {
			fs.files = append(fs.files, selectedFiles...)
		}
		fs.updateFileCount()

		// 调用回调函数
		if fs.OnFileSelected != nil {
			fs.OnFileSelected(cloneStrings(fs.files))
		}
	}
}

// selectDirectory 选择目录
func (fs *FileSelector) selectDirectory() {
	var defFilename string
	if fs.LastFile != nil {
		defer fs.LastFile.Close()
		defFilename = fs.LastFile.Value
	}
	selectedDir, err := zenity.SelectFile(zenity.Filename(defFilename),
		zenity.Title("选择目录"),
		zenity.Directory(),
	)

	if err != nil {
		if !errors.Is(err, zenity.ErrCanceled) {
			dialog.ShowError(err, fs.window)
		}
		return
	}

	if selectedDir != "" {
		if fs.LastFile != nil {
			fs.LastFile.Value = selectedDir
		}
		fs.files = []string{selectedDir}
		fs.updateFileCount()

		if fs.OnFileSelected != nil {
			fs.OnFileSelected(cloneStrings(fs.files))
		}
	}
}

// convertFilters 将内部过滤器格式转换为 zenity 格式
func (fs *FileSelector) convertFilters() []zenity.FileFilter {
	filters := fs.filters
	if fs.defaultFilter != nil {
		for i, filter := range filters {
			if filter.Name == fs.defaultFilter.Name {
				ordered := make([]FileFilter, 0, len(filters))
				ordered = append(ordered, filter)
				ordered = append(ordered, filters[:i]...)
				ordered = append(ordered, filters[i+1:]...)
				filters = ordered
				break
			}
		}
	}

	zenityFilters := make([]zenity.FileFilter, 0, len(filters))
	for _, filter := range filters {
		zenityFilters = append(zenityFilters, zenity.FileFilter{
			Name:     filter.Name,
			Patterns: append([]string(nil), filter.Patterns...),
		})
	}
	return zenityFilters
}

// showFileList 显示文件列表
func (fs *FileSelector) showFileList() {
	if len(fs.files) == 0 {
		var message string
		switch fs.mode {
		case ModeSingle:
			message = "没有选择任何文件"
		case ModeDirectory:
			message = "没有选择任何目录"
		default:
			message = "没有选择任何文件"
		}
		dialog.ShowInformation("文件列表", message, fs.window)
		return
	}

	// 关闭之前可能打开的对话框
	if fs.currentDialog != nil {
		fs.currentDialog.Hide()
	}

	// 创建文件列表对话框
	fileListDialog := dialog.NewCustom(
		"已选择的文件",
		"关闭",
		fs.createFileListContent(),
		fs.window,
	)

	fs.currentDialog = fileListDialog
	fileListDialog.Resize(fyne.NewSize(600, 400)) // 增加宽度以容纳更多按钮
	fileListDialog.SetOnClosed(func() {
		fs.currentDialog = nil
	})
	fileListDialog.Show()
}

// createFileListContent 创建文件列表内容
func (fs *FileSelector) createFileListContent() fyne.CanvasObject {
	if len(fs.files) == 0 {
		return widget.NewLabel("没有选择任何文件")
	}

	// 创建文件列表
	fileList := widget.NewList(
		func() int {
			return len(fs.files)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("000."),
				widget.NewLabel("文件名会显示在这里"),
				layout.NewSpacer(),
				widget.NewButtonWithIcon("", theme.DocumentIcon(), nil), // 打开按钮
				widget.NewButtonWithIcon("", theme.DeleteIcon(), nil),   // 删除按钮
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			box := item.(*fyne.Container)
			numberLabel := box.Objects[0].(*widget.Label)
			fileLabel := box.Objects[1].(*widget.Label)
			openBtn := box.Objects[3].(*widget.Button)
			removeBtn := box.Objects[4].(*widget.Button)

			numberLabel.SetText(fmt.Sprintf("%d.", id+1))
			fileLabel.SetText(filepath.Base(fs.files[id]))

			// 更新打开按钮的回调
			openBtn.OnTapped = func() {
				filePath := fs.files[id]
				if fs.mode == ModeDirectory {
					// 对于目录，显示目录信息而不是尝试打开
					dialog.ShowInformation("目录信息",
						fmt.Sprintf("目录路径: %s", filePath), fs.window)
				} else {
					err := browser.OpenFile(filePath)
					if err != nil {
						dialog.ShowError(fmt.Errorf("无法打开文件: %v", err), fs.window)
					}
				}
			}

			// 单文件模式和目录模式不允许删除单个项目
			if fs.mode == ModeSingle || fs.mode == ModeDirectory {
				removeBtn.Hide()
			} else {
				removeBtn.Show()
				// 更新移除按钮的回调
				removeBtn.OnTapped = func() {
					// 确认删除对话框
					confirmDialog := dialog.NewConfirm(
						"确认删除",
						fmt.Sprintf("确定要移除文件 '%s' 吗？", filepath.Base(fs.files[id])),
						func(confirmed bool) {
							if confirmed {
								fs.removeFile(id)
								// 刷新对话框内容
								if fs.currentDialog != nil {
									fs.currentDialog.Hide()
									fs.showFileList() // 重新打开更新后的对话框
								}
							}
						},
						fs.window,
					)
					confirmDialog.Show()
				}
			}
		},
	)

	// 添加列表项点击事件（点击文件项本身）
	fileList.OnSelected = func(id widget.ListItemID) {
		filePath := fs.files[id]
		var info string
		if fs.mode == ModeDirectory {
			info = fmt.Sprintf("目录名: %s\n完整路径: %s", filepath.Base(filePath), filePath)
		} else {
			info = fmt.Sprintf("文件名: %s\n完整路径: %s", filepath.Base(filePath), filePath)
		}
		dialog.ShowInformation("文件信息", info, fs.window)
		fileList.Unselect(id) // 取消选中状态
	}

	// 创建顶部信息栏
	var infoText string
	switch fs.mode {
	case ModeSingle:
		infoText = "已选择的文件:"
	case ModeDirectory:
		infoText = "已选择的目录:"
	default:
		infoText = fmt.Sprintf("共 %d 个文件:", len(fs.files))
	}

	infoLabel := widget.NewLabel(infoText)
	infoLabel.TextStyle = fyne.TextStyle{Bold: true}

	// 创建底部按钮栏
	var bottomButtons fyne.CanvasObject

	if fs.mode == ModeSingle || fs.mode == ModeDirectory {
		// 单文件或目录模式，只显示重新选择按钮
		var buttonText string
		if fs.mode == ModeDirectory {
			buttonText = "重新选择目录"
		} else {
			buttonText = "重新选择文件"
		}

		bottomButtons = container.NewHBox(
			widget.NewButtonWithIcon(buttonText, theme.DocumentCreateIcon(), func() {
				fs.selectFiles()
				if fs.currentDialog != nil {
					fs.currentDialog.Hide()
				}
			}),
			layout.NewSpacer(),
			widget.NewButtonWithIcon("清空", theme.DeleteIcon(), func() {
				confirmDialog := dialog.NewConfirm(
					"确认清空",
					"确定要清空选择吗？",
					func(confirmed bool) {
						if confirmed {
							fs.ClearFiles()
							if fs.currentDialog != nil {
								fs.currentDialog.Hide()
							}
						}
					},
					fs.window,
				)
				confirmDialog.Show()
			}),
		)
	} else {
		// 多文件模式，显示完整的按钮组
		bottomButtons = container.NewHBox(
			widget.NewButtonWithIcon("添加文件", theme.ContentAddIcon(), func() {
				fs.selectMultipleFiles(false)
				if fs.currentDialog != nil {
					fs.currentDialog.Hide()
					fs.showFileList()
				}
			}),
			widget.NewButtonWithIcon("替换所有", theme.DocumentCreateIcon(), func() {
				fs.selectMultipleFiles(true)
				if fs.currentDialog != nil {
					fs.currentDialog.Hide()
					fs.showFileList()
				}
			}),
			layout.NewSpacer(),
			widget.NewButtonWithIcon("清空所有", theme.DeleteIcon(), func() {
				confirmDialog := dialog.NewConfirm(
					"确认清空",
					fmt.Sprintf("确定要清空所有 %d 个文件吗？", len(fs.files)),
					func(confirmed bool) {
						if confirmed {
							fs.ClearFiles()
							if fs.currentDialog != nil {
								fs.currentDialog.Hide()
							}
						}
					},
					fs.window,
				)
				confirmDialog.Show()
			}),
		)
	}

	return container.NewBorder(
		infoLabel,
		bottomButtons,
		nil, nil,
		fileList,
	)
}

// removeFile 移除单个文件（仅多文件模式使用）
func (fs *FileSelector) removeFile(index int) {
	if index >= 0 && index < len(fs.files) {
		removedFile := fs.files[index]
		fs.files = append(fs.files[:index], fs.files[index+1:]...)
		fs.updateFileCount()

		if fs.OnFileSelected != nil {
			fs.OnFileSelected(cloneStrings(fs.files))
		}

		// 显示移除成功的消息
		dialog.ShowInformation("移除成功", fmt.Sprintf("已移除文件: %s", filepath.Base(removedFile)), fs.window)

		// 刷新显示
		fs.Refresh()
	}
}

// updateFileCount 更新文件数量显示
func (fs *FileSelector) updateFileCount() {
	count := len(fs.files)

	if count == 0 {
		switch fs.mode {
		case ModeSingle:
			fs.countLabel.SetText("未选择")
		case ModeDirectory:
			fs.countLabel.SetText("未选择")
		default:
			fs.countLabel.SetText("0 个文件")
		}
	} else if count == 1 {
		switch fs.mode {
		case ModeSingle:
			fs.countLabel.SetText("1 个文件")
		case ModeDirectory:
			fs.countLabel.SetText("1 个目录")
		default:
			fs.countLabel.SetText("1 个文件")
		}
	} else {
		fs.countLabel.SetText(fmt.Sprintf("%d 个文件", count))
	}

	// 刷新显示
	fs.Refresh()
}

// GetFiles 获取当前选择的文件列表
func (fs *FileSelector) GetFiles() []string {
	return cloneStrings(fs.files)
}

// ClearFiles 清空文件列表
func (fs *FileSelector) ClearFiles() {
	fs.files = []string{}
	fs.updateFileCount()

	if fs.OnFileSelected != nil {
		fs.OnFileSelected(cloneStrings(fs.files))
	}
}

// SetFiles 设置文件列表
func (fs *FileSelector) SetFiles(files []string) {
	// 单文件模式只取第一个文件
	if fs.mode == ModeSingle && len(files) > 0 {
		fs.files = []string{files[0]}
	} else {
		fs.files = cloneStrings(files)
	}
	fs.updateFileCount()

	if fs.OnFileSelected != nil {
		fs.OnFileSelected(cloneStrings(fs.files))
	}
}

// SetMode 设置选择器模式
func (fs *FileSelector) SetMode(mode FileSelectorMode) {
	fs.mode = mode

	// 更新按钮文本和图标
	switch mode {
	case ModeSingle:
		fs.button.SetText("选择文件")
		fs.button.SetIcon(theme.FileIcon())
	case ModeDirectory:
		fs.button.SetText("选择目录")
		fs.button.SetIcon(theme.FolderIcon())
	default: // 多文件模式
		fs.button.SetText("选择文件")
		fs.button.SetIcon(theme.FolderOpenIcon())
	}

	// 清空现有选择（因为模式改变了）
	fs.ClearFiles()
	fs.Refresh()
}

// GetMode 获取当前模式
func (fs *FileSelector) GetMode() FileSelectorMode {
	return fs.mode
}

// SetFilters 设置文件过滤器
func (fs *FileSelector) SetFilters(filters []FileFilter) {
	fs.filters = cloneFileFilters(filters)
	if len(filters) > 0 {
		fs.defaultFilter = &filters[0]
	} else {
		fs.defaultFilter = nil
	}
}

// GetFilters 获取当前文件过滤器
func (fs *FileSelector) GetFilters() []FileFilter {
	return cloneFileFilters(fs.filters)
}

// AddFilter 添加文件过滤器
func (fs *FileSelector) AddFilter(filter FileFilter) {
	filter.Patterns = append([]string(nil), filter.Patterns...)
	fs.filters = append(fs.filters, filter)
	if fs.defaultFilter == nil {
		fs.defaultFilter = &fs.filters[len(fs.filters)-1]
	}
}

// RemoveFilter 移除文件过滤器
func (fs *FileSelector) RemoveFilter(name string) bool {
	for i, filter := range fs.filters {
		if filter.Name == name {
			defaultName := ""
			if fs.defaultFilter != nil {
				defaultName = fs.defaultFilter.Name
			}
			fs.filters = append(fs.filters[:i], fs.filters[i+1:]...)
			fs.defaultFilter = nil
			for j := range fs.filters {
				if fs.filters[j].Name == defaultName {
					fs.defaultFilter = &fs.filters[j]
					break
				}
			}
			if fs.defaultFilter == nil && len(fs.filters) > 0 {
				fs.defaultFilter = &fs.filters[0]
			}
			return true
		}
	}
	return false
}

// SetDefaultFilter 设置默认过滤器
func (fs *FileSelector) SetDefaultFilter(name string) bool {
	for i, filter := range fs.filters {
		if filter.Name == name {
			fs.defaultFilter = &fs.filters[i]
			return true
		}
	}
	return false
}

// GetDefaultFilter 获取默认过滤器
func (fs *FileSelector) GetDefaultFilter() *FileFilter {
	if fs.defaultFilter == nil {
		return nil
	}
	filter := *fs.defaultFilter
	filter.Patterns = append([]string(nil), filter.Patterns...)
	return &filter
}

func cloneFileFilters(filters []FileFilter) []FileFilter {
	cloned := make([]FileFilter, len(filters))
	for i, filter := range filters {
		cloned[i] = filter
		cloned[i].Patterns = append([]string(nil), filter.Patterns...)
	}
	return cloned
}
