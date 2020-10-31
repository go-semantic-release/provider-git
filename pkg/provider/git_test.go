package provider

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-semantic-release/semantic-release/v2/pkg/provider"
	"github.com/stretchr/testify/require"
)

func TestNewRepository(t *testing.T) {
	require := require.New(t)
	repo := &Repository{}
	err := repo.Init(map[string]string{})
	require.EqualError(err, "repository does not exist")

	gitPath, err := setupRepo()
	require.NoError(err)

	repo = &Repository{}
	err = repo.Init(map[string]string{
		"git_path":       gitPath,
		"default_branch": "development",
		"tagger_name":    "test",
		"tagger_email":   "test@test.com",
	})
	require.NoError(err)
	require.Equal(repo.defaultBranch, "development")
	require.Equal(repo.taggerName, "test")
	require.Equal(repo.taggerEmail, "test@test.com")
}

func setupRepo() (string, error) {
	dir, err := ioutil.TempDir("", "provider-git")
	if err != nil {
		return "", err
	}
	repo, err := git.PlainInit(dir, false)
	if err != nil {
		return "", err
	}
	w, err := repo.Worktree()
	if err != nil {
		return "", err
	}

	author := &object.Signature{
		Name:  "test",
		Email: "test@test.com",
		When:  time.Now(),
	}
	versionCount := 0
	for i := 0; i < 100; i++ {
		commit, err := w.Commit(fmt.Sprintf("feat: commit %d", i), &git.CommitOptions{Author: author})
		if err != nil {
			return "", err
		}
		if i%10 == 0 {
			if _, err := repo.CreateTag(fmt.Sprintf("v1.%d.0", versionCount), commit, nil); err != nil {
				return "", err
			}
			versionCount++
		}
	}

	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("new-fix"),
		Create: true,
	})
	if err != nil {
		return "", err
	}

	if _, err = w.Commit("fix: error", &git.CommitOptions{Author: author}); err != nil {
		return "", err
	}
	if err = w.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName("master")}); err != nil {
		return "", err
	}

	return dir, nil
}

func createRepo() (*Repository, string, error) {
	gitPath, err := setupRepo()
	if err != nil {
		return nil, "", err
	}

	repo := &Repository{}
	err = repo.Init(map[string]string{
		"git_path": gitPath,
	})
	if err != nil {
		return nil, "", err
	}

	return repo, gitPath, nil
}

func TestGetInfo(t *testing.T) {
	require := require.New(t)
	repo, _, err := createRepo()
	require.NoError(err)
	repoInfo, err := repo.GetInfo()
	require.NoError(err)
	require.Equal("master", repoInfo.DefaultBranch)
}

func TestGetCommits(t *testing.T) {
	require := require.New(t)
	repo, _, err := createRepo()
	require.NoError(err)
	commits, err := repo.GetCommits("", "master")
	require.NoError(err)
	require.Len(commits, 100)

	for _, c := range commits {
		require.True(strings.HasPrefix(c.RawMessage, "feat: commit"))
	}
}

func TestGithubCreateRelease(t *testing.T) {
	require := require.New(t)
	repo, gitPath, err := createRepo()
	require.NoError(err)

	gRepo, err := git.PlainOpen(gitPath)
	require.NoError(err)
	head, err := gRepo.Head()
	require.NoError(err)

	err = repo.CreateRelease(&provider.CreateReleaseConfig{
		NewVersion: "2.0.0",
		SHA:        head.Hash().String(),
		Changelog:  "new feature",
	})
	require.NoError(err)

	tagRef, err := gRepo.Tag("v2.0.0")
	require.NoError(err)

	tagObj, err := gRepo.TagObject(tagRef.Hash())
	require.NoError(err)

	require.Equal("new feature\n", tagObj.Message)
}
