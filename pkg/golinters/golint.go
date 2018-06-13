package golinters

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"

	"github.com/golangci/golangci-lint/pkg/lint/linter"
	"github.com/golangci/golangci-lint/pkg/result"
	lintAPI "github.com/golangci/lint-1"
	"github.com/sirupsen/logrus"
)

type Golint struct{}

func (Golint) Name() string {
	return "golint"
}

func (Golint) Desc() string {
	return "Golint differs from gofmt. Gofmt reformats Go source code, whereas golint prints out style mistakes"
}

func (g Golint) Run(ctx context.Context, lintCtx *linter.Context) ([]result.Issue, error) {
	var issues []result.Issue
	var lintErr error
	for _, pkg := range lintCtx.PkgProgram.Packages() {
		files, fset, err := getASTFilesForPkg(lintCtx, &pkg)
		if err != nil {
			return nil, err
		}

		i, err := g.lintPkg(lintCtx.Settings().Golint.MinConfidence, files, fset)
		if err != nil {
			lintErr = err
			continue
		}
		issues = append(issues, i...)
	}
	if lintErr != nil {
		logrus.Warnf("golint: %s", lintErr)
	}

	return issues, nil
}

func (g Golint) lintPkg(minConfidence float64, files []*ast.File, fset *token.FileSet) ([]result.Issue, error) {
	l := new(lintAPI.Linter)
	ps, err := l.LintASTFiles(files, fset)
	if err != nil {
		return nil, fmt.Errorf("can't lint %d files: %s", len(files), err)
	}

	if len(ps) == 0 {
		return nil, nil
	}

	issues := make([]result.Issue, 0, len(ps)) // This is worst case
	for _, p := range ps {
		if p.Confidence >= minConfidence {
			issues = append(issues, result.Issue{
				Pos:        p.Position,
				Text:       p.Text,
				FromLinter: g.Name(),
			})
			// TODO: use p.Link and p.Category
		}
	}

	return issues, nil
}
