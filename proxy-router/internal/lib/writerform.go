package lib

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"reflect"
	"strings"
)

// WriteForm walks v, looks at its `form` tags, and writes fields.
// Slice/array fields are sent as repeated keys. `Extra` RawMessages are appended.
func WriteForm(w *multipart.Writer, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("WriteForm expects a struct, got %T", v)
	}

	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		tag := f.Tag.Get("form")
		if tag == "" || tag == "-" {
			continue
		}
		name, opts := parseTag(tag)
		if name == "" {
			name = f.Name
		}
		fv := rv.Field(i)
		if opts.Contains("omitempty") && fv.IsZero() {
			continue
		}
		if err := writeValue(w, name, fv); err != nil {
			return err
		}
	}

	// Extras: map[string]json.RawMessage
	if extraField := rv.FieldByName("Extra"); extraField.IsValid() && !extraField.IsZero() {
		if extra, ok := extraField.Interface().(map[string]json.RawMessage); ok {
			for k, raw := range extra {
				// If the RawMessage is a JSON string, drop the quotes for form field.
				val := raw
				if len(raw) >= 2 && raw[0] == '"' && raw[len(raw)-1] == '"' {
					val = raw[1 : len(raw)-1]
				}
				if err := w.WriteField(k, string(val)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

type tagOptions map[string]struct{}

func parseTag(tag string) (name string, opts tagOptions) {
	parts := strings.Split(tag, ",")
	name = parts[0]
	opts = make(tagOptions)
	for _, p := range parts[1:] {
		if p != "" {
			opts[p] = struct{}{}
		}
	}
	return
}
func (o tagOptions) Contains(k string) bool { _, ok := o[k]; return ok }

// writeValue writes a single struct field to the multipart writer.
func writeValue(w *multipart.Writer, key string, v reflect.Value) error {
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		// Treat []byte/json.RawMessage specially (avoid per-byte loop)
		if v.Type().Elem().Kind() == reflect.Uint8 {
			// Raw []byte: write as a single field
			return w.WriteField(key, string(v.Bytes()))
		}
		for i := 0; i < v.Len(); i++ {
			if err := w.WriteField(key, fmt.Sprint(v.Index(i).Interface())); err != nil {
				return err
			}
		}
	default:
		return w.WriteField(key, fmt.Sprint(v.Interface()))
	}
	return nil
}
