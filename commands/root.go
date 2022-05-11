package commands

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/vorlif/xspreak/config"
	"github.com/vorlif/xspreak/extract"
	"github.com/vorlif/xspreak/tmpl"
)

type Executor struct {
	rootCmd    *cobra.Command
	extractCmd *cobra.Command

	cfg *config.Config
	log *logrus.Entry

	contextLoader *extract.ContextLoader
}

func NewExecutor() *Executor {
	e := &Executor{
		cfg: config.NewDefault(),
		log: logrus.WithField("service", "executor"),
	}

	e.rootCmd = &cobra.Command{
		Use:     "xspreak",
		Version: VersionName,
		Short:   "String extraction for spreak.",
		Long:    `Simple tool to extract strings and create POT files for application translations.`,
		Run:     e.executeRun,
	}

	fs := pflag.NewFlagSet("config flag set", pflag.ContinueOnError)
	fs.ParseErrorsWhitelist = pflag.ParseErrorsWhitelist{UnknownFlags: true}
	fs.Usage = func() {}
	fs.SortFlags = false
	e.rootCmd.PersistentFlags().SortFlags = false
	initRootFlags(e.rootCmd.PersistentFlags(), config.NewDefault())
	initRootFlags(fs, e.cfg)

	if err := fs.Parse(os.Args); err != nil {
		if err != pflag.ErrHelp {
			logrus.WithError(err).Fatal("Args could not be parsed")
		}
	}

	if rawKeywords, err := fs.GetStringArray("template-keyword"); err != nil {
		logrus.WithError(err).Fatal("Args could not be parsed")
	} else if len(rawKeywords) == 0 {
		if keywordPrefix, errP := fs.GetString("template-prefix"); errP != nil {
			logrus.WithError(errP).Fatal("Args could not be parsed")
		} else {
			e.cfg.Keywords = tmpl.DefaultKeywords(keywordPrefix)
		}
	} else {
		for _, raw := range rawKeywords {
			if kw, errKw := tmpl.ParseKeywords(raw); errKw != nil {
				logrus.WithError(errKw).Fatal("Args could not be parsed")
			} else {
				e.cfg.Keywords = append(e.cfg.Keywords, kw)
			}
		}
	}

	if err := e.cfg.Prepare(); err != nil {
		logrus.Fatalf("Configuration could not be processed %v", err)
	}

	if e.cfg.IsVerbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	e.log.Debug("Starting execution...")

	e.initExtract()

	e.contextLoader = extract.NewContextLoader(e.cfg)

	return e
}

func (e *Executor) Execute() error {
	return e.rootCmd.Execute()
}

func initRootFlags(fs *pflag.FlagSet, cfg *config.Config) {
	def := config.NewDefault()

	fs.BoolVarP(&cfg.IsVerbose, "verbose", "V", def.IsVerbose, "Author name for copyright attribution")
	fs.StringVarP(&cfg.SourceDir, "directory", "D", def.SourceDir, "Directory with the Go source files")
	fs.StringVarP(&cfg.OutputDir, "output-dir", "p", def.OutputDir, "Directory in which the pot files are stored.")
	fs.StringVarP(&cfg.OutputFile, "output", "o", def.OutputFile, "Write output to specified file")
	fs.StringSliceVarP(&cfg.CommentPrefixes, "add-comments", "c", def.CommentPrefixes, "Place comment blocks starting with TAG and preceding keyword lines in output file")
	fs.BoolVarP(&cfg.ExtractErrors, "extract-errors", "e", def.ExtractErrors, "Strings from errors.New(STRING) are extracted")
	fs.StringVar(&cfg.ErrorContext, "errors-context", def.ErrorContext, "Context which is automatically assigned to extracted errors")

	fs.StringArrayVarP(&cfg.TemplatePatterns, "template-directory", "t", def.TemplatePatterns, "Set a list of paths to which the template files contain. Regular expressions can be used.")
	fs.String("template-prefix", ".T", "Sets a prefix for the translation functions, which is used within the templates")
	fs.StringArrayP("template-keyword", "k", []string{}, "Sets a keyword that is used within templates to identify translation functions")

	fs.StringVarP(&cfg.DefaultDomain, "default-domain", "d", def.DefaultDomain, "Use name.pot for output (instead of messages.pot)")
	fs.BoolVar(&cfg.WriteNoLocation, "no-location", def.WriteNoLocation, "Do not write '#: filename:line' lines")

	fs.IntVarP(&cfg.WrapWidth, "width", "w", def.WrapWidth, "Set output page width")
	fs.BoolVar(&cfg.DontWrap, "no-wrap", def.DontWrap, "Do not break long message lines, longer than the output page width, into several lines")

	fs.BoolVar(&cfg.OmitHeader, "omit-header", def.OmitHeader, "Don't write header with 'msgid \"\"' entry")
	fs.StringVar(&cfg.CopyrightHolder, "copyright-holder", def.CopyrightHolder, "Set copyright holder in output")
	fs.StringVar(&cfg.PackageName, "package-name", def.PackageName, "Set package name in output")
	fs.StringVar(&cfg.BugsAddress, "msgid-bugs-address", def.BugsAddress, "Set report address for msgid bugs")

	fs.DurationVar(&cfg.Timeout, "timeout", def.Timeout, "Timeout for total work")
}
