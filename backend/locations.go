package backend

import (
	"path/filepath"
	"sort"
	"strings"
)

// Location is one directory on PATH, with how many binaries it contributes.
type Location struct {
	Path  string `json:"path"`
	Label string `json:"label"`
	Logo  string `json:"logo"` // brand mark in /logos, empty when none is known
	Count int    `json:"count"`
}

// knownDir labels the directories people actually look for, and gives each a
// brand mark. Matching is a single substring test against the lowercased path,
// so the first match wins: deeper fragments have to come before their parents
// (flutter\bin before dart-sdk\bin, windows kits before windows).
var knownDir = []struct{ fragment, label, logo string }{
	// Windows. \windows kits starts with \windows, so it has to come first.
	{`\windows\system32\windowspowershell`, "PowerShell", ""},
	{`\windows\system32\openssh`, "OpenSSH", ""},
	{`\windows\system32\wbem`, "WMI", "windows"},
	{`\windows\system32`, "System32", "windows"},
	{`\windows\syswow64`, "SysWOW64", "windows"},
	{`\windows kits`, "Windows SDK", "windows"},
	{`\windowsapps`, "WindowsApps", "windows"},
	{`\windows`, "Windows", "windows"},

	// Editors and IDEs.
	{`\microsoft vs code insiders`, "VS Code Insiders", ""},
	{`\microsoft vs code\bin`, "VS Code", ""},
	{`\cursor\resources\app\bin`, "Cursor", "cursor"},
	{`\jetbrains`, "JetBrains", "jetbrains"},
	{`\nvim`, "Neovim", "neovim"},
	{`\notepad++`, "Notepad++", "notepadplusplus"},
	{`\sublime text`, "Sublime Text", "sublimetext"},

	// Git and source control.
	{`\github cli`, "GitHub CLI", "github"},
	{`\githubdesktop`, "GitHub Desktop", "github"},
	{`\tortoisegit`, "TortoiseGit", "git"},
	{`\github\copilot`, "Copilot", "githubcopilot"},
	{`\git\`, "Git", "git"},

	// Languages and runtimes.
	{`\.deno\bin`, "Deno", "deno"},
	{`\.volta\bin`, "Volta", ""},
	{`\nvm`, "nvm", "nodedotjs"},
	{`\fnm`, "fnm", "nodedotjs"},
	{`\yarn\bin`, "Yarn", "yarn"},
	{`\pnpm`, "pnpm", "pnpm"},
	{`\npm`, "npm", "npm"},
	{`\nodejs`, "Node.js", "nodedotjs"},
	{`\.bun\bin`, "Bun", "bun"},
	{`\eclipse adoptium`, "Temurin JDK", "openjdk"},
	{`\java`, "Java", "openjdk"},
	{`\jdk`, "JDK", "openjdk"},
	{`\apache-maven`, "Maven", "apachemaven"},
	{`\gradle`, "Gradle", "gradle"},
	{`\ruby`, "Ruby", "ruby"},
	{`\composer`, "Composer", "composer"},
	{`\php`, "PHP", "php"},
	{`\strawberry`, "Strawberry Perl", "perl"},
	{`\flutter\bin`, "Flutter", "flutter"},
	{`\dart-sdk\bin`, "Dart", "dart"},
	{`\ghcup\bin`, "Haskell", "haskell"},
	{`\julia`, "Julia", "julia"},
	{`\r\r-`, "R", "r"},
	{`\zig`, "Zig", "zig"},
	{`\pyenv`, "pyenv", "python"},
	{`\anaconda3`, "Anaconda", "anaconda"},
	{`\miniconda3`, "Miniconda", "anaconda"},
	{`\python`, "Python", "python"},
	{`\.rustup`, "Rustup", "rust"},
	{`.cargo\bin`, "Cargo", "rust"},
	{`\go\bin`, "Go", "go"},
	{`\.dotnet\tools`, ".NET Tools", "dotnet"},
	{`\dotnet`, ".NET", "dotnet"},
	{`\nsis`, "NSIS", "nsis"},
	{`\mingw64`, "MinGW", ""},
	{`\msys64`, "MSYS2", ""},
	{`\cygwin`, "Cygwin", ""},

	// Build and toolchains.
	{`\microsoft visual studio`, "Visual Studio", ""},
	{`\msbuild`, "MSBuild", ""},
	{`\cmake\bin`, "CMake", "cmake"},
	{`\llvm\bin`, "LLVM", "llvm"},
	{`\ninja`, "Ninja", ""},
	{`\android\sdk`, "Android SDK", "android"},

	// Package managers.
	{`\chocolatey`, "Chocolatey", "chocolatey"},
	{`\scoop`, "Scoop", ""},

	// Containers, cloud, infrastructure.
	{`\docker`, "Docker", "docker"},
	{`\rancher desktop`, "Rancher Desktop", "rancher"},
	{`\kubectl`, "kubectl", "kubernetes"},
	{`\minikube`, "minikube", "kubernetes"},
	{`\helm`, "Helm", "helm"},
	{`\terraform`, "Terraform", "terraform"},
	{`\cloud sdk`, "Google Cloud SDK", "googlecloud"},
	{`\amazon\awscli`, "AWS CLI", ""},
	{`\microsoft sdks\azure`, "Azure CLI", ""},
	{`\tailscale`, "Tailscale", "tailscale"},
	{`\ngrok`, "ngrok", "ngrok"},

	// Databases.
	{`\postgresql`, "PostgreSQL", "postgresql"},
	{`\mysql`, "MySQL", "mysql"},
	{`\mongodb`, "MongoDB", "mongodb"},
	{`\redis`, "Redis", "redis"},
	{`\sqlite`, "SQLite", "sqlite"},

	// AI and GPU. The CUDA toolkit wins over the vendor folder.
	{`\ollama`, "Ollama", "ollama"},
	{`\nvidia gpu computing toolkit`, "CUDA", "nvidia"},
	{`\nvidia corporation`, "NVIDIA", "nvidia"},
	{`\amd\`, "AMD", "amd"},

	// Media, docs and utilities.
	{`\7-zip`, "7-Zip", "7zip"},
	{`\imagemagick`, "ImageMagick", ""},
	{`\pandoc`, "Pandoc", ""},
	{`\graphviz`, "Graphviz", ""},
	{`\miktex`, "MiKTeX", "latex"},
	{`\texlive`, "TeX Live", "latex"},
	{`\yt-dlp`, "yt-dlp", ""},
	{`\obs-studio`, "OBS Studio", "obsstudio"},
	{`\wireshark`, "Wireshark", "wireshark"},
	{`\putty`, "PuTTY", ""},
	{`\openssl`, "OpenSSL", "openssl"},
	{`\postman`, "Postman", "postman"},
	{`\ffmpeg`, "FFmpeg", "ffmpeg"},
	{`\cloudflare`, "Cloudflare WARP", "cloudflare"},

	// Last on purpose: a package-manager cache wraps every tool it installed,
	// so \microsoft\winget\packages\<vendor>\...\mingw64\bin should read as
	// MinGW, and only paths nothing else claimed fall through to here.
	{`\microsoft\winget`, "winget", "windows"},
}

// installRoots are the parents installers choose between. They are stripped
// from both the path and the fragments, so one entry covers a machine-wide and
// a per-user install: \program files\nodejs also matches the AppData copy.
var installRoots = []string{
	`\program files (x86)\`,
	`\program files\`,
	`\appdata\local\programs\`,
	`\appdata\local\`,
	`\appdata\roaming\`,
}

func init() {
	for i := range knownDir {
		knownDir[i].fragment = trimRoot(knownDir[i].fragment)
	}
}

// Locations groups the PATH scan by directory: the folders that keep a family
// of binaries together, which is what the grid view shows.
func (b *Binaries) Locations() ([]Location, error) {
	binaries, err := b.List()
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int, len(binaries))
	for _, binary := range binaries {
		counts[filepath.Dir(binary.Path)]++
	}

	out := make([]Location, 0, len(counts))
	for dir, count := range counts {
		label, logo := describeDir(dir)
		out = append(out, Location{Path: dir, Label: label, Logo: logo, Count: count})
	}

	// Biggest clusters first: System32 before a single-tool folder.
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Path < out[j].Path
	})
	return out, nil
}

func describeDir(dir string) (string, string) {
	lower := trimRoot(strings.ToLower(strings.ReplaceAll(dir, "/", `\`)))
	for _, known := range knownDir {
		if strings.Contains(lower, known.fragment) {
			return known.label, known.logo
		}
	}
	return filepath.Base(dir), ""
}

// trimRoot drops the first installer root it finds, keeping a separator so
// fragments still match on a folder boundary.
func trimRoot(path string) string {
	for _, root := range installRoots {
		if i := strings.Index(path, root); i >= 0 {
			return path[:i] + `\` + path[i+len(root):]
		}
	}
	return path
}
