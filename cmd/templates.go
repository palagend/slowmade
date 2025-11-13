// cmd/templates.go
package cmd

import (
	"os"
	"path/filepath"

	"github.com/palagend/slowmade/internal/mvc/views"
	"github.com/spf13/cobra"
)

var templateCmd = &cobra.Command{
	Use:   "templates",
	Short: "Manage custom templates",
}

var templateExtractCmd = &cobra.Command{
	Use:   "extract [output-dir]",
	Short: "Extract default templates to directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		outputDir := "."
		if len(args) > 0 {
			outputDir = args[0]
		}

		return extractDefaultTemplates(outputDir)
	},
}

func extractDefaultTemplates(outputDir string) error {
	templates := []string{
		"wallet_created.tmpl",
		"wallet_list.tmpl",
		"wallet_info.tmpl",
		"qr_ascii.tmpl",
		"transaction.tmpl",
		// ... 其他模板
	}

	for _, tmplName := range templates {
		content, err := views.GetDefaultTemplate(tmplName)
		if err != nil {
			return err
		}

		outputPath := filepath.Join(outputDir, tmplName)
		if err := os.WriteFile(outputPath, content, 0644); err != nil {
			return err
		}

		templateCmd.Printf("Extracted: %s\n", outputPath)
	}

	templateCmd.Println("Default templates extracted. Modify them and enable custom templates in config.")
	return nil
}

func init() {
	templateCmd.AddCommand(templateExtractCmd)
	rootCmd.AddCommand(templateCmd)
}
