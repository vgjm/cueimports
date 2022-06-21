/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var (
	grepEntry *regexp.Regexp
)

type ImportList []string

func (l ImportList) Len() int {
	return len(l)
}

func (l ImportList) Less(i, j int) bool {
	var iEntry, jEntry string
	res := grepEntry.FindStringSubmatch(l[i])
	if res != nil {
		iEntry = res[1]
	}
	res = grepEntry.FindStringSubmatch(l[j])
	if res != nil {
		jEntry = res[1]
	}
	return iEntry < jEntry
}

func (l ImportList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cueimports",
	Short: "Sort the imports of cue file",
	Args:  cobra.ExactArgs(1),
	// Uncomment the following line if your bare application
	// has an action associated with it:
	RunE: func(cmd *cobra.Command, args []string) error {
		grepEntry = regexp.MustCompile(`[\s\S]*\"([\s\S]+)\"`)
		root := args[0]
		fi, err := os.Stat(root)
		if err != nil {
			return err
		}
		if fi.IsDir() {
			fileSystem := os.DirFS(root)
			fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					log.Fatal(err)
				}
				rex := regexp.MustCompile(`\.cue$`)
				if !rex.Match([]byte(path)) {
					return nil
				}
				rex = regexp.MustCompile(`cue.mod/pkg`)
				if rex.Match([]byte(path)) {
					return nil
				}
				absPath := filepath.Join(root, path)
				if err := fmtCueImports(absPath); err != nil {
					log.Fatal(err)
				}
				return nil
			})
			return nil
		}
		return fmtCueImports(root)
	},
}

func fmtCueImports(name string) error {
	b, err := os.ReadFile(name)
	if err != nil {
		return fmt.Errorf("could not read file %s", name)
	}
	rex := regexp.MustCompile(`import[\s]*\(([\s\S]*?)\)`)
	res := rex.FindSubmatch(b)
	if res == nil {
		return nil
	}
	imp := res[1]
	lines := strings.Split(string(imp), "\n")
	for i, line := range lines {
		lines[i] = "\t" + strings.TrimSpace(line)
	}
	sl := ImportList(lines)
	sort.Sort(sl)
	spi := -1
	for i := range lines {
		if lines[i] == "\t" {
			spi = i
		}
	}
	if spi != -1 {
		lines = lines[spi+1:]
	}
	nimp := fmt.Sprintf("import (\n%s\n)", strings.Join(lines, "\n"))
	newb := rex.ReplaceAll(b, []byte(nimp))
	if err := os.WriteFile(name, newb, 0644); err != nil {
		return fmt.Errorf("could not write file %s", name)
	}
	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cueimports.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("save", "s", false, "Save changes")
}
