package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/google/shlex"
	"gopkg.in/alecthomas/kingpin.v3-unstable"
)

// Severity of linter message.
type Severity string

// Linter message severity levels.
const ( // nolint
	Warning Severity = "warning"
	Error   Severity = "error"
)

var (
	// Locations to look for vendored linters.
	vendoredSearchPaths = [][]string{
		{"github.com", "alecthomas", "gometalinter", "_linters"},
		{"gopkg.in", "alecthomas", "gometalinter.v1", "_linters"},
	}
)

type Linter struct {
	Name             string   `json:"name"`
	Command          string   `json:"command"`
	CompositeCommand string   `json:"composite_command,omitempty"`
	Pattern          string   `json:"pattern"`
	InstallFrom      string   `json:"install_from"`
	SeverityOverride Severity `json:"severity,omitempty"`
	MessageOverride  string   `json:"message_override,omitempty"`

	regex *regexp.Regexp
}

func (l *Linter) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.Name)
}

func (l *Linter) String() string {
	return l.Name
}

func LinterFromName(name string) *Linter {
	s := linterDefinitions[name]
	parts := strings.SplitN(s, ":", 2)
	if len(parts) < 2 {
		kingpin.Fatalf("invalid linter: %q", name)
	}

	pattern := parts[1]
	if p, ok := predefinedPatterns[pattern]; ok {
		pattern = p
	}
	re, err := regexp.Compile("(?m:" + pattern + ")")
	kingpin.FatalIfError(err, "invalid regex for %q", name)
	return &Linter{
		Name:             name,
		Command:          s[0:strings.Index(s, ":")],
		Pattern:          pattern,
		InstallFrom:      installMap[name],
		SeverityOverride: Severity(config.Severity[name]),
		MessageOverride:  config.MessageOverride[name],
		regex:            re,
	}
}

type sortedIssues struct {
	issues []*Issue
	order  []string
}

func (s *sortedIssues) Len() int      { return len(s.issues) }
func (s *sortedIssues) Swap(i, j int) { s.issues[i], s.issues[j] = s.issues[j], s.issues[i] }

// nolint: gocyclo
func (s *sortedIssues) Less(i, j int) bool {
	l, r := s.issues[i], s.issues[j]
	for _, key := range s.order {
		switch key {
		case "path":
			if l.Path > r.Path {
				return false
			}
		case "line":
			if l.Line > r.Line {
				return false
			}
		case "column":
			if l.Col > r.Col {
				return false
			}
		case "severity":
			if l.Severity > r.Severity {
				return false
			}
		case "message":
			if l.Message > r.Message {
				return false
			}
		case "linter":
			if l.Linter.Name > r.Linter.Name {
				return false
			}
		}
	}
	return true
}

func init() {
	kingpin.Flag("config", "Load JSON configuration from file.").Action(loadConfig).String()
	kingpin.Flag("disable", "Disable previously enabled linters.").PlaceHolder("LINTER").Short('D').Action(disableAction).Strings()
	kingpin.Flag("enable", "Enable previously disabled linters.").PlaceHolder("LINTER").Short('E').Action(enableAction).Strings()
	kingpin.Flag("linter", "Define a linter.").PlaceHolder("NAME:COMMAND:PATTERN").StringMapVar(&config.Linters)
	kingpin.Flag("message-overrides", "Override message from linter. {message} will be expanded to the original message.").PlaceHolder("LINTER:MESSAGE").StringMapVar(&config.MessageOverride)
	kingpin.Flag("severity", "Map of linter severities.").PlaceHolder("LINTER:SEVERITY").StringMapVar(&config.Severity)
	kingpin.Flag("disable-all", "Disable all linters.").Action(disableAllAction).Bool()
	kingpin.Flag("enable-all", "Enable all linters.").Action(enableAllAction).Bool()
	kingpin.Flag("format", "Output format.").PlaceHolder(config.Format).StringVar(&config.Format)
	kingpin.Flag("vendored-linters", "Use vendored linters (recommended).").BoolVar(&config.VendoredLinters)
	kingpin.Flag("fast", "Only run fast linters.").BoolVar(&config.Fast)
	kingpin.Flag("install", "Attempt to install all known linters.").Short('i').BoolVar(&config.Install)
	kingpin.Flag("update", "Pass -u to go tool when installing.").Short('u').BoolVar(&config.Update)
	kingpin.Flag("force", "Pass -f to go tool when installing.").Short('f').BoolVar(&config.Force)
	kingpin.Flag("download-only", "Pass -d to go tool when installing.").BoolVar(&config.DownloadOnly)
	kingpin.Flag("debug", "Display messages for failed linters, etc.").Short('d').BoolVar(&config.Debug)
	kingpin.Flag("concurrency", "Number of concurrent linters to run.").PlaceHolder(fmt.Sprintf("%d", runtime.NumCPU())).Short('j').IntVar(&config.Concurrency)
	kingpin.Flag("exclude", "Exclude messages matching these regular expressions.").Short('e').PlaceHolder("REGEXP").StringsVar(&config.Exclude)
	kingpin.Flag("include", "Include messages matching these regular expressions.").Short('I').PlaceHolder("REGEXP").StringsVar(&config.Include)
	kingpin.Flag("skip", "Skip directories with this name when expanding '...'.").Short('s').PlaceHolder("DIR...").StringsVar(&config.Skip)
	kingpin.Flag("vendor", "Enable vendoring support (skips 'vendor' directories and sets GO15VENDOREXPERIMENT=1).").BoolVar(&config.Vendor)
	kingpin.Flag("cyclo-over", "Report functions with cyclomatic complexity over N (using gocyclo).").PlaceHolder("10").IntVar(&config.Cyclo)
	kingpin.Flag("line-length", "Report lines longer than N (using lll).").PlaceHolder("80").IntVar(&config.LineLength)
	kingpin.Flag("min-confidence", "Minimum confidence interval to pass to golint.").PlaceHolder(".80").FloatVar(&config.MinConfidence)
	kingpin.Flag("min-occurrences", "Minimum occurrences to pass to goconst.").PlaceHolder("3").IntVar(&config.MinOccurrences)
	kingpin.Flag("min-const-length", "Minimumum constant length.").PlaceHolder("3").IntVar(&config.MinConstLength)
	kingpin.Flag("dupl-threshold", "Minimum token sequence as a clone for dupl.").PlaceHolder("50").IntVar(&config.DuplThreshold)
	kingpin.Flag("sort", fmt.Sprintf("Sort output by any of %s.", strings.Join(sortKeys, ", "))).PlaceHolder("none").EnumsVar(&config.Sort, sortKeys...)
	kingpin.Flag("tests", "Include test files for linters that support this option").Short('t').BoolVar(&config.Test)
	kingpin.Flag("deadline", "Cancel linters if they have not completed within this duration.").PlaceHolder("30s").DurationVar(&config.Deadline)
	kingpin.Flag("errors", "Only show errors.").BoolVar(&config.Errors)
	kingpin.Flag("json", "Generate structured JSON rather than standard line-based output.").BoolVar(&config.JSON)
	kingpin.Flag("checkstyle", "Generate checkstyle XML rather than standard line-based output.").BoolVar(&config.Checkstyle)
	kingpin.Flag("enable-gc", "Enable GC for linters (useful on large repositories).").BoolVar(&config.EnableGC)
	kingpin.Flag("aggregate", "Aggregate issues reported by several linters.").BoolVar(&config.Aggregate)
	kingpin.CommandLine.GetFlag("help").Short('h')
}

func loadConfig(app *kingpin.Application, element *kingpin.ParseElement, ctx *kingpin.ParseContext) error {
	r, err := os.Open(*element.Value)
	if err != nil {
		return err
	}
	defer r.Close() // nolint: errcheck
	err = json.NewDecoder(r).Decode(config)
	if err != nil {
		return err
	}
	if config.DeadlineJSONCrutch != "" {
		config.Deadline, err = time.ParseDuration(config.DeadlineJSONCrutch)
	}
	for _, disable := range config.Disable {
		for i, enable := range config.Enable {
			if enable == disable {
				config.Enable = append(config.Enable[:i], config.Enable[i+1:]...)
				break
			}
		}
	}
	return err
}

func disableAction(app *kingpin.Application, element *kingpin.ParseElement, ctx *kingpin.ParseContext) error {
	out := []string{}
	for _, linter := range config.Enable {
		if linter != *element.Value {
			out = append(out, linter)
		}
	}
	config.Enable = out
	return nil
}

func enableAction(app *kingpin.Application, element *kingpin.ParseElement, ctx *kingpin.ParseContext) error {
	config.Enable = append(config.Enable, *element.Value)
	return nil
}

func disableAllAction(app *kingpin.Application, element *kingpin.ParseElement, ctx *kingpin.ParseContext) error {
	config.Enable = []string{}
	return nil
}

func enableAllAction(app *kingpin.Application, element *kingpin.ParseElement, ctx *kingpin.ParseContext) error {
	for linter := range linterDefinitions {
		config.Enable = append(config.Enable, linter)
	}
	return nil
}

type Issue struct {
	Linter   *Linter  `json:"linter"`
	Severity Severity `json:"severity"`
	Path     string   `json:"path"`
	Line     int      `json:"line"`
	Col      int      `json:"col"`
	Message  string   `json:"message"`
}

func (i *Issue) String() string {
	buf := new(bytes.Buffer)
	err := formatTemplate.Execute(buf, i)
	kingpin.FatalIfError(err, "Invalid output format")
	return buf.String()
}

func debug(format string, args ...interface{}) {
	if config.Debug {
		fmt.Fprintf(os.Stderr, "DEBUG: "+format+"\n", args...)
	}
}

func warning(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "WARNING: "+format+"\n", args...)
}

func formatLinters() string {
	w := bytes.NewBuffer(nil)
	for name := range linterDefinitions {
		linter := LinterFromName(name)
		install := "(" + linter.InstallFrom + ")"
		if install == "()" {
			install = ""
		}
		fmt.Fprintf(w, "  %s  %s\n        %s\n        %s\n", name, install, linter.Command, linter.Pattern)
	}
	return w.String()
}

func formatSeverity() string {
	w := bytes.NewBuffer(nil)
	for name, severity := range config.Severity {
		fmt.Fprintf(w, "  %s -> %s\n", name, severity)
	}
	return w.String()
}

type Vars map[string]string

func (v Vars) Copy() Vars {
	out := Vars{}
	for k, v := range v {
		out[k] = v
	}
	return out
}

func (v Vars) Replace(s string) string {
	for k, v := range v {
		prefix := regexp.MustCompile(fmt.Sprintf("{%s=([^}]*)}", k))
		if v != "" {
			s = prefix.ReplaceAllString(s, "$1")
		} else {
			s = prefix.ReplaceAllString(s, "")
		}
		s = strings.Replace(s, fmt.Sprintf("{%s}", k), v, -1)
	}
	return s
}

func main() {
	// Linters are by their very nature, short lived, so disable GC.
	// Reduced (user) linting time on kingpin from 0.97s to 0.64s.
	kingpin.CommandLine.Help = fmt.Sprintf(`Aggregate and normalise the output of a whole bunch of Go linters.

PlaceHolder linters:

%s

Severity override map (default is "warning"):

%s
`, formatLinters(), formatSeverity())
	kingpin.Parse()

	configureEnvironment()

	if config.Install {
		installLinters()
		return
	}

	include, exclude := processConfig(config)

	start := time.Now()
	paths := expandPaths(*pathsArg, config.Skip)

	linters := lintersFromFlags()
	status := 0
	issues, errch := runLinters(linters, paths, *pathsArg, config.Concurrency, exclude, include)
	if config.JSON {
		status |= outputToJSON(issues)
	} else if config.Checkstyle {
		status |= outputToCheckstyle(issues)
	} else {
		status |= outputToConsole(issues)
	}
	for err := range errch {
		warning("%s", err)
		status |= 2
	}
	elapsed := time.Since(start)
	debug("total elapsed time %s", elapsed)
	os.Exit(status)
}

// nolint: gocyclo
func processConfig(config *Config) (include *regexp.Regexp, exclude *regexp.Regexp) {
	// Move configured linters into linterDefinitions.
	for name, definition := range config.Linters {
		linterDefinitions[name] = definition
	}

	tmpl, err := template.New("output").Parse(config.Format)
	kingpin.FatalIfError(err, "invalid format %q", config.Format)
	formatTemplate = tmpl
	if !config.EnableGC {
		_ = os.Setenv("GOGC", "off")
	}
	if config.VendoredLinters && config.Install && config.Update {
		warning(`Linters are now vendored by default, --update ignored. The original
behaviour can be re-enabled with --no-vendored-linters.

To request an update for a vendored linter file an issue at:
https://github.com/alecthomas/gometalinter/issues/new
`)
		config.Update = false
	}
	// Force sorting by path if checkstyle mode is selected
	// !jsonFlag check is required to handle:
	// 	gometalinter --json --checkstyle --sort=severity
	if config.Checkstyle && !config.JSON {
		config.Sort = []string{"path"}
	}

	// PlaceHolder to skipping "vendor" directory if GO15VENDOREXPERIMENT=1 is enabled.
	// TODO(alec): This will probably need to be enabled by default at a later time.
	if os.Getenv("GO15VENDOREXPERIMENT") == "1" || config.Vendor {
		if err := os.Setenv("GO15VENDOREXPERIMENT", "1"); err != nil {
			warning("setenv GO15VENDOREXPERIMENT: %s", err)
		}
		config.Skip = append(config.Skip, "vendor")
		config.Vendor = true
	}
	if len(config.Exclude) > 0 {
		exclude = regexp.MustCompile(strings.Join(config.Exclude, "|"))
	}

	if len(config.Include) > 0 {
		include = regexp.MustCompile(strings.Join(config.Include, "|"))
	}

	runtime.GOMAXPROCS(config.Concurrency)
	return include, exclude
}

func outputToConsole(issues chan *Issue) int {
	status := 0
	for issue := range issues {
		if config.Errors && issue.Severity != Error {
			continue
		}
		fmt.Println(issue.String())
		status = 1
	}
	return status
}

func outputToJSON(issues chan *Issue) int {
	fmt.Println("[")
	status := 0
	for issue := range issues {
		if config.Errors && issue.Severity != Error {
			continue
		}
		if status != 0 {
			fmt.Printf(",\n")
		}
		d, err := json.Marshal(issue)
		kingpin.FatalIfError(err, "")
		fmt.Printf("  %s", d)
		status = 1
	}
	fmt.Printf("\n]\n")
	return status
}

func runLinters(linters map[string]*Linter, paths, ellipsisPaths []string, concurrency int, exclude *regexp.Regexp, include *regexp.Regexp) (chan *Issue, chan error) {
	errch := make(chan error, len(linters)*(len(paths)+len(ellipsisPaths)))
	concurrencych := make(chan bool, config.Concurrency)
	incomingIssues := make(chan *Issue, 1000000)
	directives := newDirectiveParser(paths)
	processedIssues := filterIssuesViaDirectives(directives, maybeSortIssues(maybeAggregateIssues(incomingIssues)))
	wg := &sync.WaitGroup{}
	for _, linter := range linters {
		// Recreated in each loop because it is mutated by executeLinter().
		vars := Vars{
			"duplthreshold":    fmt.Sprintf("%d", config.DuplThreshold),
			"mincyclo":         fmt.Sprintf("%d", config.Cyclo),
			"maxlinelength":    fmt.Sprintf("%d", config.LineLength),
			"min_confidence":   fmt.Sprintf("%f", config.MinConfidence),
			"min_occurrences":  fmt.Sprintf("%d", config.MinOccurrences),
			"min_const_length": fmt.Sprintf("%d", config.MinConstLength),
			"tests":            "",
		}
		if config.Test {
			vars["tests"] = "-t"
		}
		linterPaths := paths
		// Most linters don't exclude vendor paths when recursing, so we don't use ... paths.
		if acceptsEllipsis[linter.Name] && !config.Vendor && len(ellipsisPaths) > 0 {
			linterPaths = ellipsisPaths
		}
		for _, path := range linterPaths {
			wg.Add(1)
			deadline := time.After(config.Deadline)
			state := &linterState{
				Linter:   linter,
				issues:   incomingIssues,
				path:     path,
				vars:     vars.Copy(),
				exclude:  exclude,
				include:  include,
				deadline: deadline,
			}
			go func() {
				concurrencych <- true
				err := executeLinter(state)
				if err != nil {
					errch <- err
				}
				<-concurrencych
				wg.Done()
			}()
		}
	}

	go func() {
		wg.Wait()
		close(incomingIssues)
		close(errch)
	}()
	return processedIssues, errch
}

// nolint: gocyclo
func expandPaths(paths, skip []string) []string {
	if len(paths) == 0 {
		paths = []string{"."}
	}
	skipMap := map[string]bool{}
	for _, name := range skip {
		skipMap[name] = true
	}
	dirs := map[string]bool{}
	for _, path := range paths {
		if strings.HasSuffix(path, "/...") {
			root := filepath.Dir(path)
			_ = filepath.Walk(root, func(p string, i os.FileInfo, err error) error {
				if err != nil {
					warning("invalid path %q: %s", p, err)
					return err
				}

				base := filepath.Base(p)
				skip := skipMap[base] || skipMap[p] || (strings.ContainsAny(base[0:1], "_.") && base != "." && base != "..")
				if i.IsDir() {
					if skip {
						return filepath.SkipDir
					}
				} else if !skip && strings.HasSuffix(p, ".go") {
					dirs[filepath.Clean(filepath.Dir(p))] = true
				}
				return nil
			})
		} else {
			dirs[filepath.Clean(path)] = true
		}
	}
	out := make([]string, 0, len(dirs))
	for d := range dirs {
		out = append(out, d)
	}
	sort.Strings(out)
	for _, d := range out {
		debug("linting path %s", d)
	}
	return out
}

func makeInstallCommand(linters ...string) []string {
	cmd := []string{"get"}
	if config.VendoredLinters {
		cmd = []string{"install"}
	} else {
		if config.Update {
			cmd = append(cmd, "-u")
		}
		if config.Force {
			cmd = append(cmd, "-f")
		}
		if config.DownloadOnly {
			cmd = append(cmd, "-d")
		}
	}
	if config.Debug {
		cmd = append(cmd, "-v")
	}
	cmd = append(cmd, linters...)
	return cmd
}

func installLintersWithOneCommand(targets []string) error {
	cmd := makeInstallCommand(targets...)
	debug("go %s", strings.Join(cmd, " "))
	c := exec.Command("go", cmd...) // nolint: gas
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func installLintersIndividually(targets []string) {
	failed := []string{}
	for _, target := range targets {
		cmd := makeInstallCommand(target)
		debug("go %s", strings.Join(cmd, " "))
		c := exec.Command("go", cmd...) // nolint: gas
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			warning("failed to install %s: %s", target, err)
			failed = append(failed, target)
		}
	}
	if len(failed) > 0 {
		kingpin.Fatalf("failed to install the following linters: %s", strings.Join(failed, ", "))
	}
}

func installLinters() {
	names := make([]string, 0, len(installMap))
	targets := make([]string, 0, len(installMap))
	for name, target := range installMap {
		names = append(names, name)
		targets = append(targets, target)
	}
	namesStr := strings.Join(names, "\n  ")
	if config.DownloadOnly {
		fmt.Printf("Downloading:\n  %s\n", namesStr)
	} else {
		fmt.Printf("Installing:\n  %s\n", namesStr)
	}
	err := installLintersWithOneCommand(targets)
	if err == nil {
		return
	}
	warning("failed to install one or more linters: %s (installing individually)", err)
	installLintersIndividually(targets)
}

func maybeAggregateIssues(issues chan *Issue) chan *Issue {
	if !config.Aggregate {
		return issues
	}
	return aggregateIssues(issues)
}

func maybeSortIssues(issues chan *Issue) chan *Issue {
	if reflect.DeepEqual([]string{"none"}, config.Sort) {
		return issues
	}
	out := make(chan *Issue, 1000000)
	sorted := &sortedIssues{
		issues: []*Issue{},
		order:  config.Sort,
	}
	go func() {
		for issue := range issues {
			sorted.issues = append(sorted.issues, issue)
		}
		sort.Sort(sorted)
		for _, issue := range sorted.issues {
			out <- issue
		}
		close(out)
	}()
	return out
}

type linterState struct {
	*Linter
	path     string
	issues   chan *Issue
	vars     Vars
	exclude  *regexp.Regexp
	include  *regexp.Regexp
	deadline <-chan time.Time
}

func (l *linterState) InterpolatedCommand() string {
	vars := l.vars.Copy()
	if l.ShouldChdir() {
		vars["path"] = "."
	} else {
		vars["path"] = l.path
	}
	return vars.Replace(l.Command)
}

func (l *linterState) ShouldChdir() bool {
	return config.Vendor || !strings.HasSuffix(l.path, "/...") || !strings.Contains(l.Command, "{path}")
}

func parseCommand(dir, command string) (string, []string, error) {
	args, err := shlex.Split(command)
	if err != nil {
		return "", nil, err
	}
	if len(args) == 0 {
		return "", nil, fmt.Errorf("invalid command %q", command)
	}
	exe, err := exec.LookPath(args[0])
	if err != nil {
		return "", nil, err
	}
	out := []string{}
	for _, arg := range args[1:] {
		if strings.Contains(arg, "*") {
			pattern := filepath.Join(dir, arg)
			globbed, err := filepath.Glob(pattern)
			if err != nil {
				return "", nil, err
			}
			for i, g := range globbed {
				if strings.HasPrefix(g, dir+string(filepath.Separator)) {
					globbed[i] = g[len(dir)+1:]
				}
			}
			out = append(out, globbed...)
		} else {
			out = append(out, arg)
		}
	}
	return exe, out, nil
}

func executeLinter(state *linterState) error {
	debug("linting with %s: %s (on %s)", state.Name, state.Command, state.path)

	start := time.Now()
	command := state.InterpolatedCommand()
	exe, args, err := parseCommand(state.path, command)
	if err != nil {
		return err
	}
	debug("executing %s %q", exe, args)
	buf := bytes.NewBuffer(nil)
	cmd := exec.Command(exe, args...) // nolint: gas
	if state.ShouldChdir() {
		cmd.Dir = state.path
	}
	cmd.Stdout = buf
	cmd.Stderr = buf
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to execute linter %s: %s", command, err)
	}

	done := make(chan bool)
	go func() {
		err = cmd.Wait()
		done <- true
	}()

	// Wait for process to complete or deadline to expire.
	select {
	case <-done:

	case <-state.deadline:
		err = fmt.Errorf("deadline exceeded by linter %s on %s (try increasing --deadline)",
			state.Name, state.path)
		kerr := cmd.Process.Kill()
		if kerr != nil {
			warning("failed to kill %s: %s", state.Name, kerr)
		}
		return err
	}

	if err != nil {
		debug("warning: %s returned %s", command, err)
	}

	processOutput(state, buf.Bytes())
	elapsed := time.Since(start)
	debug("%s linter took %s", state.Name, elapsed)
	return nil
}

func (l *linterState) fixPath(path string) string {
	lpath := strings.TrimSuffix(l.path, "...")
	labspath, _ := filepath.Abs(lpath)

	if !l.ShouldChdir() {
		path = strings.TrimPrefix(path, lpath)
	}

	if !filepath.IsAbs(path) {
		path, _ = filepath.Abs(filepath.Join(labspath, path))
	}
	if strings.HasPrefix(path, labspath) {
		return filepath.Join(lpath, strings.TrimPrefix(path, labspath))
	}
	return path
}

func lintersFromFlags() map[string]*Linter {
	out := map[string]*Linter{}
	for _, linter := range config.Enable {
		out[linter] = LinterFromName(linter)
	}
	for _, linter := range config.Disable {
		delete(out, linter)
	}
	if config.Fast {
		for _, linter := range slowLinters {
			delete(out, linter)
		}
	}
	return out
}

// nolint: gocyclo
func processOutput(state *linterState, out []byte) {
	re := state.regex
	all := re.FindAllSubmatchIndex(out, -1)
	debug("%s hits %d: %s", state.Name, len(all), state.Pattern)
	for _, indices := range all {
		group := [][]byte{}
		for i := 0; i < len(indices); i += 2 {
			var fragment []byte
			if indices[i] != -1 {
				fragment = out[indices[i]:indices[i+1]]
			}
			group = append(group, fragment)
		}

		issue := &Issue{Line: 1}
		issue.Linter = LinterFromName(state.Name)
		for i, name := range re.SubexpNames() {
			if group[i] == nil {
				continue
			}
			part := string(group[i])
			if name != "" {
				state.vars[name] = part
			}
			switch name {
			case "path":
				issue.Path = state.fixPath(part)

			case "line":
				n, err := strconv.ParseInt(part, 10, 32)
				kingpin.FatalIfError(err, "line matched invalid integer")
				issue.Line = int(n)

			case "col":
				n, err := strconv.ParseInt(part, 10, 32)
				kingpin.FatalIfError(err, "col matched invalid integer")
				issue.Col = int(n)

			case "message":
				issue.Message = part

			case "":
			}
		}
		if m, ok := config.MessageOverride[state.Name]; ok {
			issue.Message = state.vars.Replace(m)
		}
		if sev, ok := config.Severity[state.Name]; ok {
			issue.Severity = Severity(sev)
		} else {
			issue.Severity = "warning"
		}
		if state.exclude != nil && state.exclude.MatchString(issue.String()) {
			continue
		}
		if state.include != nil && !state.include.MatchString(issue.String()) {
			continue
		}
		state.issues <- issue
	}
	return
}

func findVendoredLinters() string {
	gopaths := strings.Split(getGoPath(), string(os.PathListSeparator))
	for _, home := range vendoredSearchPaths {
		for _, p := range gopaths {
			joined := append([]string{p, "src"}, home...)
			vendorRoot := filepath.Join(joined...)
			if _, err := os.Stat(vendorRoot); err == nil {
				return vendorRoot
			}
		}
	}
	return ""

}

// Go 1.8 compatible GOPATH.
func getGoPath() string {
	path := os.Getenv("GOPATH")
	if path == "" {
		user, err := user.Current()
		kingpin.FatalIfError(err, "")
		path = filepath.Join(user.HomeDir, "go")
	}
	return path
}

// addPath appends p to paths and returns it if:
// 1. p is not a blank string
// 2. p doesn't already exist in paths
// Otherwise paths is returned unchanged.
func addPath(p string, paths []string) []string {
	if p == "" {
		return paths
	}
	for _, path := range paths {
		if p == path {
			return paths
		}
	}
	return append(paths, p)
}

// Ensure all "bin" directories from GOPATH exists in PATH, as well as GOBIN if set.
func configureEnvironment() {
	gopaths := strings.Split(getGoPath(), string(os.PathListSeparator))
	paths := strings.Split(os.Getenv("PATH"), string(os.PathListSeparator))
	gobin := os.Getenv("GOBIN")

	if config.VendoredLinters && config.Install {
		vendorRoot := findVendoredLinters()
		if vendorRoot == "" {
			kingpin.Fatalf("could not find vendored linters in GOPATH=%q", getGoPath())
		}
		debug("found vendored linters at %s, updating environment", vendorRoot)
		if gobin == "" {
			gobin = filepath.Join(gopaths[0], "bin")
		}
		// "go install" panics when one GOPATH element is beneath another, so we just set
		// our vendor root instead.
		gopaths = []string{vendorRoot}
	}

	for _, p := range gopaths {
		paths = addPath(filepath.Join(p, "bin"), paths)
	}
	paths = addPath(gobin, paths)

	path := strings.Join(paths, string(os.PathListSeparator))
	gopath := strings.Join(gopaths, string(os.PathListSeparator))

	if err := os.Setenv("PATH", path); err != nil {
		warning("setenv PATH: %s", err)
	}
	debug("PATH=%s", os.Getenv("PATH"))

	if err := os.Setenv("GOPATH", gopath); err != nil {
		warning("setenv GOPATH: %s", err)
	}
	debug("GOPATH=%s", os.Getenv("GOPATH"))

	if err := os.Setenv("GOBIN", gobin); err != nil {
		warning("setenv GOBIN: %s", err)
	}
	debug("GOBIN=%s", os.Getenv("GOBIN"))
}
