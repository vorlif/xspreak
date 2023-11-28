package encoder

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/vorlif/spreak/catalog/po"

	"github.com/vorlif/xspreak/config"
	"github.com/vorlif/xspreak/result"
	"github.com/vorlif/xspreak/util"
)

type potEncoder struct {
	cfg *config.Config
	w   *po.Encoder
}

func NewPotEncoder(cfg *config.Config, w io.Writer) Encoder {
	enc := po.NewEncoder(w)
	enc.SetWrapWidth(cfg.WrapWidth)
	enc.SetWriteHeader(!cfg.OmitHeader)
	enc.SetWriteReferences(!cfg.WriteNoLocation)

	return &potEncoder{cfg: cfg, w: enc}
}

func (e *potEncoder) Encode(issues []result.Issue) error {
	file := &po.File{
		Header:   e.buildHeader(),
		Messages: make(map[string]map[string]*po.Message),
	}

	for _, msg := range e.buildMessages(issues) {
		file.AddMessage(msg)
	}

	return e.w.Encode(file)
}

func (e *potEncoder) buildMessages(issues []result.Issue) []*po.Message {
	util.TrackTime(time.Now(), "Build messages")
	messages := make([]*po.Message, 0, len(issues))

	absOut, errA := filepath.Abs(e.cfg.OutputDir)
	if errA != nil {
		absOut = e.cfg.OutputDir
	}

	for _, iss := range issues {
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

		msg := &po.Message{
			Comment: &po.Comment{
				Extracted:  strings.Join(iss.Comments, "\n"),
				References: []*po.Reference{ref},
				Flags:      iss.Flags,
			},
			Context:  iss.Context,
			ID:       iss.MsgID,
			IDPlural: iss.PluralID,
		}

		messages = append(messages, msg)
	}

	return messages
}

func (e *potEncoder) buildHeader() *po.Header {
	headerComment := fmt.Sprintf(`SOME DESCRIPTIVE TITLE.
Copyright (C) YEAR %s
This file is distributed under the same license as the %s package.
FIRST AUTHOR <EMAIL@ADDRESS>, YEAR.
`, e.cfg.CopyrightHolder, e.cfg.PackageName)
	return &po.Header{
		Comment: &po.Comment{
			Translator:     headerComment,
			Extracted:      "",
			References:     nil,
			Flags:          []string{"fuzzy"},
			PrevMsgContext: "",
			PrevMsgID:      "",
		},
		ProjectIDVersion:        e.cfg.PackageName,
		ReportMsgidBugsTo:       e.cfg.BugsAddress,
		POTCreationDate:         time.Now().Format("2006-01-02 15:04-0700"),
		PORevisionDate:          "YEAR-MO-DA HO:MI+ZONE",
		LastTranslator:          "FULL NAME <EMAIL@ADDRESS>",
		LanguageTeam:            "LANGUAGE <LL@li.org>",
		Language:                "",
		MimeVersion:             "1.0",
		ContentType:             "text/plain; charset=UTF-8",
		ContentTransferEncoding: "8bit",
		PluralForms:             "", // alternative  "nplurals=INTEGER; plural=EXPRESSION;"
	}
}
