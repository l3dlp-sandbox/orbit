// Copyright (c) 2021 Guy A. Ross
// This source code is licensed under the GNU GPLv3 found in the
// license file in the root directory of this source tree.

package internal

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/GuyARoss/orbit/internal/srcpack"
	"github.com/GuyARoss/orbit/pkg/fsutils"
	"github.com/GuyARoss/orbit/pkg/jsparse"
	"github.com/GuyARoss/orbit/pkg/webwrap"
	ewrap "github.com/GuyARoss/orbit/pkg/webwrap/embed"
)

type StaticBuild struct {
	buildOpts         *BuildOpts
	staticBuildOut    string
	SkipResourceCheck bool
}

type ComponentStaticContext struct {
	StaticMap   map[ewrap.PageRender]bool
	Pages       map[ewrap.PageRender]*ewrap.DocumentRenderer
	BundlePaths map[ewrap.PageRender]string
}

func (opts *StaticBuild) createStaticContext(components srcpack.PackedComponentList) *ComponentStaticContext {
	staticMap := make(map[ewrap.PageRender]bool)
	pages := make(map[ewrap.PageRender]*ewrap.DocumentRenderer)
	bundlePaths := make(map[ewrap.PageRender]string)

	if opts.SkipResourceCheck {
		ssrWrapMethod := webwrap.NewReactSSRPartial(&webwrap.NewReactSSROpts{
			Bundler: &webwrap.BaseBundler{
				Mode:           webwrap.DevelopmentBundle,
				WebDir:         opts.buildOpts.ApplicationDir,
				PageOutputDir:  ".orbit/base/pages",
				NodeModulesDir: opts.buildOpts.NodeModulePath,
				Logger:         nil,
			},
			SourceMapDoc: jsparse.NewEmptyDocument(),
			InitDoc:      jsparse.NewEmptyDocument(),
		})

		for _, c := range components {
			page, err := ssrWrapMethod.Apply(c.JsDocument())
			if err != nil {
				continue
			}

			resource, _ := ssrWrapMethod.Setup(context.TODO(), &webwrap.BundleOpts{
				FileName:  c.OriginalFilePath(),
				BundleKey: c.BundleKey() + "_ssr",
				Name:      page.Name(),
			})

			for _, d := range resource.BundleOpFileDescriptor {
				_, err = os.Stat(d)
				if errors.Is(err, os.ErrNotExist) {
					bundlePageErr := page.WriteFile(d)
					if bundlePageErr != nil {
						continue
					}
				}
			}

			// TODO: this may be redundant to do each iteration.
			for _, r := range resource.Configurators {
				configErr := r.Page.WriteFile(r.FilePath)
				if configErr != nil {
					fmt.Println(configErr)
					break
				}
			}

			pages[ewrap.PageRender(c.BundleKey())] = ewrap.NewEmptyDocumentRenderer(c.WebWrapper().Stats().Bundler)
			bundlePaths[ewrap.PageRender(c.BundleKey())] = fsutils.LastPathIndex(c.OriginalFilePath()) + ".html"

			switch c.WebWrapper().Stats().WebVersion {
			case "react":
				staticMap[ewrap.PageRender(c.BundleKey())] = true
			}
		}
	}

	for _, c := range components {
		pages[ewrap.PageRender(c.BundleKey())] = ewrap.NewEmptyDocumentRenderer(c.WebWrapper().Stats().WebVersion)
		bundlePaths[ewrap.PageRender(c.BundleKey())] = fsutils.LastPathIndex(c.OriginalFilePath()) + ".html"

		if c.IsStaticResource() {
			staticMap[ewrap.PageRender(c.BundleKey())] = true
		}
	}

	return &ComponentStaticContext{
		StaticMap:   staticMap,
		Pages:       pages,
		BundlePaths: bundlePaths,
	}
}

// StaticBuild builds the given components into its static file (html) counterpart
// this method does not account for the javascript bundles that may be present in process
func (opts *StaticBuild) Build(components srcpack.PackedComponentList) error {
	staticCtx := opts.createStaticContext(components)

	if len(staticCtx.StaticMap) == 0 {
		return nil
	}

	doc := ewrap.DocFromFile(opts.buildOpts.PublicPath)

	defer ewrap.Close()
	ewrap.StartupTaskReactSSR(opts.staticBuildOut, staticCtx.Pages, staticCtx.StaticMap, staticCtx.BundlePaths, *doc)()

	return nil
}

func NewStaticBuild(buildOpts *BuildOpts, staticBuildOut string) *StaticBuild {
	return &StaticBuild{
		buildOpts:         buildOpts,
		staticBuildOut:    staticBuildOut,
		SkipResourceCheck: false,
	}
}
