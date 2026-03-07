package utils

import (
	"testing"
)

func TestMD5V(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "空字符串",
			input:    []byte(""),
			expected: "d41d8cd98f00b204e9800998ecf8427e",
		},
		{
			name:     "简单密码",
			input:    []byte("123456"),
			expected: "e10adc3949ba59abbe56e057f20f883e",
		},
		{
			name:     "复杂字符串",
			input:    []byte("Hello, World!"),
			expected: "65a8e27d8879283831b664bd8b7f0ad4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MD5V(tt.input)
			if result != tt.expected {
				t.Errorf("MD5V() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func BenchmarkMD5V(b *testing.B) {
	input := []byte("test password 123456")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MD5V(input)
	}
}
