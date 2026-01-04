package feconf

import (
	"flag"
	"reflect"
	"strings"
)

type flagValue struct {
	ptr   any
	isSet bool
}

var globalFlags = make(map[any]map[string]*flagValue)

func (c *ConfOpt[T]) registerFlags() {
	globalFlags[c] = make(map[string]*flagValue)

	if c.flagName != "" {
		flag.StringVar(&c.uri, c.flagName, c.uri, "configuration file path or URI")
	}

	var t T
	c.registerStructFlags(reflect.TypeOf(t))
}

func (c *ConfOpt[T]) registerStructFlags(t reflect.Type) {
	if t == nil {
		return
	}
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return
	}

	flags := globalFlags[c]
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		usage := field.Tag.Get("usage")
		if usage == "" {
			continue
		}

		name := field.Tag.Get("mapstructure")
		if name == "" {
			name = field.Tag.Get("json")
		}
		if name == "" {
			name = field.Name
		}
		name = strings.SplitN(name, ",", 2)[0]

		switch field.Type.Kind() {
		case reflect.String:
			v := new(string)
			flag.StringVar(v, name, "", usage)
			flags[name] = &flagValue{ptr: v}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			v := new(int)
			flag.IntVar(v, name, 0, usage)
			flags[name] = &flagValue{ptr: v}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			v := new(uint)
			flag.UintVar(v, name, 0, usage)
			flags[name] = &flagValue{ptr: v}
		case reflect.Bool:
			v := new(bool)
			flag.BoolVar(v, name, false, usage)
			flags[name] = &flagValue{ptr: v}
		case reflect.Float32, reflect.Float64:
			v := new(float64)
			flag.Float64Var(v, name, 0, usage)
			flags[name] = &flagValue{ptr: v}
		}
	}
}

func (c *ConfOpt[T]) parseFlags() {
	if !flag.Parsed() {
		flag.Parse()
	}
	flags := globalFlags[c]
	flag.Visit(func(f *flag.Flag) {
		if fv, ok := flags[f.Name]; ok {
			fv.isSet = true
		}
	})
}

func (c *ConfOpt[T]) mergeFlagValues() {
	flags := globalFlags[c]
	if flags == nil {
		return
	}
	if c.parsedData == nil {
		c.parsedData = make(map[string]any)
	}

	for key, fv := range flags {
		if !fv.isSet {
			continue
		}
		switch v := fv.ptr.(type) {
		case *string:
			c.parsedData[key] = *v
		case *int:
			c.parsedData[key] = *v
		case *uint:
			c.parsedData[key] = *v
		case *bool:
			c.parsedData[key] = *v
		case *float64:
			c.parsedData[key] = *v
		}
	}
}
