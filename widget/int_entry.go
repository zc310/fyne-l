package widget

import (
	"fmt"
	"strconv"
)

// IntEntry 整数输入框组件
type IntEntry struct {
	MinWidthEntry
	min, max     int
	hasMin       bool
	hasMax       bool
	defaultValue *int // 默认值
}

// NewIntEntry 创建新的整数输入框
func NewIntEntry(min, max int) *IntEntry {
	// 参数验证和自动校正
	if min > max && max != 0 {
		min, max = max, min
	}

	return newIntEntry(min, max, min != 0 || max != 0, max != 0)
}

func newIntEntry(min, max int, hasMin, hasMax bool) *IntEntry {
	entry := &IntEntry{min: min, max: max, hasMin: hasMin, hasMax: hasMax}
	entry.ExtendBaseWidget(entry)
	entry.minWidth = 72
	entry.Validator = entry.validate
	entry.PlaceHolder = entry.getDefaultPlaceholder()
	return entry
}
func NewIntEntryWithNoValidator(min, max, defaultValue int) *IntEntry {
	entry := NewIntEntryWithDefault(min, max, defaultValue)
	entry.Validator = nil
	return entry
}

// NewIntEntryWithDefault 创建带默认值的整数输入框
func NewIntEntryWithDefault(min, max, defaultValue int) *IntEntry {
	entry := NewIntEntry(min, max)
	entry.defaultValue = &defaultValue
	entry.SetText(strconv.Itoa(defaultValue))
	return entry
}

// getDefaultPlaceholder 获取默认占位符文本
func (e *IntEntry) getDefaultPlaceholder() string {
	if e.hasMin && e.hasMax && e.min == e.max {
		return fmt.Sprintf("%d", e.min)
	}

	if !e.hasMin && !e.hasMax {
		return "请输入整数"
	}

	if e.hasMin && e.hasMax {
		return fmt.Sprintf("%d - %d", e.min, e.max)
	}

	if e.hasMin {
		return fmt.Sprintf("≥ %d", e.min)
	}

	return fmt.Sprintf("≤ %d", e.max)
}

// validate 验证输入
func (e *IntEntry) validate(s string) error {
	if s == "" {
		if e.defaultValue != nil {
			return nil // 允许为空，因为有默认值
		}
		return fmt.Errorf("请输入一个整数值")
	}

	val, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("请输入有效的整数")
	}

	// 范围验证
	if e.hasMin || e.hasMax {
		if e.hasMin && val < e.min {
			return fmt.Errorf("数值不能小于 %d", e.min)
		}
		if e.hasMax && val > e.max {
			return fmt.Errorf("数值不能大于 %d", e.max)
		}
	}

	return nil
}

// GetInt 获取整数值
func (e *IntEntry) GetInt() (int, error) {
	if err := e.Validate(); err != nil {
		return 0, err
	}

	if e.Text == "" {
		if e.defaultValue != nil {
			return *e.defaultValue, nil
		}
		return 0, nil
	}

	return strconv.Atoi(e.Text)
}

// MustGetInt 获取整数值，如果无效则返回默认值或0
func (e *IntEntry) MustGetInt() int {
	val, err := e.GetInt()
	if err != nil {
		if e.defaultValue != nil {
			return *e.defaultValue
		}
		return 0
	}
	return val
}

// SetInt 设置整数值
func (e *IntEntry) SetInt(value int) error {
	// 验证范围
	if e.hasMin && value < e.min {
		return fmt.Errorf("数值 %d 小于最小值 %d", value, e.min)
	}
	if e.hasMax && value > e.max {
		return fmt.Errorf("数值 %d 大于最大值 %d", value, e.max)
	}

	e.SetText(strconv.Itoa(value))
	return nil
}

// SetValueRange 设置数值范围
func (e *IntEntry) SetValueRange(min, max int) {
	// 参数验证
	if min > max && max != 0 {
		min, max = max, min
	}

	e.min = min
	e.max = max
	e.hasMin = min != 0 || max != 0
	e.hasMax = max != 0
	e.PlaceHolder = e.getDefaultPlaceholder()
	e.Refresh()
}

// GetValueRange 获取数值范围
func (e *IntEntry) GetValueRange() (int, int) {
	return e.min, e.max
}

// SetDefaultValue 设置默认值
func (e *IntEntry) SetDefaultValue(value int) {
	e.defaultValue = &value
	e.SetText(strconv.Itoa(value))
}

// ClearDefaultValue 清除默认值
func (e *IntEntry) ClearDefaultValue() {
	e.defaultValue = nil
}

// Clear 清空输入框，如果有默认值则设置为默认值
func (e *IntEntry) Clear() {
	if e.defaultValue != nil {
		e.SetText(strconv.Itoa(*e.defaultValue))
	} else {
		e.SetText("")
	}
}

// Increment 增加指定值
func (e *IntEntry) Increment(step int) {
	current, err := e.GetInt()
	if err != nil {
		if e.defaultValue != nil {
			current = *e.defaultValue
		} else {
			current = 0
		}
	}

	newValue := current + step

	// 边界检查
	if e.hasMin && newValue < e.min {
		newValue = e.min
	}
	if e.hasMax && newValue > e.max {
		newValue = e.max
	}

	e.SetText(strconv.Itoa(newValue))
}

// Decrement 减少指定值
func (e *IntEntry) Decrement(step int) {
	e.Increment(-step)
}

// BindInt 绑定整数值到变量
func (e *IntEntry) BindInt(target *int) {
	e.OnChanged = func(s string) {
		if val, err := e.GetInt(); err == nil {
			*target = val
		}
	}
}

// SetMinWidth 设置最小宽度
func (e *IntEntry) SetMinWidth(width float32) {
	e.minWidth = width
	e.Refresh()
}

// NewPositiveIntEntry 创建正整数输入框
func NewPositiveIntEntry() *IntEntry {
	return NewIntEntry(1, 0)
}

// NewNonNegativeIntEntry 创建非负整数输入框
func NewNonNegativeIntEntry() *IntEntry {
	return newIntEntry(0, 0, true, false)
}

// NewBoundedIntEntry 创建有界整数输入框
func NewBoundedIntEntry(min, max int) *IntEntry {
	return NewIntEntry(min, max)
}

// NewPercentageEntry 创建百分比输入框 (0-100)
func NewPercentageEntry() *IntEntry {
	return newIntEntry(0, 100, true, true)
}

// NewPortEntry 创建端口号输入框 (1-65535)
func NewPortEntry() *IntEntry {
	return NewIntEntry(1, 65535)
}
