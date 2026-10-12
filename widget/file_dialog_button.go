package widget

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/zc310/lottery/fyne/icons/fontawesome"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/ncruces/zenity"
)

// FileDialogButton 文件对话框按钮组件
type FileDialogButton struct {
	button     *widget.Button
	onSelected func(filePath string)
	fileExt    *string
}

// SaveLoadButtons 保存和导入按钮组件
type SaveLoadButtons struct {
	container  *fyne.Container
	saveButton *FileDialogButton
	loadButton *FileDialogButton
	fileExt    string // 文件扩展名
}

// ButtonConfig 按钮配置
type ButtonConfig struct {
	Text     string
	Icon     fyne.Resource
	FileName string
}

// FileTypeConfig 文件类型配置
type FileTypeConfig struct {
	Name     string
	Patterns []string
}

// NewFileDialogButton 创建文件对话框按钮
func NewFileDialogButton(config ButtonConfig, isSave bool, fileExt string, onSelected func(string)) *FileDialogButton {
	fileExt = normalizeFileExtension(fileExt)
	btn := &FileDialogButton{
		onSelected: onSelected,
		fileExt:    &fileExt,
	}

	button := widget.NewButtonWithIcon(config.Text, config.Icon, func() {
		var filePath string
		var err error
		lastFile := NewLastFile("_last_file"+config.FileName, config.FileName)
		defer lastFile.Close()
		// 构建文件过滤器
		filters := []zenity.FileFilter{{
			Name:     "所有文件",
			Patterns: []string{"*"},
		}}
		ext := *btn.fileExt
		if ext != "" {
			filters = append([]zenity.FileFilter{{
				Name:     strings.ToUpper(strings.TrimPrefix(ext, ".")) + "文件",
				Patterns: []string{"*" + ext},
			}}, filters...)
		}

		if isSave {
			// 保存文件对话框
			filePath, err = zenity.SelectFileSave(
				zenity.Filename(lastFile.Value),
				zenity.Title("保存文件"),
				zenity.FileFilters(filters),
			)
		} else {
			// 打开文件对话框
			filePath, err = zenity.SelectFile(
				zenity.Filename(lastFile.Value),
				zenity.Title("选择文件"),
				zenity.FileFilters(filters),
			)
		}

		if err != nil {
			if !errors.Is(err, zenity.ErrCanceled) {
				_ = zenity.Error("操作失败: "+err.Error(), zenity.Title("错误"))
			}
			return
		}
		lastFile.Value = filePath

		if filePath != "" {
			if isSave && ext != "" && filepath.Ext(filePath) == "" {
				filePath += ext
			}
			if btn.onSelected != nil {
				btn.onSelected(filePath)
			}
		}
	})

	btn.button = button
	return btn
}

func normalizeFileExtension(fileExt string) string {
	fileExt = strings.TrimSpace(fileExt)
	if fileExt == "" || fileExt == "." {
		return ""
	}
	if !strings.HasPrefix(fileExt, ".") {
		fileExt = "." + fileExt
	}
	return fileExt
}

// NewSaveLoadButtons 创建保存和导入按钮组件
func NewSaveLoadButtons(fileExt, fileName string, onSave func(string), onLoad func(string)) *SaveLoadButtons {
	comp := &SaveLoadButtons{
		fileExt: normalizeFileExtension(fileExt),
	}

	// 默认配置
	saveConfig := ButtonConfig{
		Text:     "保存",
		FileName: fileName,
	}

	loadConfig := ButtonConfig{
		Text:     "导入",
		FileName: fileName,
	}

	// 创建按钮
	comp.saveButton = NewFileDialogButton(saveConfig, true, fileExt, onSave)
	comp.loadButton = NewFileDialogButton(loadConfig, false, fileExt, onLoad)

	// 创建容器
	comp.container = container.NewHBox(
		comp.loadButton.button,
		comp.saveButton.button,
	)

	return comp
}

// NewSaveLoadButtonsWithConfig 使用自定义配置创建按钮
func NewSaveLoadButtonsWithConfig(saveConfig, loadConfig ButtonConfig, fileExt string, onSave, onLoad func(string)) *SaveLoadButtons {
	comp := &SaveLoadButtons{
		fileExt: normalizeFileExtension(fileExt),
	}

	comp.saveButton = NewFileDialogButton(saveConfig, true, fileExt, onSave)
	comp.loadButton = NewFileDialogButton(loadConfig, false, fileExt, onLoad)

	comp.container = container.NewHBox(
		comp.loadButton.button,
		comp.saveButton.button,
	)

	return comp
}

// NewSaveLoadButtonsWithFileTypes 使用自定义文件类型创建按钮
func NewSaveLoadButtonsWithFileTypes(fileTypes []FileTypeConfig, defaultFileName, defaultExt string, onSave, onLoad func(string)) *SaveLoadButtons {
	comp := &SaveLoadButtons{
		fileExt: normalizeFileExtension(defaultExt),
	}

	saveConfig := ButtonConfig{
		Text: "保存配置",
		Icon: theme.NewThemedResource(fontawesome.Save),
	}

	loadConfig := ButtonConfig{
		Text: "导入配置",
		Icon: theme.NewThemedResource(fontawesome.FolderOpen),
	}

	// 创建带文件类型过滤的按钮
	saveBtn := widget.NewButtonWithIcon(saveConfig.Text, saveConfig.Icon, func() {
		var filters []zenity.FileFilter
		for _, ft := range fileTypes {
			filters = append(filters, zenity.FileFilter{
				Name:     ft.Name,
				Patterns: ft.Patterns,
			})
		}

		filePath, err := zenity.SelectFileSave(
			zenity.Title("保存配置文件"),
			zenity.FileFilters(filters),
			zenity.Filename(defaultFileName),
		)

		if err != nil {
			if !errors.Is(err, zenity.ErrCanceled) {
				_ = zenity.Error("保存失败: "+err.Error(), zenity.Title("错误"))
			}
			return
		}

		if filePath != "" {
			// 如果没有扩展名，添加默认扩展名
			if filepath.Ext(filePath) == "" && comp.fileExt != "" {
				filePath += comp.fileExt
			}
			if onSave != nil {
				onSave(filePath)
			}
		}
	})

	loadBtn := widget.NewButtonWithIcon(loadConfig.Text, loadConfig.Icon, func() {
		var filters []zenity.FileFilter
		for _, ft := range fileTypes {
			filters = append(filters, zenity.FileFilter{
				Name:     ft.Name,
				Patterns: ft.Patterns,
			})

		}

		filePath, err := zenity.SelectFile(
			zenity.Title("选择配置文件"),
			zenity.FileFilters(filters),
			zenity.Filename(defaultFileName),
		)

		if err != nil {
			if !errors.Is(err, zenity.ErrCanceled) {
				_ = zenity.Error("导入失败: "+err.Error(), zenity.Title("错误"))
			}
			return
		}

		if filePath != "" && onLoad != nil {
			onLoad(filePath)
		}
	})

	comp.saveButton = &FileDialogButton{button: saveBtn}
	comp.loadButton = &FileDialogButton{button: loadBtn}
	comp.container = container.NewHBox(loadBtn, saveBtn)

	return comp
}

// GetContainer 获取按钮容器
func (s *SaveLoadButtons) GetContainer() fyne.CanvasObject {
	return s.container
}

// SetFileExtension 设置文件扩展名
func (s *SaveLoadButtons) SetFileExtension(fileExt string) {
	s.fileExt = normalizeFileExtension(fileExt)
	if s.saveButton != nil {
		s.saveButton.setFileExtension(s.fileExt)
	}
	if s.loadButton != nil {
		s.loadButton.setFileExtension(s.fileExt)
	}
}

func (b *FileDialogButton) setFileExtension(fileExt string) {
	if b.fileExt != nil {
		*b.fileExt = fileExt
	}
}

// GetFileExtension 获取当前文件扩展名
func (s *SaveLoadButtons) GetFileExtension() string {
	return s.fileExt
}

// SetSaveButtonText 设置保存按钮文字
func (s *SaveLoadButtons) SetSaveButtonText(text string) {
	s.saveButton.button.SetText(text)
}

// SetLoadButtonText 设置导入按钮文字
func (s *SaveLoadButtons) SetLoadButtonText(text string) {
	s.loadButton.button.SetText(text)
}

// SetSaveButtonIcon 设置保存按钮图标
func (s *SaveLoadButtons) SetSaveButtonIcon(icon fyne.Resource) {
	s.saveButton.button.SetIcon(icon)
}

// SetLoadButtonIcon 设置导入按钮图标
func (s *SaveLoadButtons) SetLoadButtonIcon(icon fyne.Resource) {
	s.loadButton.button.SetIcon(icon)
}

// Enable 启用所有按钮
func (s *SaveLoadButtons) Enable() {
	s.saveButton.button.Enable()
	s.loadButton.button.Enable()
}

// Disable 禁用所有按钮
func (s *SaveLoadButtons) Disable() {
	s.saveButton.button.Disable()
	s.loadButton.button.Disable()
}

// SetSaveEnabled 设置保存按钮启用状态
func (s *SaveLoadButtons) SetSaveEnabled(enabled bool) {
	if enabled {
		s.saveButton.button.Enable()
	} else {
		s.saveButton.button.Disable()
	}
}

// SetLoadEnabled 设置导入按钮启用状态
func (s *SaveLoadButtons) SetLoadEnabled(enabled bool) {
	if enabled {
		s.loadButton.button.Enable()
	} else {
		s.loadButton.button.Disable()
	}
}

// 预定义的文件类型配置
var (
	// ZCP 文件类型
	ZCPFileTypes = []FileTypeConfig{
		{Name: "ZCP文件", Patterns: []string{"*.zcp"}},
		{Name: "所有文件", Patterns: []string{"*"}},
	}

	// JSON 文件类型
	JSONFileTypes = []FileTypeConfig{
		{Name: "JSON文件", Patterns: []string{"*.json"}},
		{Name: "所有文件", Patterns: []string{"*"}},
	}
)

// NewSaveLoadButtonsForZCP 创建用于ZCP文件的按钮（便捷函数）
func NewSaveLoadButtonsForZCP(fileName string, onSave, onLoad func(string)) *SaveLoadButtons {
	return NewSaveLoadButtons(".zcp", fileName, onSave, onLoad)
}

// NewSaveLoadButtonsForJSON 创建用于JSON文件的按钮（便捷函数）
func NewSaveLoadButtonsForJSON(fileName string, onSave, onLoad func(string)) *SaveLoadButtons {
	return NewSaveLoadButtons(".json", fileName, onSave, onLoad)
}

// NewSaveLoadButtonsForConfig 创建用于配置文件的按钮（便捷函数）
func NewSaveLoadButtonsForConfig(fileName string, onSave, onLoad func(string)) *SaveLoadButtons {
	return NewSaveLoadButtons(".conf", fileName, onSave, onLoad)
}

// NewSaveLoadButtonsForText 创建用于文本文件的按钮（便捷函数）
func NewSaveLoadButtonsForText(fileName string, onSave, onLoad func(string)) *SaveLoadButtons {
	return NewSaveLoadButtons(".txt", fileName, onSave, onLoad)
}
