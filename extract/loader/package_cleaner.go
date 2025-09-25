package loader

import (
	"strings"
	"time"

	"golang.org/x/tools/go/packages"

	"github.com/vorlif/xspreak/config"
	"github.com/vorlif/xspreak/util"
)

type packageCleaner struct {
	// All loaded packages
	allPackages []*packages.Package

	// Packages that were visited
	visited map[string]bool

	cleanedPackages []*packages.Package
}

func cleanPackages(originalPackages []*packages.Package) []*packages.Package {
	pc := &packageCleaner{
		allPackages:     originalPackages,
		visited:         make(map[string]bool),
		cleanedPackages: make([]*packages.Package, 0, len(originalPackages)),
	}

	pc.performCleanup()

	return pc.cleanedPackages
}

func (pc *packageCleaner) performCleanup() {
	defer util.TrackTime(time.Now(), "Collect packages")

	for _, startPck := range pc.allPackages {
		pc.collectPackages(startPck)
	}
}

func (pc *packageCleaner) collectPackages(startPck *packages.Package) {
	queue := []*packages.Package{startPck}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if pc.visited[current.ID] {
			continue
		}

		pc.visited[current.ID] = true
		pc.cleanedPackages = append(pc.cleanedPackages, current)

		for _, importedPackage := range current.Imports {
			if pc.visited[importedPackage.ID] {
				continue
			}

			if !pc.isPartOfDirectory(importedPackage) {
				continue
			}

			queue = append(queue, importedPackage)
		}
	}
}

func (pc *packageCleaner) isPartOfDirectory(pkg *packages.Package) bool {
	if config.IsValidSpreakPackage(pkg.PkgPath) {
		return true
	}

	for _, src := range pc.allPackages {
		if strings.HasPrefix(pkg.PkgPath, src.PkgPath) {
			return true
		}
	}

	return false
}
