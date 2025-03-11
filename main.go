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

	"github.com/microsoft/typescript-go/core/ast"
	"github.com/microsoft/typescript-go/core/core"
	"github.com/microsoft/typescript-go/core/parser"
	"github.com/microsoft/typescript-go/core/repo"
	"github.com/microsoft/typescript-go/core/scanner"
	"github.com/microsoft/typescript-go/core/tspath"
)

func main() {
	sourcePath := os.Args[1]
	sourceText, _ := os.ReadFile(sourcePath)

	fileName := tspath.GetNormalizedAbsolutePath(sourcePath, repo.TypeScriptSubmodulePath)
	path := tspath.ToPath(sourcePath, repo.TypeScriptSubmodulePath, true)

	var sourceFile *ast.SourceFile

	sourceFile = parser.ParseSourceFile(fileName, path, string(sourceText), core.ScriptTargetESNext, scanner.JSDocParsingModeParseAll)

	diagnostics := sourceFile.Diagnostics()
	if len(diagnostics) != 0 {
		for _, diagnostic := range diagnostics {
			fmt.Printf("error at %d:%d: %s", diagnostic.Pos(), diagnostic.End(), diagnostic.Message())
		}
	}

	module := C.LLVMModuleCreateWithName(C.CString("main"))
	C.LLVMDisposeModule(module)
}
