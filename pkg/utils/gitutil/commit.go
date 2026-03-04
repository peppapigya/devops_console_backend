package gitutil

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// GetCommitID 获取当前仓库的commitID
func GetCommitID(path string, branch string) (string, error) {
	if path == "" {
		return "", nil
	}
	repo, err := git.PlainOpen(path)
	if err != nil {
		return "", err
	}
	head, err := repo.Reference(plumbing.NewBranchReferenceName(branch), true)
	if err != nil {
		return "", err
	}
	return head.Hash().String(), nil
}
