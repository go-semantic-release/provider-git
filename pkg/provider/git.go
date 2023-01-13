package provider

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-semantic-release/semantic-release/v2/pkg/provider"
	"github.com/go-semantic-release/semantic-release/v2/pkg/semrel"
)

var PVERSION = "dev"

type Repository struct {
	defaultBranch string
	taggerName    string
	taggerEmail   string
	remoteName    string
	auth          transport.AuthMethod
	repo          *git.Repository
}

func (repo *Repository) Init(config map[string]string) error {
	repo.defaultBranch = config["default_branch"]
	if repo.defaultBranch == "" {
		repo.defaultBranch = "master"
	}

	repo.taggerName = config["tagger_name"]
	if repo.taggerName == "" {
		repo.taggerName = "semantic-release"
	}

	repo.taggerEmail = config["tagger_email"]
	if repo.taggerEmail == "" {
		repo.taggerEmail = "git@go-semantic-release.xyz"
	}

	repo.remoteName = config["remote_name"]

	if config["auth_username"] == "" {
		config["auth_username"] = "git"
	}

	switch config["auth"] {
	case "basic":
		repo.auth = &http.BasicAuth{
			Username: config["auth_username"],
			Password: config["auth_password"],
		}
	case "ssh":
		auth, err := ssh.NewPublicKeysFromFile(config["auth_username"], config["auth_private_key"], config["auth_password"])
		if err != nil {
			return err
		}
		repo.auth = auth
	default:
		repo.auth = nil
	}

	gitPath := config["git_path"]
	if gitPath == "" {
		gitPath = "."
	}
	gr, err := git.PlainOpen(gitPath)
	if err != nil {
		return err
	}
	repo.repo = gr

	return nil
}

func (repo *Repository) GetInfo() (*provider.RepositoryInfo, error) {
	return &provider.RepositoryInfo{
		Owner:         "",
		Repo:          "",
		DefaultBranch: repo.defaultBranch,
		Private:       false,
	}, nil
}

func (repo *Repository) GetCommits(fromSha, toSha string) ([]*semrel.RawCommit, error) {
	allCommits := make([]*semrel.RawCommit, 0)
	toHash, err := repo.repo.ResolveRevision(plumbing.Revision(toSha))
	if err != nil {
		return nil, err
	}

	commits, err := repo.repo.Log(&git.LogOptions{
		From: *toHash,
	})
	if err != nil {
		return nil, err
	}

	err = commits.ForEach(func(commit *object.Commit) error {
		if commit.Hash.String() == fromSha {
			return storer.ErrStop
		}
		allCommits = append(allCommits, &semrel.RawCommit{
			SHA:        commit.Hash.String(),
			RawMessage: commit.Message,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return allCommits, nil
}

func (repo *Repository) GetReleases(rawRe string) ([]*semrel.Release, error) {
	re := regexp.MustCompile(rawRe)
	allReleases := make([]*semrel.Release, 0)
	tags, err := repo.repo.References()
	if err != nil {
		return nil, err
	}

	err = tags.ForEach(func(reference *plumbing.Reference) error {
		ref := reference.Name().String()
		if !strings.HasPrefix(ref, "refs/tags/") {
			return nil
		}
		tag := strings.TrimPrefix(ref, "refs/tags/")
		if rawRe != "" && !re.MatchString(tag) {
			return nil
		}
		version, semverErr := semver.NewVersion(tag)
		if semverErr != nil {
			return nil
		}

		// resolve annotated tags
		sha := reference.Hash()
		if tagObj, tagErr := repo.repo.TagObject(sha); tagErr == nil {
			if com, commitErr := tagObj.Commit(); commitErr == nil {
				sha = com.Hash
			}
		}

		allReleases = append(allReleases, &semrel.Release{
			SHA:     sha.String(),
			Version: version.String(),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return allReleases, nil
}

func (repo *Repository) CreateRelease(release *provider.CreateReleaseConfig) error {
	var hash plumbing.Hash
	if plumbing.IsHash(release.SHA) {
		hash = plumbing.NewHash(release.SHA)
	} else {
		// hash is not valid, let's assume it is a branch name
		resolvedRef, err := repo.repo.Reference(plumbing.NewBranchReferenceName(release.SHA), true)
		if err != nil {
			return err
		}
		hash = resolvedRef.Hash()
	}
	tag := fmt.Sprintf("v%s", release.NewVersion)
	_, err := repo.repo.CreateTag(tag, hash, &git.CreateTagOptions{
		Message: release.Changelog,
		Tagger: &object.Signature{
			Name:  repo.taggerName,
			Email: repo.taggerEmail,
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}
	err = repo.repo.Push(&git.PushOptions{
		RemoteName: repo.remoteName,
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("refs/tags/%s:refs/tags/%s", tag, tag)),
		},
		Auth: repo.auth,
	})
	return err
}

func (repo *Repository) Name() string {
	return "git"
}

func (repo *Repository) Version() string {
	return PVERSION
}
