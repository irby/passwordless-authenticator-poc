package persistence

import (
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type PostPersister interface {
	Create(models.Post) error
	List(page int, perPage int) ([]models.Post, error)
}

type postPersister struct {
	db *pop.Connection
}

func NewPostPersister(db *pop.Connection) PostPersister {
	return &postPersister{db: db}
}

func (p *postPersister) Create(post models.Post) error {
	vErr, err := p.db.ValidateAndCreate(&post)
	if err != nil {
		return fmt.Errorf("failed to store user: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("user object validation failed: %w", vErr)
	}

	return nil
}

func (p *postPersister) List(page int, perPage int) ([]models.Post, error) {
	post := []models.Post{}
	err := p.db.Q().Order("created_at desc").Paginate(page, perPage).All(&post)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch posts: %w", err)
	}

	return post, nil
}
