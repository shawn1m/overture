/*
 * Copyright (c) 2019 shawn1m. All rights reserved.
 * Use of this source code is governed by The MIT License (MIT) that can be
 * found in the LICENSE file..
 */

package mix

import (
	"errors"
	"regexp"
	"strings"
)

type Data struct {
	Type    string
	Content string
}

type List struct {
	DataList []Data
}

func (s *List) Insert(str string) error {
	kv := strings.Split(str, ":")
	if len(kv) > 2 {
		return errors.New("Invalid format: " + str)
	}
	if len(kv) == 1 {
		s.DataList = append(s.DataList,
			Data{
				Type:    "domain",
				Content: strings.ToLower(kv[0])})
	}
	if len(kv) == 2 {
		s.DataList = append(s.DataList,
			Data{
				Type:    strings.ToLower(kv[0]),
				Content: strings.ToLower(kv[1])})
	}

	return nil
}

func (s *List) Has(str string) bool {
	for _, data := range s.DataList {
		switch data.Type {
		case "domain":
			idx := len(str) - len(data.Content)
			if idx > 0 && data.Content == str[idx:] {
				return true
			}
		case "regex":
			reg := regexp.MustCompile(data.Content)
			if reg.MatchString(str) {
				return true
			}
		case "keyword":
			if strings.Contains(str, data.Content) {
				return true
			}
		case "full":
			if data.Content == str {
				return true
			}
		}
	}
	return false
}

func (s *List) Name() string {
	return "mix-list"
}
