package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var goos = []string{"linux", "darwin", "windows"}
var goarch = []string{"amd64", "arm64"}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "build all cloned repos",
	RunE: func(cmd *cobra.Command, args []string) error {
		files, err := ioutil.ReadDir("./")
		if err != nil {
			log.Fatal(err)
		}

		if _, err := os.Stat("release-builds"); os.IsNotExist(err) {
			if err := os.Mkdir("release-builds", 0775); err != nil {
				return err
			}
		}

		for _, f := range files {
			if err := os.Chdir(f.Name()); err != nil {
				return err
			}

			for _, o := range goos {
				for _, a := range goarch {
					fmt.Printf("Building %s (%s, %s)\n", f.Name(), o, a)

					cmd := exec.Command("make", "install")
					cmd.Env = os.Environ()
					cmd.Env = append(cmd.Env, "GOOS="+o)
					cmd.Env = append(cmd.Env, "GOARCH="+a)
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					err = cmd.Run()
					if err != nil {
						fmt.Printf("build failed for %s\n (%s, %s)\n", f.Name(), o, a)
						continue
					}

					whichCmd := exec.Command("which ")
					fn := bf[0].Name()
					ext := filepath.Ext(fn)
					newName := fmt.Sprintf("%s-%s-%s%s", strings.TrimSuffix(fn, ext), o, a, ext)
					if err := os.Rename("./build/"+bf[0].Name(), "../release-builds/"+newName); err != nil {
						return err
					}
				}
			}

			if err := os.Chdir(".."); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
