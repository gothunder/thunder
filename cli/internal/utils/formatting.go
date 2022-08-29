package utils

import (
	"regexp"
	"strings"
)

func NormalizeToKebabOrSnakeCase(s string) string {
	matchFirstCap := regexp.MustCompile("([A-Z])([A-Z][a-z])")
	matchAllCap := regexp.MustCompile("([a-z0-9])([A-Z])")

	output := matchFirstCap.ReplaceAllString(s, "${1}-${2}")
	output = matchAllCap.ReplaceAllString(output, "${1}-${2}")

	dasherRegexp := regexp.MustCompile("\\s")

	output = dasherRegexp.ReplaceAllString(output, "-")
	return strings.ToLower(output)
}
