// Copyright 2019-2024 go-tcap authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package tcap

import "fmt"

// InvalidCodeError indicates that Code in TCAP message is invalid.
type InvalidCodeError struct {
	Code int
}

// Error returns error message with violating content.
func (e *InvalidCodeError) Error() string {
	return fmt.Sprintf("tcap: got invalid code: %d", e.Code)
}
