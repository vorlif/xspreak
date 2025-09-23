package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vorlif/xspreak/tmpl"
)

const (
	ExtractFormatPot  = "pot"
	ExtractFormatJSON = "json"
)

type Config struct {
	IsVerbose       bool
	CurrentDir      string
	SourceDir       string
	OutputDir       string
	OutputFile      string
	CommentPrefixes []string
	ExtractErrors   bool
	ErrorContext    string

	TemplatePatterns []string
	Keywords         []*tmpl.Keyword

	DefaultDomain   string
	WriteNoLocation bool
	WrapWidth       int
	DontWrap        bool

	OmitHeader      bool
	CopyrightHolder string
	PackageName     string
	BugsAddress     string

	MaxDepth int
	Args     []string

	LoadedPackages []string

	Timeout time.Duration

	ExtractFormat     string
	TmplIsMonolingual bool
}

func NewDefault() *Config {
	return &Config{
		IsVerbose:       false,
		SourceDir:       "",
		OutputDir:       filepath.Clean("./"),
		OutputFile:      "",
		CommentPrefixes: []string{"TRANSLATORS"},
		ExtractErrors:   false,
		ErrorContext:    "errors",

		DefaultDomain:   "messages",
		WriteNoLocation: false,
		WrapWidth:       80,
		DontWrap:        false,

		OmitHeader:      false,
		CopyrightHolder: "THE PACKAGE'S COPYRIGHT HOLDER",
		PackageName:     "PACKAGE VERSION",
		BugsAddress:     "",

		MaxDepth: 20,
		Timeout:  15 * time.Minute,

		ExtractFormat: ExtractFormatPot,
	}
}

func (c *Config) Prepare() error {
	c.ErrorContext = strings.TrimSpace(c.ErrorContext)
	c.DefaultDomain = strings.TrimSpace(c.DefaultDomain)
	if c.DefaultDomain == "" {
		return errors.New("a default domain is required")
	}

	if c.Timeout < 1*time.Minute {
		return errors.New("the value for Timeout must be at least one minute")
	}

	currentDir, errC := os.Getwd()
	if errC != nil {
		return errC
	}
	c.CurrentDir = currentDir

	if c.SourceDir == "" {
		c.SourceDir = c.CurrentDir
	}

	var err error
	c.SourceDir, err = filepath.Abs(c.SourceDir)
	if err != nil {
		return err
	}

	if c.OutputFile != "" {
		c.OutputDir = filepath.Dir(c.OutputFile)
		c.OutputFile = filepath.Base(c.OutputFile)
	} else {
		c.OutputFile = c.DefaultDomain + "." + c.ExtractFormat
	}

	c.OutputDir, err = filepath.Abs(c.OutputDir)
	if err != nil {
		return err
	}

	if c.DontWrap {
		c.WrapWidth = -1
	}

	if c.ExtractFormat == "po" {
		c.ExtractFormat = ExtractFormatPot
	}
	if c.ExtractFormat != ExtractFormatPot && c.ExtractFormat != ExtractFormatJSON {
		return fmt.Errorf("only the JSON and pot format is supported, you want %v", c.ExtractFormat)
	}

	if len(c.TemplatePatterns) > 0 && len(c.Keywords) == 0 {
		c.Keywords = tmpl.DefaultKeywords("T", c.TmplIsMonolingual)
	}

	return nil
}
