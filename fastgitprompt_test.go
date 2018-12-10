package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	git "gopkg.in/src-d/go-git.v4"
)

func run(t testing.TB, path, name string, arg ...string) {
	t.Helper()
	cmd := exec.Command(name, arg...)
	cmd.Dir = path
	err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}
	return
}

func initialize(t testing.TB) string {
	dir, err := ioutil.TempDir("", "fast-git-prompt")
	if err != nil {
		t.Fatal(err.Error())
	}

	run(t, dir, "ls")
	run(t, dir, "git", "init")
	run(t, dir, "touch", "README")
	run(t, dir, "git", "add", "README")
	run(t, dir, "git", "commit", "-m", "'first'")
	run(t, dir, "git", "checkout", "-b", "fork")
	run(t, dir, "touch", "test.txt")
	run(t, dir, "git", "add", "test.txt")
	run(t, dir, "git", "commit", "-m", "'second'")
	return dir
}

func BenchmarkPrompt(b *testing.B) {
	dir := initialize(b)
	defer os.RemoveAll(dir)
	repository, err := git.PlainOpen(dir)
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < b.N; i++ {
		_, err = Prompt(repository)
		if err != nil {
			b.Error(err)
		}
	}

	err = os.RemoveAll(dir)
	if err != nil {
		b.Error(err)
	}
}

func TestPrompt(t *testing.T) {
	dir := initialize(t)
	defer os.RemoveAll(dir)
	repository, err := git.PlainOpen(dir)
	if err != nil {
		t.Error(err)
	}

	prompt, err := Prompt(repository)
	if err != nil {
		t.Error(err)
	}
	fmt.Println("prompt:", prompt)

	if len(prompt) == 0 {
		t.Error("x")
	}
}
