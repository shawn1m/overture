/*
 * Copyright (c) 2019 shawn1m. All rights reserved.
 * Use of this source code is governed by The MIT License (MIT) that can be
 * found in the LICENSE file..
 */

package matcher

type Matcher interface {
	Insert(string) error
	Has(string) bool
	Name() string
}
