package arg

import (
	"encoding"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	scalar "github.com/alexflint/go-scalar"
)

// spec represents a command line option
type spec struct {
	dest       reflect.Value
	long       string
	short      string
	multiple   bool
	required   bool
	positional bool
	separate   bool
	help       string
	env        string
	wasPresent bool
	boolean    bool
}

// ErrHelp indicates that -h or --help were provided
var ErrHelp = errors.New("help requested by user")

// ErrVersion indicates that --version was provided
var ErrVersion = errors.New("version requested by user")

// MustParse processes command line arguments and exits upon failure
func MustParse(dest ...interface{}) *Parser {
	p, err := NewParser(Config{}, dest...)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	err = p.Parse(flags())
	if err == ErrHelp {
		p.WriteHelp(os.Stdout)
		os.Exit(0)
	}
	if err == ErrVersion {
		fmt.Println(p.version)
		os.Exit(0)
	}
	if err != nil {
		p.Fail(err.Error())
	}
	return p
}

// Parse processes command line arguments and stores them in dest
func Parse(dest ...interface{}) error {
	p, err := NewParser(Config{}, dest...)
	if err != nil {
		return err
	}
	return p.Parse(flags())
}

// flags gets all command line arguments other than the first (program name)
func flags() []string {
	if len(os.Args) == 0 { // os.Args could be empty
		return nil
	}
	return os.Args[1:]
}

// Config represents configuration options for an argument parser
type Config struct {
	Program string // Program is the name of the program used in the help text
}

// Parser represents a set of command line options with destination values
type Parser struct {
	spec        []*spec
	config      Config
	version     string
	description string
}

// Versioned is the interface that the destination struct should implement to
// make a version string appear at the top of the help message.
type Versioned interface {
	// Version returns the version string that will be printed on a line by itself
	// at the top of the help message.
	Version() string
}

// Described is the interface that the destination struct should implement to
// make a description string appear at the top of the help message.
type Described interface {
	// Description returns the string that will be printed on a line by itself
	// at the top of the help message.
	Description() string
}

// walkFields calls a function for each field of a struct, recursively expanding struct fields.
func walkFields(v reflect.Value, visit func(field reflect.StructField, val reflect.Value, owner reflect.Type) bool) {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		val := v.Field(i)
		expand := visit(field, val, t)
		if expand && field.Type.Kind() == reflect.Struct {
			walkFields(val, visit)
		}
	}
}

// NewParser constructs a parser from a list of destination structs
func NewParser(config Config, dests ...interface{}) (*Parser, error) {
	p := Parser{
		config: config,
	}
	for _, dest := range dests {
		if dest, ok := dest.(Versioned); ok {
			p.version = dest.Version()
		}
		if dest, ok := dest.(Described); ok {
			p.description = dest.Description()
		}
		v := reflect.ValueOf(dest)
		if v.Kind() != reflect.Ptr {
			panic(fmt.Sprintf("%s is not a pointer (did you forget an ampersand?)", v.Type()))
		}
		v = v.Elem()
		if v.Kind() != reflect.Struct {
			panic(fmt.Sprintf("%T is not a struct pointer", dest))
		}

		var errs []string
		walkFields(v, func(field reflect.StructField, val reflect.Value, t reflect.Type) bool {
			// Check for the ignore switch in the tag
			tag := field.Tag.Get("arg")
			if tag == "-" {
				return false
			}

			// If this is an embedded struct then recurse into its fields
			if field.Anonymous && field.Type.Kind() == reflect.Struct {
				return true
			}

			spec := spec{
				long: strings.ToLower(field.Name),
				dest: val,
			}

			// Check whether this field is supported. It's good to do this here rather than
			// wait until setScalar because it means that a program with invalid argument
			// fields will always fail regardless of whether the arguments it received
			// exercised those fields.
			var parseable bool
			parseable, spec.boolean, spec.multiple = canParse(field.Type)
			if !parseable {
				errs = append(errs, fmt.Sprintf("%s.%s: %s fields are not supported",
					t.Name(), field.Name, field.Type.String()))
				return false
			}

			// Look at the tag
			if tag != "" {
				for _, key := range strings.Split(tag, ",") {
					var value string
					if pos := strings.Index(key, ":"); pos != -1 {
						value = key[pos+1:]
						key = key[:pos]
					}

					switch {
					case strings.HasPrefix(key, "---"):
						errs = append(errs, fmt.Sprintf("%s.%s: too many hyphens", t.Name(), field.Name))
					case strings.HasPrefix(key, "--"):
						spec.long = key[2:]
					case strings.HasPrefix(key, "-"):
						if len(key) != 2 {
							errs = append(errs, fmt.Sprintf("%s.%s: short arguments must be one character only",
								t.Name(), field.Name))
							return false
						}
						spec.short = key[1:]
					case key == "required":
						spec.required = true
					case key == "positional":
						spec.positional = true
					case key == "separate":
						spec.separate = true
					case key == "help":
						spec.help = value
					case key == "env":
						// Use override name if provided
						if value != "" {
							spec.env = value
						} else {
							spec.env = strings.ToUpper(field.Name)
						}
					default:
						errs = append(errs, fmt.Sprintf("unrecognized tag '%s' on field %s", key, tag))
						return false
					}
				}
			}
			p.spec = append(p.spec, &spec)

			// if this was an embedded field then we already returned true up above
			return false
		})

		if len(errs) > 0 {
			return nil, errors.New(strings.Join(errs, "\n"))
		}
	}
	if p.config.Program == "" {
		p.config.Program = "program"
		if len(os.Args) > 0 {
			p.config.Program = filepath.Base(os.Args[0])
		}
	}
	return &p, nil
}

// Parse processes the given command line option, storing the results in the field
// of the structs from which NewParser was constructed
func (p *Parser) Parse(args []string) error {
	// If -h or --help were specified then print usage
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			return ErrHelp
		}
		if arg == "--version" {
			return ErrVersion
		}
		if arg == "--" {
			break
		}
	}

	// Process all command line arguments
	err := process(p.spec, args)
	if err != nil {
		return err
	}

	// Validate
	return validate(p.spec)
}

// process goes through arguments one-by-one, parses them, and assigns the result to
// the underlying struct field
func process(specs []*spec, args []string) error {
	// construct a map from --option to spec
	optionMap := make(map[string]*spec)
	for _, spec := range specs {
		if spec.positional {
			continue
		}
		if spec.long != "" {
			optionMap[spec.long] = spec
		}
		if spec.short != "" {
			optionMap[spec.short] = spec
		}
		if spec.env != "" {
			if value, found := os.LookupEnv(spec.env); found {
				err := setScalar(spec.dest, value)
				if err != nil {
					return fmt.Errorf("error processing environment variable %s: %v", spec.env, err)
				}
				spec.wasPresent = true
			}
		}
	}

	// process each string from the command line
	var allpositional bool
	var positionals []string

	// must use explicit for loop, not range, because we manipulate i inside the loop
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			allpositional = true
			continue
		}

		if !isFlag(arg) || allpositional {
			positionals = append(positionals, arg)
			continue
		}

		// check for an equals sign, as in "--foo=bar"
		var value string
		opt := strings.TrimLeft(arg, "-")
		if pos := strings.Index(opt, "="); pos != -1 {
			value = opt[pos+1:]
			opt = opt[:pos]
		}

		// lookup the spec for this option
		spec, ok := optionMap[opt]
		if !ok {
			return fmt.Errorf("unknown argument %s", arg)
		}
		spec.wasPresent = true

		// deal with the case of multiple values
		if spec.multiple {
			var values []string
			if value == "" {
				for i+1 < len(args) && !isFlag(args[i+1]) {
					values = append(values, args[i+1])
					i++
					if spec.separate {
						break
					}
				}
			} else {
				values = append(values, value)
			}
			err := setSlice(spec.dest, values, !spec.separate)
			if err != nil {
				return fmt.Errorf("error processing %s: %v", arg, err)
			}
			continue
		}

		// if it's a flag and it has no value then set the value to true
		// use boolean because this takes account of TextUnmarshaler
		if spec.boolean && value == "" {
			value = "true"
		}

		// if we have something like "--foo" then the value is the next argument
		if value == "" {
			if i+1 == len(args) || isFlag(args[i+1]) {
				return fmt.Errorf("missing value for %s", arg)
			}
			value = args[i+1]
			i++
		}

		err := setScalar(spec.dest, value)
		if err != nil {
			return fmt.Errorf("error processing %s: %v", arg, err)
		}
	}

	// process positionals
	for _, spec := range specs {
		if spec.positional {
			if spec.multiple {
				if spec.required && len(positionals) == 0 {
					return fmt.Errorf("%s is required", spec.long)
				}
				err := setSlice(spec.dest, positionals, true)
				if err != nil {
					return fmt.Errorf("error processing %s: %v", spec.long, err)
				}
				positionals = nil
			} else if len(positionals) > 0 {
				err := setScalar(spec.dest, positionals[0])
				if err != nil {
					return fmt.Errorf("error processing %s: %v", spec.long, err)
				}
				positionals = positionals[1:]
			} else if spec.required {
				return fmt.Errorf("%s is required", spec.long)
			}
		}
	}
	if len(positionals) > 0 {
		return fmt.Errorf("too many positional arguments at '%s'", positionals[0])
	}
	return nil
}

// isFlag returns true if a token is a flag such as "-v" or "--user" but not "-" or "--"
func isFlag(s string) bool {
	return strings.HasPrefix(s, "-") && strings.TrimLeft(s, "-") != ""
}

// validate an argument spec after arguments have been parse
func validate(spec []*spec) error {
	for _, arg := range spec {
		if !arg.positional && arg.required && !arg.wasPresent {
			return fmt.Errorf("--%s is required", arg.long)
		}
	}
	return nil
}

// parse a value as the appropriate type and store it in the struct
func setSlice(dest reflect.Value, values []string, trunc bool) error {
	if !dest.CanSet() {
		return fmt.Errorf("field is not writable")
	}

	var ptr bool
	elem := dest.Type().Elem()
	if elem.Kind() == reflect.Ptr {
		ptr = true
		elem = elem.Elem()
	}

	// Truncate the dest slice in case default values exist
	if trunc && !dest.IsNil() {
		dest.SetLen(0)
	}

	for _, s := range values {
		v := reflect.New(elem)
		if err := setScalar(v.Elem(), s); err != nil {
			return err
		}
		if !ptr {
			v = v.Elem()
		}
		dest.Set(reflect.Append(dest, v))
	}
	return nil
}

// canParse returns true if the type can be parsed from a string
func canParse(t reflect.Type) (parseable, boolean, multiple bool) {
	parseable, boolean = isScalar(t)
	if parseable {
		return
	}

	// Look inside pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// Look inside slice types
	if t.Kind() == reflect.Slice {
		multiple = true
		t = t.Elem()
	}

	parseable, boolean = isScalar(t)
	if parseable {
		return
	}

	// Look inside pointer types (again, in case of []*Type)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	parseable, boolean = isScalar(t)
	if parseable {
		return
	}

	return false, false, false
}

var textUnmarshalerType = reflect.TypeOf([]encoding.TextUnmarshaler{}).Elem()

// isScalar returns true if the type can be parsed from a single string
func isScalar(t reflect.Type) (parseable, boolean bool) {
	parseable = scalar.CanParse(t)
	switch {
	case t.Implements(textUnmarshalerType):
		return parseable, false
	case t.Kind() == reflect.Bool:
		return parseable, true
	case t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Bool:
		return parseable, true
	default:
		return parseable, false
	}
}

// set a value from a string
func setScalar(v reflect.Value, s string) error {
	return scalar.ParseValue(v, s)
}
