package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/alecthomas/kong"
)

var cli struct {
	Apply  bool `help:"実際に変更を適用する"`
	DryRun bool `help:"変更内容を表示するのみ（デフォルト）"`
}

type Change struct {
	Kind   string `json:"kind"`
	File   string `json:"file"`
	Detail string `json:"detail"`
}

func (c Change) Output() {
	b, _ := json.Marshal(c)
	fmt.Println(string(b))
}

func promptInput(promptMsg string, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", promptMsg, defaultVal)
	} else {
		fmt.Printf("%s: ", promptMsg)
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())

	if input == "" && defaultVal != "" {
		return defaultVal
	}
	return input
}

type Inputs struct {
	AppName     string
	ModulePath  string
	EnvName     string
	XdgPath     string
	PackageName string
}

func gatherInputs() Inputs {
	rawAppName := promptInput("アプリケーション名（バイナリ名）", "myapp")
	appName := strings.ToLower(rawAppName)
	re := regexp.MustCompile(`[^a-z0-9-]`)
	appName = re.ReplaceAllString(appName, "-")
	appName = strings.Trim(appName, "-")

	if appName == "" {
		fmt.Fprintln(os.Stderr, "Error: アプリケーション名は必須です")
		os.Exit(1)
	}

	githubUser := promptInput("GitHubユーザー名", "username")
	defaultModule := fmt.Sprintf("github.com/%s/%s", githubUser, appName)
	modulePath := promptInput("Go module path", defaultModule)

	if modulePath == "" {
		fmt.Fprintln(os.Stderr, "Error: module pathは必須です")
		os.Exit(1)
	}

	envName := strings.ToUpper(appName)
	reEnv := regexp.MustCompile(`[^A-Z0-9]`)
	envName = reEnv.ReplaceAllString(envName, "_")
	envName = strings.Trim(envName, "_")

	xdgPath := appName

	rawPkg := promptInput("ルートパッケージ名（config.go等）", appName)
	rePkg := regexp.MustCompile(`[^a-z0-9_]`)
	packageName := rePkg.ReplaceAllString(rawPkg, "_")
	packageName = strings.Trim(packageName, "_")

	return Inputs{
		AppName:     appName,
		ModulePath:  modulePath,
		EnvName:     envName,
		XdgPath:     xdgPath,
		PackageName: packageName,
	}
}

func replaceInFile(filePath, old, new, desc string, dryRun bool) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	if !strings.Contains(string(content), old) {
		return nil
	}

	Change{Kind: "change", File: filePath, Detail: desc}.Output()

	if !dryRun {
		newContent := strings.ReplaceAll(string(content), old, new)
		if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
			return err
		}
	}

	return nil
}

func renameDir(oldPath, newPath string, dryRun bool) error {
	Change{Kind: "rename", File: oldPath, Detail: "→ " + newPath}.Output()

	if !dryRun {
		if _, err := os.Stat(oldPath); err == nil {
			if err := os.Rename(oldPath, newPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func main() {
	ctx := kong.Parse(&cli)
	_ = ctx

	dryRun := cli.DryRun || !cli.Apply

	inputs := gatherInputs()

	oldModule := "github.com/nurazon59/go-template"
	oldApp := "go-template"
	oldEnv := "GO_TEMPLATE_CONFIG"
	oldXdg := "go-template"
	oldPkg := "template"

	rootDir, _ := filepath.Abs(".")

	// go.mod
	replaceInFile(filepath.Join(rootDir, "go.mod"), oldModule, inputs.ModulePath, "module path更新", dryRun)

	// cmd/template → cmd/<app>
	oldCmdDir := filepath.Join(rootDir, "cmd", "template")
	newCmdDir := filepath.Join(rootDir, "cmd", inputs.AppName)

	replaceInFile(filepath.Join(oldCmdDir, "main.go"), oldModule+"/cmd", inputs.ModulePath+"/cmd", "main.goのimportパス更新", dryRun)
	replaceInFile(filepath.Join(oldCmdDir, "main_test.go"), oldApp, inputs.AppName, "テストのバイナリ名更新", dryRun)
	renameDir(oldCmdDir, newCmdDir, dryRun)

	// cmd/run.go
	runFile := filepath.Join(rootDir, "cmd", "run.go")
	replaceInFile(runFile, oldModule, inputs.ModulePath, "run.goのimportパス更新", dryRun)
	replaceInFile(runFile, oldPkg+"\"", inputs.PackageName+"\"", "run.goのimportエイリアス更新", dryRun)
	replaceInFile(runFile, oldEnv, inputs.EnvName, "環境変数名更新", dryRun)
	replaceInFile(runFile, oldXdg, inputs.XdgPath, "XDGパス更新", dryRun)
	replaceInFile(runFile, fmt.Sprintf("kong.Name(\"%s\")", oldApp), fmt.Sprintf("kong.Name(\"%s\")", inputs.AppName), "kong.Name更新", dryRun)
	replaceInFile(runFile, oldPkg+".Load", inputs.PackageName+".Load", "run.goのLoad呼び出し更新", dryRun)

	// config.go
	replaceInFile(filepath.Join(rootDir, "config.go"), "package "+oldPkg, "package "+inputs.PackageName, "config.goパッケージ名更新", dryRun)

	// config_test.go
	replaceInFile(filepath.Join(rootDir, "config_test.go"), "package "+oldPkg, "package "+inputs.PackageName, "config_test.goパッケージ名更新", dryRun)

	// Taskfile.yml
	replaceInFile(filepath.Join(rootDir, "Taskfile.yml"), oldApp, inputs.AppName, "Taskfileのバイナリ名更新", dryRun)
	replaceInFile(filepath.Join(rootDir, "Taskfile.yml"), "cmd/template", "cmd/"+inputs.AppName, "Taskfileのbuildパス更新", dryRun)

	// .goreleaser.yaml
	replaceInFile(filepath.Join(rootDir, ".goreleaser.yaml"), "cmd/template", "cmd/"+inputs.AppName, "goreleaserのmainパス更新", dryRun)

	// README.md
	replaceInFile(filepath.Join(rootDir, "README.md"), oldApp, inputs.AppName, "READMEのアプリ名更新", dryRun)
	replaceInFile(filepath.Join(rootDir, "README.md"), oldEnv, inputs.EnvName, "READMEの環境変数名更新", dryRun)
	replaceInFile(filepath.Join(rootDir, "README.md"), oldXdg, inputs.XdgPath, "READMEのXDGパス更新", dryRun)
	replaceInFile(filepath.Join(rootDir, "README.md"), "cmd/template", "cmd/"+inputs.AppName, "READMEのパス更新", dryRun)

	// bin/template → bin/<app>
	oldBin := filepath.Join(rootDir, "bin", "template")
	newBin := filepath.Join(rootDir, "bin", inputs.AppName)
	if _, err := os.Stat(oldBin); err == nil {
		renameDir(oldBin, newBin, dryRun)
	}

	if dryRun {
		fmt.Println()
		fmt.Println("上記の変更が適用されます。実際に適用するには --apply を付けて実行してください。")
	} else {
		// setupスクリプトと一時ファイルを削除
		setupSh := filepath.Join(rootDir, "scripts", "setup.sh")
		if _, err := os.Stat(setupSh); err == nil {
			os.Remove(setupSh)
			fmt.Println("setup.shを削除しました")
		}

		zzzDir := filepath.Join(rootDir, "zzz")
		if _, err := os.Stat(zzzDir); err == nil {
			os.RemoveAll(zzzDir)
			fmt.Println("zzz/ディレクトリを削除しました")
		}

		// setupディレクトリを削除
		setupDir := filepath.Join(rootDir, "cmd", "setup")
		if err := os.RemoveAll(setupDir); err == nil {
			fmt.Println("cmd/setup/ディレクトリを削除しました")
		} else {
			fmt.Printf("cmd/setup/ディレクトリを手動で削除してください: %v\n", err)
		}

		fmt.Println()
		fmt.Println("変更を適用しました。以下の確認を実行してください:")
		fmt.Println("  1. go mod tidy")
		fmt.Println("  2. task ci")
		fmt.Println("  3. git status で変更確認")
	}
}
