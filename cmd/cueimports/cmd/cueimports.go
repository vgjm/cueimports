/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func sortCueFile(name string, module string, save bool) error {
	b, err := os.ReadFile(name)
	if err != nil {
		return errors.New("could not read file")
	}
	// Find the import entries.
	rex := regexp.MustCompile(`import[\s]*\([\s]*([\s\S]*?)[\s]*\)`)
	res := rex.FindSubmatch(b)
	if len(res) < 2 {
		// can't find import entries
		return nil
	}
	// Sort entries.
	lines := strings.Split(string(res[1]), "\n")
	if err := sortEntries(lines, module); err != nil {
		return fmt.Errorf("sort imports for %s failed: %w", name, err)
	}
	for i, line := range lines {
		lines[i] = "\t" + strings.TrimSpace(line)
	}
	// Write file if save is true.
	oldImp := res[0]
	newImp := []byte(fmt.Sprintf("import (\n%s\n)", strings.Join(lines, "\n")))
	if !save {
		if !isSameBytes(oldImp, newImp) {
			fmt.Println(color.HiWhiteString(name), ":")
			fmt.Println(color.RedString(string(oldImp)))
			fmt.Println(color.GreenString(string(newImp)))
		}
		return nil
	}
	newb := rex.ReplaceAll(b, newImp)
	if err := os.WriteFile(name, newb, 0644); err != nil {
		return errors.New("could not write file")
	}
	return nil
}

func isSameBytes(b1, b2 []byte) bool {
	if len(b1) != len(b2) {
		return false
	}
	for i := range b1 {
		if b1[i] != b2[i] {
			return false
		}
	}
	return true
}

// If it is failed to get module, will return empty string
func getModule(name string) string {
	module := ""
	modRe := regexp.MustCompile(`module:\s*\"(\S+)\"`)
	for cwd, prewd := name, ""; cwd != prewd; cwd, prewd = filepath.Dir(cwd), cwd {
		cur := filepath.Join(cwd, "cue.mod", "module.cue")
		b, err := os.ReadFile(cur)
		if err != nil {
			continue
		}
		mods := modRe.FindSubmatch(b)
		if len(mods) > 1 {
			module = string(mods[1])
		}
	}
	return module
}

// root should be absolute path
func walk(root, module string, save bool) []error {
	errs := make([]error, 0)
	fileSystem := os.DirFS(root)
	fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Skip non-cue files
		_, base := filepath.Split(path)
		rex := regexp.MustCompile(`^[^\.]\S*\.cue$`)
		if !rex.Match([]byte(base)) {
			return nil
		}
		// Skip third party files
		absPath := filepath.Join(root, path)
		if strings.Contains(absPath, filepath.Join("cue.mod", "pkg")) {
			return nil
		}
		fi, err := os.Stat(absPath)
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}
		if err := sortCueFile(absPath, module, save); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w, skipped", absPath, err))
		}
		return nil
	})
	return errs
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cueimports",
	Short: "Sort the imports of cue file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := args[0]
		save := cmd.Flag("save").Value.String() == "true"

		pwd, err := os.Getwd()
		if err != nil {
			return err
		}
		absRoot, err := filepath.Abs(filepath.Join(pwd, root))
		if err != nil {
			return err
		}
		mod := getModule(absRoot)

		fi, err := os.Stat(root)
		if err != nil {
			return err
		}
		if fi.IsDir() {
			errs := walk(absRoot, mod, save)
			if len(errs) > 0 {
				color.Red(fmt.Sprintf("\nCueimports completed, but failed with %d files:\n\n", len(errs)))
			}
			for _, err := range errs {
				fmt.Println(err)
			}
			return nil
		}
		return sortCueFile(absRoot, mod, save)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("save", "s", false, "Save changes")
}
