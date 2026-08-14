package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	signtool = "C:\\Program Files (x86)\\Windows Kits\\10\\bin\\" +
		"10.0.26100.0\\x64\\signtool.exe"
	certSha1      = "c088a20413edc0fdfe17f5905af624c68a33144c"
	timestampUrl  = "http://timestamp.digicert.com"
	innoSetupIscc = "C:\\Program Files (x86)\\Inno Setup 6\\ISCC.exe"
	changesPath   = "CHANGES"
	changesVerTag = "<%= version %>"
)

var versionSources = []struct {
	path string
	expr string
}{
	{changesPath, `(?m)^Version (\S+)`},
	{"client/package.json", `"version": *"(.*?)"`},
	{"client/package-lock.json", `"version": *"(.*?)"`},
	{"service/constants/constants.go", `Version = "(.*?)"`},
	{"cli/constants/constants.go", `Version = "(.*?)"`},
	{"resources_win/setup.iss", `MyAppVersion "(.*?)"`},
}

var signArgs = []string{
	"sign",
	"/sha1", certSha1,
	"/tr", timestampUrl,
	"/td", "sha256",
	"/fd", "sha256",
}

var fuseArgs = []string{
	"RunAsNode=off",
	"EnableNodeOptionsEnvironmentVariable=off",
	"EnableNodeCliInspectArguments=off",
	"EnableEmbeddedAsarIntegrityValidation=on",
	"OnlyLoadAppFromAsar=on",
}

func run(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}

func runEnv(env []string, name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}

func walkPeFiles(root string) (paths []string) {
	err := filepath.WalkDir(root, func(path string,
		d fs.DirEntry, err error) error {

		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		switch strings.ToLower(filepath.Ext(path)) {
		case ".exe", ".dll", ".node":
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	return
}

func signFiles(paths []string) {
	if len(paths) == 0 {
		return
	}
	run(signtool, append(signArgs, paths...)...)
}

func signDir(root string) {
	signFiles(walkPeFiles(root))
}

func verifyDir(root string) {
	paths := walkPeFiles(root)
	if len(paths) == 0 {
		return
	}
	run(signtool, append([]string{"verify", "/pa", "/all"}, paths...)...)
}

func signExe(path string) {
	signFiles([]string{path})
}

func writeFuses(app string) {
	for _, fuse := range fuseArgs {
		run("npx",
			"@electron/fuses",
			"write", "--app", app,
			fuse,
		)
	}
}

func readVersion(path string, expr string) string {
	data, err := os.ReadFile(filepath.FromSlash(path))
	if err != nil {
		panic(err)
	}

	match := regexp.MustCompile(expr).FindSubmatch(data)
	if match == nil {
		panic("build: unable to read version from " + path)
	}

	return string(match[1])
}

// The installer version comes from resources_win/setup.iss and the app version
// from client/package.json, building before builder.py set-version stamps the
// installer, the app and both Go binaries with the previous release version
func checkVersion() {
	changes, err := os.ReadFile(changesPath)
	if err != nil {
		panic(err)
	}

	if bytes.Contains(changes, []byte(changesVerTag)) {
		panic("build: unreleased version placeholder in CHANGES, run " +
			"tools/builder.py set-version before building")
	}

	version := ""
	for _, source := range versionSources {
		ver := readVersion(source.path, source.expr)

		if version == "" {
			version = ver
		} else if ver != version {
			panic("build: source tree version mismatch, expected " +
				version + " but " + source.path + " has " + ver +
				", run tools/builder.py set-version")
		}
	}

	fmt.Println("Building version " + version)
}

func main() {
	checkVersion()

	err := os.RemoveAll("build")
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}

	err = os.Chdir("service")
	if err != nil {
		panic(err)
	}

	run("go", "get")

	run("go", "build", "-v", "-ldflags", "-H windowsgui")
	signExe("service.exe")
	err = os.Rename("service.exe", "service_amd64.exe")
	if err != nil {
		panic(err)
	}

	runEnv([]string{"GOOS=windows", "GOARCH=arm64"},
		"go", "build", "-v", "-ldflags", "-H windowsgui")
	signExe("service.exe")
	err = os.Rename("service.exe", "service_arm64.exe")
	if err != nil {
		panic(err)
	}

	err = os.Chdir(filepath.Join("..", "cli"))
	if err != nil {
		panic(err)
	}

	run("go", "get")

	run("go", "build", "-v")
	signExe("cli.exe")
	err = os.Rename("cli.exe", "cli_amd64.exe")
	if err != nil {
		panic(err)
	}

	runEnv([]string{"GOOS=windows", "GOARCH=arm64"},
		"go", "build", "-v")
	signExe("cli.exe")
	err = os.Rename("cli.exe", "cli_arm64.exe")
	if err != nil {
		panic(err)
	}

	err = os.Chdir(filepath.Join("..", "client"))
	if err != nil {
		panic(err)
	}

	run("npm", "install")

	run(".\\node_modules\\.bin\\electron-rebuild")

	run(
		".\\node_modules\\.bin\\electron-packager",
		".\\",
		"pritunl",
		"--platform=win32",
		"--arch=x64",
		"--icon=www\\img\\logo.ico",
		"--out=..\\build\\win",
		"--prune",
		"--asar",
	)

	run(
		".\\node_modules\\.bin\\electron-packager",
		".\\",
		"pritunl",
		"--platform=win32",
		"--arch=arm64",
		"--icon=www\\img\\logo.ico",
		"--out=..\\build\\win",
		"--prune",
		"--asar",
	)

	err = os.Chdir(filepath.Join("..", "build", "win",
		"pritunl-win32-x64"))
	if err != nil {
		panic(err)
	}

	writeFuses("pritunl.exe")
	signDir(".")
	verifyDir(".")

	err = os.Chdir(filepath.Join("..", "pritunl-win32-arm64"))
	if err != nil {
		panic(err)
	}

	writeFuses("pritunl.exe")
	signDir(".")
	verifyDir(".")

	err = os.Chdir(filepath.Join("..", "..", "..",
		"resources_win"))
	if err != nil {
		panic(err)
	}

	run(innoSetupIscc,
		"/Ssigntool=$q"+signtool+"$q sign /sha1 "+certSha1+
			" /tr "+timestampUrl+" /td sha256 /fd sha256 $q$f$q",
		"setup.iss")
}
