package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/renderer"
	"github.com/spf13/cobra"
)

type generateCmdOptions struct {
	outputPath string
	dummy      bool
}

// generateCmd represents the generate command
func NewGenerateCmd() *cobra.Command {
	opts := new(generateCmdOptions)
	cmd := &cobra.Command{
		Use:   "generate [oapi_file]",
		Short: "Generate message ",
		Long:  ``,
		Args:  cobra.ExactArgs(1),
		RunE:  opts.runGenerateCmd,
	}
	cmd.Flags().StringVarP(&opts.outputPath, "output", "o", "", "-o [outputPath]")
	cmd.Flags().BoolVar(&opts.dummy, "dummy", false, "--dummy")

	return cmd
}

func (opts *generateCmdOptions) runGenerateCmd(cmd *cobra.Command, args []string) error {
	// create a new JSON mock generator
	mg := renderer.NewMockGenerator(renderer.JSON)

	// tell the mock generator to pretty print the output
	mg.SetPretty()

	openapiFile, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}

	document, err := libopenapi.NewDocument(openapiFile)
	if err != nil {
		return err
	}

	v3doc, errors := document.BuildV3Model()
	if len(errors) > 0 {
		for i := range errors {
			fmt.Printf("error: %e\n", errors[i])
		}
		panic(fmt.Sprintf("cannot create v3 model from document: %d errors reported", len(errors)))
	}

	if err := os.MkdirAll(opts.outputPath, os.ModePerm); err != nil {
		return err
	}

	// Iterate through schemas and write each to a JSON file
	for modelName, schemaRef := range v3doc.Model.Components.Schemas.FromNewest() {
		schema := schemaRef.Schema()
		mock, err := mg.GenerateMock(schema, modelName)
		if err != nil {
			return err
		}

		// Save JSON to a file
		filePath := filepath.Join(opts.outputPath, modelName+".json")
		if err := os.WriteFile(filePath, mock, 0644); err != nil {
			fmt.Printf("Error writing model '%s' to file: %v\n", modelName, err)
			continue
		}

		fmt.Printf("Model '%s' written to '%s'\n", modelName, filePath)
	}
	return nil
}
