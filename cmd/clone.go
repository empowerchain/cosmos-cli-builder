package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

type GetAllChainsResponseJSON struct {
	Chains []struct {
		ChainName string `json:"chain_name"`
	} `json:"chains"`
}

type GetSingleChainResponseJSON struct {
	Chain ChainResponseJSON `json:"chain"`
}

type ChainResponseJSON struct {
	ChainName string       `json:"chain_name"`
	Codebase  CodebaseJSON `json:"codebase"`
}

type CodebaseJSON struct {
	GitRepo            string `json:"git_repo"`
	RecommendedVersion string `json:"recommended_version"`
}

var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "Clones all the repos",
	RunE: func(cmd *cobra.Command, args []string) error {
		url := "https://chains.cosmos.directory/"
		allRes, err := http.Get(url)
		if err != nil {
			return err
		}
		defer allRes.Body.Close()

		if allRes.StatusCode != 200 {
			panic("Status code was not 200")
		}

		chainsRes := GetAllChainsResponseJSON{}
		err = json.NewDecoder(allRes.Body).Decode(&chainsRes)
		if err != nil {
			return err
		}

		for _, c := range chainsRes.Chains {
			cRes, err := http.Get(url + c.ChainName)
			if err != nil {
				panic(err)
			}
			defer cRes.Body.Close()

			if cRes.StatusCode != 200 {
				fmt.Printf("Something went wrong getting %s with code %d\n", url+c.ChainName, cRes.StatusCode)
				continue
			}

			cjson := GetSingleChainResponseJSON{}
			err = json.NewDecoder(cRes.Body).Decode(&cjson)
			if err != nil {
				panic(err)
			}

			if cjson.Chain.Codebase.GitRepo == "" {
				fmt.Printf("No git repo found for %s\n", c.ChainName)
				continue
			}

			if cjson.Chain.Codebase.RecommendedVersion == "" {
				fmt.Printf("No recommended version found for %s\n", c.ChainName)
				continue
			}

			if _, err := os.Stat(c.ChainName); os.IsNotExist(err) {
				gitURI := strings.TrimSuffix(strings.TrimSuffix(cjson.Chain.Codebase.GitRepo, "/"), ".git") + ".git"
				cmd := exec.Command("gh", "repo", "clone", gitURI, c.ChainName, "--", "-b", cjson.Chain.Codebase.RecommendedVersion)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err = cmd.Run()
				if err != nil {
					fmt.Printf("git clone failed for %s (%s)\n", c.ChainName, gitURI)
					continue
				}
			}

		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(cloneCmd)
}
