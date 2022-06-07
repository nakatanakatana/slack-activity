package main

import (
	"strings"
)

func GetChannelIDs(text string) []string {
	result := make([]string, 0)

	spl := strings.Split(text, "<#")
	if len(spl) <= 1 {
		return []string{}
	}

	for _, s := range spl[1:] {
		tmp := strings.Split(s, ">")
		if len(tmp) <= 1 {
			continue
		}

		id := tmp[0]
		if strings.Contains(id, "|") {
			tmp := strings.Split(id, "|")
			id = tmp[0]
		}

		result = append(result, id)
	}

	return result
}
