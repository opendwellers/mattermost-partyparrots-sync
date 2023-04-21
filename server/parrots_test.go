package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetchParrots(t *testing.T) {
	assert := assert.New(t)
	for _, parrotType := range parrotTypes {
		list, err := fetchParrotList(parrotType)
		assert.Nil(err)
		assert.Greater(len(list), 0)
	}
}

func TestFetchParrotGif(t *testing.T) {
	assert := assert.New(t)
	parrot := Parrot{
		name: "parrot",
		file: "hd/parrot.gif",
	}
	err := fetchParrotGif(&parrot, "parrots")
	assert.Nil(err)
	assert.NotEmpty(parrot.gif)
}
