package processors

import (
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"unicode"

	log "github.com/sirupsen/logrus"

	"github.com/vorlif/xspreak/extract"
)

var workingDir, _ = os.Getwd()

type unprintableChecker struct{}

func NewUnprintableCheck() Processor {
	return &unprintableChecker{}
}

func (u unprintableChecker) Process(issues []extract.Issue) ([]extract.Issue, error) {

	for _, iss := range issues {
		checkForUnprintableChars(iss.MsgID, iss.Pos)
		checkForUnprintableChars(iss.PluralID, iss.Pos)
	}

	return issues, nil
}

func (u unprintableChecker) Name() string { return "unprintable check" }

func checkForUnprintableChars(s string, pos token.Position) {
	for _, r := range s {
		if !unicode.IsPrint(r) && !unicode.IsSpace(r) {
			filename := pos.Filename
			if relPath, err := filepath.Rel(workingDir, filename); err == nil {
				filename = relPath
			}

			log.Warnf("%s:%d internationalized messages should not contain the %s character", filename, pos.Line, strconv.QuoteRune(r))
		}
	}
}
