package watcher

import (
	"errors"
	"os"
	"reflect"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/shawn1m/overture/core/config"
	log "github.com/sirupsen/logrus"
)

type Watcher struct {
	Config   *config.Config
	Watch    *fsnotify.Watcher
	Callback func()
}

// NewWatcher returns a watcher struct, c is config, f is reload callback function
// callback function should be blocked
func NewWatcher(c *config.Config, f func()) (*Watcher, error) {
	var err error
	w := &Watcher{
		Config:   c,
		Callback: f,
	}
	w.Watch, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	err = w.AddWatchList()
	if err != nil {
		return nil, err
	}
	return w, nil
}

// AddWatchList initialize watch list
func (w *Watcher) AddWatchList() error {
	if w.Config == nil {
		return errors.New("Not set variable Config")
	}
	err := w.Watch.Add(w.Config.FilePath)
	if err != nil {
		return err
	}
	err = w.Watch.Add(w.Config.DomainTTLFile)
	if err != nil {
		return err
	}
	err = w.Watch.Add(w.Config.DomainFile.Primary)
	if err != nil {
		return err
	}
	err = w.Watch.Add(w.Config.DomainFile.Alternative)
	if err != nil {
		return err
	}
	err = w.Watch.Add(w.Config.IPNetworkFile.Primary)
	if err != nil {
		return err
	}
	err = w.Watch.Add(w.Config.IPNetworkFile.Alternative)
	if err != nil {
		return err
	}
	err = w.Watch.Add(w.Config.HostsFile.HostsFile)
	if err != nil {
		return err
	}
	return nil
}

// WalkConfig walks two config and rewatch changed filepath.
func (w *Watcher) WalkConfig(c1Value reflect.Value, c2Value reflect.Value, prefix string) error {
	c1Type := c1Value.Type()
	for i := 0; i < c1Value.NumField(); i++ {
		var err error
		switch c1Value.Field(i).Kind() {
		case reflect.String:
			log.Debugf("Comparing config: %s = pre: %s, after: %s", c1Type.Field(i).Name, c1Value.Field(i).Interface(), c2Value.Field(i).Interface())
			if c1Value.Field(i).Interface() != c2Value.Field(i).Interface() && (strings.Contains(c1Type.Field(i).Name, "File") || strings.Contains(prefix, "File")) {
				err = w.Watch.Remove(c1Value.Field(i).Interface().(string))
				if err != nil {
					return err
				}
				err = w.Watch.Add(c2Value.Field(i).Interface().(string))
				if err != nil {
					return err
				}
			}
		case reflect.Struct:
			err = w.WalkConfig(c1Value.Field(i), c2Value.Field(i), c1Type.Field(i).Name)
			if err != nil {
				return err
			}
		default:
			continue
		}
	}
	return nil
}

// ReloadConfig will compare two config file, rewatch changed sub config files
func (w *Watcher) ReloadConfig(c *config.Config) error {
	err := w.WalkConfig(reflect.ValueOf(w.Config).Elem(), reflect.ValueOf(c).Elem(), "")
	if err != nil {
		return err
	}
	return nil
}

// StartWatch begins to track config file changes.
func (w *Watcher) StartWatch() error {
	reloadmux := false
	if w.Config == nil {
		return errors.New("Not set variable Config")
	}
	go func() {
		for {
			select {
			case event, ok := <-w.Watch.Events:
				if event.Op&fsnotify.Write == fsnotify.Write && ok && !reloadmux {
					log.Warnf("%s file changed, reloading.", event.Name)
					reloadmux = true
					w.Callback()
					reloadmux = false
				}
			case err, ok := <-w.Watch.Errors:
				if !ok && err != nil {
					log.Fatalf("File watch error: %s", err)
					os.Exit(1)
				}
			}
		}
	}()
	return nil
}
