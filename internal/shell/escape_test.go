package shell

import (
	"testing"
)

func TestSingleQuoteEscape(t *testing.T) {
	cases := []struct {
		input, want string
	}{
		{"hello", "hello"},
		{"", ""},
		{"it's here", `it'\''s here`},
		{"a'b'c", `a'\''b'\''c`},
		{"no quotes", "no quotes"},
		{"$VAR", "$VAR"},
		{`back\slash`, `back\slash`},
	}
	for _, c := range cases {
		got := SingleQuoteEscape(c.input)
		if got != c.want {
			t.Errorf("SingleQuoteEscape(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestFormatExport(t *testing.T) {
	got := FormatExport("FOO", "bar baz")
	want := "export FOO='bar baz'"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	got2 := FormatExport("KEY", "it's a value")
	want2 := `export KEY='it'\''s a value'`
	if got2 != want2 {
		t.Errorf("got %q, want %q", got2, want2)
	}
}
