// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

package interpreter

import "fmt"

type signalKind int

const (
	signalReturn signalKind = iota
	signalBreak
	signalContinue
)

type Signal struct {
	Kind  signalKind
	Value Value
}

func (s *Signal) Error() string {
	switch s.Kind {
	case signalReturn:
		return fmt.Sprintf("return %s", valueString(s.Value))
	case signalBreak:
		return "break"
	case signalContinue:
		return "continue"
	default:
		return "signal"
	}
}
