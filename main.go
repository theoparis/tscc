package main

/*
#cgo LDFLAGS: -lLLVM

#include <llvm-c/Core.h>
#include <llvm-c/Analysis.h>
*/
import "C"
import (
	"fmt"
	"os"

	"flag"

	"code.tinted.dev/tinted/tscc/codegen"
	"code.tinted.dev/tinted/tscc/llvm"
	"github.com/microsoft/typescript-go/core/ast"
	"github.com/microsoft/typescript-go/core/core"
	"github.com/microsoft/typescript-go/core/parser"
	"github.com/microsoft/typescript-go/core/repo"
	"github.com/microsoft/typescript-go/core/scanner"
	"github.com/microsoft/typescript-go/core/tspath"
)

var (
	sourcePath = flag.String("source", "", "source file path")
	outputPath = flag.String("output", "", "output file path")
	target     = flag.String("target", "esnext", "target")
	triple     = flag.String("triple", "aarch64-unknown-linux-gnu", "triple")
	optLevel   = flag.Int("opt", 0, "optimization level")
	entryPoint = flag.String("entry", "main", "entry point name")
)

func main() {
	flag.Parse()

	if *sourcePath == "" {
		fmt.Println("source file path is required")
		os.Exit(1)
	}

	if *outputPath == "" {
		// set to the source file path with the .o extension
		*outputPath = *sourcePath + ".ll"
	}

	// Parse command line arguments
	// source path, output path, target, optimization level, entry point name
	sourceText, _ := os.ReadFile(*sourcePath)

	fileName := tspath.GetNormalizedAbsolutePath(*sourcePath, repo.TypeScriptSubmodulePath)
	path := tspath.ToPath(*sourcePath, repo.TypeScriptSubmodulePath, true)

	var sourceFile *ast.SourceFile

	sourceFile = parser.ParseSourceFile(fileName, path, string(sourceText), core.ScriptTargetESNext, scanner.JSDocParsingModeParseAll)

	diagnostics := sourceFile.Diagnostics()
	if len(diagnostics) != 0 {
		for _, diagnostic := range diagnostics {
			fmt.Printf("error at %d:%d: %s", diagnostic.Pos(), diagnostic.End(), diagnostic.Message())
		}
	}

	llvm.Initialize()
	module := llvm.NewModule("main", *triple)

	// Create a function
	functionType := llvm.NewFunctionType(llvm.VoidType(), []llvm.Type{}, false)
	function := llvm.NewFunction(module, *entryPoint, functionType)

	entryBlock := function.AppendBasicBlock("entry")
	builder := llvm.NewBuilder(module)

	builder.PositionAtEnd(entryBlock)

	// Translate the AST
	codegen.TranslateSourceFile(sourceFile, module, function, builder)

	builder.CreateRetVoid()

	module.Dump()
	module.Verify()
	success, err := module.WriteToFile(*outputPath)
	if !success {
		fmt.Println(err)
		builder.Dispose()
		module.Dispose()
		os.Exit(1)
	}

	builder.Dispose()
	module.Dispose()
}
