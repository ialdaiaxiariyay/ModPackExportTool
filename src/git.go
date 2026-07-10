// git.go
package pack

import (
    "fmt"
    "os/exec"
)

func InitGitRepo(dir, remote, branch string) error {
    if isGitRepo(dir) {
        fmt.Println("Directory is already a Git repository, skipping init")
        return nil
    }

    cmd := exec.Command("git", "init")
    cmd.Dir = dir
    if out, err := cmd.CombinedOutput(); err != nil {
        return fmt.Errorf("git init failed: %s, output: %s", err, out)
    }
    fmt.Println("Git repository initialised")

    if remote != "" {
        delCmd := exec.Command("git", "remote", "remove", "origin")
        delCmd.Dir = dir
        _ = delCmd.Run()

        addCmd := exec.Command("git", "remote", "add", "origin", remote)
        addCmd.Dir = dir
        if out, err := addCmd.CombinedOutput(); err != nil {
            return fmt.Errorf("adding remote failed: %s, output: %s", err, out)
        }
        fmt.Printf("Remote %s added (branch: %s)\n", remote, branch)
    }
    return nil
}

func isGitRepo(dir string) bool {
    cmd := exec.Command("git", "rev-parse", "--git-dir")
    cmd.Dir = dir
    return cmd.Run() == nil
}