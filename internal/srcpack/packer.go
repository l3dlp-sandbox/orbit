package srcpack

import (
	"context"
	"fmt"
	"sync"

	"github.com/GuyARoss/orbit/pkg/bundler"
	"github.com/GuyARoss/orbit/pkg/fs"
	"github.com/GuyARoss/orbit/pkg/jsparse"
	"github.com/GuyARoss/orbit/pkg/log"
	webwrapper "github.com/GuyARoss/orbit/pkg/web_wrapper"
)

// packer is the primary struct used for packing a directory of javascript files into
// valid web components.
type Packer struct {
	Bundler          bundler.Bundler
	JsParser         jsparse.JSParser
	ValidWebWrappers webwrapper.JSWebWrapperMap
	Logger           log.Logger

	AssetDir string
	WebDir   string
}

// copies the required assets to the asset directory
func (s *Packer) CopyAssets() ([]*fs.CopyResults, error) {
	results := fs.CopyDir(s.AssetDir, s.AssetDir, ".orbit/assets", false)

	return results, nil
}

// concpack is a private packing mechanism embedding the packer to pack a set of files concurrently.
type concPack struct {
	*Packer
	m sync.Mutex

	packedPages []*Component
	packMap     map[string]bool
}

// pack single packs a single file path into a usable web component
// this process includes the following:
// 1. wrapping the component with the specified front-end web framework.
// 2. bundling the component with the specified javascript bundler.
func (p *concPack) PackSingle(errchan chan error, wg *sync.WaitGroup, path string) {
	// @@todo: we should validate if these components exist on our source map yet, if so we should
	// inherit the metadata, rather than generate new metadata.
	page, err := NewComponent(context.TODO(), &NewComponentOpts{
		FilePath:      path,
		WebDir:        p.WebDir,
		JSWebWrappers: p.ValidWebWrappers,
		Bundler:       p.Bundler,
		JSParser:      p.JsParser,
	})

	if err != nil {
		errchan <- err
		fmt.Println(err)

		wg.Done()
		return
	}

	if p.packMap[page.Name] {
		// this page has already been packed before
		// and does not need to be repacked.
		wg.Done()
		return
	}

	p.m.Lock()
	p.packedPages = append(p.packedPages, page)
	p.packMap[page.Name] = true
	p.m.Unlock()

	wg.Done()
}

// packs the provoided file paths into the orbit root directory
func (s *Packer) PackMany(pages []string) ([]*Component, error) {
	cp := &concPack{
		Packer:      s,
		packedPages: make([]*Component, 0),
		packMap:     make(map[string]bool),
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(pages))

	errchan := make(chan error)

	go func() {
		err := <-errchan
		// @@todo: do something more with this error?
		fmt.Println("error occurred", err.Error())
	}()

	sh := NewSyncHook(s.Logger)

	defer sh.Close()

	for _, dir := range pages {
		// we copy dir here to avoid the pointer of dir being passed to our wrap func.
		t := dir
		// go routine to pack every page found in the pages directory
		// we wrap this routine with the sync hook to measure & log time deltas.
		go sh.WrapFunc(dir, func() { cp.PackSingle(errchan, wg, t) })
	}

	wg.Wait()

	return cp.packedPages, nil
}
