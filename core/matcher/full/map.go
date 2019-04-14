/*
 * Copyright (c) 2019 shawn1m. All rights reserved.
 * Use of this source code is governed by The MIT License (MIT) that can be
 * found in the LICENSE file..
 */

package full

type Map struct {
	DataMap map[string]struct{}
}

func (m *Map) Insert(str string) error {
	m.DataMap[str] = struct{}{}
	return nil
}

func (m *Map) Has(str string) bool {
	if _, ok := m.DataMap[str]; ok {
		return true
	}
	return false
}

func (m *Map) Name() string {
	return "full-map"
}
