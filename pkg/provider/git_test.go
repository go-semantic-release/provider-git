package provider

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-semantic-release/semantic-release/v2/pkg/provider"
	"github.com/stretchr/testify/require"
)

var testGitPath string

func TestGit(t *testing.T) {
	var err error
	testGitPath, err = setupRepo()
	require.NoError(t, err)
	t.Run("NewRepository", newRepository)
	t.Run("GetInfo", getInfo)
	t.Run("GetReleases", getReleases)
	t.Run("GetCommits", getCommits)
	t.Run("GetCommitsNoFFMerge", getCommitsNoFFMerge)
	t.Run("GetCommitsNoFFMergeCTime", getCommitsNoFFMergeCTime)
	t.Run("CreateRelease", createRelease)
}

func newRepository(t *testing.T) {
	require := require.New(t)
	repo := &Repository{}
	err := repo.Init(map[string]string{})
	require.EqualError(err, "repository does not exist")

	repo = &Repository{}
	err = repo.Init(map[string]string{
		"git_path":       testGitPath,
		"default_branch": "development",
		"tagger_name":    "test",
		"tagger_email":   "test@test.com",
		"auth":           "basic",
		"auth_username":  "test",
		"auth_password":  "test",
	})

	require.NoError(err)
	require.Equal("development", repo.defaultBranch)
	require.Equal("test", repo.taggerName)
	require.Equal("test@test.com", repo.taggerEmail)
	require.NotNil(repo.auth)
}

var gitCommitAuthor = &object.Signature{
	Name:  "test",
	Email: "test@test.com",
	When:  time.Now(),
}

//gocyclo:ignore
func setupRepo() (string, error) {
	dir, err := os.MkdirTemp("", "provider-git")
	if err != nil {
		return "", err
	}
	repo, err := git.PlainInit(dir, false)
	if err != nil {
		return "", err
	}

	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{"http://localhost:3000/test/test.git"},
	})
	if err != nil {
		return "", err
	}
	w, err := repo.Worktree()
	if err != nil {
		return "", err
	}

	versionCount := 0
	betaCount := 1
	for i := 0; i < 100; i++ {
		commit, commitErr := w.Commit(fmt.Sprintf("feat: commit %d", i), &git.CommitOptions{Author: gitCommitAuthor, AllowEmptyCommits: true})
		if commitErr != nil {
			return "", err
		}
		if i%10 == 0 {
			if _, tagErr := repo.CreateTag(fmt.Sprintf("v1.%d.0", versionCount), commit, nil); tagErr != nil {
				return "", tagErr
			}
			versionCount++
		}
		if i%5 == 0 {
			if _, tagErr := repo.CreateTag(fmt.Sprintf("v2.0.0-beta.%d", betaCount), commit, nil); tagErr != nil {
				return "", tagErr
			}
			betaCount++
		}
	}

	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("new-fix"),
		Create: true,
	})
	if err != nil {
		return "", err
	}

	if _, err = w.Commit("fix: error", &git.CommitOptions{Author: gitCommitAuthor, AllowEmptyCommits: true}); err != nil {
		return "", err
	}
	if err = w.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName("master")}); err != nil {
		return "", err
	}

	err = repo.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs: []config.RefSpec{
			"refs/heads/*:refs/heads/*",
			"refs/tags/*:refs/tags/*",
		},
		Auth: &http.BasicAuth{
			Username: "test",
			Password: "test",
		},
		Force: true,
	})
	if err != nil {
		return "", err
	}
	return dir, nil
}

func createRepo() (*Repository, error) {
	repo := &Repository{}
	err := repo.Init(map[string]string{
		"git_path":      testGitPath,
		"auth":          "basic",
		"auth_username": "test",
		"auth_password": "test",
	})
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func cloneRepo(path, url string) (*Repository, error) {
	_, err := git.PlainClone(path, false, &git.CloneOptions{
		Auth: &http.BasicAuth{
			Username: "test",
			Password: "test",
		},
		URL: url,
	})
	if err != nil {
		return nil, err
	}
	repo := &Repository{}
	err = repo.Init(map[string]string{
		"git_path":      path,
		"auth":          "basic",
		"auth_username": "test",
		"auth_password": "test",
	})
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func getInfo(t *testing.T) {
	require := require.New(t)
	repo, err := createRepo()
	require.NoError(err)
	repoInfo, err := repo.GetInfo()
	require.NoError(err)
	require.Equal("master", repoInfo.DefaultBranch)
}

func getCommits(t *testing.T) {
	require := require.New(t)
	repo, err := createRepo()
	require.NoError(err)
	commits, err := repo.GetCommits("", "master")
	require.NoError(err)
	require.Len(commits, 100)

	for _, c := range commits {
		require.True(strings.HasPrefix(c.RawMessage, "feat: commit"))
		require.Equal(gitCommitAuthor.Name, c.Annotations["author_name"])
		require.Equal(gitCommitAuthor.Email, c.Annotations["author_email"])
		require.Equal(gitCommitAuthor.When.Format(time.RFC3339), c.Annotations["author_date"])
		require.Equal(gitCommitAuthor.When.Format(time.RFC3339), c.Annotations["committer_date"])
		require.Equal(gitCommitAuthor.Name, c.Annotations["committer_name"])
		require.Equal(gitCommitAuthor.Email, c.Annotations["committer_email"])
	}
}

func getCommitsNoFFMerge(t *testing.T) {
	require := require.New(t)
	dir, err := os.MkdirTemp("", "provider-git")
	require.NoError(err)
	repo, err := cloneRepo(dir, "http://localhost:3000/test/no_ff_merge.git")
	require.NoError(err)
	releases, err := repo.GetReleases("")
	require.NoError(err)
	require.Len(releases, 1)
	initialCommitSha := releases[0].GetSHA()
	commits, err := repo.GetCommits(initialCommitSha, "master")
	require.NoError(err)
	require.Len(commits, 1)
}

func getCommitsNoFFMergeCTime(t *testing.T) {
	require := require.New(t)
	dir, err := os.MkdirTemp("", "provider-git")
	require.NoError(err)
	repo, err := cloneRepo(dir, "http://localhost:3000/test/no_ff_merge.git")
	repo.logOrder = git.LogOrderCommitterTime
	require.NoError(err)
	releases, err := repo.GetReleases("")
	require.NoError(err)
	require.Len(releases, 1)
	initialCommitSha := releases[0].GetSHA()
	commits, err := repo.GetCommits(initialCommitSha, "master")
	require.NoError(err)
	require.Len(commits, 2)
}

func createRelease(t *testing.T) {
	require := require.New(t)
	repo, err := createRepo()
	require.NoError(err)

	gRepo, err := git.PlainOpen(testGitPath)
	require.NoError(err)
	head, err := gRepo.Head()
	require.NoError(err)

	testCases := []struct {
		version, sha, changelog string
	}{
		{
			version:   "2.0.0",
			sha:       head.Hash().String(),
			changelog: "new feature",
		},
		{
			version:   "3.0.0",
			sha:       "master",
			changelog: "breaking change",
		},
	}

	for _, testCase := range testCases {
		err = repo.CreateRelease(&provider.CreateReleaseConfig{
			NewVersion: testCase.version,
			SHA:        testCase.sha,
			Changelog:  testCase.changelog,
		})
		require.NoError(err)
		tagName := "v" + testCase.version

		tagRef, err := gRepo.Tag(tagName)
		require.NoError(err)

		tagObj, err := gRepo.TagObject(tagRef.Hash())
		require.NoError(err)

		require.Equal(testCase.changelog+"\n", tagObj.Message)

		// Clean up tags so future test runs succeed
		tagRefName := ":refs/tags/" + tagName
		err = gRepo.Push(&git.PushOptions{
			RemoteName: "origin",
			RefSpecs:   []config.RefSpec{config.RefSpec(tagRefName)},
			Auth: &http.BasicAuth{
				Username: "test",
				Password: "test",
			},
		})
		require.NoError(err)
	}
}

func getReleases(t *testing.T) {
	require := require.New(t)
	repo, err := createRepo()
	require.NoError(err)

	releases, err := repo.GetReleases("")
	require.NoError(err)
	require.Len(releases, 30)

	releases, err = repo.GetReleases("^v2")
	require.NoError(err)
	require.Len(releases, 20)
}
