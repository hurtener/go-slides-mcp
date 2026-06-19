// Package contracts holds this server's tool input and output contracts.
//
// These typed Go structs are the SOURCE OF TRUTH for the tool's schema
// (Dockyard P1 — contract-first, RFC §6). The JSON Schema and TypeScript
// alongside this file are GENERATED from these structs by `dockyard generate`;
// never hand-edit a generated file. Change a contract here, then regenerate.
package contracts
