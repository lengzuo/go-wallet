package middlewares

import "testing"

func Test_getAPIName(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{name: "Trailing slash", args: "https://example.com/api/a/b/c/d/", want: "d"},
		{name: "No trailing slash", args: "https://example.com/api/a/b/c", want: "c"},
		{name: "Single component after domain", args: "https://example.com/api", want: "api"},
		{name: "Path only with trailing slash", args: "/some/path/component/", want: "component"},
		{name: "Path only no trailing slash", args: "/some/path/component", want: "component"},
		{name: "Root path only", args: "https://example.com/", want: ""},
		{name: "Domain only", args: "https://example.com", want: ""},
		{name: "No slashes in path part", args: "filename.txt", want: "filename.txt"},
		{name: "Path is just a filename", args: "/filename.txt", want: "filename.txt"},
		{name: "Empty string", args: "", want: ""},
		{name: "Only slashes", args: "///", want: ""},
		{name: "Multiple slashes between components", args: "http://host/a//b///c", want: "c"},
		{name: "URL with query", args: "https://example.com/api/a/b/item?id=123", want: "item"},
		{name: "URL with fragment", args: "https://example.com/api/a/b/section#details", want: "section"},
		{name: "Single leading slash and component", args: "/rootfile", want: "rootfile"},
		{name: "Only component, no domain (parsed as path)", args: "justcomponent", want: "justcomponent"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getAPIName(tt.args); got != tt.want {
				t.Errorf("getAPIName() = %v, want %v", got, tt.want)
			}
		})
	}
}
