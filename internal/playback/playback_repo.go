package playback

import (
	"context"
	"errors"
	"sync"
	"time"
)

const (
	sessionDocumentId  = "default"
	sessionIdleTimeout = 60 * time.Minute
)

var errSessionNotFound = errors.New("playback session not found")

type playbackRepository interface {
	Get(ctx context.Context) (*playbackSession, error)
	Create(ctx context.Context, session *playbackSession) error
	Save(ctx context.Context, session *playbackSession) error
}

type repo struct {
	mu        sync.RWMutex
	session   *playbackSession
	expiresAt time.Time
}

var playbackRepoInstance = &repo{}

func (r *repo) Get(ctx context.Context) (*playbackSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.session == nil || r.isExpired(time.Now()) {
		return nil, errSessionNotFound
	}

	sessionCopy := *r.session
	return &sessionCopy, nil
}

func (r *repo) Create(ctx context.Context, session *playbackSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	sessionCopy := *session
	r.session = &sessionCopy
	r.expiresAt = time.Now().Add(sessionIdleTimeout)

	return nil
}

func (r *repo) Save(ctx context.Context, session *playbackSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	sessionCopy := *session
	r.session = &sessionCopy
	r.expiresAt = time.Now().Add(sessionIdleTimeout)

	return nil
}

func (r *repo) isExpired(now time.Time) bool {
	return !r.expiresAt.IsZero() && now.After(r.expiresAt)
}

func newPlaybackRepository(ctx context.Context) (playbackRepository, error) {
	return playbackRepoInstance, nil
}
