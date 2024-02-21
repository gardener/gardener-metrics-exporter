// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

// NewLogger creates a new logger to log to stdout.
func NewLogger() *logrus.Logger {
	return &logrus.Logger{
		Out:   os.Stdout,
		Level: logrus.InfoLevel,
		Formatter: &logrus.TextFormatter{
			DisableColors: true,
			FullTimestamp: true,
		},
	}
}
