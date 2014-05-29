// Web resource repository manager
package wiro

import (
	"fmt"
    "time"
)

//------------------------------------------------------------
// Library of managed repositories
//------------------------------------------------------------

// Global map of managed repositories
var _library = map[string]*Repo{}

//------------------------------------------------------------
// Repository
//------------------------------------------------------------

// Describes a repository located under Dir.
// Resources is a map, where key is the name of each resource,
// (ie, home.ini, help/info.ini), they are provided by parse.
// Each map entry contains a slice of resources
// each having its individual key (ie. _ _ _, com _ _,... )
type Repo struct {
    Id        string
	Dir       string
	WatchIds  []int
    Parsers   *ParserLib
	Resources map[string][]*Resource
	resources map[string][]*Resource
    onReload  func()
}

//------------------------------------------------------------
// Resource interface
//------------------------------------------------------------

// Interface that resource object must satisfy.
type Resource interface {
	GetId() string
	GetKey() *Key
	Get() interface{}
}

//------------------------------------------------------------
// Key
//------------------------------------------------------------

// Key is unique identifier for each loaded resource.
// Id can be file name. File is full path to that file.
// Domain, Language are strings, empty "" means default.
// Weight ranges from 0 to 100 (measured in %).
type Key struct {
	Id       string
	File     string
    ModTime  time.Time
	Domain   string
	Language string
    Version  string
}

// Parser converts file into resource according to key.
type Parser func(Key)(interface{}, error)
// Parser library maps file name to a parser.
type ParserLib map[string]Parser


//------------------------------------------------------------
// Main method: Directory loader
//------------------------------------------------------------

// Creates resource from specified directory. All resources are of same homogenous type
// which means same single parser used for all.
func CreateHomogenous(id string, dir string, files []string, parser Parser) (err error) {
    parsers := &ParserLib{}
    for _, f := range files {
        (*parsers)[f] = parser
    }

    return Create(id, dir, parsers)
}

// Creates resource from specified directory and set of parsers.
func Create(id string, dir string, parsers *ParserLib) (err error) {
    repo := &Repo{
        Id: id,
        Dir: dir, 
        WatchIds: []int{},
        Parsers: parsers,
        Resources: map[string][]*Resource{},
        resources: map[string][]*Resource{},
    }
    err = load(repo)
    if err != nil {
        return
    }
    // Install loaded repository into library
    _library[id] = repo
    return
}

// Sets optional reload listener.
func SetOnReload(repoId string, fn func()) (err error) {
    if repo, ok := _library[repoId]; ok {
        repo.onReload = fn
        _library[repoId] = repo
        return

    } else {
        return fmt.Errorf("Repository not found: %s", repoId)
    }
}

// Retrieves resource from specified repository.
// dlv is domain, language, version which can be omitted meaning default.
func Get(repoId, rsrcId string, dlv ...string) (rsrc *Resource) {
    r := _library[repoId]
    if r == nil {
        return nil
    }

    var domain, language, version string
    switch len(dlv) {
    case 0:
    case 1:
        domain = dlv[0]
    case 2:
        domain = dlv[0]
        language = dlv[1]
    case 3:
        domain = dlv[0]
        language = dlv[1]
        version = dlv[2]
    default:
        domain = dlv[0]
        language = dlv[1]
        version = dlv[2]
    }

    return r.Get(rsrcId, domain, language, version)
}


// Retrieves repository
func getRepo(id string) (repo *Repo, ok bool) {
    repo, ok = _library[id]
    return
}

//------------------------------------------------------------
// Key methods
//------------------------------------------------------------

func (k *Key) GetId() string {
	return k.Id
}

func (k *Key) GetKey() *Key {
	return k
}

func (k *Key) Dump() {
	fmt.Println("KEY =", k.Id, ",", k.Domain, ",", k.Language, ",", k.Version)
}

//------------------------------------------------------------
// Repository methods
//------------------------------------------------------------

// Adds resource to temporary repository.
func (r *Repo) addTemp(k *Key, rsrc Resource) {
    if k == nil {
        SOS("add", "Adding nil resource key")
        return
    }

    if rsrc == nil {
        SOS("add", "Adding nil resource")
        return
    }
    // Ensure no duplicate resources
    for _, rsrcs := range r.resources {
        for _, r := range rsrcs {
            if r == &rsrc {
                SOS("addTemp", 
                "This resource already present, skipping duplicate", 
                "rsrc", rsrc)
                return
            }
        }
    }
    // Add resource
    id := k.GetId()
    r.resources[id] = append(r.resources[id], &rsrc)
    return
}

// Activates repository via hot swap
func (r * Repo) hotSwapAll() {
    r.overlayTemp()
    r.Resources = r.resources
    r.resources = map[string][]*Resource{}
}


// Overlays repository resources (specific over default).
func (r * Repo) overlayTemp() {
    // For each resource
    var defrsrc *Resource
    for _, rsrcs := range r.resources {

        // Find default resource
        defrsrc = nil
        for _, rsrc := range rsrcs {
            if (*rsrc).GetKey().Domain == "" && 
            (*rsrc).GetKey().Language == "" && 
            (*rsrc).GetKey().Version == "" {
                defrsrc = rsrc
                break
            }
        }

        if defrsrc == nil {
            continue
        }

        // Overlay each other resource over default
        for _, rsrc := range rsrcs {
            // Skip default
            if (*rsrc).GetKey().Domain == "" && 
            (*rsrc).GetKey().Language == "" && 
            (*rsrc).GetKey().Version == "" {
                continue
            }
            overlay(rsrc, defrsrc)
            // Adjust modified time stamp to the latest of the two
            if (*rsrc).GetKey().ModTime.Before((*defrsrc).GetKey().ModTime) {
                (*rsrc).GetKey().ModTime = (*defrsrc).GetKey().ModTime
            }
        }
    }
}

func (r *Repo) dump() {
    fmt.Println("Resource for:", r.Dir)
    for i, rsrcs := range r.Resources {
        fmt.Println("\n---- RESOURCE ID =", i, "---------------------------------")
        for _, rsrc := range rsrcs {
            if rsrc != nil {
                (*rsrc).GetKey().Dump()
                fmt.Println("\n", *rsrc, "\n")
            }
        }
    }
    fmt.Println("----------------------------------------------------------------")
}

func (r *Repo) dumpTemp() {
    fmt.Println("Resource for:", r.Dir)
    for i, rsrcs := range r.resources {
        fmt.Println("\n---- TEMP RESOURCE ID =", i, "---------------------------------")
        fmt.Println(rsrcs)
        for _, rsrc := range rsrcs {
            if rsrc != nil {
                (*rsrc).GetKey().Dump()
                fmt.Println("\n", *rsrc, "\n")
            }
        }
    }
    fmt.Println("----------------------------------------------------------------")
}

func (r *Repo) Get(id string, domain, language, version string) (rsrc *Resource) {
    rsrcs, ok := r.Resources[id]
    if !ok {
        return
    }

    // Strategy:
    // 1. Exact match
    //    Try: D L V
    // 2. Ignore L
    //    Try: D _ V
    // 3. Ignore V
    //    Try: D _ _
    // 4. Only V
    //    Try: _ _ V
    // 5. Default: _ _ _

    // 1. Exact match: Domain - Lang - Ver
    for _, rsrc = range rsrcs {
        k := (*rsrc).GetKey()
        if k.Domain == domain && 
            k.Language == language && 
            k.Version == version {
                return rsrc
        }
    }

    // 2. Ignore language: Domain - _ - Ver
    for _, rsrc = range rsrcs {
        k := (*rsrc).GetKey()
        if k.Domain == domain && 
            k.Language == "" && 
            k.Version == version {
                return rsrc
        }
    }

    // 3. Ignore version: Domain - _ - _
    for _, rsrc = range rsrcs {
        k := (*rsrc).GetKey()
        if k.Domain == domain &&
            k.Language == "" && 
            k.Version == "" {
                return rsrc
        }
    }

    // 3. Only version: _ - _ - Ver
    for _, rsrc = range rsrcs {
        k := (*rsrc).GetKey()
        if k.Domain == "" &&
            k.Language == "" && 
            k.Version == version {
                return rsrc
        }
    }

    // 5. Default: _ - _ - _
    for _, rsrc := range rsrcs {
        k := (*rsrc).GetKey()
        if k.Domain == "" && 
            k.Language == "" && 
            k.Version == "" {
                return rsrc
        }
    }

    return nil
}

