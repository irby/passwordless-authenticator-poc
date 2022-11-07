package test

import (
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

func NewPostPersister(init []models.Post) persistence.PostPersister {
	return &postPersister{append([]models.Post{}, init...)}
}

type postPersister struct {
	posts []models.Post
}

func (p *postPersister) Create(post models.Post) error {
	p.posts = append(p.posts, post)
	return nil
}

func (p *postPersister) List(page int, perPage int) ([]models.Post, error) {
	if len(p.posts) == 0 {
		return p.posts, nil
	}

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	var result [][]models.Post
	var j int
	for i := 0; i < len(p.posts); i += perPage {
		j += perPage
		if j > len(p.posts) {
			j = len(p.posts)
		}
		result = append(result, p.posts[i:j])
	}

	if page > len(result) {
		return []models.Post{}, nil
	}
	return result[page-1], nil
}
