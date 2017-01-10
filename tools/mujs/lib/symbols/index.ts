// Copyright 2016 Marapongo, Inc. All rights reserved.

// Tokens.
export type Token = string;        // a valid symbol token.
export type ModuleToken = Token;   // a symbol token that resolves to a module.
export type TypeToken = Token;     // a symbol token that resolves to a type.
export type VariableToken = Token; // a symbol token that resolves to a variable.
export type FunctionToken = Token; // a symbol token that resolves to a function.

// Identifiers.
export type Identifier = string;   // a valid identifier:  (letter|"_") (letter | digit | "_")*

// Accessibility modifiers.
export type Accessibility            = "public" | "private";        // accessibility modifiers common to all.
export type ClassMemberAccessibility = Accessibility | "protected"; // accessibility modifiers for class members.

// Accessibility modifier constants.
export const publicAccessibility    = "public";
export const privateAccessibility   = "private";
export const protectedAccessibility = "protected";

// Special function tokens.
export const specialFunctionEntryPoint  = ".main"; // the special package entrypoint function.
export const specialFunctionInitializer = ".init"; // the special module/class initialize function.
export const specialFunctionConstructor = ".ctor"; // the special class instance constructor function.

