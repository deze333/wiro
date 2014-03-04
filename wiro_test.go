// Tester
package wiro

import (
	"fmt"
	"github.com/deze333/skini"
    "io/ioutil"
	"os"
	"path"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	var err error
	fmt.Println("--------------------------------------------------------")
	fmt.Println("|")
	fmt.Println("|")
	fmt.Println("|")
	fmt.Println("V")
	fmt.Println("Loading resource...")

	var pwd string
	pwd, err = os.Getwd()
	if err != nil {
		t.Errorf("Error defining working directory: %s", err)
		return
	}

	var rsrc *Resource

    //------------------------------------------------------------ 
	// Create repo for page templates
    err = CreateHomogenous("tpl", path.Join(pwd, "sample/tpl"), tplFiles,  tplParser)
	if err != nil {
		t.Errorf("Error while parsing templates: %s", err)
		return
	}

	// Get default template
    var tpl *PageTpl
	rsrc = Get("tpl", "info.html", "", "", "")
    if rsrc == nil {
		t.Errorf("Template resource not found")
        return
    }

    // Get def template 
    tpl = (*Get("tpl", "info.html", "", "", "")).Get().(*PageTpl)
    NOTE2("Tpl", "id", "info.html", "key", "_ _ _", "val", tpl.Html)

    // Get def template version
    tpl = (*Get("tpl", "info.html", "", "", "blue-elephant")).Get().(*PageTpl)
    NOTE2("Tpl", "id", "info.html", "key", "_ _ blue-elephant", "val", tpl.Html)

    // Get COM template 
    tpl = (*Get("tpl", "info.html", "com", "", "")).Get().(*PageTpl)
    NOTE2("Tpl", "id", "info.html", "key", "com _ _", "val", tpl.Html)

    // Get COM template version
    tpl = (*Get("tpl", "info.html", "com", "", "blue-elephant")).Get().(*PageTpl)
    NOTE2("Tpl", "id", "info.html", "key", "com _ blue-elephant", "val", tpl.Html)

    // Get COM.AU template 
    tpl = (*Get("tpl", "info.html", "com.au", "", "")).Get().(*PageTpl)
    NOTE2("Tpl", "id", "info.html", "key", "com.au _ _", "val", tpl.Html)

    // Get COM.AU template version
    tpl = (*Get("tpl", "info.html", "com.au", "", "blue-elephant")).Get().(*PageTpl)
    NOTE2("Tpl", "id", "info.html", "key", "com.au _ blue-elephant", "val", tpl.Html)

    //------------------------------------------------------------ 
	// Create repo for texts
	err = Create("txt", path.Join(pwd, "sample/txt"), &textParsers)
	if err != nil {
		t.Errorf("Error while parsing texts: %s", err)
		return
	}

	// Get COM resource
	var txt *InfoText
	rsrc = Get("txt", "info.ini", "com", "", "blue-elephant")
    if rsrc == nil {
		t.Errorf("Text resource not found")
        return
    }

    // Get def resource 
    txt = (*Get("txt", "info.ini", "", "", "")).Get().(*InfoText)
    NOTE2("Text", "id", "info.ini", "key", "_ _ _")
    fmt.Println(txt)
    fmt.Println()

    // Get resource version
    txt = (*Get("txt", "info.ini", "", "", "blue-elephant")).Get().(*InfoText)
    NOTE2("Text", "id", "info.ini", "key", "_ _ blue-elephant")
    fmt.Println(txt)
    fmt.Println()

    // Get COM resource 
    txt = (*Get("txt", "info.ini", "com", "", "")).Get().(*InfoText)
    NOTE2("Text", "id", "info.ini", "key", "com _ _")
    fmt.Println(txt)
    fmt.Println()

    // Get resource version
    txt = (*Get("txt", "info.ini", "com", "", "blue-elephant")).Get().(*InfoText)
    NOTE2("Text", "id", "info.ini", "key", "com _ blue-elephant")
    fmt.Println(txt)
    fmt.Println()

    // Get COM.AU resource 
    txt = (*Get("txt", "info.ini", "com.au", "", "")).Get().(*InfoText)
    NOTE2("Text", "id", "info.ini", "key", "com.au _ _")
    fmt.Println(txt)
    fmt.Println()

    // Get COM.AU resource version
    txt = (*Get("txt", "info.ini", "com.au", "", "blue-elephant")).Get().(*InfoText)
    NOTE2("Text", "id", "info.ini", "key", "com.au _ blue-elephant")
    fmt.Println(txt)
    fmt.Println()

    // STOP FOR NOW
    return

	rsrc = Get("txt", "info.ini", "com", "", "")
	txt = (*rsrc).Get().(*InfoText)
	NOTE2("Text", "com _ _")
	fmt.Println(txt)
	fmt.Println()

	rsrc = Get("txt", "home/info.ini", "com", "", "")
	txt = (*rsrc).Get().(*InfoText)
	NOTE2("Text", "com _ _")
	fmt.Println(txt)
	fmt.Println()


	go periodicRequests(10, "txt", "", "", "")
	go periodicRequests(3000, "txt", "info.ini", "com.au", "")
	go periodicUpdates(4000, path.Join(pwd, "sample/txt", "_ _ _/info.ini"))

	// Forever loop to test reloading
	NOTE("Running continuous tests, use Ctrl-C to exit...")
	ch := make(chan bool, 1)
	<-ch
}

func periodicRequests(ms time.Duration, repoId, rsrcId, domain, lang string) {
	timer := time.NewTimer(ms * time.Millisecond)
    for {
        <-timer.C
        if repoId == "" || rsrcId == "" {
            if !getAllResources() {
                return
            }
        } else {
            if !printResource(repoId, rsrcId, domain, lang) {
                return
            }
        }
        timer = time.NewTimer(ms * time.Millisecond)
    }
}

func getAllResources() bool {
	var rsrc *Resource
    for repoId, _ := range _library {
        for rsrcId, _ := range _library[repoId].Resources {
            rsrc = Get(repoId, rsrcId, "", "", "")
            if rsrc == nil {
                ERROR("getAllResources", "No resource found")
                return false
            }
        }
    }
	return true
}

func printResource(repoId, rsrcId, domain, lang string) bool {
	var rsrc *Resource
	rsrc = Get(repoId, rsrcId, domain, lang, "")
    if rsrc == nil {
		ERROR("printResource", 
        "No resource found",
        "repoId", repoId,
        "rsrcId", rsrcId,
        "domain", domain,
        "lang", lang)
        return false
    }

    txt := (*rsrc).Get().(*InfoText)
	NOTE2("Text", txt)
    return true
}

func periodicUpdates(ms time.Duration, file string) {
	timer := time.NewTimer(ms * time.Millisecond)
    for {
        <-timer.C
        if !updateResource(file) {
            return
        }
        timer = time.NewTimer(ms * time.Millisecond)
    }
}

func updateResource(file string) bool {
    NOTE("Updating resource file", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		ERROR("updateResource", "Error opening file", "err", err)
		return false
	}
	defer f.Close()

	if _, err := f.WriteString(""); err != nil {
		ERROR("updateResource", "Error appending to file", "err", err)
		return false
	}

	return true
}

//------------------------------------------------------------
// Parser lib
//------------------------------------------------------------

var tplFiles = []string{
	"info.html",
}

var textParsers = ParserLib{
	"info.ini":      infoTextParser,
	"unknown.ini":   nil,
}

//------------------------------------------------------------
// Info page
//------------------------------------------------------------

type PageTpl struct {
	Key
	Html     string
}

func (t *PageTpl) Get() interface{} {
	return t
}

//------------------------------------------------------------
// Info text
//------------------------------------------------------------

type InfoText struct {
	Key
	Name     string
	Phone    string
	Colors   []string
	Friends  map[string]map[string]string
	Articles map[string]map[string]string

	Extra struct {
		Info string
		Mask string
	}
}

func (t *InfoText) Get() interface{} {
	return t
}

//------------------------------------------------------------
// Info HTML parser
//------------------------------------------------------------

func tplParser(key Key) (rsrc interface{}, err error) {
	page := &PageTpl{}
    var bytes []byte
    bytes, err = ioutil.ReadFile(key.File) 
	if err != nil {
		return
	}
	page.Key = key
    page.Html = string(bytes)
	return page, err
}
//------------------------------------------------------------
// Info text parser
//------------------------------------------------------------

func infoTextParser(key Key) (rsrc interface{}, err error) {
	txt := &InfoText{}
	err = skini.ParseFile(txt, key.File)
	if err != nil {
		return
	}
	txt.Key = key

    /* For debug only
	d, _ := path.Split(key.File)
	d = path.Clean(d)
	_, d = path.Split(d)
    */
	return txt, err
}
