// Converts type of a function to accept only its first argument
// if original function has single or no arguments it will return function with no arguments
type FnWithFirstArg<F extends (...args: any[]) => any> = F extends (
  arg1: infer A,
  ...args: infer R
) => infer Return
  ? R extends [] // Check if there are no additional arguments
    ? F // If no extra arguments, return the original function
    : (arg1: A) => Return
  : never

// Applies FnWithFirstArg to all functions in an object
type ApplyFnWithFirstArg<T> = {
  [K in keyof T]: T[K] extends (...args: any[]) => any ? FnWithFirstArg<T[K]> : T[K]
}

export type { ApplyFnWithFirstArg }

// Type tests
type ZeroArg = () => string
type OneArg = (x: number) => string
type TwoArgs = (x: number, y: string) => string
type ThreeArgs = (x: number, y: string, z: boolean) => string

// Test cases
type test1 = Expect<Equal<FnWithFirstArg<ZeroArg>, () => string>>
type test2 = Expect<Equal<FnWithFirstArg<OneArg>, (x: number) => string>>
type test3 = Expect<Equal<FnWithFirstArg<TwoArgs>, (x: number) => string>>
type test4 = Expect<Equal<FnWithFirstArg<ThreeArgs>, (x: number) => string>>

// Negative test case
// @ts-expect-error - Should error because NotAFunction is not a function type
type test5 = FnWithFirstArg<string>

// Helper types for testing
type Equal<X, Y> =
  (<T>() => T extends X ? 1 : 2) extends <T>() => T extends Y ? 1 : 2 ? true : false
type Expect<T extends true> = T
