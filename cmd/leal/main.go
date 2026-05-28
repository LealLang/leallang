// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"

	"github.com/LealLang/leallang/internal/diagnostics"
	"github.com/LealLang/leallang/internal/lexer"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: leal <file.ll>\n")
		os.Exit(1)
	}

	filename := os.Args[1]
	src, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	diag := diagnostics.New()
	l := lexer.New(filename, string(src), diag)
	tokens := l.Tokenize()

	for _, tok := range tokens {
		fmt.Println(tok.String())
	}

	if diag.HasErrors() {
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, diag.Format())
		os.Exit(1)
	}
}
