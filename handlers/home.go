// Copyright 2023 Harran Ali <harran.m@gmail.com>. All rights reserved.
// Use of this source code is governed by MIT-style
// license that can be found in the LICENSE file.

package handlers

import (
	"github.com/gocondor/core"
)

// Show home page
func HomeGet(c *core.Context) {
	message := "{\"message\": \"Welcome to GoCondor\"}"
	c.Response.WriteJson([]byte(message))
}