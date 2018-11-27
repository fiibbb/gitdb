package handler

import (
	"github.com/fiibbb/gitdb/gitpb"
	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"io/ioutil"
)

/*
 * This file most exists because we want to convert between gitpb defined types
 * and go-git defined types.
 */

func convertBlob(b *object.Blob) (*gitpb.Blob, error) {
	reader, err := b.Reader()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer reader.Close()
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	blob := &gitpb.Blob{
		Hash:    b.Hash.String(),
		Content: content,
	}
	return blob, nil
}

func convertTree(t *object.Tree) (*gitpb.Tree, error) {
	var entries []*gitpb.TreeEntry
	for _, entry := range t.Entries {
		entries = append(entries, &gitpb.TreeEntry{
			Name: entry.Name,
			Hash: entry.Hash.String(),
			Mode: uint32(entry.Mode),
		})
	}
	tree := &gitpb.Tree{
		Hash:    t.Hash.String(),
		Entries: entries,
	}
	return tree, nil
}

func convertCommit(c *object.Commit) (*gitpb.Commit, error) {
	root, err := c.Tree()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	convertedRoot, err := convertTree(root)
	if err != nil {
		return nil, err
	}
	var parents []string
	for _, hash := range c.ParentHashes {
		parents = append(parents, hash.String())
	}
	convertSignature := func(a *object.Signature) *gitpb.Signature {
		return &gitpb.Signature{
			Name:  a.Name,
			Email: a.Email,
			Time:  a.When.Unix(),
		}
	}
	commit := &gitpb.Commit{
		Hash:      c.Hash.String(),
		Author:    convertSignature(&c.Author),
		Committer: convertSignature(&c.Committer),
		Message:   c.Message,
		Tree:      root.Hash.String(),
		Parents:   parents,
		TreeObject: &gitpb.Object{
			Obj: &gitpb.Object_Tree{
				Tree: convertedRoot,
			},
		},
	}
	return commit, nil
}

func convertBlobObject(b *object.Blob) (*gitpb.Object, error) {
	blob, err := convertBlob(b)
	if err != nil {
		return nil, err
	}
	return &gitpb.Object{Obj: &gitpb.Object_Blob{Blob: blob}}, nil
}

func convertTreeObject(t *object.Tree) (*gitpb.Object, error) {
	tree, err := convertTree(t)
	if err != nil {
		return nil, err
	}
	return &gitpb.Object{Obj: &gitpb.Object_Tree{Tree: tree}}, nil
}

func convertCommitObject(c *object.Commit) (*gitpb.Object, error) {
	commit, err := convertCommit(c)
	if err != nil {
		return nil, err
	}
	return &gitpb.Object{Obj: &gitpb.Object_Commit{Commit: commit}}, nil
}
