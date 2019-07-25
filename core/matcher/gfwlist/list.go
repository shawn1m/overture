/*
 * Copyright (c) 2019 shawn1m. All rights reserved.
 * Use of this source code is governed by The MIT License (MIT) that can be
 * found in the LICENSE file..
 */

package gfwlist

import (
	"fmt"
	"strings"

	"github.com/shawn1m/overture/core/common"
)

type List struct {
	RegexList []string
}

func (r *List) Insert(s string) error {
	if str := convertLine(s); str != "" {
		r.RegexList = append(r.RegexList, convertLine(s))
	}
	return nil
}

func (r *List) Has(s string) bool {
	for _, regex := range r.RegexList {
		if common.IsDomainMatchRule(regex, s) {
			return true
		}
	}
	return false
}

func (r *List) Name() string {
	return "gfwlist"
}

// function rewritten in Go from Python script https://gist.githubusercontent.com/sorz/5577181/raw/d065899b319242b03eaee8e1215eeeeda1102306/gfwlist2regex.py
// with changes adapted to Overture format
func convertLine(line string) string {
	// skip empty lines and comments, and whitelists
	if line == "" || line[:1] == "!" || line[:1] == "[" || strings.HasPrefix(line, "@@") {
		return ""
	}

	// if line is already a regex, return as is
	if strings.HasPrefix(line, "/") && strings.HasPrefix(line, "/") {
		return line
	}

	// wildcards
	line = strings.ReplaceAll(line, "*", ".+")

	// escaping characters for regex
	line = strings.ReplaceAll(line, "(", `\(`)
	line = strings.ReplaceAll(line, ")", `\)`)
	line = strings.ReplaceAll(line, ".", `\.`)

	// we don't need schemes in domains
	line = strings.TrimPrefix(line, "http://")
	line = strings.TrimPrefix(line, "https://")
	line = strings.TrimPrefix(line, "//")

	// process line into regex
	switch {
	case strings.HasPrefix(line, "||"):
		return fmt.Sprintf("^%s.*", line[2:])
	case strings.HasPrefix(line, "|"):
		return fmt.Sprintf("^%s.*", line[1:])
	case line[len(line)-1:] == "|":
		return fmt.Sprintf(".*%s$", line)
	default:
		return fmt.Sprintf(".*%s.*", line)
	}
}
