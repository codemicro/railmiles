package util

import (
	"errors"
	"fmt"
)

type UserError error

func Wrap(err error, msg string, args ...any) error {
	e := err
	for e != nil {
		if _, ok := e.(UserError); ok {
			return err
		}
		e = errors.Unwrap(err)
	}
	return fmt.Errorf(msg+": %w", append(args, err)...)
}

func Deduplicate[T comparable](x []T) []T {
	var m = make(map[T]struct{})
	var res []T
	for _, v := range x {
		if _, found := m[v]; !found {
			res = append(res, v)
			m[v] = struct{}{}
		}
	}
	return res
}

func Map[T, Q comparable](x []T, y func(T) Q) []Q {
	var res []Q
	for _, z := range x {
		res = append(res, y(z))
	}
	return res
}

func ChainsToMiles(chains int) float32 {
	// 80 chains to a mile
	return float32(chains) / 80
}
