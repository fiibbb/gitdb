package extended_plumbing

import (
	"github.com/fiibbb/gitdb/consts"
	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"path"
	"time"
)

// ResolveCommit finds the commit in the ancestry chain identified by `reference`
// immediately before or equal to the specified `time`.
func ResolveCommit(repo *git.Repository, reference string, time time.Time) (*object.Commit, error) {
	refName := plumbing.ReferenceName(reference)
	if refName == "" { // default to master if no ref is specified
		refName = consts.RefNameMaster
	}
	ref, err := repo.Reference(refName, false)
	if err != nil {
		return nil, errors.Wrapf(err, "ref `%s`", reference)
	}
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, errors.Wrapf(err, "hash `%s`", ref.Hash())
	}
	for {
		if commit.Committer.When.Before(time) {
			break
		}
		if commit.NumParents() != 1 {
			return nil, errors.Errorf("no commit found for ref %s time %d, commit %s has %d parent(s)", reference, time.String(), commit.Hash.String(), commit.NumParents())
		}
		commit, err = commit.Parent(0)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	return commit, nil
}

func CurrentTree(repo *git.Repository, reference string) (*object.Tree, error) {
	commit, err := ResolveCommit(repo, reference, time.Now())
	if err != nil {
		return nil, err
	}
	tree, err := commit.Tree()
	return tree, errors.Wrapf(err, "reference `%s`", reference)
}

func MakeTree(tree *object.Tree, upserts map[string][]byte, deletes []string) (*object.Tree, error) {
	// Verify upserts and deletes do not overlap.
	for upsertPath := range upserts {
		for _, deletePath := range deletes {
			if upsertPath == deletePath {
				return nil, errors.Errorf("overlapping upsert and delete path `%s`", upsertPath)
			}
		}
	}

	// newAllPaths will hold all paths in the new tree.
	newAllPaths := map[string]struct{}{}
	// 1) Add all existing paths from the old tree.
	if err := tree.Files().ForEach(func(file *object.File) error {
		newAllPaths[path.Clean(file.Name)] = struct{}{}
		return nil
	}); err != nil {
		return nil, errors.WithStack(err)
	}
	// 2) Add all paths from upserts.
	for upsertPath := range upserts {
		newAllPaths[path.Clean(upsertPath)] = struct{}{}
	}
	// 3) Delete all paths from deletes.
	for _, deletePath := range deletes {
		delete(newAllPaths, path.Clean(deletePath))
	}

	return nil, consts.ErrNYI
}
