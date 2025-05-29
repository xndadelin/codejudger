package judger

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type TestCase struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

type IsolateConfig struct {
	BoxID     int
	Memory    int
	Runtime   int
	Command   string
	Code      string
	Language  string
	Input     string
	File      string
	TestCases []TestCase
	Compile   string
	Token     string
}

type JudgeResult struct {
	ExitCode         string
	Time             string
	Memory           string
	Stdout           string
	Stderr           string
	Passed           bool
	CompilationError string
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
		"isolate",
		fmt.Sprintf("--box-id=%d", boxID),
		"--init",
	}
	cmd := exec.Command("sudo", args...)
	cmd.Dir = sandboxRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("isolate init error: %v, output: %s", err, string(output))
	}
	return nil
}

func WriteCode(sandboxRoot, code string, boxID int, fileName string) error {
	boxPath := fmt.Sprintf("%s/%d/box", sandboxRoot, boxID)
	codePath := fmt.Sprintf("%s/%s", boxPath, fileName)
	err := os.WriteFile(codePath, []byte(code), 0644)
	if err != nil {
		return fmt.Errorf("failed to write code file: %v", err)
	}
	return nil
}

func WriteInput(sandboxRoot string, boxID int, input string) error {
	inputPath := fmt.Sprintf("%s/%d/box/input.txt", sandboxRoot, boxID)
	err := os.WriteFile(inputPath, []byte(input), 0644)
	if err != nil {
		return fmt.Errorf("failed to write input file: %v", err)
	}
	return nil
}

func RunCommand(sandboxRoot string, boxID int, command string) error {
	args := []string{
		"isolate",
		fmt.Sprintf("--box-id=%d", boxID),
		"--stdin=input.txt",
		"--stdout=output.txt",
		"--stderr=cerr.txt",
		"--meta=meta.txt",
		"--processes=4",
		"--run",
		"--",
	}

	args = append(args, strings.Fields(command)...)

	cmd := exec.Command("sudo", args...)
	cmd.Dir = fmt.Sprintf("%s/%d/box", sandboxRoot, boxID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("isolate run error: %v, output: %s", err, string(output))
	}
	return nil
}

func CleanupSandbox(sandboxRoot string, boxID int) error {
	args := []string{
		"isolate",
		fmt.Sprintf("--box-id=%d", boxID),
		"--cleanup",
	}
	cmd := exec.Command("sudo", args...)
	cmd.Dir = sandboxRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("isolate cleanup error: %v, output: %s", err, string(output))
	}
	return nil
}

func GetStdout(sandboxRoot string, boxID int) (string, error) {
	stdoutPath := fmt.Sprintf("%s/%d/box/output.txt", sandboxRoot, boxID)
	data, err := os.ReadFile(stdoutPath)
	if err != nil {
		return "", fmt.Errorf("failed to read stdout file: %v", err)
	}
	return string(data), nil
}

func GetMeta(sandboxRoot string, boxID int) (string, error) {
	metadataPath := fmt.Sprintf("%s/%d/box/meta.txt", sandboxRoot, boxID)
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return "", fmt.Errorf("failed to read metadata file: %v", err)
	}
	return string(data), nil
}

func GetStderr(sandboxRoot string, boxID int) (string, error) {
	stderrPath := fmt.Sprintf("%s/%d/box/cerr.txt", sandboxRoot, boxID)
	data, err := os.ReadFile(stderrPath)
	if err != nil {
		return "", fmt.Errorf("failed to read stderr file: %v", err)
	}
	return string(data), nil
}

func Compile(sandboxRoot string, boxID int, command string) error {
	boxPath := fmt.Sprintf("%s/%d/box", sandboxRoot, boxID)
	cmdParts := strings.Fields(command)

	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	cmd.Dir = boxPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("compile error: %v, output: %s", err, string(output))
	}
	return nil
}

func ParseMeta(meta string) map[string]string {
	result := make(map[string]string)
	lines := strings.Split(meta, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

func RunIsolate(cfg IsolateConfig) ([]JudgeResult, error) {
	sandboxRoot := "/var/lib/isolate"
	boxID, err := NextBoxID(sandboxRoot)
	if err != nil {
		return nil, err
	}
	if err := InitSandbox(sandboxRoot, boxID); err != nil {
		return nil, fmt.Errorf("failed to initialize sandbox: %v", err)
	}

	boxPath := fmt.Sprintf("%s/%d/box", sandboxRoot, boxID)
	if _, err := os.Stat(boxPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("box directory does not exist after init: %v", boxPath)
	}

	if err := WriteCode(sandboxRoot, cfg.Code, boxID, cfg.File); err != nil {
		return nil, err
	}

	if cfg.Compile != "" {
		if err := Compile(sandboxRoot, boxID, cfg.Compile); err != nil {
			return nil, fmt.Errorf("failed to compile code: %v", err)
		}
	}

	var results []JudgeResult

	for _, tc := range cfg.TestCases {
		if err := WriteInput(sandboxRoot, boxID, tc.Input); err != nil {
			return nil, err
		}
		if err := RunCommand(sandboxRoot, boxID, cfg.Command); err != nil {
			return nil, err
		}

		stdout, _ := GetStdout(sandboxRoot, boxID)
		stderr, _ := GetStderr(sandboxRoot, boxID)
		meta, _ := GetMeta(sandboxRoot, boxID)

		metaMap := ParseMeta(meta)
		exitcode := metaMap["exitcode"]
		time := metaMap["time"]
		memory := metaMap["max-rss"]

		fmt.Println("Exit Code:", exitcode)
		fmt.Println("Time:", time)
		fmt.Println("Memory:", memory)
		fmt.Println("Stdout:", stdout)

		results = append(results, JudgeResult{
			ExitCode: exitcode,
			Time:     time,
			Memory:   memory,
			Stdout:   stdout,
			Stderr:   stderr,
			Passed:   strings.TrimSpace(stdout) == strings.TrimSpace(tc.Output),
		})
	}

	defer CleanupSandbox(sandboxRoot, boxID)
	return results, nil
}
