package handler

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwt"
	jwt2 "github.com/teamhanko/hanko/backend/crypto/jwt"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"net/http"
	"time"
)

type PostHandler struct {
	persister persistence.Persister
}

func NewPostHandler(persister persistence.Persister) *PostHandler {
	return &PostHandler{persister: persister}
}

type GetPostsDto struct {
	Posts []PostDto `json:"posts"`
}

type PostDto struct {
	ID                 uuid.UUID `json:"id"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	CreatedByEmail     string    `json:"created_by"`
	CreatedBySurrogate *string   `json:"created_by_surrogate"`
	UpdatedBySurrogate *string   `json:"updated_by_surrogate"`
	Data               string    `json:"data"`
}

func (h *PostHandler) GetPosts(c echo.Context) error {
	posts, err := h.persister.GetPostPersister().List(0, 20)
	if err != nil {
		return dto.NewHTTPError(http.StatusInternalServerError).SetInternal(fmt.Errorf("unable to fetch posts"))
	}

	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return dto.NewHTTPError(http.StatusUnauthorized)
	}
	surrogateId, err := jwt2.GetSurrogateKeyFromToken(sessionToken)
	if err != nil {
		return dto.NewHTTPError(http.StatusUnauthorized)
	}
	user, err := h.persister.GetUserPersister().Get(uuid.FromStringOrNil(sessionToken.Subject()))
	emailMaps := map[uuid.UUID]string{}
	result := GetPostsDto{}

	for _, post := range posts {
		if !post.IsActive {
			continue
		}

		dto := PostDto{
			ID:             post.ID,
			CreatedAt:      post.CreatedAt,
			UpdatedAt:      post.UpdatedAt,
			CreatedByEmail: h.GetUserEmail(post.CreatedByUserId, emailMaps),
			Data:           post.Data,
		}

		if user.IsAdmin && (sessionToken.Subject() == surrogateId) {
			createdBySurrogate := h.GetUserEmail(post.CreatedBySurrogateId, emailMaps)
			updatedBySurrogate := h.GetUserEmail(post.UpdatedBySurrogateId, emailMaps)
			dto.CreatedBySurrogate = &createdBySurrogate
			dto.UpdatedBySurrogate = &updatedBySurrogate
		}

		result.Posts = append(result.Posts, dto)
	}

	return c.JSON(http.StatusOK, result)
}

type CreatePostDto struct {
	Body string `json:"body" validate:"required"`
}

func (h *PostHandler) CreatePost(c echo.Context) error {
	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return dto.NewHTTPError(http.StatusUnauthorized)
	}
	surrogateId, err := jwt2.GetSurrogateKeyFromToken(sessionToken)
	if err != nil {
		return dto.NewHTTPError(http.StatusForbidden)
	}
	newPost := CreatePostDto{}
	err = c.Bind(&newPost)
	c.Validate(newPost)
	if err != nil {
		return dto.NewHTTPError(http.StatusBadRequest)
	}
	uId, _ := uuid.NewV4()
	post := models.Post{
		ID:                   uId,
		IsActive:             true,
		CreatedAt:            time.Now().UTC(),
		UpdatedAt:            time.Now().UTC(),
		CreatedByUserId:      uuid.FromStringOrNil(sessionToken.Subject()),
		UpdatedByUserId:      uuid.FromStringOrNil(sessionToken.Subject()),
		CreatedBySurrogateId: uuid.FromStringOrNil(surrogateId),
		UpdatedBySurrogateId: uuid.FromStringOrNil(surrogateId),
		Data:                 newPost.Body,
	}
	err = h.persister.GetPostPersister().Create(post)
	if err != nil {
		return dto.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, struct{}{})
}

func (h *PostHandler) GetUserEmail(userId uuid.UUID, emailMaps map[uuid.UUID]string) string {
	val, exists := emailMaps[userId]
	if exists {
		return val
	}
	user, err := h.persister.GetUserPersister().Get(userId)
	if err != nil {
		return ""
	}
	emailMaps[userId] = user.Email
	return user.Email
}
