package commands

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/text/language"

	"github.com/vorlif/spreak/catalog/cldrplural"

	"github.com/vorlif/xspreak/merger"
)

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Create or update JSON translation files",
	Long: `Merge creates new translation files or merges existing files.
Already existing translations in the target file are preserved.
Available translations from the source file will be taken over. 
The language of the target file must be specified to create the correct templates for the plural forms.`,
	Run:     mergeCmdF,
	Example: `  xspreak merge -i locale/httptempl.json -o locale/de.json -l de`,
}

func init() {
	fs := mergeCmd.Flags()
	fs.SortFlags = false
	fs.StringP("input", "i", "", "source file")
	fs.StringP("output", "o", "", "output file")
	fs.StringP("lang", "l", "", "destination language")

	rootCmd.AddCommand(mergeCmd)
}

func mergeCmdF(cmd *cobra.Command, _ []string) {
	targetLang, errL := cmd.Flags().GetString("lang")
	if errL != nil {
		log.WithError(errL).Fatal("Invalid source file")
	} else if targetLang == "" {
		log.Fatal("Target language must be specified")
	}

	lang, errP := language.Parse(targetLang)
	if errP != nil {
		log.WithError(errP).Fatal("Language could not be parsed")
	}
	ruleSet, found := cldrplural.ForLanguage(lang)
	if !found {
		log.Fatal("No rules for language found")
	}

	srcPath, errS := cmd.Flags().GetString("input")
	if errS != nil {
		log.WithError(errS).Fatal("Invalid source file")
	} else if srcPath == "" {
		log.Fatal("Source required")
	}

	dstPath, errD := cmd.Flags().GetString("output")
	if errD != nil {
		log.WithError(errD).Fatal("Invalid destination file")
	} else if dstPath == "" {
		log.Fatal("Destination required")
	}

	var sourceContent []byte
	if fi, err := os.Stat(srcPath); err != nil {
		log.WithError(err).Fatal("Source file could not be verified")
	} else if fi.IsDir() {
		log.Fatal("Source file must be a file, but is a folder")
	} else {
		sourceContent, err = os.ReadFile(srcPath)
		if err != nil {
			log.WithError(err).Fatal("Source file could not be read")
		}
	}

	var destinationContent []byte
	if fi, err := os.Stat(dstPath); err != nil {
		if !os.IsNotExist(err) {
			log.WithError(err).Fatal("Destination file could not be verified")
		}
	} else if fi.IsDir() {
		log.Fatal("Destination file must be a file, but is a folder")
	} else {
		destinationContent, err = os.ReadFile(dstPath)
		if err != nil {
			log.WithError(err).Fatal("Destination file could not be read")
		}
	}

	newContent := merger.MergeJSON(sourceContent, destinationContent, ruleSet.Categories)
	if err := os.WriteFile(dstPath, newContent, 0666); err != nil {
		log.WithError(err).Fatal("Target file could not be written")
	}
	log.Printf("Target file written %s\n", dstPath)
}
