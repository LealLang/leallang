// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"

	"github.com/LealLang/leallang/internal/checker"
	"github.com/LealLang/leallang/internal/diagnostics"
	"github.com/LealLang/leallang/internal/interpreter"
	"github.com/LealLang/leallang/internal/lexer"
	"github.com/LealLang/leallang/internal/parser"
)

func runRun(filename string, args []string) {
	src, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	diag := diagnostics.New()
	diag.SetSource(string(src))
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

	interp := interpreter.New(diag)
	interp.SetArgs(args)
	if err := interp.Run(program); err != nil {
		if diag.HasErrors() {
			fmt.Fprintln(os.Stderr, diag.Format())
		} else {
			fmt.Fprintf(os.Stderr, "runtime error: %s\n", err)
		}
		os.Exit(1)
	}
	if diag.HasErrors() {
		fmt.Fprintln(os.Stderr, diag.Format())
		os.Exit(1)
	}
	if code := interp.ExitCode(); code != 0 {
		os.Exit(code)
	}
}
