# fyne-l

[![Go Reference](https://pkg.go.dev/badge/github.com/zc310/fyne-l/widget.svg)](https://pkg.go.dev/github.com/zc310/fyne-l/widget)

基于 [Fyne](https://fyne.io/) 的自用 Go GUI 组件库，收集常用的输入框、范围设置、文件与目录选择、系统监控和帮助组件，并提供打包好的 Font Awesome 图标资源。

## 安装

当前模块要求 **Go 1.26.0 或更高版本**，使用 **Fyne v2.8.3**。

```sh
go get github.com/zc310/fyne-l
```

- 构建桌面应用前，请按 [Fyne 环境准备指南](https://docs.fyne.io/started/) 安装对应平台的编译工具和图形开发依赖。
- 文件与目录对话框使用 [`github.com/ncruces/zenity`](https://github.com/ncruces/zenity)。Linux 等 Unix 桌面环境通常需要安装 `zenity`；其他平台要求请参考其说明。
- 组件中包含中文文案，应用所用字体需支持中文。

## 包与组件

| 包 | 内容 |
| --- | --- |
| [`widget`](https://pkg.go.dev/github.com/zc310/fyne-l/widget) | 扩展组件与辅助函数 |
| [`icons/fontawesome`](https://pkg.go.dev/github.com/zc310/fyne-l/icons/fontawesome) | Solid、Regular、Brands SVG 图标资源 |

`widget` 包中的主要组件：

| 组件 | 用途 |
| --- | --- |
| `MinWidthEntry` | 可设置最小宽度的文本输入框 |
| `IntEntry` | 整数校验、默认值、范围限制及递增/递减；提供百分比、端口号等便捷构造函数 |
| `IntRange` / `DoubleRange` | 整数 / 浮点数范围输入、变化回调与范围有效性检查 |
| `CheckGroup` | 根据标签生成一组复选框，以 `[]bool` 读取和更新状态 |
| `FileSelector` | 单文件、多文件或目录选择，支持过滤器与已选列表管理 |
| `DirectorySelector` | 单目录 / 多目录选择，支持路径详情、复制和列表管理 |
| `SimpleFileDirSelector` | 切换文件 / 目录模式，支持递归检索目录中的文件 |
| `SaveLoadButtons` | 保存 / 导入按钮组合，支持扩展名和文件类型过滤 |
| `CPUSettingWidget` | 1–100 的 CPU 使用比例参数输入，可附带系统监控 |
| `SysInfoWidget` | 每 3 秒更新 CPU 使用率、可用 / 总内存及内存使用比例 |
| `MarkdownHelpWidget` | 点击按钮打开 Markdown 帮助对话框 |

此外还提供 `ShowCopyableDialog`、`RunPopUp`、`LastFile` 和可替换的缓存接口。

## 快速开始

在自己的应用目录中将以下内容保存为 `main.go`，安装依赖后运行 `go run .`：

```go
package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/zc310/fyne-l/icons/fontawesome"
	lwidget "github.com/zc310/fyne-l/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("fyne-l 示例")

	valueRange := lwidget.NewIntRange("数量", 1, 100, func(min, max int) {
		log.Printf("范围：%d ~ %d", min, max)
	}, false)
	valueRange.SetEntryMinWidth(90)

	files := lwidget.NewFileSelector(w, func(paths []string) {
		log.Println("已选文件：", paths)
	})
	files.SetFilters([]lwidget.FileFilter{
		{Name: "文本文件", Patterns: []string{"*.txt", "*.csv"}},
		{Name: "所有文件", Patterns: []string{"*"}},
	})

	cpu := lwidget.NewCPUSettingWidget(50, true)
	w.SetOnClosed(cpu.Stop)

	help := lwidget.NewMarkdownHelpWithText(
		w, "使用说明",
		"# 示例\n\n设置数量范围、选择文件，并查看系统状态。",
		"帮助",
	)
	save := widget.NewButtonWithIcon(
		"读取范围", theme.NewThemedResource(fontawesome.Save),
		func() {
			if !valueRange.Validate() {
				lwidget.ShowCopyableDialog(w, "提示", "最小值不能大于最大值")
				return
			}
			min, max := valueRange.Range()
			log.Printf("当前范围：%d ~ %d", min, max)
		},
	)

	w.SetContent(container.NewVBox(valueRange, files, cpu, help, save))
	w.Resize(fyne.NewSize(640, 360))
	w.ShowAndRun()
}
```

使用 `lwidget` 作为导入别名，可以避免与 Fyne 自带的 `widget` 包重名。

## 常见用法

以下片段沿用上例的 `lwidget` 别名和窗口变量 `w`，可按需加入应用。

### 整数与范围输入

```go
entry := lwidget.NewIntEntryWithDefault(1, 100, 10)
entry.SetMinWidth(100)
value, err := entry.GetInt() // 返回校验后的值；空输入使用默认值
if err != nil {
	log.Println(err)
} else {
	log.Println(value)
}

ratio := lwidget.NewDoubleRange("比例", 0.1, 0.9, func(min, max float64) {
	log.Println(min, max)
})
ratio.SetEntryWidth(90)
ratio.SetPrecision(2) // 格式化当前显示值
w.SetContent(container.NewVBox(entry, ratio))
```

- `NewIntEntry(0, 0)` 表示不限制上下界；`max == 0` 表示不设上界。需要非负整数、百分比或端口号时，可直接使用 `NewNonNegativeIntEntry()`、`NewPercentageEntry()`、`NewPortEntry()`。
- `IntRange` 构造函数最后一个参数为 `true` 时，标签显示在两个输入框之间，形式为「最小值 ≤ 标签 ≤ 最大值」。
- 范围组件不会自动调整大小关系；使用 `Validate()` 检查是否满足 `min <= max`。`SetRange()` / `SetValue()` 会触发变化回调。

### 文件 / 目录切换与检索

```go
selector := lwidget.NewSimpleFileDirSelector(w, "input")
selector.SetFileFilters([]lwidget.FileFilter{
	{Name: "文本文件", Patterns: []string{"*.txt"}},
})
selector.SetFileMode(false) // 目录模式；默认为文件模式
selector.SetRecursive(true)
selector.OnChanged = func(selection lwidget.FileDirSelection) {
	if selection.IsEmpty() {
		return
	}
	paths := selection.GetFiles("*.txt", "*.csv")
	log.Println(paths)
}
w.SetContent(selector.Container)
```

`GetFiles(patterns...)` 在目录模式下按文件名通配符过滤，并根据 `Recursive` 决定是否检索子目录；在文件模式下直接返回已选文件，不再应用这些过滤规则。无法访问的目录或文件会被跳过，目录扫描是同步执行的，大目录建议放到后台处理。

### 保存与导入

```go
buttons := lwidget.NewSaveLoadButtonsForJSON(
	"config.json",
	func(path string) {
		log.Println("保存到：", path) // 在这里实现写入逻辑
	},
	func(path string) {
		log.Println("导入自：", path) // 在这里实现读取逻辑
	},
)
w.SetContent(buttons.GetContainer())
```

按钮只负责选择路径并调用回调，**不会自动读写文件**。保存时若文件名没有扩展名，会补上配置的默认扩展名。也可使用 `NewSaveLoadButtonsWithFileTypes` 自定义过滤规则。

### 图标资源

图标以 `fyne.Resource` 形式内嵌，无需额外分发 SVG 文件。单色图标可通过 `theme.NewThemedResource` 适配当前主题：

```go
button := widget.NewButtonWithIcon(
	"打开", theme.NewThemedResource(fontawesome.FolderOpenSolid),
	func() { log.Println("打开") },
)
w.SetContent(button)
```

具体资源名称请查阅 [`icons/fontawesome`](icons/fontawesome) 中的导出变量。

## 使用注意

- `SysInfoWidget` 创建后会启动后台采样；不再使用时调用 `Stop()`。`CPUSettingWidget` 开启系统监控后也应调用 `Stop()`，如快速开始中绑定窗口关闭事件的写法。
- `CPUSettingWidget` 只收集比例参数，不会限制进程 CPU 使用率；实际调度逻辑由应用实现。
- `LastFile` 默认使用内存缓存，仅在当前进程内记忆路径。需要跨启动保存时，实现 `Cache` 接口并通过 `SetCache()` 注入持久化存储。
- `SimpleFileDirSelector` 通过 `.Container` 加入布局，`SaveLoadButtons` 通过 `.GetContainer()` 加入布局；其他主要组件可直接作为 Fyne 控件使用。
- 从后台 goroutine 更新界面时，使用 `fyne.Do` 切回 UI 线程。

## 测试

在仓库根目录运行：

```sh
go test ./...
```

## 许可证

项目代码使用 [Apache License 2.0](LICENSE)。Font Awesome 图标的使用和分发还需遵守其[上游许可证](https://fontawesome.com/license/free)及商标要求。
