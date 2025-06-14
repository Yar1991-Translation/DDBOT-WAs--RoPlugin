package utils

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input            string
		expectedFilename string
		expectError      bool
	}{
		{
			input:            `attachment; filename="cool.html"`,
			expectedFilename: "cool.html",
			expectError:      false,
		},
		{
			input:       `inline`,
			expectError: true,
		},
		{
			input:            `form-data; name="fieldName"; filename="example.txt"`,
			expectedFilename: "example.txt",
			expectError:      false,
		},
		{
			input:            `attachment; filename="example.txt"; filename*=utf-8''example2.txt`,
			expectedFilename: "example2.txt",
			expectError:      false,
		},
		{
			input:            `attachment; filename=example.txt`,
			expectedFilename: "example.txt",
			expectError:      false,
		},
		{
			input:            `attachment; filename*=utf-8''example2.txt`,
			expectedFilename: "example2.txt",
			expectError:      false,
		},
		{
			input:            `attachment; filename*=utf-8''example2.txt`,
			expectedFilename: "example2.txt",
			expectError:      false,
		},
		{
			input:            `attachment; filename="[]  .pdf"; filename*=UTF-8''%E9%A5%B1%E9%A5%B1.pdf`,
			expectedFilename: "饱饱.pdf",
			expectError:      false,
		},
	}

	for _, test := range tests {
		result, err := ParseDisposition(test.input)
		if test.expectError {
			if err == nil {
				t.Errorf("预期输入 %q 返回错误，但实际没有错误", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("输入 %q 不应返回错误，但得到错误：%v", test.input, err)
			}
			if result != test.expectedFilename {
				t.Errorf("对于输入 %q，预期文件名 %q，但实际得到 %q", test.input, test.expectedFilename, result)
			}
		}
	}
}
