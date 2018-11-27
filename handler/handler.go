package handler

import (
	"context"
	"github.com/fiibbb/gitdb/config"
	"github.com/fiibbb/gitdb/gitpb"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"sync"
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

func (g *GitHandler) read(name string, f func(*git.Repository) error) error {
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

func (g *GitHandler) write(name string, f func(*git.Repository) error) error {
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

func (g *GitHandler) WriteCommit(ctx context.Context, repoName string, branchName string, updates map[string][]byte, deletes []string, msg string) (*gitpb.Commit, error) {
	//if err := g.write(repoName, func(repo *git.Repository) error {
	//	wt, err := repo.Worktree()
	//	if err != nil {
	//		return errors.WithStack(err)
	//	}
	//	if err := wt.Clean(&git.CleanOptions{Dir: true}); err != nil {
	//		return errors.WithStack(err)
	//	}
	//	for p, data := range updates {
	//		if err := writeFile(repo, p, data); err != nil {
	//			return err
	//		}
	//		if _, err := wt.Add(p); err != nil {
	//			return errors.Wrapf(err, "path `%s`", p)
	//		}
	//	}
	//	for _, p := range deletes {
	//		if err := deleteFile(repo, p); err != nil {
	//			return err
	//		}
	//		if _, err := wt.Add(p); err != nil {
	//			return errors.Wrapf(err, "path `%s`", p)
	//		}
	//	}
	//	wt.Commit(msg, &git.CommitOptions{All: true})
	//	return nil
	//}); err != nil {
	//	return nil, err
	//}
	return nil, ErrNYI
}

func (g *GitHandler) GetObject(ctx context.Context, id *gitpb.ObjectIdentifier) (*gitpb.Object, error) {
	var obj *gitpb.Object
	if err := g.read(id.Repo, func(repo *git.Repository) error {
		commit, err := resolveCommit(repo, id)
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

// resolveCommit finds the commit in the ancestry chain identified by refTime.ref
// immediately before or equal to the specified refTime.Time.
func resolveCommit(repo *git.Repository, id *gitpb.ObjectIdentifier) (*object.Commit, error) {
	refName := plumbing.ReferenceName(id.Ref)
	if refName == "" { // default to master if no ref is specified
		refName = refNameMaster
	}
	ref, err := repo.Reference(refName, false)
	if err != nil {
		return nil, errors.Wrapf(err, "ref `%s`", id.Ref)
	}
	commit, err := repo.CommitObject(ref.Hash())
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
