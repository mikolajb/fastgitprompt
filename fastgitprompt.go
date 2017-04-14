package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	git "gopkg.in/libgit2/git2go.v24"
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
	head, err := repository.Head()
	if err != nil {
		if git.IsErrorCode(err, git.ErrUnbornBranch) {
			fmt.Print(black("git:(") + "no head" + black(")"))
			os.Exit(0)
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
			branch_name = append([]string{"m ↔ "}, branch_name...)
		} else {
			if behind > 0 {
				branch_name = append([]string{"m → "}, branch_name...)
			}
			if ahead > 0 {
				branch_name = append([]string{"m ← "}, branch_name...)
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
		ahead_string := fmt.Sprintf("%d", ahead)
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

	untracked := 0
	new_file := 0
	deletion := 0
	deletion_staged := 0
	modification := 0
	modification_staged := 0
	rename := 0
	rename_staged := 0
	conflict_both := 0
	conflict_our := 0
	conflict_their := 0
	for i := 0; i < size; i++ {
		status, err := statusList.ByIndex(i)
		if err != nil {
			panic(err)
		}
		switch status.Status {
		case git.StatusIndexModified:
			modification_staged++
		case git.StatusWtModified:
			modification++
		case git.StatusIndexNew:
			new_file++
		case git.StatusWtNew:
			untracked++
		case git.StatusIndexRenamed:
			rename_staged++
		case git.StatusWtRenamed:
			rename++
		case git.StatusIndexDeleted:
			deletion_staged++
		case git.StatusWtDeleted:
			deletion++
		case git.StatusConflicted:
			if status.HeadToIndex.Status > 0 && status.IndexToWorkdir.Status > 0 {
				conflict_both++
			} else if status.HeadToIndex.Status > 0 {
				conflict_their++
			} else if status.IndexToWorkdir.Status > 0 {
				conflict_our++
			}
		default:
			fmt.Println(status)
		}
	}

	result := append([]string{black("git:(")}, branch_name...)
	result = append(result, black(")"))

	if conflict_both > 0 {
		result = append(result, fmt.Sprintf(" %d", conflict_both), blue("B"))
	} else if conflict_our > 0 {
		result = append(result, fmt.Sprintf(" %d", conflict_our), blue("U"))
	} else if conflict_their > 0 {
		result = append(result, fmt.Sprintf(" %d", conflict_their), blue("T"))
	}

	staged := []string{}
	if new_file > 0 {
		staged = append(staged, fmt.Sprintf("%d", new_file), green("A"))
	}
	if modification_staged > 0 {
		staged = append(staged, fmt.Sprintf("%d", modification_staged), green("M"))
	}
	if rename_staged > 0 {
		staged = append(staged, fmt.Sprintf("%d", rename_staged), green("R"))
	}
	if deletion_staged > 0 {
		staged = append(staged, fmt.Sprintf("%d", deletion_staged), green("D"))
	}
	if len(staged) > 0 {
		result = append(result, staged...)
	}

	unstaged := []string{}
	if modification > 0 {
		unstaged = append(unstaged, fmt.Sprintf("%d", modification), red("M"))
	}
	if rename > 0 {
		unstaged = append(unstaged, fmt.Sprintf("%d", rename), red("R"))
	}
	if deletion > 0 {
		unstaged = append(unstaged, fmt.Sprintf("%d", deletion), red("D"))
	}
	if len(unstaged) > 0 {
		result = append(result, " ")
		result = append(result, unstaged...)
	}

	rest := []string{}
	if untracked > 0 {
		rest = append(rest, fmt.Sprintf("%d", untracked), blue("A"))
	}
	if len(rest) > 0 {
		result = append(result, []string{" "}...)
		result = append(result, rest...)
	}
	fmt.Print(" " + strings.Join(result, ""))
}
