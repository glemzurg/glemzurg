package identity

func (suite *KeySuite) TestNormalizeSubKey() {
	tests := []struct {
		testName string
		input    string
		expected string
	}{
		{testName: "simple", input: "hello", expected: "hello"},
		{testName: "spaces to underscores", input: "hello world", expected: "hello_world"},
		{testName: "hyphens to underscores", input: "hello-world", expected: "hello_world"},
		{testName: "trim and lower", input: " Hello World ", expected: "hello_world"},
		{testName: "mixed spaces and hyphens", input: "Some-Name Here", expected: "some_name_here"},
		{testName: "already valid", input: "my_key_123", expected: "my_key_123"},
		{testName: "uppercase", input: "MyKey", expected: "mykey"},
	}

	for _, tt := range tests {
		suite.Run(tt.testName, func() {
			result := NormalizeSubKey(tt.input)
			suite.Equal(tt.expected, result)
		})
	}
}
