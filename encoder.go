package mdencoder

import (
	"fmt"
	"io"
	"reflect"
)

type Marshaler interface {
	MarshalMarkdown() ([]byte, error)
}

var marshalerType = reflect.TypeFor[Marshaler]()

type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w}
}

func (e *Encoder) Encode(obj any) error {
	if marshaler, ok := obj.(Marshaler); ok {
		if bs, err := marshaler.MarshalMarkdown(); err != nil {
			return err
		} else if _, err := e.w.Write(bs); err != nil {
			return err
		}
		return nil
	}
	val := reflect.ValueOf(obj)
	// Handle pointer values by dereferencing them
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}
	var err error
	switch val.Kind() {
	case reflect.Slice:
		_, err = sliceToMarkdown(e.w, val, 0)
	case reflect.Map:
		_, err = mapToMarkdown(e.w, val, StructMdTitleStyle, 0)
	case reflect.Struct:
		_, err = structToMarkdown(e.w, val, StructMdTitleStyle, 0)
	default:
		fmt.Fprintf(e.w, "%v", val.Interface())
	}
	return err
}
