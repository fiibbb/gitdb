package handler

import (
	"context"
	"github.com/fiibbb/gitdb/config"
	"github.com/fiibbb/gitdb/proto"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

type GitHandler struct {
	cfg    *config.AppConfig
	repo   *git.Repository
	logger *zap.Logger
}

func NewGitHandler(cfg *config.AppConfig, logger *zap.Logger) (*GitHandler, error) {
	repo, err := git.PlainOpen(cfg.RepoPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &GitHandler{
		cfg:    cfg,
		repo:   repo,
		logger: logger,
	}, nil
}

func (g *GitHandler) Health(ctx context.Context) error {
	return nil
}

func (g *GitHandler) GetObject(ctx context.Context, id *gitdbpb.ObjectIdentifier) (*gitdbpb.Object, error) {
	commit, err := g.resolveCommit(ctx, id)
	if err != nil {
		return nil, err
	}
	root, err := commit.Tree()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	switch id.Type {
	case gitdbpb.ObjectType_BLOB:
		file, err := root.File(id.Path)
		if err != nil {
			return nil, errors.Wrapf(err, "path `%s`", id.Path)
		}
		return convertBlobObject(&file.Blob)
	case gitdbpb.ObjectType_TREE:
		folder, err := root.Tree(id.Path)
		if err != nil {
			return nil, errors.Wrapf(err, "path `%s`", id.Path)
		}
		return convertTreeObject(folder)
	case gitdbpb.ObjectType_COMMIT:
		return convertCommitObject(commit)
	default:
		return nil, errors.Errorf("unrecognized object type `%v`", id.Type)
	}
}

// resolveCommit finds the commit in the ancestry chain identified by refTime.ref
// immediately before or equal to the specified refTime.Time.
func (g *GitHandler) resolveCommit(ctx context.Context, id *gitdbpb.ObjectIdentifier) (*object.Commit, error) {
	refName := plumbing.ReferenceName(id.Ref)
	if refName == "" { // default to master if no ref is specified
		refName = refNameMaster
	}
	ref, err := g.repo.Reference(refName, false)
	if err != nil {
		return nil, errors.Wrapf(err, "ref `%s`", id.Ref)
	}
	commit, err := g.repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, errors.Wrapf(err, "hash `%s`", ref.Hash())
	}
	for {
		if commit.Committer.When.Unix() <= id.Time {
			break
		}
		if commit.NumParents() != 1 {
			return nil, errors.Errorf("no commit found for ref %s time %d, commit %s has %d parent(s)", id.Ref, id.Time, commit.Hash.String(), commit.NumParents())
		}
		commit, err = commit.Parent(0)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	return commit, nil
}
