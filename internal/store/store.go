package store

import (
	"crypto/rand"
	"encoding/hex"
	"factual-docs/internal/config"
	client "factual-docs/internal/drivers/redis"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/redis/go-redis/v9"
)

const gothicSessionName = "_gothic_session"

// redisStore implements sessions.Store (New, Get and Save)
type redisStore struct {
	config    *config.Config
	client    client.Service
	keyPrefix string
	maxAge    int
	codecs    []securecookie.Codec
}

func New(
	config *config.Config,
	client client.Service,
	keyPrefix string,
	maxAge int,
	keyPairs ...[]byte) *redisStore {

	store := &redisStore{
		config:    config,
		client:    client,
		keyPrefix: keyPrefix,
		maxAge:    maxAge,
		codecs:    securecookie.CodecsFromPairs(keyPairs...),
	}

	// Add this store to gothic
	gothic.Store = store

	protocol := "https"
	if config.Debug {
		protocol = "http"
	}

	// Add providers to goth
	goth.UseProviders(
		google.New(
			config.GoogleOAuthClientID,
			config.GoogleOAuthClientSecret,
			fmt.Sprintf("%s://%s/auth/google/callback", protocol, config.Domain),
			config.GoogleOAuthScopes...,
		),
	)

	return store
}

// New creates a new session without loading it from the store
func (rs *redisStore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := rs.newSession(name)
	session.IsNew = true
	return session, nil
}

// Get fetches session from Redis or if none creates a new session
func (rs *redisStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	// Create new session object
	session := rs.newSession(name)

	// Get the cookie
	cookie, err := r.Cookie(name)
	if err != nil {
		session.IsNew = true
		return session, nil // New session
	}

	// Get from Redis
	key := rs.buildKey(session.Name(), cookie.Value)
	val, err := rs.client.Get(r.Context(), key)
	if err == redis.Nil {
		session.IsNew = true
		return session, nil // New session
	}

	if err != nil {
		return session, fmt.Errorf("could not get the session from Redis: %w", err)
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

// Save saves a session into Redis and a corresponding session ID in a cookie
func (rs *redisStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {

	// If MaxAge is negative, delete the session
	if session.Options.MaxAge < 0 {
		return rs.deleteSession(r, w, session)
	}

	// Encode session data
	encoded, err := securecookie.EncodeMulti(session.Name(), session.Values, rs.codecs...)
	if err != nil {
		return fmt.Errorf("could not encode the session data: %w", err)
	}

	var sessionID string

	// Get session ID from cookie if it exists
	if cookie, err := r.Cookie(session.Name()); err == nil {
		sessionID = cookie.Value
	} else {
		// Generate new session ID
		sessionID = rs.generateSessionID()
	}

	// Save to Redis
	key := rs.buildKey(session.Name(), sessionID)
	expiration := time.Duration(session.Options.MaxAge) * time.Second
	err = rs.client.Set(r.Context(), key, encoded, expiration)
	if err != nil {
		return fmt.Errorf("could not save the session to Redis: %w", err)
	}

	// Set cookie with session ID
	http.SetCookie(w, &http.Cookie{
		Name:     session.Name(),
		Value:    sessionID,
		Path:     session.Options.Path,
		Domain:   session.Options.Domain,
		MaxAge:   session.Options.MaxAge,
		Secure:   session.Options.Secure,
		HttpOnly: session.Options.HttpOnly,
	})

	return nil
}

// newSession creates a new session object
func (rs *redisStore) newSession(name string) *sessions.Session {
	session := sessions.NewSession(rs, name)

	// Small max age for the gothic session
	maxAge := rs.maxAge
	if session.Name() == gothicSessionName {
		maxAge = 600
	}

	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   !rs.config.Debug,
	}
	return session
}

// deleteSession deletes a session from Redis and deletes the cookie
func (rs *redisStore) deleteSession(
	r *http.Request,
	w http.ResponseWriter,
	session *sessions.Session) error {

	// Delete from redis
	if cookie, err := r.Cookie(session.Name()); err == nil {
		key := rs.buildKey(session.Name(), cookie.Value)
		rs.client.Delete(r.Context(), key)
	}

	// Delete cookie
	http.SetCookie(w, &http.Cookie{
		Name:     session.Name(),
		Value:    "",
		Path:     session.Options.Path,
		Domain:   session.Options.Domain,
		MaxAge:   -1,
		Secure:   session.Options.Secure,
		HttpOnly: session.Options.HttpOnly,
	})

	return nil
}

// generateSessionID generates a random session ID
func (rs *redisStore) generateSessionID() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// buildKey is building a Redis key for the session
func (rs *redisStore) buildKey(sessionName, sessionID string) string {
	return fmt.Sprintf("%s:%s:%s", rs.keyPrefix, sessionName, sessionID)
}
