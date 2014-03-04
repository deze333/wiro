// Loads resource tree
package wiro

import (
    "fmt"
    "io/ioutil"
    "os"
    "path"
    "path/filepath"
    "regexp"
    "strconv"
    "github.com/deze333/wiro/fwatch"
)

// Signature of qualified resource directory name.
// <domain> <language> <[comment]<weight>>
// Root defined by: _ _ _
// Children examples:
// com _ _
// com _ big-banner 
// com es even-bigger-banner
// co.uk _ _
//var reDataDir = regexp.MustCompile(`^(?P<domain>_|[^\s]+)\s(?P<lang>_|[^\s]+)\s(?P<ver>_|[a-zA-Z]*(?P<weight>\d+)%)$`)

// This regex works with version that is specified by sting made of letters, numbers and '-'
var reDataDir = regexp.MustCompile(`^(?P<domain>_|[^\s]+)\s(?P<lang>_|[^\s]+)\s(?P<ver>_|[a-zA-Z0-9\-]*)$`)


//------------------------------------------------------------
// Directory Loader
//------------------------------------------------------------

// Loads directories residing in given root directory.
// Subdirectories can have inner directories.
func load(repo *Repo) (err error) {
    
    err = loadRoot(repo)
    if err != nil {
        return
    }

    // Watch root directory for changes.
    // No need to remember watcher id as root doesn't change
    // _, err = fwatch.WatchDir(dir, reloadRoot)
    return
}

//------------------------------------------------------------
// Directory reloaders
//------------------------------------------------------------

func loadRoot(repo *Repo) (err error) {
    // Discover all subdirectories
    var fis []os.FileInfo
    fis, err = ioutil.ReadDir(repo.Dir)
    if err != nil {
        SOS("loadRoot", "Error reading directory", "err", err, "dir", repo.Dir)
        return
    }

    // XXX: Stop all old watches
    fwatch.CloseMany(repo.WatchIds)

    // Map of watched subdirs (dir:key)
    watches := map[string]string{}

    // Load each key subdirectory
    for _, fi := range fis {
        if fi.IsDir() {
            d := path.Join(repo.Dir, fi.Name())
            subdirs := loadKeyDir(repo, d, fi.Name())
            // XXX Add to watched
            for sd, _ := range subdirs {
                watches[path.Join(repo.Dir, fi.Name(), sd)] = path.Join(fi.Name(), sd)
            }
        }
    }

    // Activate this repository
    repo.hotSwapAll()
    //repo.dump()

    // XXX Add root to watched
    var id int
    id, err = fwatch.WatchDir(repo.Dir, repo.Id, "", onDirChanged)
    if err != nil {
        SOS("loadRoot", "Error adding watch", "dir", repo.Dir, "err", err)
    } else {
        repo.WatchIds = append(repo.WatchIds, id)
    }

    // XXX Add root subdirs to watched
    for dir, subdir := range watches {
        //NOTE("WATCH", "d", dir, "key", subdir)
        id, err = fwatch.WatchDir(dir, repo.Id, subdir, onDirChanged)
        if err != nil {
            SOS("loadRoot", "Error adding watch", "dir", dir, "err", err)
        } else {
            repo.WatchIds = append(repo.WatchIds, id)
        }
    }

    return
}

// Callback on root directory changed.
// id is the repository id (ie, text),
// id2 is key subdirectory inside Repo.Dir.
func onDirChanged(id, id2 string) {
    DEBUG("onDirChanged", "Reloading repository", "id", id)
    //NOTE2("Reloading repository", id)
    if repo, ok := getRepo(id); ok {
        if id2 == "" {
            //NOTE("Reloading repo root", "id", repo.Id)
            loadRoot(repo)
        } else {
            //NOTE("Reloading repo subdir", "id", repo.Id, "subdir", id2)
            loadRoot(repo)
            /* XXX Too complicated, reload the whole repo instead.
            var subdir = id2
            for id2 != "." {
                subdir = id2
                id2, _ = path.Split(id2)
                id2 = path.Clean(id2)
            }
            loadKeyDir(repo, path.Join(repo.Dir, subdir), subdir)
            repo.hotSwap(???)
            */
        }

        // Notify optional onReload listener
        if repo.onReload != nil {
            go repo.onReload()
        }
    }
    return
}

// Loads single key directory (ie, `com de 50%`)
// and adds to temporary repository.
// dir is key directory full path (ie, /home/alex/www/txt/_ _ _),
// subdir is last part of dir (ie, _ _ _)
// Returns map of directories to watch.
func loadKeyDir(repo *Repo, dir, subdir string) (subdirs map[string]bool) {
    //NOTE2("Parsing key", subdir)
    var err error

    // Parse subdirectory name
    var domain, lang, ver string
    domain, lang, ver, err = parseDirName(subdir)
    if err != nil {
        WARNING("loadKeyDir", "Skipping directory", "dir", subdir, "err", err)
        return
    }

    subdirs = map[string]bool{}
    subdirs["."] = true

    // Parse each file with its parser
    var key Key
    var fi os.FileInfo
    var resource Resource
    var ok bool
    var obj interface{}

    for f, parser := range *(repo.Parsers) {
        if parser == nil {
            continue
        }
        //NOTE("Parsing", "subdir", subdir , "file", f)

        // Check if file exists in this key directory
        fpath := path.Join(dir, f)
        if fi, err = os.Stat(fpath); os.IsNotExist(err) {
            continue
        }

        // Parse and add resource
        key = Key{
            Id: f,
            File: fpath,
            ModTime: fi.ModTime(),
            Domain: domain,
            Language: lang,
            Version: ver,
        }
        obj, err = parser(key)
        if err != nil {
            SOS("loadKeyDir", 
            "Error parsing resource, file skipped", 
            "dir", subdir, "file", f,
            "err", err)
            continue
        }
        if resource, ok = obj.(Resource); !ok {
            SOS("loadKeyDir", 
            "Returned struct is not of type Resource, file skipped",
            "dir", subdir, "file", f,
            "err", err)
            continue
        }
        repo.addTemp(&key, resource)
        subdirs[path.Dir(f)] = true
    }
    return
}

//------------------------------------------------------------
// Directory name parsing
//------------------------------------------------------------

// Directory is not relevant if:
// (a) does not match the pattern; or
// (b) matches the pattern but has weight == 0
func isRelevant(path string) bool {
    match := reDataDir.FindStringSubmatch(path)
    if match == nil {
        return false
    }

    for i, name := range reDataDir.SubexpNames() {
        switch name {
        case "ver":
            if match[i] == "_" {
                return true
            }
        case "weight":
            weight, _ := strconv.Atoi(match[i])
            if weight == 0 {
                return false
            }
        }
    }

    return true
}

//------------------------------------------------------------
// Directory name parsing
//------------------------------------------------------------

// Parse directory name
func parseDirName(path string) (domain, lang, ver string, err error) {
    match := reDataDir.FindStringSubmatch(path)
    if match == nil {
        err = fmt.Errorf("error parsing text directory name: %s", path)
        return
    }

    var val string
    for i, name := range reDataDir.SubexpNames() {
        switch name {
        case "domain":
            val = match[i]
            if val != "_" {
                domain = val
            }
        case "lang":
            val = match[i]
            if val != "_" {
                lang = val
            }
        case "ver":
            val = match[i]
            if val != "_" {
                ver = val
            }
        /* 
        case "weight":
            if weight != 100 {
                weight, _ = strconv.Atoi(match[i])
                if weight > 99 {
                    weight = 99
                }
            }
        */
        }
    }
    return
}

//------------------------------------------------------------
// OFF: Walker, not currently used
//------------------------------------------------------------

func OFF_newWalker(root string) filepath.WalkFunc {
    return func(path string, info os.FileInfo, err error) error {
        if err != nil {
            WARNING("texts:loader", "Could not scan directory, ignoring it", 
                "dir", path, "err", err)
            // Keep scanning
            return nil
        }

        // Extract subpath
        subpath, err := filepath.Rel(root, path)
        if err != nil {
            return err
        }

        // Do not process files in the root
        if subpath == "." {
            return nil
        }

        if info.IsDir() {
            // Skip irrelevant directories
            if ! isRelevant(subpath) {
                return filepath.SkipDir
            }
            // Valid directory but no further processing is needed
            return nil
        }

        // Valid file
        dir, file := filepath.Split(subpath)
        dir = filepath.Clean(dir)
        fmt.Println(dir, ":", file)
        var domain, lang, ver string
        domain, lang, ver, err = parseDirName(dir)
        if err != nil {
            WARNING("texts:loader", "Could not parse directory name, ignoring it", 
                "dir", path, "err", err)
            // Keep scanning
            return nil
        }
        fmt.Println("Domain =", domain, "Lang =", lang, "Ver =", ver)

        // 
        return nil
    }
}

