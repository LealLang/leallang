// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/LealLang/leallang/internal/ast"
	"github.com/LealLang/leallang/internal/diagnostics"
	"github.com/LealLang/leallang/internal/lexer"
	"github.com/LealLang/leallang/internal/parser"
)

func main() {
	showTokens := flag.Bool("tokens", false, "print token stream instead of AST")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "usage: leal [--tokens] <file.ll>\n")
		os.Exit(1)
	}

	filename := flag.Arg(0)
	src, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	diag := diagnostics.New()
	l := lexer.New(filename, string(src), diag)
	tokens := l.Tokenize()

	if *showTokens {
		for _, tok := range tokens {
			fmt.Println(tok.String())
		}
		if diag.HasErrors() {
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, diag.Format())
			os.Exit(1)
		}
		return
	}

	program := parser.New(tokens, diag).Parse()

	if diag.HasErrors() {
		fmt.Fprintln(os.Stderr, diag.Format())
		os.Exit(1)
	}

	ast.Print(os.Stdout, program)
}
