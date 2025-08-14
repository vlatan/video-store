package store

import (
	"crypto/rand"
	"encoding/hex"
	"factual-docs/internal/config"
	client "factual-docs/internal/drivers/redis"
	"fmt"
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	config    *config.Config
	client    client.Service
	keyPrefix string
	maxAge    int
	codecs    []securecookie.Codec
}

func NewRedisStore(
	config *config.Config,
	client client.Service,
	keyPrefix string,
	maxAge int,
	keyPairs ...[]byte) *RedisStore {
	return &RedisStore{
		config:    config,
		client:    client,
		keyPrefix: keyPrefix,
		maxAge:    maxAge,
		codecs:    securecookie.CodecsFromPairs(keyPairs...),
	}
}

// New creates a new session without loading it from the store
func (rs *RedisStore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := rs.newSession(name)
	session.IsNew = true
	return session, nil
}

// Get fetches session from Redis or if none creates a new session
func (rs *RedisStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	// Create new session object
	session := rs.newSession(name)

	// Get the cookie
	cookie, err := r.Cookie(name)
	if err != nil {
		session.IsNew = true
		return session, nil // New session
	}

	// Get from Redis
	val, err := rs.client.Get(r.Context(), fmt.Sprintf("%s:%s", rs.keyPrefix, cookie.Value))
	if err == redis.Nil {
		session.IsNew = true
		return session, nil // New session
	}

	if err != nil {
		return session, err
	}

	// Decode session data
	err = securecookie.DecodeMulti(name, val, &session.Values, rs.codecs...)
	if err != nil {
		session.IsNew = true
		return session, nil // New session
	}

	session.IsNew = false
	return session, nil
}

func (s *RedisStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	return nil
}

// newSession creates a new session object
func (rs *RedisStore) newSession(name string) *sessions.Session {
	session := sessions.NewSession(rs, name)
	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   rs.maxAge,
		HttpOnly: true,
		Secure:   !rs.config.Debug,
	}
	return session
}

func (rs *RedisStore) generateSessionID() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
