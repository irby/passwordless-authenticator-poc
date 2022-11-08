package handler

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/test"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPostHandler_ListPosts_WhenNotAdmin(t *testing.T) {
	handler := newPostHandler()
	actingUser := generateUser(t)
	handler.persister.GetUserPersister().Create(actingUser)
	seedPostsAndUsers(t, handler)

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodGet, "/posts", strings.NewReader(""))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("session", generateJwt(t, actingUser.ID, actingUser.ID, 60))

	if assert.NoError(t, handler.GetPosts(c)) {
		assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
		response := GetPostsDto{}
		assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
		assert.Equal(t, 2, len(response.Posts))
	}
}

func TestPostHandler_ListPosts_WhenAdmin(t *testing.T) {
	handler := newPostHandler()
	actingUser := generateUser(t)
	actingUser.IsAdmin = true
	handler.persister.GetUserPersister().Create(actingUser)
	seedPostsAndUsers(t, handler)

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodGet, "/posts", strings.NewReader(""))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("session", generateJwt(t, actingUser.ID, actingUser.ID, 60))

	if assert.NoError(t, handler.GetPosts(c)) {
		assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
		var response GetPostsDto
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(response.Posts))
		assert.NotNil(t, response.Posts[1].CreatedBySurrogate)
		assert.NotNil(t, response.Posts[1].UpdatedBySurrogate)
	}
}

func TestPostHandler_CreatePost_WhenPrimaryAccountHolder(t *testing.T) {
	handler := newPostHandler()
	actingUser := generateUser(t)
	handler.persister.GetUserPersister().Create(actingUser)
	seedPostsAndUsers(t, handler)

	body := `{"body":"hello, world! this is a post"}`
	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("session", generateJwt(t, actingUser.ID, actingUser.ID, 60))

	if assert.NoError(t, handler.CreatePost(c)) {
		assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
		posts, err := handler.persister.GetPostPersister().List(0, 10)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(posts))
		assert.Equal(t, actingUser.ID, posts[2].CreatedBySurrogateId)
		assert.Equal(t, actingUser.ID, posts[2].CreatedByUserId)
	}
}

func TestPostHandler_CreatePost_WhenGuest(t *testing.T) {
	handler := newPostHandler()
	actingUser := generateUser(t)
	parentUser := generateUser(t)
	handler.persister.GetUserPersister().Create(actingUser)
	handler.persister.GetUserPersister().Create(parentUser)
	seedPostsAndUsers(t, handler)

	body := `{"body":"hello, world! this is a post"}`
	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("session", generateJwt(t, parentUser.ID, actingUser.ID, 60))

	if assert.NoError(t, handler.CreatePost(c)) {
		assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
		posts, err := handler.persister.GetPostPersister().List(0, 10)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(posts))
		assert.Equal(t, actingUser.ID, posts[2].CreatedBySurrogateId)
		assert.Equal(t, parentUser.ID, posts[2].CreatedByUserId)
	}
}

func newPostHandler() *PostHandler {
	p := test.NewPersister(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	return NewPostHandler(p)
}

func seedPostsAndUsers(t *testing.T, h *PostHandler) {
	user1 := generateUser(t)
	user2 := generateUser(t)
	posts := []models.Post{
		func() models.Post {
			return models.Post{
				ID:                   generateUuid(t),
				CreatedAt:            time.Now().UTC(),
				IsActive:             true,
				CreatedByUserId:      user1.ID,
				UpdatedByUserId:      user1.ID,
				CreatedBySurrogateId: user1.ID,
				UpdatedBySurrogateId: user1.ID,
				Data:                 "hello, world!",
			}
		}(),
		func() models.Post {
			return models.Post{
				ID:                   generateUuid(t),
				CreatedAt:            time.Now().UTC(),
				IsActive:             true,
				CreatedByUserId:      user1.ID,
				UpdatedByUserId:      user1.ID,
				CreatedBySurrogateId: user2.ID,
				UpdatedBySurrogateId: user2.ID,
				Data:                 "foo, bar!",
			}
		}(),
	}

	h.persister.GetUserPersister().Create(user1)
	h.persister.GetUserPersister().Create(user2)

	for _, post := range posts {
		h.persister.GetPostPersister().Create(post)
	}
}
