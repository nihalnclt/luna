package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type Repository struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type PackageJSON struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Main        string            `json:"main"`
	License     string            `json:"license"`
	Scripts     map[string]string `json:"scripts"`
	Author      string            `json:"author"`
	Repository  *Repository       `json:"repository,omitempty"`
	Description string            `json:"description"`
}

var yesFlag bool

func (p PackageJSON) MarshalJSON() ([]byte, error) {
	type Alias PackageJSON
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(&p), // Use the alias to avoid recursion
	})
}

func isURLFriendly(s string) bool {
	// Use a regular expression to allow only URL-friendly characters
	regexp := regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	return regexp.MatchString(s)
}

func GetBaseFolderName() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return filepath.Base(cwd), nil
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init command short",
	Long:  "init command long",
	Run: func(cmd *cobra.Command, args []string) {
		start := time.Now()

		packageName, err := GetBaseFolderName()
		if err != nil {
			fmt.Println("Error getting package name:", err)
			return
		}

		packageJSON := PackageJSON{
			Name:    packageName,
			Version: "1.0.0",
			Main:    "index.js",
			License: "MIT",
			// TODO:
			// Fix the JSON marshalling issue on &&
			Scripts: map[string]string{
				"test": "echo \"Error: no test specified\" && exit 1",
			},
			Author:      "",
			Description: "",
		}

		if !yesFlag {
			reader := bufio.NewReader(os.Stdin)

			for {
				fmt.Printf("package name: (%s) ", packageJSON.Name)
				input, _ := reader.ReadString('\n')
				projectName := strings.TrimSpace(input)

				if projectName == "" {
					break
				}

				if isURLFriendly(projectName) {
					packageJSON.Name = strings.ToLower(projectName)
					break
				}

				fmt.Println("Error: Project name must be URL-friendly (lowercase letters, numbers, and hyphens only)")
			}

			fmt.Printf("version (%s): ", packageJSON.Version)
			versionInput, _ := reader.ReadString('\n')
			version := strings.TrimSpace(versionInput)
			if version != "" {
				packageJSON.Version = version
			}

			fmt.Printf("description: ")
			descriptionInput, _ := reader.ReadString('\n')
			description := strings.TrimSpace(descriptionInput)
			if description != "" {
				packageJSON.Description = description
			}

			fmt.Printf("git repository: ")
			repositoryInput, _ := reader.ReadString('\n')
			repository := strings.TrimSpace(repositoryInput)
			if repository != "" {
				packageJSON.Repository = &Repository{}

				packageJSON.Repository.Type = "git"
				packageJSON.Repository.URL = repository
			}

			fmt.Printf("author: ")
			authorInput, _ := reader.ReadString('\n')
			author := strings.TrimSpace(authorInput)
			if author != "" {
				packageJSON.Author = author
			}

			// TODO:
			// Ask License here (There is some validation for license in npm).
		}

		file, err := os.Create("package.json")
		if err != nil {
			fmt.Println("Error creating package.json file:", err)
			return
		}
		defer file.Close()

		jsonData, err := json.MarshalIndent(packageJSON, "", "  ")
		if err != nil {
			fmt.Println("Error marshalling JSON:", err)
			return
		}

		if _, err := file.Write(jsonData); err != nil {
			fmt.Println("Error writing to package.json file:", err)
			return
		}

		elapsed := time.Since(start).Seconds()
		fmt.Printf("%ssuccess%s Saved package.json\n", Green, Reset)
		fmt.Printf("✨ Finished in %.2fs.\n", elapsed)
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Use default value for package.json")
}
