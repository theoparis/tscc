package llvm

/*
#cgo LDFLAGS: -lLLVM

#include <llvm-c/Core.h>
#include <llvm-c/Target.h>
#include <llvm-c/Analysis.h>
#include <llvm-c/BitWriter.h>
#include <stdlib.h>
#include <stdint.h>
*/
import "C"
import (
	"os"
	"unsafe"
)

type Type struct {
	handle C.LLVMTypeRef
}

func (t Type) Handle() C.LLVMTypeRef {
	return t.handle
}

func (t Type) Dump() {
	C.LLVMDumpType(t.handle)
}

func ConstInt(t Type, value uint64, signExtend bool) Value {
	signExtendValue := 0
	if signExtend {
		signExtendValue = 1
	}

	return Value{
		handle: C.LLVMConstInt(t.Handle(), C.ulonglong(value), C.int(signExtendValue)),
	}
}

func ConstFloat(t Type, value float64) Value {
	return Value{
		handle: C.LLVMConstReal(t.Handle(), C.double(value)),
	}
}

func ConstNull(t Type) Value {
	return Value{
		handle: C.LLVMConstNull(t.Handle()),
	}
}

func ConstUndef(t Type) Value {
	return Value{
		handle: C.LLVMGetUndef(t.Handle()),
	}
}

func (builder *Builder) ConstString(value string, nullTerminate bool) Value {
	global := C.LLVMAddGlobal(
		builder.module.handle,
		C.LLVMArrayType(C.LLVMInt8Type(), C.uint(len(value)+1)),
		C.CString(""),
	)

	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))

	C.LLVMSetInitializer(global, C.LLVMConstString(cValue, C.uint(len(value)), C.int(0)))
	C.LLVMSetGlobalConstant(global, C.int(1))
	C.LLVMSetLinkage(global, C.LLVMPrivateLinkage)
	C.LLVMSetUnnamedAddr(global, C.LLVMGlobalUnnamedAddr)
	C.LLVMSetAlignment(global, 1)

	// LLVMBuildInBoundsGEP2

	zeroIndex := C.LLVMConstInt(C.LLVMInt32Type(), 0, 0)
	indices := []C.LLVMValueRef{zeroIndex, zeroIndex}

	return Value{
		handle: C.LLVMBuildInBoundsGEP2(
			builder.Handle(),
			C.LLVMPointerType(C.LLVMInt8Type(), 0),
			global,
			&indices[0],
			C.uint(len(indices)),
			C.CString(""),
		),
	}
}

func ConstArray(t Type, values []Value) Value {
	vals := make([]C.LLVMValueRef, len(values))
	for i, v := range values {
		vals[i] = v.Handle()
	}

	return Value{
		handle: C.LLVMConstArray(t.Handle(), &vals[0], C.uint(len(values))),
	}
}

func ConstReal(t Type, value float64) Value {
	return Value{
		handle: C.LLVMConstReal(t.Handle(), C.double(value)),
	}
}

type Module struct {
	handle C.LLVMModuleRef
}

func Initialize() {
	C.LLVMInitializeAllTargetInfos()
	C.LLVMInitializeAllTargets()
	C.LLVMInitializeAllTargetMCs()
	C.LLVMInitializeAllAsmPrinters()
	C.LLVMInitializeAllAsmParsers()
}

func NewModule(name string, targetTriple string) *Module {
	m := &Module{
		handle: C.LLVMModuleCreateWithName(C.CString(name)),
	}

	C.LLVMSetTarget(m.handle, C.CString(targetTriple))

	return m
}

func (m *Module) Handle() C.LLVMModuleRef {
	return m.handle
}

func (m *Module) Dispose() {
	C.LLVMDisposeModule(m.handle)
}

func (m *Module) Verify() bool {
	var err *C.char
	defer C.free(unsafe.Pointer(err))
	return C.LLVMVerifyModule(m.handle, C.LLVMAbortProcessAction, &err) == 0
}

func (m *Module) Dump() {
	C.LLVMDumpModule(m.handle)
}

func (m *Module) WriteToObjectFile(path string) bool {
	return C.LLVMWriteBitcodeToFile(m.handle, C.CString(path)) == 0
}

func (m *Module) WriteToFile(path string) (bool, error) {
	// write llvm ir
	var ir *C.char
	defer C.free(unsafe.Pointer(ir))
	ir = C.LLVMPrintModuleToString(m.handle)
	irStr := C.GoString(ir)
	irBytes := []byte(irStr)

	os.WriteFile(path, irBytes, 0644)

	return true, nil
}

type Context struct {
	handle C.LLVMContextRef
}

func NewContext() *Context {
	return &Context{
		handle: C.LLVMContextCreate(),
	}
}

func (c *Context) Handle() C.LLVMContextRef {
	return c.handle
}

func (c *Context) Dispose() {
	C.LLVMContextDispose(c.handle)
}

type FunctionType struct {
	Type
}

func NewFunctionType(returnType Type, paramTypes []Type, hasVarArg bool) *FunctionType {
	params := make([]C.LLVMTypeRef, len(paramTypes))

	for i, t := range paramTypes {
		params[i] = t.Handle()
	}

	var varArgValue C.int = 0

	if hasVarArg {
		varArgValue = 1
	}

	if len(paramTypes) == 0 {
		return &FunctionType{
			Type: Type{
				handle: C.LLVMFunctionType(returnType.Handle(), nil, 0, varArgValue),
			},
		}
	}

	return &FunctionType{
		Type: Type{
			handle: C.LLVMFunctionType(returnType.Handle(), &params[0], C.uint(len(paramTypes)), varArgValue),
		},
	}
}

func IntType(width uint) Type {
	return Type{
		handle: C.LLVMIntType(C.uint(width)),
	}
}

func VoidType() Type {
	return Type{
		handle: C.LLVMVoidType(),
	}
}

func FloatType() Type {
	return Type{
		handle: C.LLVMFloatType(),
	}
}

func DoubleType() Type {
	return Type{
		handle: C.LLVMDoubleType(),
	}
}

func PointerType(elemType Type, addressSpace uint) Type {
	return Type{
		handle: C.LLVMPointerType(elemType.Handle(), C.uint(addressSpace)),
	}
}

func StructType(elementTypes []Type, packed bool) Type {
	elements := make([]C.LLVMTypeRef, len(elementTypes))
	for i, t := range elementTypes {
		elements[i] = t.Handle()
	}

	packedValue := 0
	if packed {
		packedValue = 1
	}

	return Type{
		handle: C.LLVMStructType(&elements[0], C.uint(len(elementTypes)), C.int(packedValue)),
	}
}

type Builder struct {
	module *Module
	handle C.LLVMBuilderRef
}

func (b *Builder) CreateGetElementPtr(stringValue Value, value []Value, s string) Value {
	values := make([]C.LLVMValueRef, len(value))
	for i, v := range value {
		values[i] = v.Handle()
	}

	return Value{
		handle: C.LLVMBuildInBoundsGEP2(
			b.handle,
			C.LLVMPointerType(C.LLVMInt8Type(), 0),
			// string value
			stringValue.Handle(),
			&values[0],
			C.uint(len(value)),
			C.CString(s),
		),
	}
}

func (b *Builder) CreateCall(function *Function, arguments []Value, s string) Value {
	args := make([]C.LLVMValueRef, len(arguments))
	for i, a := range arguments {
		args[i] = a.Handle()
	}

	return Value{
		handle: C.LLVMBuildCall2(b.handle, function.Type(), function.Handle(), &args[0], C.uint(len(arguments)), C.CString(s)),
	}
}

func (b *Builder) CreateRet(value Value) Value {
	return Value{handle: C.LLVMBuildRet(b.handle, value.Handle())}
}

func (b *Builder) CreateRetVoid() Value {
	return Value{handle: C.LLVMBuildRetVoid(b.handle)}
}

func NewBuilder(module *Module) *Builder {
	return &Builder{
		handle: C.LLVMCreateBuilder(),
		module: module,
	}
}

func (b *Builder) Handle() C.LLVMBuilderRef {
	return b.handle
}

func (b *Builder) Dispose() {
	C.LLVMDisposeBuilder(b.handle)
}

func (b *Builder) PositionAtEnd(block C.LLVMBasicBlockRef) {
	C.LLVMPositionBuilderAtEnd(b.handle, block)
}

func (b *Builder) Insert(block C.LLVMValueRef) {
	C.LLVMInsertIntoBuilder(b.handle, block)
}

type Value struct {
	handle C.LLVMValueRef
}

func (v *Value) Handle() C.LLVMValueRef {
	return v.handle
}

func (v *Value) Dump() {
	C.LLVMDumpValue(v.handle)
}

func (v *Value) Print() string {
	return C.GoString(C.LLVMPrintValueToString(v.handle))
}

func (v *Value) IsConstant() bool {
	return C.LLVMIsConstant(v.handle) == 1
}

func (v *Value) IsNull() bool {
	return C.LLVMIsNull(v.handle) == 1
}

func (v *Value) IsUndef() bool {
	return C.LLVMIsUndef(v.handle) == 1
}

func (v *Value) IsAConstantInt() bool {
	return C.LLVMIsAConstantInt(v.handle) != nil
}

func (v *Value) IsAConstantFP() bool {
	return C.LLVMIsAConstantFP(v.handle) != nil
}

type Function struct {
	value    Value
	llvmType FunctionType
}

func NewFunction(module *Module, name string, fnType *FunctionType) *Function {
	return &Function{
		value: Value{
			handle: C.LLVMAddFunction(module.Handle(), C.CString(name), fnType.Handle()),
		},
		llvmType: *fnType,
	}
}

func (f *Function) Type() C.LLVMTypeRef {
	return f.llvmType.Handle()
}

func (f *Function) IsNil() bool {
	return f.value.handle == nil
}

func (f *Function) Value() Value {
	return f.value
}

func (f *Function) Handle() C.LLVMValueRef {
	return f.value.handle
}

func (f *Function) Dump() {
	C.LLVMDumpValue(f.value.handle)
}

func (f *Function) SetCallConv(cc C.uint) {
	C.LLVMSetFunctionCallConv(f.value.handle, cc)
}

func (f *Function) SetLinkage(linkage C.LLVMLinkage) {
	C.LLVMSetLinkage(f.value.handle, linkage)
}

func CCallConv() C.uint {
	return C.LLVMCCallConv
}

func ExternalLinkage() C.LLVMLinkage {
	return C.LLVMExternalLinkage
}

func InternalLinkage() C.LLVMLinkage {
	return C.LLVMInternalLinkage
}

func (f *Function) AppendBasicBlock(name string) C.LLVMBasicBlockRef {
	return C.LLVMAppendBasicBlock(f.value.handle, C.CString(name))
}

func (f *Function) GetParam(index uint) C.LLVMValueRef {
	return C.LLVMGetParam(f.value.handle, C.uint(index))
}
