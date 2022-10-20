package session

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/teamhanko/hanko/backend/config"
	hankoJwk "github.com/teamhanko/hanko/backend/crypto/jwk"
	hankoJwt "github.com/teamhanko/hanko/backend/crypto/jwt"
	"github.com/teamhanko/hanko/backend/persistence"
	"net/http"
	"time"
)

type Manager interface {
	GenerateJWT(uuid.UUID, uuid.UUID, uuid.UUID) (string, error)
	Verify(string) (jwt.Token, error)
	GenerateCookie(token string) (*http.Cookie, error)
	DeleteCookie() (*http.Cookie, error)
}

// Manager is used to create and verify session JWTs
type manager struct {
	jwtGenerator  hankoJwt.Generator
	sessionLength time.Duration
	cookieConfig  cookieConfig
	persister     persistence.Persister
}

type cookieConfig struct {
	Domain   string
	HttpOnly bool
	SameSite http.SameSite
	Secure   bool
}

// NewManager returns a new Manager which will be used to create and verify sessions JWTs
func NewManager(jwkManager hankoJwk.Manager, config config.Session, persister persistence.Persister) (Manager, error) {
	signatureKey, err := jwkManager.GetSigningKey()
	if err != nil {
		return nil, fmt.Errorf("failed to create session generator: %w", err)
	}
	verificationKeys, err := jwkManager.GetPublicKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to create session generator: %w", err)
	}
	g, err := hankoJwt.NewGenerator(signatureKey, verificationKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to create session generator: %w", err)
	}

	duration, _ := time.ParseDuration(config.Lifespan) // error can be ignored, value is checked in config validation
	sameSite := http.SameSite(0)
	switch config.Cookie.SameSite {
	case "lax":
		sameSite = http.SameSiteLaxMode
	case "strict":
		sameSite = http.SameSiteStrictMode
	case "none":
		sameSite = http.SameSiteNoneMode
	default:
		sameSite = http.SameSiteDefaultMode
	}
	return &manager{
		jwtGenerator:  g,
		sessionLength: duration,
		cookieConfig: cookieConfig{
			Domain:   config.Cookie.Domain,
			HttpOnly: config.Cookie.HttpOnly,
			SameSite: sameSite,
			Secure:   config.Cookie.Secure,
		},
		persister: persister,
	}, nil
}

// GenerateJWT creates a new session JWT for the given user
func (g *manager) GenerateJWT(subjectUserId uuid.UUID, surrogateUserId uuid.UUID, grantId uuid.UUID) (string, error) {
	issuedAt := time.Now().UTC()
	var expiration time.Time

	token := jwt.New()
	_ = token.Set(jwt.SubjectKey, subjectUserId.String())
	_ = token.Set(hankoJwt.SurrogateKey, surrogateUserId.String())
	_ = token.Set(jwt.IssuedAtKey, issuedAt)
	_ = token.Set(jwt.ExpirationKey, expiration)

	if grantId != uuid.Nil {
		grant, err := g.persister.GetUserGuestRelationPersister().Get(grantId)
		if err != nil {
			return "", fmt.Errorf("unable to get user guest relationship: %w", err)
		}
		_ = token.Set(hankoJwt.GrantKey, grantId.String())

		grantExpiry := grant.CreatedAt.Add(time.Duration(grant.MinutesAllowed.Int32) * time.Minute)

		if issuedAt.Add(g.sessionLength).Before(grantExpiry) {
			expiration = issuedAt.Add(g.sessionLength)
		} else {
			expiration = grantExpiry
		}
	} else {
		expiration = issuedAt.Add(g.sessionLength)
	}
	_ = token.Set(jwt.ExpirationKey, expiration)
	//_ = token.Set(jwt.AudienceKey, []string{"http://localhost"})

	signed, err := g.jwtGenerator.Sign(token)
	if err != nil {
		return "", err
	}

	return string(signed), nil
}

// Verify verifies the given JWT and returns a parsed one if verification was successful
func (g *manager) Verify(token string) (jwt.Token, error) {
	parsedToken, err := g.jwtGenerator.Verify([]byte(token))
	if err != nil {
		return nil, fmt.Errorf("failed to verify session token: %w", err)
	}

	surrogateId, err := hankoJwt.GetSurrogateKeyFromToken(parsedToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get surrogate id from token: %w", err)
	}

	user, err := g.persister.GetUserPersister().Get(uuid.FromStringOrNil(parsedToken.Subject()))
	if err != nil {
		return nil, fmt.Errorf("failed to get user from database: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user does not exist")
	}

	if surrogateId != parsedToken.Subject() {
		grantId, err := hankoJwt.GetGrantKeyFromToken(parsedToken)
		if err != nil {
			return nil, fmt.Errorf("unable to pull grant key from jwt: %w", err)
		}

		grant, err := g.persister.GetUserGuestRelationPersister().Get(uuid.FromStringOrNil(grantId))
		if err != nil {
			return nil, fmt.Errorf("unable to get grant from database %w", err)
		}
		if !grant.IsActive {
			return nil, fmt.Errorf("grant %s is not active", grant.ID)
		}
	}

	return parsedToken, nil
}

// GenerateCookie creates a new session cookie for the given user
func (g *manager) GenerateCookie(token string) (*http.Cookie, error) {
	return &http.Cookie{
		Name:     "hanko",
		Value:    token,
		Domain:   g.cookieConfig.Domain,
		Path:     "/",
		Secure:   g.cookieConfig.Secure,
		HttpOnly: g.cookieConfig.HttpOnly,
		SameSite: g.cookieConfig.SameSite,
	}, nil
}

func (g *manager) DeleteCookie() (*http.Cookie, error) {
	return &http.Cookie{
		Name:     "hanko",
		Value:    "",
		Domain:   g.cookieConfig.Domain,
		Path:     "/",
		Secure:   g.cookieConfig.Secure,
		HttpOnly: g.cookieConfig.HttpOnly,
		SameSite: g.cookieConfig.SameSite,
		MaxAge:   -1,
	}, nil
}
