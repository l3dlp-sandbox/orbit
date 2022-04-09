// Copyright (c) 2021 Guy A. Ross
// This source code is licensed under the GNU GPLv3 found in the
// license file in the root directory of this source tree.

package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/GuyARoss/orbit/internal"
	"github.com/GuyARoss/orbit/internal/srcpack"
	"github.com/GuyARoss/orbit/pkg/hotreload"
	"github.com/GuyARoss/orbit/pkg/jsparse"
	"github.com/GuyARoss/orbit/pkg/log"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var devCMD = &cobra.Command{
	Use:   "dev",
	Long:  "hot-reload bundle data given the specified pages in dev mode",
	Short: "hot-reload bundle data given the specified pages in dev mode",
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.NewDefaultLogger()

		s, err := internal.New(context.Background(), &internal.SessionOpts{
			WebDir:        viper.GetString("webdir"),
			Mode:          viper.GetString("mode"),
			Pacname:       viper.GetString("pacname"),
			OutDir:        viper.GetString("out"),
			NodeModDir:    viper.GetString("nodemod"),
			PublicDir:     viper.GetString("publicdir"),
			HotReloadPort: viper.GetInt("hotreloadport"),
		})

		if err != nil {
			panic(err)
		}

		watcher, _ := fsnotify.NewWatcher()
		defer watcher.Close()

		hotReload := hotreload.New()

		if err := filepath.Walk(viper.GetString("webdir"), WatchDir(watcher)); err != nil {
			panic("invalid walk on watchDir")
		}

		timeout := time.Duration(viper.GetInt("timeout")) * time.Millisecond

		fileChangeOpts := &internal.ChangeRequestOpts{
			SafeFileTimeout: time.Duration(viper.GetInt("samefiletimeout")) * time.Millisecond,
			Hook:            srcpack.NewSyncHook(log.NewDefaultLogger()),
			HotReload:       hotReload,
			Parser:          &jsparse.JSFileParser{},
		}

		go func() {
			for {
				time.Sleep(timeout)

				select {
				case e := <-watcher.Events:
					err := s.DoFileChangeRequest(e.Name, fileChangeOpts)

					if err == nil && len(viper.GetString("depout")) > 0 {
						s.SourceMap.Write(viper.GetString("depout"))
					}
				case err := <-watcher.Errors:
					panic(fmt.Sprintf("watcher failed %s", err.Error()))
				}
			}
		}()

		go func(hr *hotreload.HotReload) {
			for {
				event := <-hr.Redirected

				changes := event.BundleKeys.Diff(event.PreviousBundleKeys)

				wg := &sync.WaitGroup{}
				wg.Add(len(changes))
				for _, r := range changes {
					go func(change string, wg *sync.WaitGroup) {
						err := s.DoBundleKeyChangeRequest(change, fileChangeOpts)
						if err != nil {
							fmt.Println(err)
						}
						wg.Done()
					}(r, wg)
				}

				wg.Wait()
			}
		}(hotReload)

		http.HandleFunc("/ws", hotReload.HandleWebSocket)
		logger.Info(fmt.Sprintf("server started on port %d", viper.GetInt("hotreloadport")))

		err = http.ListenAndServe(fmt.Sprintf("localhost:%d", viper.GetInt("hotreloadport")), nil)
		if err != nil {
			panic(err)
		}
	},
}

// watchDir is a utility function used by the file path walker that applies
// each sub directory found under a path to the file watcher
func WatchDir(watcher *fsnotify.Watcher) func(path string, fi os.FileInfo, err error) error {
	return func(path string, fi os.FileInfo, err error) error {
		if fi.Mode().IsDir() {
			return watcher.Add(path)
		}

		return nil
	}
}

func init() {
	var timeoutDuration int
	var samefileTimeout int
	var port int

	devCMD.PersistentFlags().IntVar(&timeoutDuration, "timeout", 2000, "specifies the timeout duration in milliseconds until a change will be detected")
	viper.BindPFlag("timeout", devCMD.PersistentFlags().Lookup("timeout"))

	devCMD.PersistentFlags().IntVar(&samefileTimeout, "samefiletimeout", 2000, "specifies the timeout duration in milliseconds until a change will be detected for repeating files")
	viper.BindPFlag("samefiletimeout", devCMD.PersistentFlags().Lookup("samefiletimeout"))

	devCMD.PersistentFlags().IntVar(&port, "hotreloadport", 3005, "port used for hotreload")
	viper.BindPFlag("hotreloadport", devCMD.PersistentFlags().Lookup("hotreloadport"))
}
