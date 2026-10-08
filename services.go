package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
	"golang.design/x/clipboard"
)

func newmdpreviewer() *mdpreviewer {
	m := new(mdpreviewer)

	m.windows = make(map[string]*application.WebviewWindow)
	m.docs = make(map[string]*documentWindow)

	return m
}

func (m *mdpreviewer) browse(newwindow bool) error {
	app := application.Get()
	path, err := app.Dialog.OpenFile().SetTitle("Select File").AddFilter("Markdown Files", "*.md;*.markdown;*.txt").AddFilter("All Files", "*").PromptForSingleSelection()

	if err != nil {
		return err
	}

	if path == "" {
		return errors.New("No File Selected")
	}

	var d *documentWindow

	if !newwindow {
		d = m.docs[app.Window.Current().Name()]
	} else {
		d = new(documentWindow)
		addDocWindow(app, m, d, windowOpts)
	}

	if err := d.loadFile(path); err != nil {
		return err
	}
	return nil
}

func (m *mdpreviewer) OpenExternal(url string) error {

	l := strings.ToLower(url)
	if !strings.HasPrefix(l, "http://") && !strings.HasPrefix(l, "https://") && !strings.HasPrefix(l, "mailto:") {
		return errors.New("unsupported link scheme")
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()

}

// mostly useful at startup of app when events cannot
// be used
func (m *mdpreviewer) Html() string {
	d, ok := m.docs[application.Get().Window.Current().Name()]
	if !ok {
		return ""
	}
	return d.html
}

// Clipboard actions
// We handle the clipboard on the backend.
// Wails3 has limited clipboard capabilities so
// we use clipboard package.
var fmtHtml clipboard.Format

func init() {
	//Initialize Clipboard for use
	if err := clipboard.Init(); err != nil {
		log.Fatal(err)
	}
	fmtHtml = clipboard.Register("text/html")

}

func (m *mdpreviewer) copyRichText() error {
	ctx := context.Background()
	d := m.docs[application.Get().Window.Current().Name()]
	clipboard.WriteAll(ctx,
		clipboard.Item{Format: fmtHtml, Bytes: []byte(d.html)},
		clipboard.Item{Format: clipboard.FmtText, Bytes: []byte(d.plainText)},
	)
	return nil
}

func (m *mdpreviewer) copyHtml() error {
	ctx := context.Background()
	d := m.docs[application.Get().Window.Current().Name()]
	clipboard.Write(ctx, clipboard.FmtText, []byte(d.html))
	return nil
}

func (m *mdpreviewer) copyMd() error {
	ctx := context.Background()
	d := m.docs[application.Get().Window.Current().Name()]
	clipboard.Write(ctx, clipboard.FmtText, []byte(d.markdown))
	return nil
}

// the following struct is used for sending theme changes to frontend.
type cssTheme struct {
	Target string
	Color  string
	Font   string
	Size   string
}

func init() {
	application.RegisterEvent[cssTheme]("themechange")
}

func (m *mdpreviewer) setColorScheme(theme string) {
	app := application.Get()
	app.Event.Emit("themechange", cssTheme{Color: theme})
}

func (m *mdpreviewer) setFont(fontfamily string) {
	app := application.Get()
	target := app.Window.Current().Name()
	app.Event.Emit("themechange", cssTheme{Target: target, Font: fontfamily})
}

func (m *mdpreviewer) setZoom(direction string) {
	fmt.Println("zoom " + direction)
	app := application.Get()
	target := app.Window.Current().Name()
	app.Event.Emit("themechange", cssTheme{Target: target, Size: direction})
}
