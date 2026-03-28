package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCamelCase(t *testing.T) {
	funcs := buildFuncMap()
	fn := funcs["camelCase"].(func(string) string)

	cases := []struct{ in, want string }{
		{"my_field", "myField"},
		{"my-field", "myField"},
		{"MyField", "myField"},
		{"user_id", "userId"},
		{"hello", "hello"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.in, func(t *testing.T) {
			assert.Equal(t, tc.want, fn(tc.in), "input: %q", tc.in)
		})
	}
}

func TestPascalCase(t *testing.T) {
	funcs := buildFuncMap()
	fn := funcs["pascalCase"].(func(string) string)

	cases := []struct{ in, want string }{
		{"my_field", "MyField"},
		{"my-field", "MyField"},
		{"myField", "MyField"},
		{"user_id", "UserId"},
		{"hello", "Hello"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.in, func(t *testing.T) {
			assert.Equal(t, tc.want, fn(tc.in), "input: %q", tc.in)
		})
	}
}

func TestSnakeCase(t *testing.T) {
	funcs := buildFuncMap()
	fn := funcs["snakeCase"].(func(string) string)

	cases := []struct{ in, want string }{
		{"MyField", "my_field"},
		{"myField", "my_field"},
		{"my_field", "my_field"},
		{"UserId", "user_id"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.in, func(t *testing.T) {
			assert.Equal(t, tc.want, fn(tc.in), "input: %q", tc.in)
		})
	}
}

func TestKebabCase(t *testing.T) {
	funcs := buildFuncMap()
	fn := funcs["kebabCase"].(func(string) string)

	cases := []struct{ in, want string }{
		{"MyField", "my-field"},
		{"my_field", "my-field"},
		{"myField", "my-field"},
		{"UserId", "user-id"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.in, func(t *testing.T) {
			assert.Equal(t, tc.want, fn(tc.in), "input: %q", tc.in)
		})
	}
}

func TestHasModule(t *testing.T) {
	funcs := buildFuncMap()
	fn := funcs["hasModule"].(func([]string, string) bool)

	modules := []string{"logger", "database", "api"}

	assert.True(t, fn(modules, "logger"))
	assert.True(t, fn(modules, "database"))
	assert.True(t, fn(modules, "api"))
	assert.False(t, fn(modules, "metrics"))
	assert.False(t, fn(modules, ""))
	assert.False(t, fn(nil, "logger"))
}
