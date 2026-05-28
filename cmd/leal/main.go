// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "usage: leal <command> <file.ll>\n\n")
		fmt.Fprintf(os.Stderr, "commands:\n")
		fmt.Fprintf(os.Stderr, "  tokenize   print the token stream\n")
		fmt.Fprintf(os.Stderr, "  parse      print the AST\n")
		fmt.Fprintf(os.Stderr, "  check      run the type checker\n")
		os.Exit(1)
	}

	command := os.Args[1]
	filename := os.Args[2]

	switch command {
	case "tokenize":
		runTokenize(filename)
	case "parse":
		runParse(filename)
	case "check":
		runCheck(filename)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", command)
		fmt.Fprintf(os.Stderr, "available commands: tokenize, parse, check\n")
		os.Exit(1)
	}
}
