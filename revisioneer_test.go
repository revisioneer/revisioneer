package main

import (
	_ "github.com/eaigner/hood"
	. "github.com/revisioneer/revisioneer/app/controllers"
)

func init() {
	base = &Base{}
	base.Setup()
}
