package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/vorlif/xspreak/config"
	"github.com/vorlif/xspreak/encoder"
	"github.com/vorlif/xspreak/extract"
	"github.com/vorlif/xspreak/extract/extractors"
	"github.com/vorlif/xspreak/goextractors"
	"github.com/vorlif/xspreak/result"
	"github.com/vorlif/xspreak/tmplextractors"
	"github.com/vorlif/xspreak/util"
)

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
