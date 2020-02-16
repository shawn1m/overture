/*
 * Copyright (c) 2019 shawn1m. All rights reserved.
 * Use of this source code is governed by The MIT License (MIT) that can be
 * found in the LICENSE file..
 */

package final

type Default struct {
}

func (s *Default) Insert(str string) error {
	return nil
}

func (s *Default) Has(str string) bool {
	return true
}

func (s *Default) Name() string {
	return "final"
}
