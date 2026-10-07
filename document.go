package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"
	"github.com/wailsapp/wails/v3/pkg/application"
)

func initializeDoc(d *documentWindow) {
	//load the document specified on command line (if any). We
	//discard all but the first argument.
	if len(os.Args) > 1 {
		err := d.loadFile(expandHome(os.Args[1]))
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	}
}

func (d *documentWindow) loadFile(path string) (err error) {
	d.path, err = filepath.Abs(path)
	if d.watcher != nil {
		d.watcher.Close()
		d.watcher = nil
	}
	err = d.reload()
	if err != nil {
		return err
	}
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	d.watcher = w

	dir := filepath.Dir(d.path)
	if dir != d.watchedDir {
		if d.watchedDir != "" {
			_ = d.watcher.Remove(d.watchedDir)
			d.watchedDir = ""
		}
		if err := d.watcher.Add(dir); err != nil {
			return err
		}
		d.watchedDir = dir
	}
	err = d.startWatching()
	return err
}

// We use this to send update to frontend.
// Need to specify Target because Wails3
// does not (yet?) support sending an event to just
// one Window.
type update struct {
	Target string
	Html   string
}

func init() {
	application.RegisterEvent[update]("updatefile")
}

func (d *documentWindow) reload() error {
	const (
		mdExtensions = parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
		htmlFlags    = html.CommonFlags
	)
	data, err := os.ReadFile(d.path)
	if err != nil {
		return err
	}
	d.markdown = string(data)
	p := parser.NewWithExtensions(mdExtensions)
	r := html.NewRenderer(html.RendererOptions{Flags: htmlFlags})
	d.html = string(bluemonday.UGCPolicy().SanitizeBytes(markdown.ToHTML(markdown.NormalizeNewlines(data), p, r)))
	//notify frontend of update
	d.window.EmitEvent("updatefile", update{Html: d.html, Target: d.window.Name()})
	return nil
}

// Watch the file of d or changes. The reloads are debounced.
func (d *documentWindow) watchLoop() error {
	const debounceDelay = 100 * time.Millisecond
	watcher := d.watcher
	if watcher == nil {
		return nil
	}

	var timer *time.Timer
	var fire <-chan time.Time
	const interesting = fsnotify.Write | fsnotify.Create | fsnotify.Rename | fsnotify.Remove

	for {
		select {
		case ev, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			d.mu.Lock()
			match := d.path != "" && filepath.Clean(ev.Name) == d.path
			d.mu.Unlock()
			if match && ev.Op&interesting != 0 {
				if timer != nil {
					timer.Stop()
				}
				timer = time.NewTimer(debounceDelay)
				fire = timer.C
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return err
			}
			return err
		case <-fire:
			fire = nil
			d.reload()
		}
	}
	return nil
}

// startWatching launches the event loop the first time a file is opened.
// Call it only after a file has been loaded successfully and a.w is set.
func (d *documentWindow) startWatching() error {
	//d.loopOnce.Do(func() { go d.watchLoop() })
	go d.watchLoop()
	return nil
}

func expandHome(p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, p[1:])
		}
	}
	return p
}
