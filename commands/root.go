package commands

import (
	"runtime/debug"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/vorlif/xspreak/config"
	"github.com/vorlif/xspreak/tmpl"
)

// Version is initialized via ldflags or debug.BuildInfo.
var Version = ""

var (
	extractCfg = config.NewDefault()

	rootCmd = &cobra.Command{
		Use:     "xspreak",
		Version: Version,
		Short:   "String extraction for spreak.",
		Long:    `Simple tool to extract strings and create POT/JSON files for application translations.`,
		Run:     extractCmdF,
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	initVersionNumber()

	def := config.NewDefault()

	rootCmd.PersistentFlags().BoolVarP(&extractCfg.IsVerbose, "verbose", "V", def.IsVerbose, "increase verbosity level")
	rootCmd.PersistentFlags().DurationVar(&extractCfg.Timeout, "timeout", def.Timeout, "Timeout for total work")

	fs := rootCmd.Flags()
	fs.SortFlags = false
	fs.StringVarP(&extractCfg.ExtractFormat, "format", "f", def.ExtractFormat, "Output format of the extraction. Valid values are 'pot' and 'json'.")
	fs.StringVarP(&extractCfg.SourceDir, "directory", "D", def.SourceDir, "Directory with the Go source files")
	fs.StringVarP(&extractCfg.OutputDir, "output-dir", "p", def.OutputDir, "Directory in which the pot files are stored.")
	fs.StringVarP(&extractCfg.OutputFile, "output", "o", def.OutputFile, "Write output to specified file")
	fs.StringSliceVarP(&extractCfg.CommentPrefixes, "add-comments", "c", def.CommentPrefixes, "Place comment blocks starting with TAG and preceding keyword lines in output file")
	fs.BoolVarP(&extractCfg.ExtractErrors, "extract-errors", "e", def.ExtractErrors, "Strings from errors.New(STRING) are extracted")
	fs.StringVar(&extractCfg.ErrorContext, "errors-context", def.ErrorContext, "Context which is automatically assigned to extracted errors")

	fs.StringArrayVarP(&extractCfg.TemplatePatterns, "template-directory", "t", []string{}, "Set a list of paths to which the template files contain. Regular expressions can be used.")
	fs.String("template-prefix", "", "Sets a prefix for the translation functions, which is used within the templates")
	fs.BoolVar(&extractCfg.TmplIsMonolingual, "template-use-kv", false, "Determines whether the strings from templates should be handled as key-value")
	fs.StringArrayP("template-keyword", "k", []string{}, "Sets a keyword that is used within templates to identify translation functions")

	fs.StringVarP(&extractCfg.DefaultDomain, "default-domain", "d", def.DefaultDomain, "Use name.pot for output (instead of messages.pot)")
	fs.BoolVar(&extractCfg.WriteNoLocation, "no-location", def.WriteNoLocation, "Do not write '#: filename:line' lines")

	fs.IntVarP(&extractCfg.WrapWidth, "width", "w", def.WrapWidth, "Set output page width")
	fs.BoolVar(&extractCfg.DontWrap, "no-wrap", def.DontWrap, "Do not break long message lines, longer than the output page width, into several lines")

	fs.BoolVar(&extractCfg.OmitHeader, "omit-header", def.OmitHeader, "Don't write header with 'msgid \"\"' entry")
	fs.StringVar(&extractCfg.CopyrightHolder, "copyright-holder", def.CopyrightHolder, "Set copyright holder in output")
	fs.StringVar(&extractCfg.PackageName, "package-name", def.PackageName, "Set package name in output")
	fs.StringVar(&extractCfg.BugsAddress, "msgid-bugs-address", def.BugsAddress, "Set report address for msgid bugs")
}

func initVersionNumber() {
	// If already set via ldflags, the value is retained.
	if Version != "" {
		return
	}

	info, available := debug.ReadBuildInfo()
	if available {
		Version = info.Main.Version
	} else {
		Version = "dev"
	}

	rootCmd.Version = Version
}

func extractCmdF(cmd *cobra.Command, args []string) {
	validateExtractConfig(cmd)
	extractCfg.Args = args

	extractor := NewExtractor()
	extractor.extract()
}

func validateExtractConfig(cmd *cobra.Command) {
	fs := cmd.Flags()
	if keywordPrefix, errP := fs.GetString("template-prefix"); errP != nil {
		log.WithError(errP).Fatal("Args could not be parsed")
	} else if keywordPrefix != "" {
		extractCfg.Keywords = tmpl.DefaultKeywords(keywordPrefix, extractCfg.TmplIsMonolingual)
	}

	if rawKeywords, err := fs.GetStringArray("template-keyword"); err != nil {
		log.WithError(err).Fatal("Args could not be parsed")
	} else {
		for _, raw := range rawKeywords {
			if kw, errKw := tmpl.ParseKeywords(raw, extractCfg.TmplIsMonolingual); errKw != nil {
				log.WithError(errKw).Fatalf("Arg could not be parsed %s", raw)
			} else {
				extractCfg.Keywords = append(extractCfg.Keywords, kw)
			}
		}
	}

	if err := extractCfg.Prepare(); err != nil {
		log.Fatalf("Configuration could not be processed: %v", err)
	}

	if extractCfg.IsVerbose {
		log.SetLevel(log.DebugLevel)
	}

	log.Debug("Starting execution...")
}
