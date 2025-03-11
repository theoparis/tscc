package codegen

import (
	"strconv"

	"code.tinted.dev/tinted/tscc/llvm"
	"github.com/microsoft/typescript-go/core/ast"
)

// TranslateSourceFile translates the given source file to LLVM IR.
func TranslateSourceFile(
	sourceFile *ast.SourceFile,
	module *llvm.Module,
	function *llvm.Function,
	builder *llvm.Builder,
) {
	variables := make(map[string]llvm.Value)
	functions := make(map[string]*llvm.Function)

	for _, node := range sourceFile.Statements.Nodes {
		translateNode(node, module, function, builder, variables, functions)
	}
}

func translateNode(
	node *ast.Node,
	module *llvm.Module,
	function *llvm.Function,
	builder *llvm.Builder,
	variables map[string]llvm.Value,
	functions map[string]*llvm.Function,
) llvm.Value {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		return translateFunctionDeclaration(node, module, function, builder, variables, functions)
	case ast.KindVariableStatement:
		return translateVariableStatement(node, module, function, builder, variables, functions)
	case ast.KindExpressionStatement:
		return translateExpression(node.AsExpressionStatement().Expression, module, function, builder, variables, functions)
	case ast.KindReturnStatement:
		return builder.CreateRet(translateExpression(node.AsReturnStatement().Expression, module, function, builder, variables, functions))
	default:
		panic("unimplemented node kind: " + node.Kind.String())
	}
}

func translateFunctionDeclaration(
	node *ast.Node,
	module *llvm.Module,
	function *llvm.Function,
	builder *llvm.Builder,
	variables map[string]llvm.Value,
	functions map[string]*llvm.Function,
) llvm.Value {
	functionDeclaration := node.AsFunctionDeclaration()
	functionName := functionDeclaration.Name().Text()

	var argumentTypes []llvm.Type
	var varArg bool = false

	for _, parameter := range functionDeclaration.Parameters.Nodes {
		// Check if the parameter is a rest parameter
		if parameter.AsParameterDeclaration().DotDotDotToken != nil {
			varArg = true
		} else {
			argumentTypes = append(
				argumentTypes,
				translateType(
					parameter.AsParameterDeclaration().Type,
					module,
					function,
					builder,
					variables,
				),
			)
		}
	}

	functionType := llvm.NewFunctionType(
		translateType(functionDeclaration.Type, module, function, builder, variables),
		argumentTypes,
		varArg,
	)

	declaredFunction := llvm.NewFunction(module, functionName, functionType)
	declaredFunction.SetCallConv(llvm.CCallConv())

	for _, modifier := range functionDeclaration.Modifiers().Nodes {
		switch modifier.Kind {
		case ast.KindExportKeyword:
			declaredFunction.SetLinkage(llvm.ExternalLinkage())
		case ast.KindDeclareKeyword:
			declaredFunction.SetLinkage(llvm.ExternalLinkage())
		default:
		}
	}

	if functionDeclaration.Body != nil {
		translateBlockStatement(functionDeclaration.Body, module, function, builder, variables, functions)
	}

	functions[functionName] = declaredFunction

	return function.Value()
}

func translateType(
	node *ast.Node,
	module *llvm.Module,
	function *llvm.Function,
	builder *llvm.Builder,
	variables map[string]llvm.Value,
) llvm.Type {
	text := ""

	switch node.Kind {
	case ast.KindTypeReference:
		text = node.AsTypeReference().TypeName.Text()
	case ast.KindStringKeyword:
		text = "string"
	case ast.KindNumberKeyword:
		text = "number"
	case ast.KindBooleanKeyword:
		text = "boolean"
	case ast.KindVoidKeyword:
		text = "void"
	case ast.KindAnyKeyword:
		text = "any"
	case ast.KindObjectKeyword:
		text = "object"
	case ast.KindFunctionKeyword:
		text = "function"
	default:
		if node.Kind == ast.KindArrayType {
			return llvm.PointerType(translateType(node.AsArrayTypeNode().ElementType, module, function, builder, variables), 0)
		}

		text = node.Kind.String()
	}

	switch text {
	case "i32":
		return llvm.IntType(32)
	case "i64":
		return llvm.IntType(64)
	case "void":
		return llvm.VoidType()
	case "string":
		return llvm.PointerType(llvm.IntType(8), 0)
	case "boolean":
		return llvm.IntType(1)
	case "f32":
		return llvm.FloatType()
	case "f64":
		return llvm.DoubleType()
	case "number":
		return llvm.DoubleType()
	default:
		panic("unimplemented type: " + node.Kind.String())
	}

}

func translateBlockStatement(
	node *ast.Node,
	module *llvm.Module,
	function *llvm.Function,
	builder *llvm.Builder,
	variables map[string]llvm.Value,
	functions map[string]*llvm.Function,
) {
	blockStatement := node.AsBlock()
	for _, statement := range blockStatement.Statements.Nodes {
		translateNode(statement, module, function, builder, variables, functions)
	}
}

func translateVariableStatement(
	node *ast.Node,
	module *llvm.Module,
	function *llvm.Function,
	builder *llvm.Builder,
	variables map[string]llvm.Value,
	functions map[string]*llvm.Function,
) llvm.Value {
	variableStatement := node.AsVariableStatement()

	for _, declaration := range variableStatement.AsVariableDeclarationList().Declarations.Nodes {
		translateVariableDeclaration(declaration, module, function, builder, variables, functions)
	}

	return function.Value()
}

func translateVariableDeclaration(
	node *ast.Node,
	module *llvm.Module,
	function *llvm.Function,
	builder *llvm.Builder,
	variables map[string]llvm.Value,
	functions map[string]*llvm.Function,
) llvm.Value {
	variableDeclaration := node.AsVariableDeclaration()
	name := variableDeclaration.Name().Text()
	typeName := variableDeclaration.Type.Text()

	switch typeName {
	case "i32":
		if variableDeclaration.Initializer == nil {
			variables[name] = llvm.ConstInt(llvm.IntType(32), 0, false)
		} else {
			value := translateExpression(variableDeclaration.Initializer, module, function, builder, variables, functions)
			variables[name] = value
		}
	case "i64":
		if variableDeclaration.Initializer == nil {
			variables[name] = llvm.ConstInt(llvm.IntType(64), 0, false)
		} else {
			value := translateExpression(variableDeclaration.Initializer, module, function, builder, variables, functions)
			variables[name] = value
		}
	case "number":
		if variableDeclaration.Initializer == nil {
			variables[name] = llvm.ConstReal(llvm.DoubleType(), 0)
		} else {
			value := translateExpression(variableDeclaration.Initializer, module, function, builder, variables, functions)
			variables[name] = value
		}
	default:
		panic("unimplemented initializer kind: " + variableDeclaration.Initializer.Kind.String())
	}

	return function.Value()
}

func translateExpression(
	node *ast.Node,
	module *llvm.Module,
	function *llvm.Function,
	builder *llvm.Builder,
	variables map[string]llvm.Value,
	functions map[string]*llvm.Function,
) llvm.Value {
	switch node.Kind {
	case ast.KindNumericLiteral:
		return translateNumericLiteral(node, module, function, builder, variables)
	case ast.KindStringLiteral:
		return TranslateStringLiteral(node, module, function, builder, variables)
	case ast.KindIdentifier:
		return translateIdentifier(node, module, function, builder, variables)
	case ast.KindCallExpression:
		callExpression := node.AsCallExpression()
		functionName := callExpression.Expression.AsIdentifier().Node.Text()
		function := functions[functionName]
		if function.IsNil() {
			panic("function not found: " + functionName)
		}

		var arguments []llvm.Value
		for _, argument := range callExpression.Arguments.Nodes {
			arguments = append(arguments, translateExpression(argument, module, function, builder, variables, functions))
		}

		return builder.CreateCall(function, arguments, functionName)
	default:
		panic("unimplemented expression kind: " + node.Kind.String())
	}
}

func translateNumericLiteral(
	node *ast.Node,
	module *llvm.Module,
	function *llvm.Function,
	builder *llvm.Builder,
	variables map[string]llvm.Value,
) llvm.Value {
	numericLiteral := node.AsNumericLiteral()
	numericLiteralValue, err := strconv.ParseFloat(numericLiteral.Text, 64)
	if err != nil {
		numericLiteralInt, err := strconv.ParseInt(numericLiteral.Text, 10, 64)
		if err != nil {
			panic("failed to parse numeric literal: " + numericLiteral.Text)
		}
		return llvm.ConstInt(
			llvm.IntType(32),
			uint64(numericLiteralInt),
			false,
		)
	}

	return llvm.ConstReal(llvm.DoubleType(), numericLiteralValue)
}

func TranslateStringLiteral(
	node *ast.Node,
	module *llvm.Module,
	function *llvm.Function,
	builder *llvm.Builder,
	variables map[string]llvm.Value,
) llvm.Value {
	stringLiteral := node.AsStringLiteral()
	return builder.ConstString(stringLiteral.Text, true)
}

func translateIdentifier(
	node *ast.Node,
	module *llvm.Module,
	function *llvm.Function,
	builder *llvm.Builder,
	variables map[string]llvm.Value,
) llvm.Value {
	identifier := node.AsIdentifier()
	name := identifier.Node.Text()
	value, ok := variables[name]
	if !ok {
		panic("variable not found: " + name)
	}
	return value
}
