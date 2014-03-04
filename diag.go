// Diagnostics inteface
package wiro

import(
    _D "github.com/deze333/diag"
)

// Package name for diagnostics messages
const _me = "wiro"


func DEBUG(location, title string, v ...interface{}) {
    _D.DEBUG("[" + _me +" : " + location +"]", title, v...)
}

func NOTE(msg string, v ...interface{}) {
    _D.NOTE(msg, v...)
}

func NOTE2(msg string, v ...interface{}) {
    _D.NOTE2(msg, v...)
}

func WARNING(location, title string, v ...interface{}) {
    _D.WARNING("[" + _me +" : " + location +"]", title, v...)
}

func ERROR(location, title string, v ...interface{}) {
    _D.ERROR("[" + _me +" : " + location +"]", title, v...)
}

func SOS(location, title string, v ...interface{}) {
    _D.SOS("[" + _me +" : " + location +"]", title, v...)
}
