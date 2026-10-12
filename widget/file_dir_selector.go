package widget

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// FileDirSelection 文件/目录选择的状态
type FileDirSelection struct {
	Files      []string
	Folders    []string
	IsFileMode bool // true=选文件, false=选目录
	Recursive  bool // 选目录时是否递归检索子目录
}

// GetFiles 获取文件列表（包含子目录）
// 如果是文件模式，返回 Files
// 如果是目录模式，递归遍历目录及其子目录下的文件，支持通配符过滤，如 GetFiles("*.txt", "*.zcp")
func (s *FileDirSelection) GetFiles(patterns ...string) []string {
	if s.IsFileMode {
		return cloneStrings(s.Files)
	}

	var result []string
	seen := make(map[string]bool) // 去重

	for _, folder := range s.Folders {
		if s.Recursive {
			// 递归遍历目录
			filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil // 跳过无法访问的目录
				}
				if info.IsDir() {
					return nil // 继续遍历子目录
				}

				// 检查是否已经添加过
				if seen[path] {
					return nil
				}

				fileName := info.Name()

				// 如果没有指定过滤规则，全部加入
				if len(patterns) == 0 {
					seen[path] = true
					result = append(result, path)
					return nil
				}

				// 根据过滤规则匹配
				for _, pattern := range patterns {
					matched, err := filepath.Match(pattern, fileName)
					if err == nil && matched {
						seen[path] = true
						result = append(result, path)
						break
					}
				}
				return nil
			})
		} else {
			// 仅当前目录
			entries, err := os.ReadDir(folder)
			if err != nil {
				continue
			}
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				fileName := entry.Name()
				fullPath := filepath.Join(folder, fileName)

				// 如果没有指定过滤规则，全部加入
				if len(patterns) == 0 {
					result = append(result, fullPath)
					continue
				}

				// 根据过滤规则匹配
				for _, pattern := range patterns {
					matched, err := filepath.Match(pattern, fileName)
					if err == nil && matched {
						result = append(result, fullPath)
						break
					}
				}
			}
		}
	}
	return result
}

// IsEmpty 判断是否未选择任何内容
func (s *FileDirSelection) IsEmpty() bool {
	if s.IsFileMode {
		return len(s.Files) == 0
	}
	return len(s.Folders) == 0
}

// SimpleFileDirSelector 简化的选文件和选目录组件（不带多目录合并选项）
type SimpleFileDirSelector struct {
	W                 fyne.Window
	fileSelector      *FileSelector
	directorySelector *DirectorySelector
	selectorType      *widget.RadioGroup
	ckRecursive       *widget.Check // 检索子目录复选框

	Selection FileDirSelection

	// OnChanged 统一的回调函数，选择发生变化时触发
	OnChanged func(selection FileDirSelection)

	// Container 外部可访问的容器
	Container *fyne.Container
	key       string
}

// NewSimpleFileDirSelector 创建新的简化版文件/目录选择器组件
func NewSimpleFileDirSelector(w fyne.Window, key string) *SimpleFileDirSelector {
	fds := &SimpleFileDirSelector{W: w, key: key}
	fds.initComponents()
	fds.setupLayout()
	fds.setupEventHandlers()
	return fds
}

func (fds *SimpleFileDirSelector) initComponents() {
	// 文件选择器
	fds.fileSelector = NewFileSelector(fds.W, func(s []string) {
		fds.Selection.Files = s
		if fds.OnChanged != nil {
			fds.OnChanged(fds.Selection)
		}
	})
	fds.fileSelector.LastFile = NewLastFile("simple_file_dir_selector_file"+fds.key, "")

	// 目录选择器
	fds.directorySelector = NewMultiDirectorySelector(fds.W, func(s []string) {
		fds.Selection.Folders = s
		if fds.OnChanged != nil {
			fds.OnChanged(fds.Selection)
		}
	})
	fds.directorySelector.LastFile = NewLastFile("simple_file_dir_selector_dir"+fds.key, "")

	// 检索子目录复选框
	fds.ckRecursive = widget.NewCheck("检索子目录", func(b bool) {
		fds.Selection.Recursive = b
		if fds.OnChanged != nil {
			fds.OnChanged(fds.Selection)
		}
	})
	fds.ckRecursive.SetChecked(true) // 默认勾选

	// 选择类型单选组
	fds.selectorType = widget.NewRadioGroup([]string{"选文件", "选目录"}, func(s string) {
		isFileMode := s == "选文件"
		fds.Selection.IsFileMode = isFileMode
		fds.updateVisibility()
		if fds.OnChanged != nil {
			fds.OnChanged(fds.Selection)
		}
	})
	fds.selectorType.Required = true
	fds.selectorType.Horizontal = true
	fds.selectorType.SetSelected("选文件")
	fds.Selection.IsFileMode = true
}

func (fds *SimpleFileDirSelector) setupLayout() {
	fds.Container = container.NewVBox(
		container.NewBorder(nil, nil, fds.selectorType, fds.ckRecursive),
		widget.NewSeparator(),
		fds.fileSelector,
		fds.directorySelector,
	)
}

func (fds *SimpleFileDirSelector) setupEventHandlers() {
	fds.updateVisibility()
}

func (fds *SimpleFileDirSelector) updateVisibility() {
	if fds.Selection.IsFileMode {
		fds.fileSelector.Show()
		fds.directorySelector.Hide()
		fds.ckRecursive.Hide()
	} else {
		fds.fileSelector.Hide()
		fds.directorySelector.Show()
		fds.ckRecursive.Show()
	}
}

// SetFiles 设置已选择的文件列表
func (fds *SimpleFileDirSelector) SetFiles(files []string) {
	fds.Selection.Files = cloneStrings(files)
	fds.fileSelector.SetFiles(files)
}

// SetFolders 设置已选择的目录列表
func (fds *SimpleFileDirSelector) SetFolders(folders []string) {
	fds.Selection.Folders = cloneStrings(folders)
	fds.directorySelector.SetSelectedPaths(folders)
}

// SetFileMode 设置选择模式，true=选文件, false=选目录
func (fds *SimpleFileDirSelector) SetFileMode(isFileMode bool) {
	fds.Selection.IsFileMode = isFileMode
	if isFileMode {
		fds.selectorType.SetSelected("选文件")
	} else {
		fds.selectorType.SetSelected("选目录")
	}
	fds.updateVisibility()
}

// SetRecursive 设置是否递归检索子目录
func (fds *SimpleFileDirSelector) SetRecursive(recursive bool) {
	fds.Selection.Recursive = recursive
	fds.ckRecursive.SetChecked(recursive)
}

// SetFileFilters 设置文件过滤器（可选调用）
func (fds *SimpleFileDirSelector) SetFileFilters(filters []FileFilter) {
	fds.fileSelector.SetFilters(filters)
}

// SetDefault 一次性设置所有默认值
func (fds *SimpleFileDirSelector) SetDefault(selection FileDirSelection) {
	fds.SetFiles(selection.Files)
	fds.SetFolders(selection.Folders)
	fds.SetFileMode(selection.IsFileMode)
	fds.SetRecursive(selection.Recursive)
}
