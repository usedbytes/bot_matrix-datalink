// Copyright 2017 Brian Starkey <stark3y@gmail.com>
package transport

type Transport interface {
	Transfer([]byte) ([]byte, error)
}
