package base_utils

import (
	"fmt"
	"io"
)

// StringCollector provides a way to collect the string and other objects for building string at a later point to avoid
// memory and compute usage for a potentially throwaway object. Any implementation and usage must take in to consideration
//
// 1. Stack vs Heap - Keeping pointer to items typically allocated to stack can move them to heap.
//
type StringCollector interface {
	fmt.Stringer
	io.Writer
	io.ByteWriter
	io.StringWriter
}

// StringSecure interface is used to pack string that need to be secured in memory. It provides basic method to destroy
// the given value and enforces an implementation of fmt.Stringer & fmt.GoStringer to ensure that secure string is not
// printed by mistake.
type StringSecure interface {
	// DestroyE function destroys the encrypted storage and any decrypted version of string
	DestroyE() (bool, error)
	// Stringer implementation is expected to return a generic string (e.g. "******") to identify that this is a secure value.
	// This interface should typically not return either encrypted or original value of
	fmt.Stringer
	// GoStringer may return the same value as Stringer or different value. Additional values like encryption algorithm,
	// may be provided depending on implementation.
	fmt.GoStringer
}

// StringSecureByteEnabled defines the methods to store and retrieve bytes from the StringSecure implementation
type StringSecureByteEnabled interface {
	StoreE(value []byte) (bool, error)
	Get() []byte
}

// StringSecureStringEnabled defines the methods to store and retrieve string from the StringSecure implementation
type StringSecureStringEnabled interface {
	StoreStringE(value string) (bool, error)
	GetString() string
}

// StringSecureLockEnabled defines the methods to explicitly Lock and unlock value stored. This additional step if
type StringSecureLockEnabled interface {
	LockE() (bool, error)
	IsLocked() bool
	Unlock() StringSecure
	UnlockE() (StringSecure, error)
}
