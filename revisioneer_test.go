package main

import (
	. "./app/controllers"
	_ "github.com/eaigner/hood"
)

func init() {
	base = &Base{}
	base.Setup()
}
