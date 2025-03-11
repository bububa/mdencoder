package mdencoder

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"
)

var bufferPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

func getBuffer() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

func releaseBuffer(v *bytes.Buffer) {
	v.Reset()
	bufferPool.Put(v)
}

var encoderPool = sync.Pool{
	New: func() any {
		return new(Encoder)
	},
}

func getEncoder() *Encoder {
	enc := encoderPool.Get().(*Encoder)
	enc.w = getBuffer()
	return enc
}

func releaseEncoder(enc *Encoder) {
	releaseBuffer(enc.w.(*bytes.Buffer))
	enc.w = nil
	encoderPool.Put(enc)
}

// Marshal converts a struct instance to Markdown based on `jsonschema` tags.
func Marshal(obj any) ([]byte, error) {
	enc := getEncoder()
	defer releaseEncoder(enc)
	enc.Encode(obj)
	return enc.w.(*bytes.Buffer).Bytes(), nil
}

type StructMarkdownStyle int

const (
	StructMdTitleStyle StructMarkdownStyle = iota
	StructMdListStyle
	StructMdInListStyle
)

// structToMarkdown converts a struct instance to Markdown based on `jsonschema` tags.
func structToMarkdown(w io.Writer, val reflect.Value, style StructMarkdownStyle, indentLevel int) (int, error) {
	// Handle pointer values by dereferencing them
	typ := val.Type()
	if ok := typ.Implements(marshalerType); ok {
		if marshaler, ok := val.Interface().(Marshaler); ok {
			if bs, err := marshaler.MarshalMarkdown(); err != nil {
				return 0, err
			} else if _, err := w.Write(bs); err != nil {
				return 0, err
			}
		}
		return 0, nil
	}
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return 0, nil
		}
		val = val.Elem()
	}
	typ = val.Type()
	var (
		idx   int
		total = typ.NumField()
	)
	for i := range total {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// Skip empty fields if they have `omitempty`
		if shouldOmitField(field, fieldValue) {
			continue
		}

		if fieldValue.Kind() == reflect.Ptr {
			if fieldValue.IsNil() {
				continue
			}
			// Dereference the pointer and check if it's a struct or primitive
			fieldValue = fieldValue.Elem()
		}

		if field.Anonymous {
			if v, err := structToMarkdown(w, fieldValue, style, indentLevel); err != nil {
				return idx, err
			} else if v > 0 {
				idx += v
			}
			continue
		}

		if idx > 0 {
			w.Write([]byte{'\n'})
		}

		jsonTag := field.Tag.Get("json")
		jsonSchemaTag := field.Tag.Get("jsonschema")
		title, desc := parseJSONSchemaTag(jsonSchemaTag)
		// Fallback logic for title selection
		mdTitle := title
		if mdTitle == "" {
			mdTitle = parseJSONTag(jsonTag)
		}
		if mdTitle == "" {
			mdTitle = field.Name
		}
		var (
			indent         = bytes.Repeat([]byte{' '}, indentLevel*2)
			subIndent      = indent
			prefix         string
			subIndentLevel = indentLevel
		)
		switch style {
		case StructMdTitleStyle:
			prefix = "# "
		case StructMdListStyle:
			prefix = "- "
			subIndentLevel += 1
			subIndent = append(subIndent, ' ', ' ')
		case StructMdInListStyle:
			if idx > 0 {
				indent = append(indent, ' ', ' ')
				prefix = ""
			} else {
				prefix = "- "
			}
			subIndentLevel += 1
			subIndent = append(subIndent, ' ', ' ')
		}

		switch fieldValue.Kind() {
		case reflect.Struct:
			w.Write(indent)
			fmt.Fprintf(w, "%s%s", prefix, mdTitle)
			if desc != "" {
				fmt.Fprintf(w, " (%s)", desc)
			}
			w.Write([]byte{'\n'})
			// Handle nested structs
			if _, err := structToMarkdown(w, fieldValue, StructMdListStyle, subIndentLevel); err != nil {
				return idx, err
			}
		case reflect.Slice:
			w.Write(indent)
			fmt.Fprintf(w, "%s%s", prefix, mdTitle)
			if desc != "" {
				fmt.Fprintf(w, " (%s)", desc)
			}
			w.Write([]byte{'\n'})
			if _, err := sliceToMarkdown(w, fieldValue, subIndentLevel); err != nil {
				return idx, err
			}
		case reflect.Map:
			w.Write(indent)
			fmt.Fprintf(w, "%s%s", prefix, mdTitle)
			if desc != "" {
				fmt.Fprintf(w, " (%s)", desc)
			}
			w.Write([]byte{'\n'})
			if _, err := mapToMarkdown(w, fieldValue, StructMdListStyle, subIndentLevel); err != nil {
				return idx, err
			}
		default:
			w.Write(indent)
			fmt.Fprintf(w, "%s%s", prefix, mdTitle)
			if desc != "" {
				fmt.Fprintf(w, " (%s)", desc)
			}
			if style == StructMdTitleStyle {
				fmt.Fprintf(w, "\n%s%v", subIndent, fieldValue.Interface())
			} else {
				fmt.Fprintf(w, ": %v", fieldValue.Interface())
			}
		}
		idx++
		if idx < total {
			w.Write([]byte{'\n'})
		}
	}
	return idx, nil
}

// mapToMarkdown converts a struct instance to Markdown based on `jsonschema` tags.
func mapToMarkdown(w io.Writer, val reflect.Value, style StructMarkdownStyle, indentLevel int) (int, error) {
	// Handle pointer values by dereferencing them
	typ := val.Type()
	if ok := typ.Implements(marshalerType); ok {
		if marshaler, ok := val.Interface().(Marshaler); ok {
			if bs, err := marshaler.MarshalMarkdown(); err != nil {
				return 0, err
			} else if _, err := w.Write(bs); err != nil {
				return 0, err
			}
		}
		return 0, nil
	}
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return 0, nil
		}
		val = val.Elem()
	}
	keys := val.MapKeys()
	var idx int
	for _, key := range keys {
		fieldValue := val.MapIndex(key)
		if fieldValue.Kind() == reflect.Ptr {
			if fieldValue.IsNil() {
				continue
			}
			fieldValue = fieldValue.Elem()
		}
		var (
			mdTitle        = fmt.Sprintf("%v", key.Interface())
			indent         = bytes.Repeat([]byte{' '}, indentLevel*2)
			prefix         string
			subIndentLevel = indentLevel
		)
		switch style {
		case StructMdTitleStyle:
			prefix = "# "
		case StructMdListStyle:
			prefix = "- "
			subIndentLevel += 1
		case StructMdInListStyle:
			if idx > 0 {
				indent = append(indent, ' ', ' ')
				prefix = ""
			} else {
				prefix = "- "
			}
			subIndentLevel += 1
		}
		if idx > 0 {
			w.Write([]byte{'\n'})
		}
		switch fieldValue.Kind() {
		case reflect.Struct:
			w.Write(indent)
			fmt.Fprintf(w, `%s%s`, prefix, mdTitle)
			w.Write([]byte{'\n'})
			// Handle nested structs
			if _, err := structToMarkdown(w, fieldValue, StructMdListStyle, subIndentLevel); err != nil {
				return idx, err
			}
		case reflect.Slice:
			w.Write(indent)
			fmt.Fprintf(w, `%s%s`, prefix, mdTitle)
			w.Write([]byte{'\n'})
			if _, err := sliceToMarkdown(w, fieldValue, subIndentLevel); err != nil {
				return idx, err
			}
		case reflect.Map:
			w.Write(indent)
			fmt.Fprintf(w, `%s%s`, prefix, mdTitle)
			w.Write([]byte{'\n'})
			if _, err := mapToMarkdown(w, fieldValue, StructMdListStyle, subIndentLevel); err != nil {
				return idx, err
			}
		default:
			w.Write(indent)
			fmt.Fprintf(w, "%s%s: %v", prefix, mdTitle, fieldValue.Interface())
		}
		idx++
	}
	return idx, nil
}

// sliceToMarkdown converts a struct instance to Markdown based on `jsonschema` tags.
func sliceToMarkdown(w io.Writer, val reflect.Value, indentLevel int) (int, error) {
	// Handle pointer values by dereferencing them
	typ := val.Type()
	if ok := typ.Implements(marshalerType); ok {
		if marshaler, ok := val.Interface().(Marshaler); ok {
			if bs, err := marshaler.MarshalMarkdown(); err != nil {
				return 0, err
			} else if _, err := w.Write(bs); err != nil {
				return 0, err
			}
		}
		return 0, nil
	}
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return 0, nil
		}
		val = val.Elem()
	}
	var (
		idx   int
		total = val.Len()
	)
	for i := range total {
		value := val.Index(i)
		if value.Kind() == reflect.Ptr {
			if value.IsNil() {
				continue
			}
			value = value.Elem()
		}
		if idx > 0 {
			w.Write([]byte{'\n'})
		}
		switch value.Kind() {
		case reflect.Struct:
			if _, err := structToMarkdown(w, value, StructMdInListStyle, indentLevel); err != nil {
				return idx, err
			}
		case reflect.Map:
			if _, err := mapToMarkdown(w, value, StructMdInListStyle, indentLevel); err != nil {
				return idx, err
			}
		case reflect.Slice:
			if _, err := sliceToMarkdown(w, value, indentLevel+1); err != nil {
				return idx, err
			}
		default:
			indent := strings.Repeat(" ", indentLevel*2)
			fmt.Fprintf(w, "%s- %v", indent, value)
		}
		idx += 1
	}
	return idx, nil
}

// shouldOmitField checks if a field should be omitted based on `omitempty`
func shouldOmitField(field reflect.StructField, fieldValue reflect.Value) bool {
	// Get the json tag to check for `omitempty`
	jsonTag := field.Tag.Get("json")
	omitempty := strings.Contains(jsonTag, "omitempty")

	// If there's no `omitempty`, keep the field
	if !omitempty {
		return false
	}

	// Check if the field is zero-valued
	if !fieldValue.IsValid() || fieldValue.IsZero() {
		return true
	}

	return false
}

// parseJSONSchemaTag extracts title and description from jsonschema tag.
func parseJSONSchemaTag(tag string) (title, description string) {
	parts := strings.Split(tag, ",")
	for _, part := range parts {
		if strings.HasPrefix(part, "title=") {
			title = strings.TrimPrefix(part, "title=")
		} else if strings.HasPrefix(part, "description=") {
			description = strings.TrimPrefix(part, "description=")
		}
	}
	return title, description
}

// parseJSONTag extracts title from json tag.
func parseJSONTag(tag string) string {
	parts := strings.Split(tag, ",")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}
