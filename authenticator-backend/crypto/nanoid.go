package crypto

import (
	"fmt"
	"github.com/jaevor/go-nanoid"
)

type NanoidGenerator interface {
	Generate() (string, error)
}

type nanoidGenerator struct {
}

func NewNanoidGenerator() NanoidGenerator {
	return &nanoidGenerator{}
}

func (g *nanoidGenerator) Generate() (string, error) {
	id, err := nanoid.Standard(40)
	if err != nil {
		return "", fmt.Errorf("failed to generate nanoid: %w", err)
	}

	return id(), nil
}
