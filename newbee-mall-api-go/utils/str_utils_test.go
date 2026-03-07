package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubStrLen(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		length   int
		expected string
	}{
		{
			name:     "字符串长度小于限制",
			str:      "Hello",
			length:   10,
			expected: "Hello",
		},
		{
			name:     "字符串长度等于限制",
			str:      "Hello",
			length:   5,
			expected: "Hello",
		},
		{
			name:     "字符串长度大于限制",
			str:      "Hello World",
			length:   5,
			expected: "Hello...",
		},
		{
			name:     "中文字符串截取",
			str:      "你好世界欢迎你",
			length:   4,
			expected: "你好世界...",
		},
		{
			name:     "空字符串",
			str:      "",
			length:   5,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SubStrLen(tt.str, tt.length)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNumsInList(t *testing.T) {
	tests := []struct {
		name     string
		num      int
		nums     []int
		expected bool
	}{
		{
			name:     "数字存在于列表中",
			num:      3,
			nums:     []int{1, 2, 3, 4, 5},
			expected: true,
		},
		{
			name:     "数字不存在于列表中",
			num:      6,
			nums:     []int{1, 2, 3, 4, 5},
			expected: false,
		},
		{
			name:     "空列表",
			num:      1,
			nums:     []int{},
			expected: false,
		},
		{
			name:     "列表只有一个元素且匹配",
			num:      5,
			nums:     []int{5},
			expected: true,
		},
		{
			name:     "列表只有一个元素且不匹配",
			num:      3,
			nums:     []int{5},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NumsInList(tt.num, tt.nums)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenValidateCode(t *testing.T) {
	tests := []struct {
		name  string
		width int
	}{
		{
			name:  "生成6位验证码",
			width: 6,
		},
		{
			name:  "生成4位验证码",
			width: 4,
		},
		{
			name:  "生成8位验证码",
			width: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := GenValidateCode(tt.width)
			assert.Len(t, code, tt.width)
			// 验证码应该只包含数字
			for _, c := range code {
				assert.True(t, c >= '0' && c <= '9')
			}
		})
	}
}

func TestGenOrderNo(t *testing.T) {
	t.Run("生成订单号", func(t *testing.T) {
		orderNo1 := GenOrderNo()
		orderNo2 := GenOrderNo()

		// 订单号应该是时间戳(13位) + 4位随机数 = 17位
		assert.True(t, len(orderNo1) >= 17)
		assert.True(t, len(orderNo2) >= 17)

		// 两次生成的订单号应该不同
		assert.NotEqual(t, orderNo1, orderNo2)

		// 订单号应该只包含数字
		for _, c := range orderNo1 {
			assert.True(t, c >= '0' && c <= '9')
		}
	})
}

func TestStrToInt(t *testing.T) {
	tests := []struct {
		name     string
		strNum   string
		expected []int
	}{
		{
			name:     "单个数字",
			strNum:   "5",
			expected: []int{5},
		},
		{
			name:     "多个数字",
			strNum:   "1,2,3,4,5",
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "空字符串",
			strNum:   "",
			expected: []int{0},
		},
		{
			name:     "大数字",
			strNum:   "100,200,300",
			expected: []int{100, 200, 300},
		},
		{
			name:     "包含零的数字",
			strNum:   "0,1,0,2",
			expected: []int{0, 1, 0, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StrToInt(tt.strNum)
			assert.Equal(t, tt.expected, result)
		})
	}
}
