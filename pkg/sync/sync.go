package sync

import (
	"errors"
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

type Sync struct {
	GitDir        string
	GitUser       string
	GitPassword   string
	SSHPrivateKey string
	SSHPassphrase string
	Remote        string
	Branch        string
}

// TODO(alec): do git fetching and sync processing in this package.
// I'd like to abstract this because I imagine multiple places in the operator code
// will need to touch this.
func gitUpdate(sc *Sync) error {
	repo, err := git.PlainOpen(sc.GitDir)
	if err != nil {
		return fmt.Errorf("unable to open local repository %s: %w", sc.GitDir, err)
	}

	// FetchOptions configured with: 1) password, 2) ssh private key, or 3) no auth
	opts := &git.FetchOptions{
		Auth:            nil,
		InsecureSkipTLS: true,
	}
	if sc.GitPassword != "" {
		opts.Auth = &http.BasicAuth{
			Username: sc.GitUser,
			Password: sc.GitPassword,
		}
	} else if sc.SSHPrivateKey != "" {
		opts.Auth, err = ssh.NewPublicKeysFromFile("git", sc.SSHPrivateKey, sc.SSHPassphrase)
		if err != nil {
			return fmt.Errorf("failed to read in ssh private key: %w", err)
		}
	}
	if err := repo.Fetch(opts); err != nil {
		if !errors.Is(git.NoErrAlreadyUpToDate, err) {
			return fmt.Errorf("failed to fetch remote %s: %w", sc.Remote, err)
		}
	}

	wt, err := repo.Worktree()
	if err != nil {
		return err
	}
	branch := plumbing.NewBranchReferenceName(sc.Branch)
	if branch == "" {
		return fmt.Errorf("missing git branch")
	}

	// Attempt a checkout WITH create, but throw away the error. :(
	// NOTE(cm): we throw this error away, because we haven't figured out
	// how to reliably continue when a harmless "branch exists" error is
	// returned. I find this library difficult to use, but a pure Go git
	// implementation is worth it.
	co1 := git.CheckoutOptions{
		Branch: branch,
		Create: true,
		Force:  true,
	}
	wt.Checkout(&co1)

	// Do checkout WITHOUT create. Required for a pull operation.
	co := git.CheckoutOptions{
		Branch: branch,
		Create: false,
		Force:  true,
	}
	if err := wt.Checkout(&co); err != nil {
		return fmt.Errorf("failed to successfully checkout: %w", err)
	}

	// Do the pull
	po := git.PullOptions{
		RemoteName:      "origin",
		ReferenceName:   branch,
		SingleBranch:    true,
		Auth:            opts.Auth,
		Force:           true,
		InsecureSkipTLS: true,
	}
	if err := wt.Pull(&po); err != nil {
		if !errors.Is(err, git.NoErrAlreadyUpToDate) {
			return fmt.Errorf("failed to pull changes from remote: %w", err)
		}
	}

	// Finally, perform a clean, to remove any untracked files from the tree
	if err := wt.Clean(&git.CleanOptions{Dir: true}); err != nil {
		return fmt.Errorf("failed to run git clean: %w", err)
	}

	return nil
}

func cloneIfNeeded(sc *Sync) error {
	var clone bool
	fi, err := os.Stat(sc.GitDir)
	if err != nil {
		if os.IsNotExist(err) {
			clone = true
		} else {
			return fmt.Errorf("failed to stat %s: %w", sc.GitDir, err)
		}
	} else {
		if !fi.IsDir() {
			return fmt.Errorf("%s exists, but is a regular file", sc.GitDir)
		}
	}

	if clone {
		opts := &git.CloneOptions{
			URL: sc.Remote,
		}

		if sc.GitPassword != "" {
			opts.Auth = &http.BasicAuth{
				Username: sc.GitUser,
				Password: sc.GitPassword,
			}
			opts.InsecureSkipTLS = true

			_, err = git.PlainClone(sc.GitDir, false, opts)
			if err != nil {
				return fmt.Errorf("failed to clone with username+password: %w", err)
			}
		} else if sc.SSHPrivateKey != "" {
			auth, err := ssh.NewPublicKeysFromFile("git", sc.SSHPrivateKey, sc.SSHPassphrase)
			if err != nil {
				return fmt.Errorf("failed to find private key from file: %w ", err)
			}
			opts.Auth = auth
			opts.InsecureSkipTLS = true

			_, err = git.PlainClone(sc.GitDir, false, opts)
			if err != nil {
				return fmt.Errorf("failed to clone with ssh: %w", err)
			}
		} else {
			dir, _ := os.Getwd()
			if sc.GitDir != "" {
				dir = sc.GitDir
			}

			if _, err := git.PlainClone(dir, false, opts); err != nil {
				return fmt.Errorf("failed to clone without auth: %w", err)
			}
		}
	}

	return nil
}
