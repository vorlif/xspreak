package processors

import (
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/vorlif/spreak/pkg/po"

	"github.com/vorlif/xspreak/config"
	"github.com/vorlif/xspreak/result"
	"github.com/vorlif/xspreak/util"
)

// Recognizes not all cases, but most. - See Unit tests.
var reGoStringFormat = regexp.MustCompile(`%([#+\-*0.])?(\[\d])?(([1-9])\.([1-9])|([1-9])|([1-9])\.|\.([1-9]))?[xsvTtbcdoOqXUeEfFgGp]`)

type translationBuilder struct {
	cfg *config.Config
}

func BuildTranslations(cfg *config.Config) Processor {
	return &translationBuilder{
		cfg: cfg,
	}
}

func (s translationBuilder) Process(inIssues []result.Issue) ([]result.Issue, error) {
	util.TrackTime(time.Now(), "Build messages")
	outIssues := make([]result.Issue, 0, len(inIssues))

	absOut, errA := filepath.Abs(s.cfg.OutputDir)
	if errA != nil {
		absOut = s.cfg.OutputDir
	}

	for _, iss := range inIssues {
		path, errP := filepath.Rel(absOut, iss.Pos.Filename)
		if errP != nil {
			logrus.WithError(errP).Warn("Relative path could not be created, use absolute")
			path = iss.Pos.Filename
		}

		ref := &po.Reference{
			Path:   filepath.ToSlash(path),
			Line:   iss.Pos.Line,
			Column: iss.Pos.Column,
		}

		if reGoStringFormat.MatchString(iss.MsgID) || reGoStringFormat.MatchString(iss.PluralID) {
			iss.Flags = append(iss.Flags, "go-format")
		}

		iss.Message = &po.Message{
			Comment: &po.Comment{
				Extracted:  strings.Join(iss.Comment, "\n"),
				References: []*po.Reference{ref},
				Flags:      iss.Flags,
			},
			Context:  iss.Context,
			ID:       iss.MsgID,
			IDPlural: iss.PluralID,
		}

		outIssues = append(outIssues, iss)
	}

	return outIssues, nil
}

func (s translationBuilder) Name() string {
	return "build_translation"
}
