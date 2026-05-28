// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"

	"github.com/LealLang/leallang/internal/checker"
	"github.com/LealLang/leallang/internal/diagnostics"
	"github.com/LealLang/leallang/internal/lexer"
	"github.com/LealLang/leallang/internal/parser"
)

func runCheck(filename string) {
	src, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	diag := diagnostics.New()
	tokens := lexer.New(filename, string(src), diag).Tokenize()
	program := parser.New(tokens, diag).Parse()

	if diag.HasErrors() {
		fmt.Fprintln(os.Stderr, diag.Format())
		os.Exit(1)
	}

	checker.Check(program, diag)

	if diag.HasErrors() {
		fmt.Fprintln(os.Stderr, diag.Format())
		os.Exit(1)
	}

	fmt.Println("type check passed")
}
