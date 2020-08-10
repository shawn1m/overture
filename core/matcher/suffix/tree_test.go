/*
 * Copyright (c) 2019 shawn1m. All rights reserved.
 * Use of this source code is governed by The MIT License (MIT) that can be
 * found in the LICENSE file..
 */

package suffix

import (
	"testing"
)

func TestTree_Has(t *testing.T) {
	tree := DefaultDomainTree()
	for _, d := range []string{
		"1.abc.com",
		"2.abc.com",
		"1.2.abc.com",
	} {
		tree.Insert(d)
	}
	for _, d := range []string{
		"1.abc.com",
		"2.abc.com",
		"1.2.abc.com",
	} {
		if !tree.Has(d) {
			t.Fail()
		}
	}
	if tree.Has("abc.com") {
		t.Fail()
	}
}
