// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package utils

import "errors"

const typeConversionError = "Type conversion error"

// NewTypeConversionError returns an error for a failed type conversion situation.
func NewTypeConversionError() error {
	return errors.New(typeConversionError)
}
