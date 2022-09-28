package crypto

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNanoidGenerator_Generate(t *testing.T) {
	ng := NewNanoidGenerator()
	nanoid, err := ng.Generate()

	assert.NoError(t, err)
	assert.NotEmpty(t, nanoid)
	assert.Equal(t, 40, len(nanoid))
}

func TestNanoidGenerator_GeneratesUniqueValueEachRun(t *testing.T) {
	ng := NewNanoidGenerator()
	list := make([]string, 0)

	nanoid, err1 := ng.Generate()
	assert.NoError(t, err1)
	assert.NotEmpty(t, nanoid)
	assert.Equal(t, 40, len(nanoid))

	list = append(list, nanoid)

	// Ensure our nanoid is collision resistant by making several runs
	// and verifying it's a unique value each run
	for i := 0; i < 40; i++ {

		nanoid, err1 := ng.Generate()

		assert.NoError(t, err1)
		assert.NotEmpty(t, nanoid)
		assert.Equal(t, 40, len(nanoid))

		for j := 0; j < len(list); j++ {
			if list[j] == nanoid {
				panic("Element was found in list")
			}
		}

		list = append(list, nanoid)
	}
}
