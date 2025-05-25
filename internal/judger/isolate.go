package judger

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

type IsolateConfig struct {
	BoxID   int
	Memory  int
	Runtime int
	Command string
}

func NextBoxID(sandboxRoot string) (int, error) {
	files, err := os.ReadDir(sandboxRoot)
	if err != nil {
		return 0, fmt.Errorf("failed to read sandbox directory: %v", err)
	}
	maxID := 0
	for _, file := range files {
		if file.IsDir() {
			id, err := strconv.Atoi(file.Name())
			if err == nil && id > maxID {
				maxID = id
			}
		}
	}
	return maxID + 1, nil
}

func InitSandbox(sandboxRoot string, boxID int) error {
	args := []string{
		fmt.Sprintf("--box-id=%d", boxID),
		"--init",
	}
	cmd := exec.Command("isolate", args...)
	cmd.Dir = sandboxRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("isolate init error: %v, output: %s", err, string(output))
	}
	return nil
}

func RunIsolate(cfg IsolateConfig) error {
	sandboxRoot := "/var/lib/isolate"
	boxID, err := NextBoxID(sandboxRoot)
	if err != nil {
		return err
	}
	if err := InitSandbox(sandboxRoot, boxID); err != nil {
		return fmt.Errorf("failed to initialize sandbox: %v", err)
	}
	fmt.Println("running isolate with box ID:", boxID)
	defer func() {
		cleanupCmd := exec.Command("isolate", "--box-id="+strconv.Itoa(boxID), "--cleanup")
		cleanupCmd.Dir = sandboxRoot
		if err := cleanupCmd.Run(); err != nil {
			fmt.Printf("failed to clean up box %d: %v\n", boxID, err)
		} else {
			fmt.Printf("successfully cleaned up box %d\n", boxID)
		}
	}()
	return nil
}
