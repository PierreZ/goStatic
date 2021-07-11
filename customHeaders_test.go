package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeaderConfigWithValidRegex(t *testing.T) {
	assert := assert.New(t)
	config := HeaderConfig{Regex: "/$"}

	ok := config.Init()
	assert.True(ok)
	assert.NotNil(config.CompiledRegex)
	assert.True(config.UsesRegex())
}

func TestHeaderConfigWithInvalidRegex(t *testing.T) {
	assert := assert.New(t)
	config := HeaderConfig{Regex: "["}

	ok := config.Init()
	assert.False(ok)
	assert.Nil(config.CompiledRegex)
}

func TestHeaderConfigWithoutRegex(t *testing.T) {
	assert := assert.New(t)
	config := HeaderConfig{
		Path:          "/page-data",
		FileExtension: "json",
	}

	ok := config.Init()
	assert.True(ok)
	assert.Nil(config.CompiledRegex)
	assert.False(config.UsesRegex())
}
