package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"git.gohegan.uk/kaigoh/gowww/v2/gowww"
	"git.gohegan.uk/kaigoh/gowww/v2/utilities"
	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/mux"
	"github.com/withmandala/go-log"
	"gopkg.in/yaml.v2"
)

var logger *log.Logger
var hosts map[string]string
var routes []string
var configs map[string]gowww.Config
var defaultConfig gowww.Config

func main() {

	// Banner
	fmt.Println("      -----------------------------------")
	fmt.Println("         GoWWW - (C) Kai Gohegan, 2022   ")
	fmt.Println("      -----------------------------------")
	fmt.Println("      https://git.gohegan.uk/kaigoh/gowww")
	fmt.Println("      -----------------------------------")

	// Setup logging
	logger = log.New(os.Stderr)

	// Load environment variables from disk...
	utilities.LoadEnv()

	// Check if we have a "vhosts" (or whatever has been configured as the root directory) directory...
	if !utilities.DirExists(utilities.GetEnv("GOWWW_ROOT", "vhosts")) {
		logger.Warn("gowww root directory does not exist, attempting to create it...")
		if err := os.Mkdir(utilities.GetEnv("GOWWW_ROOT", "vhosts"), os.ModeDir); err != nil {
			logger.Fatal(err)
		}
	}

	// Output the basic config to stdout...
	fmt.Println("Using \"" + utilities.GetEnv("GOWWW_ROOT", "vhosts") + "\" as gowww root")
	fmt.Println("Listening on port " + utilities.GetEnv("GOWWW_PORT", "8080"))

	// Setup our configs and hosts...
	hosts = make(map[string]string)
	configs = make(map[string]gowww.Config)
	defaultConfig = gowww.DefaultConfig

	// Setup router...
	r := mux.NewRouter()

	// Middleware to "fake" allowing the addition and removal of hosts on the fly...
	r.Use(CanRoute)

	// Watch the root directory for changes...
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Fatal(err)
	}
	defer watcher.Close()
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				base := strings.TrimPrefix(event.Name, utilities.GetEnv("GOWWW_ROOT", "vhosts")+string(os.PathSeparator))
				host := strings.TrimSuffix(base, string(os.PathSeparator)+".gowww.yml")
				switch {
				case event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create:
					// Did we get a file or a folder?
					if strings.Count(base, string(os.PathSeparator)) == 1 && strings.HasSuffix(event.Name, string(os.PathSeparator)+".gowww.yml") {
						AddHost(r, host, strings.TrimSuffix(event.Name, string(os.PathSeparator)+".gowww.yml"), watcher)
					} else {
						if strings.Count(base, string(os.PathSeparator)) == 0 {
							AddHost(r, host, event.Name, watcher)
						}
					}
				case event.Op&fsnotify.Remove == fsnotify.Remove:
					if strings.Count(base, string(os.PathSeparator)) == 0 {
						logger.Warn("Removing host \"" + host + "\"")
						hosts = gowww.RemoveHost(hosts, host)
					}
				case event.Op&fsnotify.Rename == fsnotify.Rename:
					if strings.Count(base, string(os.PathSeparator)) == 0 {
						logger.Warn("Renaming host \"" + host + "\"")
						hosts = gowww.RemoveHost(hosts, host)
					}
				}

				// watch for errors
			case err := <-watcher.Errors:
				logger.Error(err)
			}
		}
	}()
	if err := watcher.Add(utilities.GetEnv("GOWWW_ROOT", "vhosts")); err != nil {
		logger.Error(err)
	} else {
		logger.Info("Watching root folder for changes...")
	}

	// Add existing hosts...
	items, _ := ioutil.ReadDir(utilities.GetEnv("GOWWW_ROOT", "vhosts"))
	for _, item := range items {
		if item.IsDir() {
			host := gowww.CleanName(item.Name())
			AddHost(r, host, utilities.GetEnv("GOWWW_ROOT", "vhosts")+string(os.PathSeparator)+item.Name(), watcher)
		}
	}

	// Run the server...
	srv := &http.Server{
		Handler:      r,
		Addr:         ":" + utilities.GetEnv("GOWWW_PORT", "8080"),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logger.Fatal(srv.ListenAndServe())

}

// Add a host to the server...
func AddHost(router *mux.Router, host string, path string, watcher *fsnotify.Watcher) {

	logger.Info("Adding host \"" + host + "\"")

	// Do we have a config for this host? If not, fall back to a default config...
	config, set := configs[host]
	if !set {
		config := defaultConfig
		config.Host = host
		config.Path = path
		config.IsWatched = false
		config.Hosts = []string{host}
	} else {
		for _, h := range configs[host].Hosts {
			if h != host {
				logger.Warn("Removing virtual host \"" + h + "\"")
			}
			hosts = gowww.RemoveHost(hosts, h)
		}
	}

	// Do we have a custom config for the host?
	yml := path + string(os.PathSeparator) + ".gowww.yml"
	if utilities.FileExists(yml) {
		yfile, err := ioutil.ReadFile(yml)
		if err != nil {
			logger.Warn(err)
		}
		err = yaml.Unmarshal(yfile, &config)
		if err != nil {
			logger.Warn(err)
		}
		config.Host = host
		config.Path = path
		config.Hosts = append(config.Hosts, host)

		// Watch the config file (if not already being watched)...
		if !config.IsWatched {
			if err := watcher.Add(path); err != nil {
				logger.Error(err)
			} else {
				logger.Info("Watching " + host + " config for changes...")
				config.IsWatched = true
			}
		}

	}

	// Store the config...
	configs[host] = config

	// Update the host maps and router...
	for _, h := range config.Hosts {
		hosts = gowww.RemoveHost(hosts, h)
		if h != host {
			logger.Info("Adding virtual host \"" + h + "\"")
		}
		hosts[h] = host
		if !gowww.HaveRoute(routes, h) {
			router.Host(h).Handler(http.FileServer(http.Dir(path)))
		}
	}
}

// Middleware to effectively allow the on-the-fly removal and addition of routes...
func CanRoute(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		host := strings.Split(r.Host, ":")

		// Log the request
		logger.Infof("%s: (%s) %s %s", r.Host, gowww.GetClientIP(r), r.Method, r.URL)

		if gowww.HaveHost(hosts, host[0]) {
			// Get the config...
			config := configs[hosts[host[0]]]

			// Does the requested entity exist?
			if !utilities.EntityExists(utilities.GetEnv("GOWWW_ROOT", "vhosts") + string(os.PathSeparator) + config.Host + string(os.PathSeparator) + strings.ReplaceAll(r.URL.Path, "/", string(os.PathSeparator))) {
				http.Error(w, "Document not found", http.StatusNotFound)
				return
			}

			// Are we requesting the index?
			if r.URL.Path == "" || r.URL.Path == "/" || strings.HasSuffix(r.URL.Path, "/") {
				index, err := config.GetDefaultDocument(strings.ReplaceAll(r.URL.Path, "/", string(os.PathSeparator)))
				if err != nil {
					if !config.AllowDirectoryIndex {
						http.Error(w, "Directory index not allowed", http.StatusForbidden)
						logger.Warn("Directory index not allowed")
						return
					} else {
						logger.Warn(err.Error())
					}
				} else {
					r.URL.Path = "/" + index
				}
			}
			// Is the custom config file being requested? If so, deny the request...
			if strings.HasSuffix(r.URL.Path, "/.gowww.yml") {
				http.Error(w, "Document not found", http.StatusNotFound)
			} else {
				next.ServeHTTP(w, r)
			}
		} else {
			logger.Error("Host not found \"" + host[0] + "\"")
			http.Error(w, "Host not found", http.StatusNotFound)
			return
		}

	})
}
