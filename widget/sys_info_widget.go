package widget

import (
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

// SysInfoWidget 显示系统信息的组件
type SysInfoWidget struct {
	widget.BaseWidget
	container *fyne.Container
	stop      chan struct{}
	stopOnce  sync.Once
}

// NewSysInfoWidget 创建系统信息组件
func NewSysInfoWidget() *SysInfoWidget {
	s := &SysInfoWidget{stop: make(chan struct{})}
	lblCpu := widget.NewLabel("...%")
	lblRam := widget.NewLabel("...")
	row := container.NewHBox(
		widget.NewLabel(" CPU: "),
		lblCpu,
		widget.NewSeparator(),
		widget.NewLabel(" RAM: "),
		lblRam,
	)

	s.container = container.NewVBox(
		widget.NewLabel("系统监控"),
		row,
	)

	s.ExtendBaseWidget(s)
	go s.startUpdating(lblCpu, lblRam)

	return s
}

func (s *SysInfoWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(s.container)
}

// Stop 停止后台系统信息更新。
// 重复调用 Stop 是安全的。
func (s *SysInfoWidget) Stop() {
	if s.stop == nil {
		return
	}
	s.stopOnce.Do(func() { close(s.stop) })
}

func (s *SysInfoWidget) startUpdating(infoCpu, infoRam *widget.Label) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
		}

		cpuPercent, err := cpu.Percent(0, false)
		if err != nil {
			fyne.Do(func() {
				infoCpu.SetText(fmt.Sprintf("❌ CPU 使用率获取失败: %v", err))
			})
			continue
		}
		if len(cpuPercent) == 0 {
			fyne.Do(func() {
				infoCpu.SetText("CPU 使用率暂无数据")
			})
			continue
		}

		memInfo, err := mem.VirtualMemory()
		if err != nil {
			fyne.Do(func() {
				infoRam.SetText(fmt.Sprintf("❌ 内存信息获取失败: %v", err))
			})
			continue
		}
		if memInfo.Total == 0 {
			fyne.Do(func() {
				infoRam.SetText("内存总量暂无数据")
			})
			continue
		}

		cpuVal := cpuPercent[0]
		freeMemGB := float64(memInfo.Available) / (1024 * 1024 * 1024)
		totalMemGB := float64(memInfo.Total) / (1024 * 1024 * 1024)
		usedPercent := 100.0 - float64(memInfo.Available)/float64(memInfo.Total)*100.0

		textCpu := fmt.Sprintf("%.1f%%", cpuVal)
		textRam := fmt.Sprintf("%.2f GB / %.2f GB (%.1f%%)", freeMemGB, totalMemGB, usedPercent)

		fyne.Do(func() {
			infoCpu.SetText(textCpu)
			infoRam.SetText(textRam)
		})
	}
}
