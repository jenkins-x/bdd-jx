package parsers

import "strings"

func ParseJxGetQuickstarts(s string) (map[string]string, error) {
	answer := make(map[string]string)
	for _, line := range strings.Split(s, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 5 {
			if fields[0] == "NAME" {
				continue
			}
			answer[fields[0]] = line
		}
	}
	return answer, nil
}
