package logs

import "testing"

var (
	gitURLs = []string{
		"git://github.com/go-logs/logs",
		"https://github.com/go-logs/logs.git",
		"http://gitlab.com/go-logs/logs.git",
		"http://gitlab.com/go-logs/logs.git#branch",
		"http://gitlab.com/go-logs/logs.git#:dir",
	}
	invalidGitURLs = []string{
		"http://github.com/go-logs/logs.git:#branch",
	}
	socketURLs = []string{
		"tcp://example.com",
		"tcp+tls://example.com",
		"udp://example.com",
		"unix:///example",
		"unixgram:///example",
	}
)

func TestIsSocket(t *testing.T) {
	for _, url := range socketURLs {
		if !IsSocket(url) {
			t.Fatalf("%q should be detected as valid Transport url", url)
		}
	}
}
