/*
 * Copyright (c) 2019 shawn1m. All rights reserved.
 * Use of this source code is governed by The MIT License (MIT) that can be
 * found in the LICENSE file..
 */

package full

type List struct {
	DataList []string
}

func (s *List) Insert(str string) error {
	s.DataList = append(s.DataList, str)
	return nil
}

func (s *List) Has(str string) bool {
	for _, data := range s.DataList {
		if data == str {
			return true
		}
	}
	return false
}

func (s *List) Name() string {
	return "full-list"
}
