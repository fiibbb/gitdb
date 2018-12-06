package handler

import (
	"context"
	"github.com/fiibbb/gitdb/config"
	"github.com/fiibbb/gitdb/consts"
	ep "github.com/fiibbb/gitdb/extended_plumbing"
	"github.com/fiibbb/gitdb/gitpb"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/src-d/go-git.v4"
	"sync"
	"time"
)

type repository struct {
	sync.RWMutex
	*git.Repository
}

type GitHandler struct {
	cfg    *config.AppConfig
	repos  sync.Map
	logger *zap.Logger
}

func NewGitHandler(cfg *config.AppConfig, logger *zap.Logger) (*GitHandler, error) {
	return &GitHandler{
		cfg:    cfg,
		repos:  sync.Map{},
		logger: logger,
	}, nil
}

func (g *GitHandler) shared(name string, f func(*git.Repository) error) error {
	repoPtr, ok := g.repos.Load(name)
	if !ok {
		return errors.Errorf("repo not found `%s`", name)
	}
	repo, ok := repoPtr.(*repository)
	if !ok {
		panic("WTF")
	}
	repo.RLock()
	defer repo.RUnlock()
	return f(repo.Repository)
}

func (g *GitHandler) exclusive(name string, f func(*git.Repository) error) error {
	repoPtr, ok := g.repos.Load(name)
	if !ok {
		return errors.Errorf("repo not found `%s`", name)
	}
	repo, ok := repoPtr.(*repository)
	if !ok {
		panic("WTF")
	}
	repo.Lock()
	defer repo.Unlock()
	return f(repo.Repository)
}

func (g *GitHandler) Health(ctx context.Context) error {
	return nil
}

func (g *GitHandler) WriteCommit(ctx context.Context, repo string, ref string, upserts map[string][]byte, deletes []string, msg string) (*gitpb.Commit, error) {
	if ref == "" {
		ref = consts.RefNameMaster
	}
	if err := g.shared(repo, func(r *git.Repository) error {
		return nil
	}); err != nil {
		return nil, err
	}
	return nil, consts.ErrNYI
}

func (g *GitHandler) GetObject(ctx context.Context, id *gitpb.ObjectIdentifier) (*gitpb.Object, error) {
	var obj *gitpb.Object
	if err := g.shared(id.Repo, func(repo *git.Repository) error {
		commit, err := ep.ResolveCommit(repo, id.Ref, time.Unix(id.Time, 0))
		if err != nil {
			return err
		}
		root, err := commit.Tree()
		if err != nil {
			return errors.WithStack(err)
		}
		switch id.Type {
		case gitpb.ObjectType_BLOB:
			file, err := root.File(id.Path)
			if err != nil {
				return errors.Wrapf(err, "path `%s`", id.Path)
			}
			obj, err = convertBlobObject(&file.Blob)
			return err
		case gitpb.ObjectType_TREE:
			folder, err := root.Tree(id.Path)
			if err != nil {
				return errors.Wrapf(err, "path `%s`", id.Path)
			}
			obj, err = convertTreeObject(folder)
			return err
		case gitpb.ObjectType_COMMIT:
			obj, err = convertCommitObject(commit)
			return err
		default:
			return errors.Errorf("unrecognized object type `%v`", id.Type)
		}
	}); err != nil {
		return nil, err
	}
	return obj, nil
}
