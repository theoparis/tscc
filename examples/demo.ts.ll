; ModuleID = 'main'
source_filename = "main"
target triple = "aarch64-unknown-linux-gnu"

@0 = private unnamed_addr constant [15 x i8] c"Hello, world!\0A\00", align 1

define void @main() {
entry:
  %printf = call i32 (ptr, ...) @printf(ptr @0)
  ret void
}

declare i32 @printf(ptr, ...)
