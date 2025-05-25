package judger

var Languages = map[string]LanguageConfig{
	"C++": {
		Compile:   "g++ -O2 -o main main.cpp -Wall 2> error.txt",
		Extension: "cpp",
		Run:       "./main",
		File:      "main.cpp",
	},
	"C": {
		Compile:   "gcc -O2 -o main main.c -Wall 2> error.txt",
		Extension: "c",
		Run:       "./main",
		File:      "main.c",
	},
	"C#": {
		Compile:   "dotnet new console -o main && cp main.cs main/Program.cs && cd main && dotnet build -c Release 2> error.txt && cd .. && cp -r main/bin/Release/net8.0/ /var/local/lib/isolate/1/box/program",
		Extension: "cs",
		Run:       "./program/main",
		File:      "main.cs",
	},
	"Java": {
		Compile:     "javac Main.java 2> error.txt",
		Extension:   "java",
		Run:         "./Main.jar",
		File:        "Main.java",
		Requirement: `touch MANIFEST.MF && echo "Main-Class: Main \nJVM-Args: -Xmx4g -Xms2g" > MANIFEST.MF && jar cfm Main.jar MANIFEST.MF Main.class && chmod +x Main.jar`,
	},
	"Python": {
		Compile:     "python3 -m py_compile main.py 2> error.txt",
		Extension:   "py",
		Run:         "./main.py",
		File:        "main.py",
		Requirement: "chmod +x main.py",
		Shebang:     "#!/usr/bin/env python3",
	},
	"Javascript": {
		Compile:     "node main.js > /dev/null 2> error.txt",
		Extension:   "js",
		Run:         "./main.js",
		File:        "main.js",
		Requirement: "chmod +x main.js",
		Shebang:     "#!/usr/bin/env node",
	},
	"Ruby": {
		Compile:   "ruby -c main.rb 2> error.txt",
		Extension: "rb",
		Run:       "ruby main.rb",
		File:      "main.rb",
	},
	"Rust": {
		Compile:   "rustc main.rs 2> error.txt",
		Extension: "rs",
		Run:       "./main",
		File:      "main.rs",
	},
	"Go": {
		Compile:   "go build main.go 2> error.txt && go mod init main",
		Extension: "go",
		Run:       "./main",
		File:      "main.go",
	},
	"PHP": {
		Compile:     "php -l main.php 2> error.txt",
		Extension:   "php",
		Run:         "./main.php",
		File:        "main.php",
		Shebang:     "#!/usr/bin/env php",
		Requirement: "chmod +x main.php",
	},
}

/*
execSync(`isolate
    --box-id=1
    --wait
    --mem=${memory}
    --time=${runtime}
    --meta=${path.join(sandboxPath, 'box', 'meta.txt')}
    --stderr=cerr.txt
    --stdin=input.txt
    --stdout=output.txt
    --run -- "${command}"`,
{ cwd: path.join(sandboxPath, 'box') });
*/
