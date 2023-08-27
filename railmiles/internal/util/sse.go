package util

import "strings"

type SSEItem struct {
	Event   string
	Message string
}

func (s *SSEItem) String() string {
	var sb strings.Builder

	if s.Event != "" {
		sb.WriteString("event: ")
		sb.WriteString(s.Event)
		sb.WriteRune('\n')
	}

	if s.Message != "" {
		for _, line := range strings.Split(s.Message, "\n") {
			sb.WriteString("data: ")
			sb.WriteString(line)
			sb.WriteRune('\n')
		}
	}

	sb.WriteString("\n\n")

	return sb.String()
}
