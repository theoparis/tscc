declare namespace LLVM {
  export interface Indexable<T> {
    [index: number]: T;
  }

  function indexAddress<T>(value: number): Indexable<T>;
}

declare type i32 = number;
