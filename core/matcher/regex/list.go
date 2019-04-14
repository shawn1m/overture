/*
 * Copyright (c) 2019 shawn1m. All rights reserved.
 * Use of this source code is governed by The MIT License (MIT) that can be
 * found in the LICENSE file..
 */

package regex

import "github.com/shawn1m/overture/core/common"

type List struct {
	RegexList []string
}

func (r *List) Insert(s string) error {
	r.RegexList = append(r.RegexList, s)
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
	return "regex-list"
}
