// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package loader has a data loading interface and various implementations.
package loader

import (
	"do3b/xltemplate/api/git"
	"path/filepath"

	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// NewLoader returns a Loader pointed at the given target.
// If the target is remote, the loader will be restricted
// to the root and below only.  If the target is local, the
// loader will have the restrictions passed in.  Regardless,
// if a local target attempts to transitively load remote bases,
// the remote bases will all be root-only restricted.
func NewLoader(
	lr LoadRestrictorFunc,
	target string, fSys filesys.FileSystem) (*FileLoader, error) {
	repoSpec, err := git.NewRepoSpecFromURL(target)
	if err == nil {
		// The target qualifies as a remote git target.
		return newLoaderAtGitClone(
			repoSpec, fSys, nil, git.ClonerUsingGitExec)
	}
	var root filesys.ConfirmedDir
	cleanedTarget := target
	if !fSys.IsDir(target) {
		if !IsRemoteFile(target) {
			cleanedTarget = filepath.Base(target)
		}
		root, _, err = fSys.CleanedAbs(target)
	} else {
		root, err = filesys.ConfirmDir(fSys, target)
	}

	if err != nil {
		return nil, errors.WrapPrefixf(err, ErrRtNotDir.Error()) //nolint:govet
	}
	return newLoaderAtConfirmedDir(
		lr, root, fSys, nil, git.ClonerUsingGitExec, cleanedTarget), nil
}
