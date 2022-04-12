package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var overrides = map[string]*GetSingleChainResponseJSON{
	"juno": {
		Chain: ChainResponseJSON{
			ChainName:  "juno",
			DaemonName: "junod",
			Codebase: CodebaseJSON{
				GitRepo:            "https://github.com/CosmosContracts/juno",
				RecommendedVersion: "v3.1.0",
			},
		},
	},
}

var chainsToInclude = map[string]bool{
	"agoric":         true,
	"akash":          true,
	"bitsong":        true,
	"bostrom":        true,
	"cerberus":       true,
	"cheqd":          true,
	"cosmoshub":      true,
	"cryptoorgchain": true,
	"desmos":         true,
	"fetchhub":       true,
	"gravitybridge":  true,
	"juno":           true,
	"lumnetwork":     true,
	"osmosis":        true,
	"persistence":    true,
	"regen":          true,
	"rizon":          true,
	"secretnetwork":  true,
	"sentinel":       true,
	"stargaze":       true,
	"terra":          true,
	"umee":           true,
}

var targets = map[string][]string{
	"linux":   {"amd64"},
	"darwin":  {"amd64", "arm64"},
	"windows": {"amd64"},
}

type GetAllChainsResponseJSON struct {
	Chains []struct {
		ChainName string `json:"chain_name"`
	} `json:"chains"`
}

type GetSingleChainResponseJSON struct {
	Chain ChainResponseJSON `json:"chain"`
}

type ChainResponseJSON struct {
	ChainName  string       `json:"chain_name"`
	Codebase   CodebaseJSON `json:"codebase"`
	DaemonName string       `json:"daemon_name"`
}

type CodebaseJSON struct {
	GitRepo            string `json:"git_repo"`
	RecommendedVersion string `json:"recommended_version"`
}

const url = "https://chains.cosmos.directory/"

func main() {
	allRes, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer allRes.Body.Close()

	if allRes.StatusCode != 200 {
		panic("Status code was not 200")
	}

	chainsRes := GetAllChainsResponseJSON{}
	err = json.NewDecoder(allRes.Body).Decode(&chainsRes)
	if err != nil {
		panic(err)
	}

	for _, c := range chainsRes.Chains {
		if !chainsToInclude[c.ChainName] {
			continue
		}

		cjson, err := getSingleChain(c.ChainName)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		if cjson.Chain.Codebase.GitRepo == "" {
			fmt.Printf("No git repo found for %s\n", c.ChainName)
			continue
		}

		if cjson.Chain.DaemonName == "" {
			fmt.Printf("No daemon name found for %s\n", c.ChainName)
			continue
		}

		if err := clone(cjson.Chain); err != nil {
			fmt.Println(err.Error())
			continue
		}

		if err := build(cjson.Chain); err != nil {
			fmt.Println(err.Error())
			continue
		}
	}
}

func getSingleChain(chainName string) (GetSingleChainResponseJSON, error) {
	override, ok := overrides[chainName]
	if ok && override != nil {
		return *override, nil
	}

	cRes, err := http.Get(url + chainName)
	if err != nil {
		panic(err)
	}
	defer cRes.Body.Close()

	if cRes.StatusCode != 200 {
		return GetSingleChainResponseJSON{}, errors.New(fmt.Sprintf("Something went wrong getting %s with code %d\n", url+chainName, cRes.StatusCode))
	}

	cjson := GetSingleChainResponseJSON{}
	err = json.NewDecoder(cRes.Body).Decode(&cjson)
	if err != nil {
		panic(err)
	}

	return cjson, nil
}

func clone(c ChainResponseJSON) error {
	if c.Codebase.RecommendedVersion == "" {
		return errors.New(fmt.Sprintf("No recommended version found for %s\n", c.ChainName))
	}

	if _, err := os.Stat(c.ChainName); os.IsNotExist(err) {
		gitURI := strings.TrimSuffix(strings.TrimSuffix(c.Codebase.GitRepo, "/"), ".git") + ".git"
		cmd := exec.Command("gh", "repo", "clone", gitURI, c.ChainName, "--", "-b", c.Codebase.RecommendedVersion)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			return errors.New(fmt.Sprintf("git clone failed for %s (%s)\n", c.ChainName, gitURI))
		}
	}

	return nil
}

func build(c ChainResponseJSON) error {
	if _, err := os.Stat("release-builds"); os.IsNotExist(err) {
		if err := os.Mkdir("release-builds", 0775); err != nil {
			panic(err)
		}
	}

	if err := os.Chdir(c.ChainName); err != nil {
		panic(err)
	}

	cmdName := "make"
	cmdArgs := []string{"build"}
	overridePath := "../../override-build-files/" + c.ChainName + ".sh"
	if _, err := os.Stat(overridePath); !os.IsNotExist(err) {
		fmt.Println("Found override for " + c.ChainName)
		cmdName = overridePath
		cmdArgs = []string{}
	}

	goos := runtime.GOOS
	archs := targets[goos]
	for _, arch := range archs {
		if isBuilt(c.DaemonName, goos, arch, c.Codebase.RecommendedVersion) {
			fmt.Printf("%s %s (%s, %s) exists, skipping\n", c.ChainName, c.Codebase.RecommendedVersion, goos, arch)
			continue
		}
		fmt.Printf("Building %s %s (%s, %s)\n", c.ChainName, c.Codebase.RecommendedVersion, goos, arch)

		cmd := exec.Command(cmdName, cmdArgs...)
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "GOOS="+goos)
		cmd.Env = append(cmd.Env, "GOARCH="+arch)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			fmt.Printf("build failed for %s\n (%s, %s)\n", c.ChainName, goos, arch)
			continue
		}

		files, err := ioutil.ReadDir("./build")
		if err != nil {
			panic(err)
		}
		if len(files) != 1 {
			panic(errors.New("Expected exactly one binary to be found"))
		}

		fn := files[0].Name()
		ext := filepath.Ext(fn)
		newName := fmt.Sprintf("%s-%s-%s-%s%s", c.DaemonName, goos, arch, c.Codebase.RecommendedVersion, ext)
		newPath := "../release-builds/" + newName
		fmt.Println("fn", fn)
		fmt.Println("ext", ext)
		fmt.Println("newName", newName)
		fmt.Println("newPath", newPath)
		if err := os.Rename("./build/"+fn, newPath); err != nil {
			panic(err)
		}
	}

	if err := os.Chdir(".."); err != nil {
		panic(err)
	}

	return nil
}

func isBuilt(daemon string, o string, a string, version string) bool {
	files, err := ioutil.ReadDir("../release-builds")
	if err != nil {
		panic(err)
	}

	for _, f := range files {
		if strings.HasPrefix(f.Name(), fmt.Sprintf("%s-%s-%s-%s", daemon, o, a, version)) {
			return true
		}
	}

	return false
}
