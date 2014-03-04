// Uses reflection to overlay resources
package wiro

import (
    "fmt"
    "reflect"
)

//------------------------------------------------------------
// 
//------------------------------------------------------------

func overlay(child, parent *Resource) {
    if reflect.TypeOf(child) != reflect.TypeOf(parent) {
        panic("[wiro] types don't match")
    }

    if reflect.TypeOf(child).Kind() != reflect.Ptr {
        panic("[wiro] child must be pointer")
    }

    if reflect.TypeOf(parent).Kind() != reflect.Ptr {
        panic("[wiro] parent must be pointer")
    }

    childVal := reflect.ValueOf((*child).Get()).Elem()
    parentVal := reflect.ValueOf((*parent).Get()).Elem()

    if childVal.Kind() != reflect.Struct {
        panic("[wiro] child Get() must return struct")
    }

    if parentVal.Kind() != reflect.Struct {
        panic("[wiro] parent Get() must return struct")
    }

    keyType := reflect.TypeOf(*(*child).GetKey())
    overlayValues(&childVal, &parentVal, keyType)
}

func overlayValues(child, parent *reflect.Value, skip reflect.Type) {
    for i := 0; i < child.NumField(); i++ {
        f := child.Field(i)

        if f.Type() == skip {
            continue
        }

        fp := parent.Field(i)
        switch f.Kind() {
        case reflect.String:
            overlayString(&f, &fp)

        case reflect.Slice:
            overlaySlice(&f, &fp)

        case reflect.Map:
            overlayMap(&f, &fp)

        case reflect.Struct:
            overlayValues(&f, &fp, skip)

        default:
            fmt.Println("[wiro] skipping not supported field type:", f.Type())
        }
    }
}

func overlayString(dst, src *reflect.Value) {
    if dst.Len() == 0 && src.Len() > 0 {
        if dst.CanSet() {
            dst.Set(*src)
        } else {
            fmt.Println("[wiro] skipping setting string field:", dst.String())
        }
    }
}

func overlaySlice(dst, src *reflect.Value) {
    if dst.IsNil() && !src.IsNil() {
        if dst.CanSet() {
            dst.Set(*src)
        } else {
            fmt.Println("[wiro] skipping setting slice field:", dst.String())
        }
    }
}

func overlayMap(dst, src *reflect.Value) {
    if dst.IsNil() && !src.IsNil() {
        if dst.CanSet() {
            dst.Set(*src)
        } else {
            fmt.Println("[wiro] skipping setting map field:", dst.String())
        }
    }
}
