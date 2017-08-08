package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	git "gopkg.in/libgit2/git2go.v26"
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

type GitPromptError string

func (gpe GitPromptError) Error() string {
	return string(gpe)
}

func Branch(repository *git.Repository) ([]string, error) {
	is_detached, err := repository.IsHeadDetached()
	if err != nil {
		panic(err)
	} else if is_detached {
		return []string{"head detached"}, GitPromptError("head detached")
	}

	head, err := repository.Head()
	if err != nil {
		if git.IsErrorCode(err, git.ErrUnbornBranch) {
			return []string{"no head"}, GitPromptError("no head")
		}
		panic(err)
	}

	branch_name_string, err := head.Branch().Name()
	if err != nil {
		panic(err)
	}
	branch_name := []string{branch_name_string}

	masterBranch, err := repository.LookupBranch("master", git.BranchLocal)
	if err != nil {
		panic(err)
	}

	isHead, err := masterBranch.IsHead()
	if err != nil {
		panic(err)
	}
	if !isHead {
		ahead, behind, err := repository.AheadBehind(head.Target(), masterBranch.Target())
		if err != nil {
			panic(err)
		}
		if ahead > 0 && behind > 0 {
			branch_name = append([]string{"m " + magenta("↔") + fmt.Sprintf(" %d/%d ", ahead, behind)}, branch_name...)
		} else {
			if behind > 0 {
				branch_name = append([]string{"m " + magenta("→") + fmt.Sprintf(" %d ", behind)}, branch_name...)
			}
			if ahead > 0 {
				branch_name = append([]string{"m " + magenta("←") + fmt.Sprintf(" %d ", ahead)}, branch_name...)
			}
		}
	}

	upstream, err := head.Branch().Upstream()
	if err != nil {
		if git.IsErrorCode(err, git.ErrNotFound) {
			branch_name = append([]string{" "}, branch_name...)
			branch_name = append([]string{red("⚡")}, branch_name...)
			branch_name = append([]string{"upstream "}, branch_name...)
		} else {
			panic(err)
		}
	} else {
		ahead, behind, err := repository.AheadBehind(head.Target(), upstream.Target())
		if err != nil {
			panic(err)

		}
		behind_string := fmt.Sprintf(" %d", behind)
		ahead_string := fmt.Sprintf(" %d", ahead)
		if behind > 0 && ahead > 0 {
			branch_name = append(branch_name, behind_string, yellow("⇵"), ahead_string)
		} else {
			if behind > 0 {
				branch_name = append(branch_name, behind_string, red("↓"))
			}
			if ahead > 0 {
				branch_name = append(branch_name, ahead_string, green("↑"))
			}
		}

	}
	return branch_name, nil
}

type RepoState struct {
	Untracked,
	NewFiles,
	Deletions,
	DeletionsStaged,
	Modifications,
	ModificationsStaged,
	Renames,
	RenamesStaged,
	ConflictsBoth,
	ConflictsOur,
	ConflictsTheir int
}

func (repoState RepoState) Format() []string {
	result := []string{}

	if repoState.ConflictsBoth > 0 {
		result = append(result, fmt.Sprintf(" %d", repoState.ConflictsBoth), blue("B"))
	} else if repoState.ConflictsOur > 0 {
		result = append(result, fmt.Sprintf(" %d", repoState.ConflictsOur), blue("U"))
	} else if repoState.ConflictsTheir > 0 {
		result = append(result, fmt.Sprintf(" %d", repoState.ConflictsTheir), blue("T"))
	}

	staged := []string{}
	if repoState.NewFiles > 0 {
		staged = append(staged, fmt.Sprintf("%d", repoState.NewFiles), green("N"))
	}
	if repoState.ModificationsStaged > 0 {
		staged = append(staged, fmt.Sprintf("%d", repoState.ModificationsStaged), green("M"))
	}
	if repoState.RenamesStaged > 0 {
		staged = append(staged, fmt.Sprintf("%d", repoState.RenamesStaged), green("R"))
	}
	if repoState.DeletionsStaged > 0 {
		staged = append(staged, fmt.Sprintf("%d", repoState.DeletionsStaged), green("D"))
	}
	if len(staged) > 0 {
		result = append(result, " ")
		result = append(result, staged...)
	}

	unstaged := []string{}
	if repoState.Modifications > 0 {
		unstaged = append(unstaged, fmt.Sprintf("%d", repoState.Modifications), red("M"))
	}
	if repoState.Renames > 0 {
		unstaged = append(unstaged, fmt.Sprintf("%d", repoState.Renames), red("R"))
	}
	if repoState.Deletions > 0 {
		unstaged = append(unstaged, fmt.Sprintf("%d", repoState.Deletions), red("D"))
	}
	if len(unstaged) > 0 {
		result = append(result, " ")
		result = append(result, unstaged...)
	}

	rest := []string{}
	if repoState.Untracked > 0 {
		rest = append(rest, fmt.Sprintf("%d", repoState.Untracked), blue("U"))
	}
	if len(rest) > 0 {
		result = append(result, " ")
		result = append(result, rest...)
	}

	return result
}

func Status(repository *git.Repository) RepoState {
	repoState := RepoState{}

	opts := &git.StatusOptions{}
	opts.Show = git.StatusShowIndexAndWorkdir
	opts.Flags = git.StatusOptIncludeUntracked
	statusList, err := repository.StatusList(opts)
	if err != nil {
		panic(err)
	}
	size, err := statusList.EntryCount()
	if err != nil {
		panic(err)
	}

	for i := 0; i < size; i++ {
		status, err := statusList.ByIndex(i)
		if err != nil {
			panic(err)
		}
		if status.Status&git.StatusIndexModified > 0 {
			repoState.ModificationsStaged++
		}
		if status.Status&git.StatusWtModified > 0 {
			repoState.Modifications++
		}
		if status.Status&git.StatusIndexNew > 0 {
			repoState.NewFiles++
		}
		if status.Status&git.StatusWtNew > 0 {
			repoState.Untracked++
		}
		if status.Status&git.StatusIndexRenamed > 0 {
			repoState.RenamesStaged++
		}
		if status.Status&git.StatusWtRenamed > 0 {
			repoState.Renames++
		}
		if status.Status&git.StatusIndexDeleted > 0 {
			repoState.DeletionsStaged++
		}
		if status.Status&git.StatusWtDeleted > 0 {
			repoState.Deletions++
		}
		if status.Status&git.StatusConflicted > 0 {
			if status.HeadToIndex.Status > 0 && status.IndexToWorkdir.Status > 0 {
				repoState.ConflictsBoth++
			} else if status.HeadToIndex.Status > 0 {
				repoState.ConflictsTheir++
			} else if status.IndexToWorkdir.Status > 0 {
				repoState.ConflictsOur++
			}
		}
	}

	return repoState
}

func main() {
	var repository *git.Repository
	for {
		wd, err := os.Getwd()
		if wd == "/" {
			os.Exit(0)
		}
		repository, err = git.OpenRepository(wd)
		if err != nil {
			if git.IsErrorCode(err, git.ErrNotFound) {
				err = os.Chdir(path.Join(wd, ".."))
				if err != nil {
					panic(err)
				}
			} else {
				panic(err)
			}
		} else {
			break
		}
	}
	branch_name, err := Branch(repository)

	result := append([]string{black("git:(")}, branch_name...)
	result = append(result, black(")"))
	if err == nil {
		state := Status(repository)
		result = append(result, state.Format()...)
	}
	fmt.Print(" " + strings.Join(result, ""))
}
