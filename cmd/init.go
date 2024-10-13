package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nihalnclt/luna/models"
	"github.com/nihalnclt/luna/utils"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init command short",
	Long:  "init command long",
	Run: func(cmd *cobra.Command, args []string) {
		start := time.Now()

		packageName, err := utils.GetBaseFolderName()
		if err != nil {
			fmt.Println("Error getting package name:", err)
			return
		}

		packageJSON := models.PackageJSON{
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

		if yesFlag, _ := cmd.PersistentFlags().GetBool("yes"); !yesFlag {
			reader := bufio.NewReader(os.Stdin)

			for {
				fmt.Printf("package name: (%s) ", packageJSON.Name)
				input, _ := reader.ReadString('\n')
				projectName := strings.TrimSpace(input)

				if projectName == "" {
					break
				}

				if utils.IsURLFriendly(projectName) {
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
				packageJSON.Repository = &models.Repository{}

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
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.PersistentFlags().BoolP("yes", "y", false, "Use default value for package.json")
}
