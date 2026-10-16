package handler

import (
	"context"

	"github.com/vlatan/video-store/internal/avatar"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/domains/user"
	"github.com/vlatan/video-store/internal/drivers/store"
	"github.com/vlatan/video-store/internal/tmpl"
	"github.com/vlatan/video-store/internal/types"
)

type favoriter interface {
	Favorites(context.Context, int, string) (types.Posts, error)
}

type Service struct {
	cfg     *config.Config
	user    *user.Repository
	post    favoriter
	avatar  *avatar.Service
	session *store.Service
	tmpl    *tmpl.Service
}

type userData struct {
	Title string
	types.PaginationInfo
	user.Users
	types.Posts
	tmpl.Base
}

func New(
	cfg *config.Config,
	user *user.Repository,
	post favoriter,
	avatar *avatar.Service,
	store *store.Service,
	tmpl *tmpl.Service,
) *Service {
	return &Service{
		cfg:     cfg,
		user:    user,
		post:    post,
		avatar:  avatar,
		session: store,
		tmpl:    tmpl,
	}
}
