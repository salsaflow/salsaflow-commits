package main

import (
	"fmt"
	"os"
)

type ErrVarNotSet struct {
	key string
}

func (err *ErrVarNotSet) Error() string {
	return fmt.Sprintf("environment variable %v not set", err.key)
}

func mustGetenv(key string) (value string) {
	v := os.Getenv(key)
	if v == "" {
		panic(&ErrVarNotSet{key})
	}
	return v
}

func recoverEnvironPanic(err *error) {
	if r := recover(); r != nil {
		if ex, ok := r.(*ErrVarNotSet); ok {
			*err = ex
		} else {
			panic(r)
		}
	}
}
