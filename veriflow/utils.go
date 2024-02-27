package veriflow

import "regexp"

// Simple utility function extract user ID for Slack. Format is <@user|name>
func extractUserID(text string) string {
	re := regexp.MustCompile(`<@([^>|]+)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
