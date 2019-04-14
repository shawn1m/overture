/*
 * Copyright (c) 2019 shawn1m. All rights reserved.
 * Use of this source code is governed by The MIT License (MIT) that can be
 * found in the LICENSE file..
 */

package full

type Map struct {
	DataMap map[string]string
}

func (m *Map) Insert(k string, v string) error {
	m.DataMap[k] = v
	return nil
}

func (m *Map) Get(k string) string {
	return m.DataMap[k]
}

func (m *Map) Name() string {
	return "full-map"
}
