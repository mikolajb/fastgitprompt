package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

func red(s string) string {
	return "%F{red}" + s + "%f"
}

func green(s string) string {
	return "%F{green}" + s + "%f"
}

func black(s string) string {
	return "%F{black}" + s + "%f"
}

func blue(s string) string {
	return "%F{blue}" + s + "%f"
}

func yellow(s string) string {
	return "%F{yellow}" + s + "%f"
}

func magenta(s string) string {
	return "%F{magenta}" + s + "%f"
}

func aheadBehind(repository *git.Repository, one, two plumbing.Hash) (ahead, behind int, err error) {
	oneVisited := make(map[plumbing.Hash]struct {
		exists bool
		count  int
	})

	oneCommitIter, err := repository.Log(&git.LogOptions{From: one})
	if err != nil {
		return 0, 0, err
	}
	oneCommitIter.ForEach(func(c *object.Commit) error {
		oneVisited[c.Hash] = struct {
			exists bool
			count  int
		}{
			exists: true,
			count:  behind,
		}
		behind++
		return nil
	})
	twoCommitIter, err := repository.Log(&git.LogOptions{From: two})
	if err != nil {
		return
	}
	for {
		var c *object.Commit
		c, err = twoCommitIter.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return
		}

		if oneVisited[c.Hash].exists {
			return ahead, oneVisited[c.Hash].count, nil
		}

		ahead++
	}
	return
}

func Prompt(repository *git.Repository) (string, error) {
	head, err := repository.Head()
	if err != nil {
		panic(err)
	}

	branchName := []string{head.Name().Short()}
	if head.Name() != plumbing.Master {
		branches := []plumbing.ReferenceName{plumbing.Master, head.Name()}
		hashes := []plumbing.Hash{}

		for _, branch := range branches {
			ref, err := repository.Reference(branch, true)
			if err != nil {
				panic(err)
			}
			hashes = append(hashes, ref.Hash())
		}
		ahead, behind, err := aheadBehind(repository, hashes[0], hashes[1])
		if err != nil {
			panic(err)
		}

		if ahead > 0 || behind > 0 {
			prefix := ""
			if behind > 0 {
				prefix += strconv.Itoa(behind) + magenta("↓")
			}
			if ahead > 0 {
				prefix += magenta("↑") + strconv.Itoa(ahead)
			}
			branchName = append([]string{prefix, " "}, branchName...)
		}
	}

	result := append([]string{black("git:(")}, branchName...)
	worktree, err := repository.Worktree()
	if err != nil {
		if err == git.ErrIsBareRepository {
			result = append(result, magenta("#bare"))
		} else {
			panic(err)
		}
	} else {
		status, err := worktree.Status()
		if err != nil {
			panic(err)
		}
		staged := make(map[git.StatusCode]int)
		unstaged := make(map[git.StatusCode]int)
		for _, fileStatus := range status {
			if fileStatus.Staging != ' ' {
				staged[fileStatus.Staging]++
			}
			if fileStatus.Worktree != ' ' {
				unstaged[fileStatus.Worktree]++
			}
		}

		for _, spec := range []struct {
			colorize func(string) string
			changes  map[git.StatusCode]int
		}{
			{
				colorize: green,
				changes:  staged,
			},
			{
				colorize: red,
				changes:  unstaged,
			},
		} {
			if len(spec.changes) > 0 {
				result = append(result, " ")
			}
			for _, mod := range []git.StatusCode{
				git.Untracked,
				git.Modified,
				git.Added,
				git.Deleted,
				git.Renamed,
				git.Copied,
				git.UpdatedButUnmerged,
			} {
				if spec.changes[mod] > 0 {
					result = append(
						result,
						strconv.Itoa(spec.changes[mod]),
						spec.colorize(string(mod)),
					)
				}
			}
		}
	}
	result = append(result, black(")"))
	return strings.Join(result, ""), nil
}

func main() {
	var repository *git.Repository
	wd, err := os.Getwd()
	if err != nil {
		os.Exit(0)
	}
	for {
		if wd == "/" {
			os.Exit(0)
		}
		repository, err = git.PlainOpen(wd)
		if err != nil {
			if err == git.ErrRepositoryNotExists {
				wd = path.Join(wd, "..")
			} else {
				panic(err)
			}
		} else {
			break
		}
	}
	prompt, err := Prompt(repository)
	if err != nil {
		panic(err)
	}
	fmt.Print(" " + prompt)
}
