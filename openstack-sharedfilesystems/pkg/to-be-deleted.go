package sharedfilesystems

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func SplitMicroversion(mv string) (major, minor int) {
	if err := ValidMicroversion(mv); err != nil {
		return
	}

	mvParts := strings.Split(mv, ".")
	major, _ = strconv.Atoi(mvParts[0])
	minor, _ = strconv.Atoi(mvParts[1])

	return
}

func ValidMicroversion(mv string) (err error) {
	mvRe := regexp.MustCompile("^\\d+\\.\\d+$")
	if v := mvRe.MatchString(mv); v {
		return
	}

	err = fmt.Errorf("invalid microversion: %q", mv)
	return
}
