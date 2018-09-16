package main

import (
	"testing"

	"github.com/kami-zh/go-capturer"
	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	output := capturer.CaptureStdout(func() {
		Run([]string{"version"})
	})
	assert.Equal(t, "heroku/7.16.0 darwin-x64 node-v10.10.0\n", output)
}
