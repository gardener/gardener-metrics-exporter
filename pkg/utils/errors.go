// SPDX-FileCopyrightText: Copyright Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package utils

import "errors"

const typeConversionError = "type conversion error"

// NewTypeConversionError returns an error for a failed type conversion situation.
func NewTypeConversionError() error {
	return errors.New(typeConversionError)
}
