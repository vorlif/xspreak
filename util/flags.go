package util

import (
	"fmt"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	flagPrefix          = "xspreak:"
	templateMarkerLong  = "template"
	templateMarkerShort = "tmpl"
)

var reRange = regexp.MustCompile(`^range:\s+\d+\.\.\d+\s*$`)

func IsInlineTemplate(comment string) bool {
	for _, line := range strings.Split(comment, "\n") {
		line = strings.ToLower(strings.TrimSpace(line))
		if strings.HasPrefix(line, flagPrefix) && (strings.Contains(line, templateMarkerLong) || strings.Contains(line, templateMarkerShort)) {
			return true
		}
	}

	return false
}

func ParseFlags(line string) []string {
	possibleFlags := strings.Split(strings.TrimPrefix(line, flagPrefix), ",")
	flags := make([]string, 0, len(possibleFlags))
	for _, flag := range possibleFlags {
		flag = strings.ToLower(strings.TrimSpace(flag))

		if strings.HasPrefix(flag, "range:") {
			if !reRange.MatchString(flag) {
				log.WithField("input", flag).Warn("Invalid range flag")
				continue
			}

			rangeFlag := fmt.Sprintf("range: %s", strings.TrimSpace(strings.TrimPrefix(flag, "range:")))
			flags = append(flags, rangeFlag)
		}

		if flag == "ignore" {
			flags = append(flags, flag)
		}
	}

	return flags
}
