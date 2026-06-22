package ui

import (
	"regexp"
	"strconv"
)

var prURLRe = regexp.MustCompile(`github\.com/([^/\s]+/[^/\s]+)/pull/(\d+)`)

// ParsePRURL extracts "owner/repo" and the number from a GitHub PR URL.
func ParsePRURL(url string) (repo string, number int, ok bool) {
	m := prURLRe.FindStringSubmatch(url)
	if m == nil {
		return "", 0, false
	}
	n, _ := strconv.Atoi(m[2])
	return m[1], n, true
}
