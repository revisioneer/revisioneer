package main

import (
	_ "github.com/eaigner/hood"
	. "github.com/revisioneer/revisioneer/controllers"
)

func init() {
	base = &Base{}
	base.Setup()
}
