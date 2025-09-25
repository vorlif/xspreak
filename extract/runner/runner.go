package runner

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/tools/go/packages"

	"github.com/vorlif/xspreak/config"
	"github.com/vorlif/xspreak/extract"
	processors2 "github.com/vorlif/xspreak/extract/processors"
	"github.com/vorlif/xspreak/util"
)

type Runner struct {
	Processors []processors2.Processor
	Log        *logrus.Entry
}

func New(cfg *config.Config, _ []*packages.Package) (*Runner, error) {
	p := []processors2.Processor{
		processors2.NewSkipEmptyMsgID(),
	}

	if !cfg.ExtractErrors {
		p = append(p, processors2.NewSkipErrors())
	}

	p = append(p,
		processors2.NewCommentCleaner(cfg),
		processors2.NewSkipIgnore(),
		processors2.NewUnprintableCheck(),
		processors2.NewPrepareKey(),
	)

	ret := &Runner{
		Processors: p,
		Log:        logrus.WithField("service", "Runner"),
	}

	return ret, nil
}

func (r Runner) Run(ctx context.Context, extractCtx *extract.Context, extractors []extract.Extractor) ([]extract.Issue, error) {
	r.Log.Debug("Start issue extracting")
	defer util.TrackTime(time.Now(), "Extracting the issues")

	issues := make([]extract.Issue, 0, 100)
	for _, extractor := range extractors {
		extractedIssues, err := extractor.Run(ctx, extractCtx)
		if err != nil {
			r.Log.Warnf("Can't run extractor %s: %v", extractor.Name(), err)
		} else {
			issues = append(issues, extractedIssues...)
		}
	}

	return r.processIssues(issues), nil
}

func (r *Runner) processIssues(issues []extract.Issue) []extract.Issue {
	defer util.TrackTime(time.Now(), "Process the issues")

	for _, p := range r.Processors {
		var newIssues []extract.Issue
		var err error

		newIssues, err = p.Process(issues)
		if err != nil {
			r.Log.Warnf("Can't process result by %s processor: %s", p.Name(), err)
		} else {
			issues = newIssues
		}

		if issues == nil {
			issues = []extract.Issue{}
		}
	}

	return issues
}
