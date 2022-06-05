package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/vorlif/xspreak/config"
	"github.com/vorlif/xspreak/encoder"
	"github.com/vorlif/xspreak/extract"
	"github.com/vorlif/xspreak/extract/extractors"
	"github.com/vorlif/xspreak/goextractors"
	"github.com/vorlif/xspreak/result"
	"github.com/vorlif/xspreak/tmpl"
	"github.com/vorlif/xspreak/tmplextractors"
	"github.com/vorlif/xspreak/util"
)

// Version can be set at link time.
var Version = "0.0.0"

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

	fs.StringArrayVarP(&extractCfg.TemplatePatterns, "template-directory", "t", def.TemplatePatterns, "Set a list of paths to which the template files contain. Regular expressions can be used.")
	fs.String("template-prefix", ".T", "Sets a prefix for the translation functions, which is used within the templates")
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

func validateExtractConfig(cmd *cobra.Command) {
	fs := cmd.Flags()
	if keywordPrefix, errP := fs.GetString("template-prefix"); errP != nil {
		log.WithError(errP).Fatal("Args could not be parsed")
	} else {
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

func extractCmdF(cmd *cobra.Command, args []string) {
	validateExtractConfig(cmd)
	extractCfg.Args = args

	extractor := NewExtractor()
	extractor.extract()
}

type Extractor struct {
	cfg *config.Config
	log *log.Entry

	contextLoader *extract.ContextLoader
}

func NewExtractor() *Extractor {
	return &Extractor{
		cfg:           extractCfg,
		log:           log.WithField("service", "extractor"),
		contextLoader: extract.NewContextLoader(extractCfg),
	}
}

func (e *Extractor) extract() {
	ctx, cancel := context.WithTimeout(context.Background(), e.cfg.Timeout)
	defer cancel()

	extractedIssues, errE := e.runExtraction(ctx)
	if errE != nil {
		e.log.Fatalf("Running error: %s", errE)
	}

	domainIssues := make(map[string][]result.Issue)
	start := time.Now()
	for _, iss := range extractedIssues {
		if _, ok := domainIssues[iss.Domain]; !ok {
			domainIssues[iss.Domain] = []result.Issue{iss}
		} else {
			domainIssues[iss.Domain] = append(domainIssues[iss.Domain], iss)
		}
	}
	log.Debugf("sort extractions took %s", time.Since(start))

	if len(extractedIssues) == 0 {
		domainIssues[""] = make([]result.Issue, 0)
		log.Println("No Strings found")
	}

	e.saveDomains(domainIssues)
}

func (e *Extractor) runExtraction(ctx context.Context) ([]result.Issue, error) {
	util.TrackTime(time.Now(), "run all extractors")
	extractorsToRun := []extractors.Extractor{
		goextractors.NewDefinitionExtractor(),
		goextractors.NewCommentsExtractor(),
		goextractors.NewFuncCallExtractor(),
		goextractors.NewFuncReturnExtractor(),
		goextractors.NewGlobalAssignExtractor(),
		goextractors.NewSliceDefExtractor(),
		goextractors.NewMapsDefExtractor(),
		goextractors.NewStructDefExtractor(),
		goextractors.NewVariablesExtractor(),
		goextractors.NewErrorExtractor(),
		goextractors.NewInlineTemplateExtractor(),
		tmplextractors.NewCommandExtractor(),
	}

	extractCtx, err := e.contextLoader.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("context loading failed: %w", err)
	}

	runner, err := extract.NewRunner(e.cfg, extractCtx.Packages)
	if err != nil {
		return nil, err
	}

	issues, err := runner.Run(ctx, extractCtx, extractorsToRun)
	if err != nil {
		return nil, err
	}

	return issues, nil
}

func (e *Extractor) saveDomains(domains map[string][]result.Issue) {
	util.TrackTime(time.Now(), "save files")
	for domainName, issues := range domains {
		var outputFile string
		if domainName == "" {
			outputFile = filepath.Join(e.cfg.OutputDir, e.cfg.OutputFile)
		} else {
			outputFile = filepath.Join(e.cfg.OutputDir, domainName+"."+e.cfg.ExtractFormat)
		}

		outputDir := filepath.Dir(outputFile)
		if _, err := os.Stat(outputDir); os.IsNotExist(err) {
			log.Printf("Output folder does not exist, trying to create it: %s\n", outputDir)
			if errC := os.MkdirAll(outputDir, os.ModePerm); errC != nil {
				log.Fatalf("Output folder does not exist and could not be created: %s", errC)
			}
		}

		dst, err := os.Create(outputFile)
		if err != nil {
			e.log.WithError(err).Fatal("Output file could not be created")
		}
		defer dst.Close()

		var enc encoder.Encoder
		if e.cfg.ExtractFormat == config.ExtractFormatPot {
			enc = encoder.NewPotEncoder(e.cfg, dst)
		} else {
			enc = encoder.NewJSONEncoder(dst, "  ")
		}

		if errEnc := enc.Encode(issues); errEnc != nil {
			e.log.WithError(errEnc).Fatal("Output file could not be written")
		}

		_ = dst.Close()
		log.Printf("File written: %s\n", outputFile)
	}
}
