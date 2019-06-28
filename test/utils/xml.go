package utils

import (
	"bytes"
	"io/ioutil"
	"strings"
)

// ReplaceElement replaces the xml element of the given name with the given value
// with count being the number of lines to replace or < 0 for no limit
func ReplaceElement(fileName string, elementName string, value string, count int) error {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	prefix := "<" + elementName + ">"
	suffix := "</" + elementName + ">"
	text := string(data)
	var buffer bytes.Buffer
	lines := strings.Split(text, "\n")
	i := 0
	for _, line := range lines {
		text := line
		if count < 0 || i < count {
			start := strings.Index(text, prefix)
			if start >= 0 {
				end := strings.Index(text, suffix)
				if end > start {
					text = line[0:start+len(prefix)] + value + line[end:]
					i++
				}
			}
		}
		buffer.WriteString(text)
		buffer.WriteString("\n")
	}
	if i > 0 {
		return ioutil.WriteFile(fileName, buffer.Bytes(), DefaultWritePermissions)
	}
	return nil
}
