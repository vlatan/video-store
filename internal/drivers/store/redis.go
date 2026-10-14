package store

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/redis/go-redis/v9"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/rdb"
)

// redisStore implements sessions.Store (New, Get and Save)
type redisStore struct {
	config    *config.Config
	rdb       *rdb.Service
	keyPrefix string
	maxAge    int
	codec     securecookie.Codec
}

func newRedisStore(
	config *config.Config,
	rdb *rdb.Service,
	keyPrefix string,
	maxAge int) *redisStore {

	return &redisStore{
		config:    config,
		rdb:       rdb,
		keyPrefix: keyPrefix,
		maxAge:    maxAge,
		codec: securecookie.New(
			config.AuthKey.Bytes,
			config.EncryptionKey.Bytes,
		),
	}

}

// New creates a new session without loading it from the store
func (rs *redisStore) New(r *http.Request, name string) (*sessions.Session, error) {

	// Create new gorilla session object, provided the custom Redis store
	session := sessions.NewSession(rs, name)

	// Small max age to 10 minutes for all sessions
	// other than for the user session
	maxAge := 600
	if session.Name() == rs.config.UserSessionName {
		maxAge = rs.maxAge
	}

	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   maxAge,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	session.IsNew = true
	return session, nil
}

// Get fetches session from Redis or if none creates a new session
func (rs *redisStore) Get(r *http.Request, name string) (*sessions.Session, error) {

	// Create new session object.
	// Err is always nil, session.IsNew is set to true.
	session, _ := rs.New(r, name)

	// Check if there's cookie
	cookie, err := r.Cookie(name)
	if err != nil {
		return session, nil
	}

	// Get from Redis
	key := rs.buildKey(session.Name(), cookie.Value)
	val, err := rs.rdb.Client.Get(r.Context(), key).Result()
	if err == redis.Nil {
		return session, nil
	}

	if err != nil {
		return session, fmt.Errorf("could not get the session from Redis: %w", err)
	}

	// Decode session data
	if err = rs.codec.Decode(name, val, &session.Values); err != nil {
		return session, fmt.Errorf("could not decode the session from Redis: %w", err)
	}

	session.ID = cookie.Value
	session.IsNew = false
	return session, nil
}

// Save saves a session into Redis and a corresponding session ID in a cookie
func (rs *redisStore) Save(
	r *http.Request,
	w http.ResponseWriter,
	session *sessions.Session) error {

	// If MaxAge is negative, delete the session
	if session.Options.MaxAge <= 0 {
		if err := rs.delete(r, w, session); err != nil {
			return fmt.Errorf("could not delete the session: %w", err)
		}
		return nil
	}

	// Encode session data
	encoded, err := rs.codec.Encode(session.Name(), session.Values)
	if err != nil {
		return fmt.Errorf("could not encode the session data: %w", err)
	}

	if session.ID == "" {
		session.ID, err = rs.generateSessionID()
		if err != nil {
			return fmt.Errorf("could not generate session ID: %w", err)
		}
	}

	// Save to Redis
	key := rs.buildKey(session.Name(), session.ID)
	expiration := time.Duration(session.Options.MaxAge) * time.Second
	err = rs.rdb.Client.Set(r.Context(), key, encoded, expiration).Err()
	if err != nil {
		return fmt.Errorf("could not save the session to Redis: %w", err)
	}

	// Set cookie with session ID
	http.SetCookie(w, sessions.NewCookie(
		session.Name(), session.ID, session.Options,
	))

	return nil
}

// deleteSession deletes a session from Redis and deletes the cookie
func (rs *redisStore) delete(
	r *http.Request,
	w http.ResponseWriter,
	session *sessions.Session) error {

	// Skip if session was never persisted
	if session.IsNew || session.ID == "" {
		return nil
	}
	// Delete from redis
	key := rs.buildKey(session.Name(), session.ID)
	if err := rs.rdb.Client.Del(r.Context(), key).Err(); err != nil {
		return err
	}

	// Delete the cookie
	http.SetCookie(w, sessions.NewCookie(
		session.Name(), "", session.Options,
	))

	return nil
}

// generateSessionID generates a random session ID
func (rs *redisStore) generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// buildKey is building a Redis key for the session
func (rs *redisStore) buildKey(sessionName, sessionID string) string {
	return fmt.Sprintf("%s:%s:%s", rs.keyPrefix, sessionName, sessionID)
}
