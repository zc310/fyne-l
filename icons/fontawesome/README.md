# Font Awesome 图标

这个目录提供打包到 Fyne `fyne.Resource` 中的 Font Awesome 免费图标资源。

## 文件说明

- `solid.go`：Solid 图标
- `regular.go`：Regular 图标
- `brands.go`：Brands 品牌图标

这些 Go 文件由 `fyne bundle` 生成，请不要直接修改。需要增加或更新图标时，应更新对应的 SVG 资源后重新生成。

## 使用方式

导入本包，并将图标资源传给 Fyne 控件。需要跟随当前主题变化时，使用 `theme.NewThemedResource` 包装资源：

```go
import (
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"your/module/icons/fontawesome"
)

button := widget.NewButtonWithIcon(
	"保存",
	theme.NewThemedResource(fontawesome.Save),
	func() {},
)
```

图标名称通常对应 Font Awesome 的名称，例如 `Save`、`FolderOpen`。数字或其他不适合作为 Go 标识符的名称会使用生成器定义的名称，例如 `F0Solid`。
