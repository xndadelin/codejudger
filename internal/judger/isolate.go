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
	BoxID       int
	Memory      int
	Runtime     int
	Code        string
	Language    string
	Input       string
	File        string
	TestCases   []TestCase
	Compile     string
	Token       string
	MemoryLimit int
	TimeLimit   int
	Run         []string
}

// swagger:model
type JudgeResult struct {
	ExitCode         string `json:"ExitCode"`
	Status           string `json:"Status"`
	Killed           string `json:"Killed"`
	Time             string `json:"Time"`
	TimeWall         string `json:"TimeWall"`
	Memory           string `json:"Memory"`
	CswVoluntary     string `json:"CswVoluntary"`
	CswForced        string `json:"CswForced"`
	Message          string `json:"Message"`
	Stdout           string `json:"Stdout"`
	Stderr           string `json:"Stderr"`
	Passed           bool   `json:"Passed"`
	CompilationError string `json:"CompilationError,omitempty"`
	Stdin            string `json:"Stdin"`
}

// swagger:model
type JudgeResponse struct {
	Status  string        `json:"status"`
	Score   int           `json:"score"`
	Slug    string        `json:"slug"`
	Results []JudgeResult `json:"results"`
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
	cmd := exec.Command(args[0], args[1:]...)
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

func RunCommand(sandboxRoot string, boxID int, runArgs []string, cfg IsolateConfig) error {
	args := []string{
		"isolate",
		fmt.Sprintf("--box-id=%d", boxID),
		fmt.Sprintf("--mem=%d", cfg.MemoryLimit),
		fmt.Sprintf("--time=%d", cfg.TimeLimit),
		fmt.Sprintf("--wall-time=%d", cfg.Runtime),
		"--stdin=input.txt",
		"--stdout=output.txt",
		"--stderr=cerr.txt",
		"--meta=meta.txt",
		"--processes=4",
		"--run",
		"--",
	}
	args = append(args, runArgs...)

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = fmt.Sprintf("%s/%d/box", sandboxRoot, boxID)
	_, err := cmd.CombinedOutput()

	cerrPath := fmt.Sprintf("%s/%d/box/cerr.txt", sandboxRoot, boxID)
	cerrData, _ := os.ReadFile(cerrPath)

	if len(cerrData) > 0 {
		return fmt.Errorf(string(cerrData))
	}

	if err != nil {
		return fmt.Errorf("isolate run error: %v", err)
	}
	return nil
}

func CleanupSandbox(sandboxRoot string, boxID int) error {
	args := []string{
		"isolate",
		fmt.Sprintf("--box-id=%d", boxID),
		"--cleanup",
	}
	cmd := exec.Command(args[0], args[1:]...)
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
	enviroment := os.Getenv("ENVIRONMENT")
	sandboxRoot := "/var/lib/isolate"
	if enviroment == "PRODUCTION" {
		sandboxRoot = "/var/local/lib/isolate"
	}
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

	if strings.TrimSpace(cfg.Compile) != "" {
		if err := Compile(sandboxRoot, boxID, cfg.Compile); err != nil {
			return nil, fmt.Errorf("failed to compile code: %v", err)
		}
	}

	var results []JudgeResult

	for _, tc := range cfg.TestCases {
		if err := WriteInput(sandboxRoot, boxID, tc.Input); err != nil {
			return nil, err
		}
		if err := RunCommand(sandboxRoot, boxID, cfg.Run, cfg); err != nil {
			return nil, err
		}

		stdout, _ := GetStdout(sandboxRoot, boxID)
		stderr, _ := GetStderr(sandboxRoot, boxID)
		meta, _ := GetMeta(sandboxRoot, boxID)

		metaMap := ParseMeta(meta)

		exitcode := 0
		if s := strings.TrimSpace(metaMap["exitcode"]); s != "" {
			if val, err := strconv.Atoi(s); err == nil {
				exitcode = val
			}
		}

		results = append(results, JudgeResult{
			ExitCode:     strconv.Itoa(exitcode),
			Status:       strings.TrimSpace(metaMap["status"]),
			Killed:       strings.TrimSpace(metaMap["killed"]),
			Time:         strings.TrimSpace(metaMap["time"]),
			TimeWall:     strings.TrimSpace(metaMap["time-wall"]),
			Memory:       strings.TrimSpace(metaMap["max-rss"]),
			CswVoluntary: strings.TrimSpace(metaMap["csw-voluntary"]),
			CswForced:    strings.TrimSpace(metaMap["csw-forced"]),
			Message:      strings.TrimSpace(metaMap["message"]),
			Stdout:       stdout,
			Stderr:       stderr,
			Passed:       strings.TrimSpace(stdout) == strings.TrimSpace(tc.Output),
			Stdin:        tc.Input,
		})
	}

	defer func() {
		CleanupSandbox(sandboxRoot, boxID)
	}()
	return results, nil
}

func RunSingleTest(code string, language string, input string) (JudgeResult, error) {
	enviroment := os.Getenv("ENVIRONMENT")
	sandboxRoot := "/var/lib/isolate"
	if enviroment == "PRODUCTION" {
		sandboxRoot = "/var/local/lib/isolate"
	}
	boxID, err := NextBoxID(sandboxRoot)
	if err != nil {
		return JudgeResult{}, err
	}
	if err := InitSandbox(sandboxRoot, boxID); err != nil {
		return JudgeResult{}, err
	}

	langCfg, ok := Languages[language]
	if !ok {
		return JudgeResult{}, fmt.Errorf("unsupported language: %s", language)
	}
	if err := WriteCode(sandboxRoot, code, boxID, langCfg.File); err != nil {
		return JudgeResult{}, err
	}
	if err := WriteInput(sandboxRoot, boxID, input); err != nil {
		return JudgeResult{}, err
	}

	if langCfg.Compile != "" {
		if err := Compile(sandboxRoot, boxID, langCfg.Compile); err != nil {
			return JudgeResult{
				CompilationError: err.Error(),
				Passed:           false,
				Stdin:            input,
			}, nil
		}
	}
	cfg := IsolateConfig{
		BoxID:       boxID,
		MemoryLimit: 128 * 1024,
		TimeLimit:   2,
		Runtime:     3,
		Run:         langCfg.Run,
	}
	if err := RunCommand(sandboxRoot, boxID, langCfg.Run, cfg); err != nil {
		return JudgeResult{}, err
	}
	stdout, _ := GetStdout(sandboxRoot, boxID)
	stderr, _ := GetStderr(sandboxRoot, boxID)
	meta, _ := GetMeta(sandboxRoot, boxID)
	metaMap := ParseMeta(meta)
	exitcode := 0
	if s := strings.TrimSpace(metaMap["exitcode"]); s != "" {
		if val, err := strconv.Atoi(s); err == nil {
			exitcode = val
		}
	}

	result := JudgeResult{
		ExitCode:     strconv.Itoa(exitcode),
		Status:       strings.TrimSpace(metaMap["status"]),
		Killed:       strings.TrimSpace(metaMap["killed"]),
		Time:         strings.TrimSpace(metaMap["time"]),
		TimeWall:     strings.TrimSpace(metaMap["time-wall"]),
		Memory:       strings.TrimSpace(metaMap["max-rss"]),
		CswVoluntary: strings.TrimSpace(metaMap["csw-voluntary"]),
		CswForced:    strings.TrimSpace(metaMap["csw-forced"]),
		Message:      strings.TrimSpace(metaMap["message"]),
		Stdout:       stdout,
		Stderr:       stderr,
		Stdin:        input,
	}

	defer func() {
		CleanupSandbox(sandboxRoot, boxID)
	}()

	return result, nil
}

type SandboxLanguageConfig struct {
	Extension string
	File      string
	Compile   string
	Run       []string
}

var Languages = map[string]SandboxLanguageConfig{
	"C++": {
		Extension: "cpp",
		File:      "main.cpp",
		Compile:   "/usr/bin/g++ -O2 -o main main.cpp -Wall",
		Run:       []string{"./main"},
	},
	"C": {
		Extension: "c",
		File:      "main.c",
		Compile:   "/usr/bin/gcc -O2 -o main main.c -Wall",
		Run:       []string{"./main"},
	},
	"Rust": {
		Extension: "rs",
		File:      "main.rs",
		Compile:   "rustc main.rs -o main",
		Run:       []string{"./main"},
	},
	"Go": {
		Extension: "go",
		File:      "main.go",
		Compile:   "go build -o main main.go",
		Run:       []string{"./main"},
	},
	"Python": {
		Extension: "py",
		File:      "main.py",
		Run:       []string{"/usr/bin/python3", "main.py"},
		Compile:   "",
	},
	"Javascript": {
		Extension: "js",
		File:      "main.js",
		Run:       []string{"/usr/bin/node", "main.js"},
	},
	"Ruby": {
		Extension: "rb",
		File:      "main.rb",
		Run:       []string{"ruby", "main.rb"},
	},
	"PHP": {
		Extension: "php",
		File:      "main.php",
		Run:       []string{"php", "main.php"},
	},
	"C#": {
		Extension: "cs",
		File:      "main.cs",
		Compile:   "dotnet build -o out main.cs",
		Run:       []string{"dotnet", "out/main.dll"},
	},
}
