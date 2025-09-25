package loader

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"path/filepath"
	"strings"
	"time"

	"github.com/mattn/go-zglob"
	"github.com/sirupsen/logrus"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/packages"

	"github.com/vorlif/xspreak/config"
	"github.com/vorlif/xspreak/extract"
	"github.com/vorlif/xspreak/tmpl"
	"github.com/vorlif/xspreak/util"
)

var loadMode = packages.NeedName |
	packages.NeedFiles |
	packages.NeedSyntax |
	packages.NeedTypes |
	packages.NeedTypesInfo |
	packages.NeedImports |
	packages.NeedDeps

type PackageLoader struct {
	config *config.Config
	log    *logrus.Entry
}

func NewPackageLoader(cfg *config.Config) *PackageLoader {
	return &PackageLoader{
		config: cfg,
		log:    logrus.WithField("service", "PackageLoader"),
	}
}

func (pl *PackageLoader) buildArgs() []string {
	args := pl.config.Args
	if len(args) == 0 {
		return []string{"./..."}
	}

	var retArgs []string
	for _, arg := range args {
		if strings.HasPrefix(arg, ".") || filepath.IsAbs(arg) {
			retArgs = append(retArgs, arg)
		} else if strings.ContainsRune(arg, filepath.Separator) {
			retArgs = append(retArgs, fmt.Sprintf(".%c%s", filepath.Separator, arg), arg)
		} else {
			retArgs = append(retArgs, arg)
		}
	}

	return retArgs
}

func (pl *PackageLoader) Load(ctx context.Context) (*extract.Context, error) {
	pkgConf := &packages.Config{
		Context: ctx,
		Mode:    loadMode,
		Dir:     pl.config.SourceDir,
		Logf:    logrus.WithField("service", "package-loader").Debugf,
		Tests:   false,
	}

	originalPkgs, err := pl.loadPackages(ctx, pkgConf)
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}

	// Add loaded packages from config to originalPkgs
	if len(pl.config.LoadedPackages) > 0 {
		loadedPkgs, err := loadPackagesFromDir(pkgConf, pl.config.LoadedPackages)
		if err != nil {
			return nil, fmt.Errorf("failed to load specified packages: %w", err)
		}
		originalPkgs = append(originalPkgs, loadedPkgs...)
	}

	if len(originalPkgs) == 0 {
		return nil, errors.New("no go files to analyze")
	}

	ret := &extract.Context{
		OriginalPackages: originalPkgs,
		Packages:         cleanPackages(originalPkgs),
		Config:           pl.config,
		Log:              pl.log,
		Definitions:      make(extract.Definitions, 200),
	}

	ret.Inspector = createInspector(ret.Packages)
	extractDefinitions(ret)
	ret.CommentMaps = extractComments(ret.Packages)

	templateFiles, errTmpl := pl.searchTemplate()
	if errTmpl != nil {
		return nil, errTmpl
	}
	ret.Templates = templateFiles

	return ret, nil
}

func (pl *PackageLoader) loadPackages(ctx context.Context, pkgCfg *packages.Config) ([]*packages.Package, error) {
	args := pl.buildArgs()
	pl.log.Debugf("Built loader args are %s", args)

	pkgs, err := loadPackagesFromDir(pkgCfg, args)
	if err != nil {
		return nil, fmt.Errorf("%w failed to load with go/packages", err)
	}

	if ctx.Err() != nil {
		return nil, fmt.Errorf("%w timed out to load packages", ctx.Err())
	}

	pl.debugPrintLoadedPackages(pkgs)
	return pkgs, nil
}

func (pl *PackageLoader) debugPrintLoadedPackages(pkgs []*packages.Package) {
	pl.log.Debugf("loaded %d pkgs", len(pkgs))
	for i, pkg := range pkgs {
		var syntaxFiles []string
		for _, sf := range pkg.Syntax {
			syntaxFiles = append(syntaxFiles, pkg.Fset.Position(sf.Pos()).Filename)
		}
		pl.log.Debugf("Loaded pkg #%d: ID=%s GoFiles=%d Syntax=%d",
			i, pkg.ID, len(pkg.GoFiles), len(syntaxFiles))
	}
}

func (pl *PackageLoader) searchTemplate() ([]*tmpl.Template, error) {
	defer util.TrackTime(time.Now(), "Template file search")
	patterns := pl.config.TemplatePatterns
	if len(patterns) == 0 {
		return []*tmpl.Template{}, nil
	}

	files := make([]*tmpl.Template, 0, len(patterns)*5)

	for _, pattern := range patterns {
		foundFiles, err := zglob.Glob(pattern)
		if err != nil {
			return nil, err
		}
		for _, file := range foundFiles {
			pl.log.Debugf("found template file %s", file)
			pathAbs, errAbs := filepath.Abs(file)
			if errAbs != nil {
				logrus.WithError(errAbs).Warn("Template could not be parsed")
				continue
			}

			parsed, errP := tmpl.ParseFile(pathAbs)
			if errP != nil {
				logrus.WithError(errP).Warn("Template could not be parsed")
				continue
			}

			files = append(files, parsed)
		}
	}

	pl.log.Debugf("found %d template files", len(files))
	return files, nil
}

func createInspector(pkgs []*packages.Package) *inspector.Inspector {
	files := make([]*ast.File, 0, 200)
	for _, pkg := range pkgs {
		files = append(files, pkg.Syntax...)
	}

	return inspector.New(files)
}

func loadPackagesFromDir(pkgCfg *packages.Config, args []string) ([]*packages.Package, error) {
	defer util.TrackTime(time.Now(), "Loading source packages")
	pkgs, err := packages.Load(pkgCfg, args...)
	if err != nil {
		return nil, err
	}

	if packages.PrintErrors(pkgs) > 0 {
		logrus.Warn("There are files with errors, the extraction may fail")
	}

	return pkgs, nil
}
