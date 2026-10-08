package main

import (
	"embed"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed all:frontend/dist
var assets embed.FS

// this is the struct type that hosts the methods for
// app services.
type mdpreviewer struct {
	windows map[string]*application.WebviewWindow //keep track of open windows
	docs    map[string]*documentWindow            //associate documents with windows
}

// documentWindow holds information for a document window.
type documentWindow struct {
	html       string
	markdown   string
	plainText  string
	path       string
	watcher    *fsnotify.Watcher
	watchedDir string
	mu         sync.Mutex
	window     *application.WebviewWindow
}

var d *documentWindow

var windowOpts application.WebviewWindowOptions

func init() {

	windowOpts = application.WebviewWindowOptions{
		Title:  "mdpreview",
		Width:  1000,
		Height: 1200,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		URL:                "/",
		UseApplicationMenu: true,
	}
}

func main() {

	m := newmdpreviewer()

	app := application.New(application.Options{
		Name:        "mdpreview",
		Description: "A simple Markdown previewer",
		Services: []application.Service{
			application.NewService(m),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	d := new(documentWindow)

	addDocWindow(app, m, d, windowOpts)

	setupMenus(app, m)

	setupEventReceivers(app, m)

	setupKeyboard(app, m)

	initializeDoc(d)

	err := app.Run()

	if err != nil {
		log.Fatal(err)
	}
}

// svc is the instance of mdpreviewer on which app Service is built.
func addDocWindow(app *application.App, svc *mdpreviewer, d *documentWindow, opts application.WebviewWindowOptions) {
	if len(svc.docs) > 0 {
		cw := app.Window.Current()
		curX, curY := cw.Position()
		fmt.Println("old pos", curX, curY)
		opts.InitialPosition = application.WindowXY
		opts.X = curX + 50
		opts.Y = curY + 50
		fmt.Println("new pos", opts.X, opts.Y)
	}
	w := app.Window.NewWithOptions(opts)
	d.window = w
	name := d.window.Name()
	svc.windows[name] = d.window
	svc.docs[name] = d
	w.OnWindowEvent(events.Common.WindowClosing, func(event *application.WindowEvent) {
		delete(svc.windows, name)
		delete(svc.docs, name)
		d.window = nil
		d = nil
	})
}

//svc is the instance *mdpreviewer on which the app service is built

func setupMenus(app *application.App, svc *mdpreviewer) {
	menu := app.NewMenu()

	// Add standard menus (platform-appropriate)
	if runtime.GOOS == "darwin" {
		menu.AddRole(application.AppMenu) // macOS only
	}

	//The top level menu items. Notice the order of declaration
	//is also the order in the menu bar
	fileMenu := menu.AddSubmenu("File")
	editMenu := menu.AddSubmenu("Edit")
	viewMenu := menu.AddSubmenu("View")

	//fileMenu options
	fileMenu.AddRole(application.Open)
	fileMenu.FindByRole(application.Open).OnClick(func(ctx *application.Context) {
		svc.browse(false)
	})
	fileMenu.Add("Open in New Window").OnClick(func(ctx *application.Context) {
		svc.browse(true)
	})
	fileMenu.AddSeparator()
	fileMenu.Add("Quit").OnClick(func(ctx *application.Context) {
		app.Quit()
	})

	//editMenu options
	editMenu.Add("Copy Rich Text").OnClick(func(ctx *application.Context) {
		svc.copyRichText()
	})
	editMenu.Add("Copy raw HTML").OnClick(func(ctx *application.Context) {
		svc.copyHtml()
	})
	editMenu.Add("Copy Markdown").OnClick(func(ctx *application.Context) {
		svc.copyMd()
	})

	//viewMenu options
	themeMenu := viewMenu.AddSubmenu("Theme")
	themeMenu.Add("Light").OnClick(func(ctx *application.Context) {
		svc.setColorScheme("light")
	})
	themeMenu.Add("Dark").OnClick(func(ctx *application.Context) {
		svc.setColorScheme("dark")
	})
	themeMenu.Add("System Setting").OnClick(func(ctx *application.Context) {
		svc.setColorScheme("system")
	})
	viewMenu.AddSeparator()
	fontMenu := viewMenu.AddSubmenu("Font")
	fontMenu.Add("Serif").OnClick(func(ctx *application.Context) {
		svc.setFont("serif")
	})
	fontMenu.Add("Sans Serif").OnClick(func(ctx *application.Context) {
		svc.setFont("sans-serif")
	})

	viewMenu.AddSeparator()
	viewMenu.AddRole(application.ZoomIn)
	viewMenu.AddRole(application.ZoomOut)
	viewMenu.AddRole(application.ResetZoom)

	app.Menu.Set(menu)
}

type metadata struct {
	Target    string
	Path      string
	WordCount int
	ShortPath string
}

func init() {
	application.RegisterEvent[string]("newplaintext")
	application.RegisterEvent[metadata]("newmetadata")
}

// svc is the instance of mdpreviewer on which app Service is built
func setupEventReceivers(app *application.App, svc *mdpreviewer) {
	app.Event.On("newplaintext", func(e *application.CustomEvent) {
		d := svc.docs[e.Sender]
		d.plainText = e.Data.(string)
		w, _ := app.Window.GetByName(e.Sender)
		w.EmitEvent("newmetadata", metadata{Target: e.Sender, Path: d.path, ShortPath: filepath.Base(d.path), WordCount: len(strings.Fields(d.plainText))})
	})
}

func setupKeyboard(app *application.App, svc *mdpreviewer) {

	kb := app.KeyBinding
	accl := "Ctrl+"
	if runtime.GOOS == "darwin" {
		accl = "Cmd+"
	}
	kb.Add(accl+"Q", func(w application.Window) { app.Quit() })

}
