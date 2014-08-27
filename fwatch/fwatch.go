// File change watcher

package fwatch

import (
	"code.google.com/p/go.exp/fsnotify"
	"fmt"
	"time"
)

//------------------------------------------------------------
// Constants
//------------------------------------------------------------

const (
    // Time to elapse without changes take place
	fileDamper time.Duration = 3 * time.Second
	callbackDamper time.Duration = 1 * time.Second
)

//------------------------------------------------------------
// Watch
//------------------------------------------------------------

// Watch definition
type Watch struct {
    id       string
    id2      string
	dir      string
	watcher  *fsnotify.Watcher
	callback func(string, string)
	tdelta   time.Duration
	timer    *time.Timer
}

// Active watches map
var _watches = map[int]*Watch{}
// Next watch map id
var _watchNextId = 0

//------------------------------------------------------------
// Callbacker
//------------------------------------------------------------

// Callback damper, prevents from multiple callbacks on same repository
type Callbacker struct {
    watches  []*Watch
	tdelta   time.Duration
	timer    *time.Timer
}

// Callbacker
var _callbacker = Callbacker{tdelta: callbackDamper}

//------------------------------------------------------------
// Callbacker methods
//------------------------------------------------------------

func (c *Callbacker) execute() {
    executed := []string{}
    for _, w := range c.watches {
        isExecuted := false
        for _, id := range executed {
            if w.id == id {
                isExecuted = true
                break
            }
        }
        if isExecuted {
            continue
        }
        executed = append(executed, w.id)
        w.callback(w.id, w.id2)
    }
    c.watches = []*Watch{}
}

//------------------------------------------------------------
// Exported functions
//------------------------------------------------------------

// Adds directory watcher. 
// rid is repository id that will be passed to callback.
func WatchDir(dir string, rid, rid2 string, callback func(string, string)) (id int, err error) {
	return addWatch(dir, rid, rid2, callback)
}

// Adds specific file watcher.
// rid is repository id that will be passed to callback.
func WatchFile(dir string, rid, rid2 string, callback func(string, string)) (id int, err error) {
	return addWatch(dir, rid, rid2, callback)
}

//------------------------------------------------------------
// Not Exported functions
//------------------------------------------------------------

// Creates customizable watcher.
func addWatch(dir string, rid, rid2 string, callback func(string, string)) (id int, err error) {
	var watcher *fsnotify.Watcher
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return -1, err
	}

	w := &Watch{rid, rid2, dir, watcher, callback, fileDamper, nil}

	err = watcher.Watch(dir)

	if err != nil {
		return -1, err
	}

	go watch(w)

	_watchNextId++
	_watches[_watchNextId] = w
	return _watchNextId, err
}

// Closes several existing watch.
func CloseMany(ids []int) {
	for _, id := range ids {
		Close(id)
	}
}

// Closes existing watch.
func Close(id int) {
	w, ok := _watches[id]
	if !ok {
		return
	}

	//fmt.Println("[fwatch] closing watch:", id)
	w.watcher.Close()
	delete(_watches, id)
}

// Go routine that waits for an change event and notifies via callback
func watch(w *Watch) {
	//fmt.Println("[fwatch] watching for changes in:", w.dir)
	for {
		select {

		case ev := <-w.watcher.Event:
			if ev == nil {
				//fmt.Println("[fwatch] closed watch for:", w.dir)
				return
			}
			//fmt.Println("[fwatch] change", ev)
			scheduleCallback(w, ev.Name)

		case err := <-w.watcher.Error:
			fmt.Println("[fwatch] error:", err)
		}
	}
}

// Event damper prevents from too many events firing at once.
// Calls callback only after duration elapsed since last change event.
func scheduleCallback(w *Watch, dir string) {
	if w.timer == nil {
		w.timer = time.NewTimer(w.tdelta)
		go func() {
			<-w.timer.C
			//w.callback(w.id, w.id2)
            queueCallback(w)
			w.timer = nil
		}()

	} else {
		w.timer.Reset(w.tdelta)
	}
}

func queueCallback(w *Watch) {
    _callbacker.watches = append(_callbacker.watches, w)
	if _callbacker.timer == nil {
		_callbacker.timer = time.NewTimer(_callbacker.tdelta)
		go func() {
			<-_callbacker.timer.C
            _callbacker.execute()
			_callbacker.timer = nil
		}()

	} else {
		_callbacker.timer.Reset(_callbacker.tdelta)
	}
}

