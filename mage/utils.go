package mage

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// ToolCommand builds an exec.Cmd for a build tool, resolving it to an absolute
// path first. Passing a bare name would leave the binary chosen by whatever
// PATH happens to be set when the build runs; resolving up front pins the
// choice and turns a missing tool into a named error instead of an opaque
// "executable file not found" from the eventual Run.
func ToolCommand(name string, args ...string) (*exec.Cmd, error) {
	resolved, err := exec.LookPath(name)
	if err != nil {
		return nil, fmt.Errorf("%s not found in PATH: %w", name, err)
	}
	return exec.Command(resolved, args...), nil
}

// CheckNodeVersion reads .nvmrc and ensures the correct Node version is active.
// Since nvm is a shell function (not a binary), we can't call "nvm use" from Go.
// Instead, if the current node version doesn't match, we look for the correct
// version in the nvm install directory and prepend its bin/ to PATH so all
// subsequent npm/npx/node calls in this process use the right version.
func CheckNodeVersion() error {
	data, err := os.ReadFile(".nvmrc")
	if err != nil {
		return fmt.Errorf("failed to read .nvmrc: %w", err)
	}
	expected := strings.TrimSpace(string(data))
	expected = strings.TrimPrefix(expected, "v")

	// Check if the current node already matches. Neither a missing node nor a
	// failed --version is fatal here: the nvm lookup below is the fallback.
	if nodeCmd, lookErr := ToolCommand("node", "--version"); lookErr == nil {
		if out, runErr := nodeCmd.Output(); runErr == nil {
			actual := strings.TrimPrefix(strings.TrimSpace(string(out)), "v")
			if actual == expected {
				return nil
			}
		}
	}

	// Current node doesn't match (or isn't found). Try to find the right
	// version in the nvm directory and prepend it to PATH.
	nvmDir := os.Getenv("NVM_DIR")
	if nvmDir == "" {
		home, _ := os.UserHomeDir()
		nvmDir = home + "/.nvm"
	}

	nodeBinDir := fmt.Sprintf("%s/versions/node/v%s/bin", nvmDir, expected)
	if _, err := os.Stat(nodeBinDir + "/node"); err != nil {
		return fmt.Errorf("node v%s is not installed via nvm (looked in %s). Run 'nvm install %s'", expected, nodeBinDir, expected)
	}

	// Prepend the correct node bin directory to PATH for this process.
	os.Setenv("PATH", nodeBinDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	fmt.Printf("Using node v%s from %s\n", expected, nodeBinDir)
	return nil
}

type wailsConfig struct {
	Info struct {
		ProductVersion string `json:"productVersion"`
		BetaExpiryDays int    `json:"betaExpiryDays"`
	} `json:"info"`
}

// Gets product version from wails.json
func getProductVersion() (string, error) {
	data, err := os.ReadFile("wails.json")
	if err != nil {
		return "", err
	}
	var wailsCfg wailsConfig
	if err := json.Unmarshal(data, &wailsCfg); err != nil {
		return "", err
	}
	return wailsCfg.Info.ProductVersion, nil
}

// If the version string contains "beta", consider it a beta version.
func isBeta(version string) bool {
	return strings.Contains(strings.ToLower(version), "beta")
}

// Gets beta expiry days from wails.json
func getBetaExpiryDays() (int, error) {
	data, err := os.ReadFile("wails.json")
	if err != nil {
		return 0, err
	}
	var wailsCfg wailsConfig
	if err := json.Unmarshal(data, &wailsCfg); err != nil {
		return 0, err
	}
	return wailsCfg.Info.BetaExpiryDays, nil
}

// GitRevParse returns the short git commit hash of the current HEAD.
func gitRevParse() string {
	cmd, err := ToolCommand("git", "rev-parse", "--short=9", "HEAD")
	if err != nil {
		return ""
	}
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// Credit to https://github.com/sfate
// https://gist.github.com/sfate/9d45f6c5405dc4c9bf63bf95fe6d1a7c
func PrettyPrint(args ...interface{}) {
	var caller string

	timeNow := time.Now().Format("01-02-2006 15:04:05")
	prefix := fmt.Sprintf("[%s] %s -- ", "PrettyPrint", timeNow)
	_, fileName, fileLine, ok := runtime.Caller(1)

	if ok {
		caller = fmt.Sprintf("%s:%d", fileName, fileLine)
	} else {
		caller = ""
	}

	fmt.Printf("\n%s%s\n", prefix, caller)

	if len(args) == 2 {
		label := args[0]
		value := args[1]

		s, _ := json.MarshalIndent(value, "", "\t")
		fmt.Printf("%s%s: %s\n", prefix, label, string(s))
	} else {
		s, _ := json.MarshalIndent(args, "", "\t")
		fmt.Printf("%s%s\n", prefix, string(s))
	}
}
