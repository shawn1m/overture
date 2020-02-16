/*
 * Copyright (c) 2019 shawn1m. All rights reserved.
 * Use of this source code is governed by The MIT License (MIT) that can be
 * found in the LICENSE file..
 */

package regex

import "github.com/shawn1m/overture/core/common"

type List struct {
	RegexMap map[string][]string
}

func (r *List) Insert(k string, v string) error {
	if r.RegexMap[k] == nil {
		r.RegexMap[k] = []string{v}
	} else {
		r.RegexMap[k] = append(r.RegexMap[k], v)
	}
	return nil
}

func (r *List) Get(str string) []string {
	var result []string
	for k, v := range r.RegexMap {
		if common.IsDomainMatchRule(k, str) {
			result = append(result, v...)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func (r *List) Name() string {
	return "regex-list"
}
