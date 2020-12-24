// +build mage
// Self-contained go-project magefile.

// nolint: deadcode
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/magefile/mage/target"

	"errors"
	"math/bits"
	"strconv"

	"github.com/mholt/archiver"
)

var curDir = func() string {
	name, _ := os.Getwd()
	return name
}()

const constCoverageDir = ".coverage"
const constToolDir = "tools"
const constBinDir = "bin"
const constReleaseDir = "release"
const constCmdDir = "cmd"
const constCoverFile = "cover.out"
const constAssets = "assets"
const constAssetsGenerated = "assets/generated"

var coverageDir = mustStr(filepath.Abs(path.Join(curDir, constCoverageDir)))
var toolDir = mustStr(filepath.Abs(path.Join(curDir, constToolDir)))
var binDir = mustStr(filepath.Abs(path.Join(curDir, constBinDir)))
var releaseDir = mustStr(filepath.Abs(path.Join(curDir, constReleaseDir)))
var cmdDir = mustStr(filepath.Abs(path.Join(curDir, constCmdDir)))
var assetsGenerated = mustStr(filepath.Abs(path.Join(curDir, constAssetsGenerated)))

// Calculate file paths
var toolsGoPath = toolDir
var toolsSrcDir = mustStr(filepath.Abs(path.Join(toolDir, "src")))
var toolsBinDir = mustStr(filepath.Abs(path.Join(toolDir, "bin")))
var toolsVendorDir = mustStr(filepath.Abs(path.Join(toolDir, "vendor")))

var outputDirs = []string{binDir, releaseDir, toolsGoPath, toolsBinDir,
	toolsVendorDir, assetsGenerated, coverageDir}

var toolsEnv = map[string]string{"GOPATH": toolsGoPath}

var containerName = func() string {
	if name := os.Getenv("CONTAINER_NAME"); name != "" {
		return name
	}
	return "wrouesnel/postgres_exporter:latest"
}()

type Platform struct {
	OS        string
	Arch      string
	BinSuffix string
}

func (p *Platform) String() string {
	return fmt.Sprintf("%s-%s", p.OS, p.Arch)
}

func (p *Platform) PlatformDir() string {
	platformDir := path.Join(binDir, fmt.Sprintf("%s_%s_%s", productName, versionShort, p.String()))
	return platformDir
}

func (p *Platform) PlatformBin(cmd string) string {
	platformBin := fmt.Sprintf("%s%s", cmd, p.BinSuffix)
	return path.Join(p.PlatformDir(), platformBin)
}

func (p *Platform) ArchiveDir() string {
	return fmt.Sprintf("%s_%s_%s", productName, versionShort, p.String())
}

func (p *Platform) ReleaseBase() string {
	return path.Join(releaseDir, fmt.Sprintf("%s_%s_%s", productName, versionShort, p.String()))
}

// Supported platforms
var platforms []Platform = []Platform{
	{"linux", "amd64", ""},
	{"linux", "386", ""},
	{"linux", "arm64", ""},
	{"linux", "mips64le", ""},
	{"darwin", "amd64", ""},
	{"darwin", "386", ""},
	{"windows", "amd64", ".exe"},
	{"windows", "386", ".exe"},
	{"freebsd", "amd64", ""},
}

// productName can be overridden by environ product name
var productName = func() string {
	if name := os.Getenv("PRODUCT_NAME"); name != "" {
		return name
	}
	name, _ := os.Getwd()
	return path.Base(name)
}()

// Source files
var goSrc []string
var goDirs []string
var goPkgs []string
var goCmds []string

var branch = func() string {
	if v := os.Getenv("BRANCH"); v != "" {
		return v
	}
	out, _ := sh.Output("git", "rev-parse", "--abbrev-ref", "HEAD")

	return out
}()

var buildDate = func() string {
	if v := os.Getenv("BUILDDATE"); v != "" {
		return v
	}
	return time.Now().Format("2006-01-02T15:04:05-0700")
}()

var revision = func() string {
	if v := os.Getenv("REVISION"); v != "" {
		return v
	}
	out, _ := sh.Output("git", "rev-parse", "HEAD")

	return out
}()

var version = func() string {
	if v := os.Getenv("VERSION"); v != "" {
		return v
	}
	out, _ := sh.Output("git", "describe", "--dirty")

	if out == "" {
		return "v0.0.0"
	}

	return out
}()

var versionShort = func() string {
	if v := os.Getenv("VERSION_SHORT"); v != "" {
		return v
	}
	out, _ := sh.Output("git", "describe", "--abbrev=0")

	if out == "" {
		return "v0.0.0"
	}

	return out
}()

var concurrency = func() int {
	if v := os.Getenv("CONCURRENCY"); v != "" {
		pv, err := strconv.ParseUint(v, 10, bits.UintSize)
		if err != nil {
			panic(err)
		}
		return int(pv)
	}
	return runtime.NumCPU()
}()

var linterDeadline = func() time.Duration {
	if v := os.Getenv("LINTER_DEADLINE"); v != "" {
		d, _ := time.ParseDuration(v)
		if d != 0 {
			return d
		}
	}
	return time.Second * 60
}()

func Log(args ...interface{}) {
	if mg.Verbose() {
		fmt.Println(args...)
	}
}

func init() {
	// Set environment
	os.Setenv("PATH", fmt.Sprintf("%s:%s", toolsBinDir, os.Getenv("PATH")))
	Log("Build PATH: ", os.Getenv("PATH"))
	Log("Concurrency:", concurrency)
	goSrc = func() []string {
		results := new([]string)
		filepath.Walk(".", func(relpath string, info os.FileInfo, err error) error {
			// Ensure absolute path so globs work
			path, err := filepath.Abs(relpath)
			if err != nil {
				panic(err)
			}

			// Look for files
			if info.IsDir() {
				return nil
			}

			// Exclusions
			for _, exclusion := range []string{toolDir, binDir, releaseDir, coverageDir} {
				if strings.HasPrefix(path, exclusion) {
					if info.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
			}

			if strings.Contains(path, "/vendor/") {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			if strings.Contains(path, ".git") {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			if !strings.HasSuffix(path, ".go") {
				return nil
			}

			*results = append(*results, path)
			return nil
		})
		return *results
	}()
	goDirs = func() []string {
		resultMap := make(map[string]struct{})
		for _, path := range goSrc {
			absDir, err := filepath.Abs(filepath.Dir(path))
			if err != nil {
				panic(err)
			}
			resultMap[absDir] = struct{}{}
		}
		results := []string{}
		for k := range resultMap {
			results = append(results, k)
		}
		return results
	}()
	goPkgs = func() []string {
		results := []string{}
		out, err := sh.Output("go", "list", "./...")
		if err != nil {
			panic(err)
		}
		for _, line := range strings.Split(out, "\n") {
			if !strings.Contains(line, "/vendor/") {
				results = append(results, line)
			}
		}
		return results
	}()
	goCmds = func() []string {
		results := []string{}

		finfos, err := ioutil.ReadDir(cmdDir)
		if err != nil {
			panic(err)
		}
		for _, finfo := range finfos {
			results = append(results, finfo.Name())
		}
		return results
	}()

	// Ensure output dirs exist
	for _, dir := range outputDirs {
		os.MkdirAll(dir, os.FileMode(0777))
	}
}

func mustStr(r string, err error) string {
	if err != nil {
		panic(err)
	}
	return r
}

func getCoreTools() []string {
	staticTools := []string{
		"github.com/kardianos/govendor",
		"github.com/wadey/gocovmerge",
		"github.com/mattn/goveralls",
		"github.com/tmthrgd/go-bindata/go-bindata",
		"github.com/GoASTScanner/gas/cmd/gas", // workaround for Ast scanner
		"github.com/alecthomas/gometalinter",
	}
	return staticTools
}

func getMetalinters() []string {
	// Gometalinter should now be on the command line
	dynamicTools := []string{}

	goMetalinterHelp, _ := sh.Output("gometalinter", "--help")
	linterRx := regexp.MustCompile(`\s+\w+:\s*\((.+)\)`)
	for _, l := range strings.Split(goMetalinterHelp, "\n") {
		linter := linterRx.FindStringSubmatch(l)
		if len(linter) > 1 {
			dynamicTools = append(dynamicTools, linter[1])
		}
	}
	return dynamicTools
}

func ensureVendorSrcLink() error {
	Log("Symlink vendor to tools dir")
	if err := sh.Rm(toolsSrcDir); err != nil {
		return err
	}
	if err := os.Symlink(toolsVendorDir, toolsSrcDir); err != nil {
		return err
	}
	return nil
}

// concurrencyLimitedBuild executes a certain number of commands limited by concurrency
func concurrencyLimitedBuild(buildCmds ...interface{}) error {
	resultsCh := make(chan error, len(buildCmds))
	concurrencyControl := make(chan struct{}, concurrency)
	for _, buildCmd := range buildCmds {
		go func(buildCmd interface{}) {
			concurrencyControl <- struct{}{}
			resultsCh <- buildCmd.(func() error)()
			<-concurrencyControl

		}(buildCmd)
	}
	// Doesn't work at the moment
	//	mg.Deps(buildCmds...)
	results := []error{}
	var resultErr error = nil
	for len(results) < len(buildCmds) {
		err := <-resultsCh
		results = append(results, err)
		if err != nil {
			fmt.Println(err)
			resultErr = errors.New("parallel build failed")
		}
		fmt.Printf("Finished %v of %v\n", len(results), len(buildCmds))
	}

	return resultErr
}

// Tools builds build tools of the project and is depended on by all other build targets.
func Tools() (err error) {
	// Catch panics and convert to errors
	defer func() {
		if perr := recover(); perr != nil {
			err = perr.(error)
		}
	}()

	if err := ensureVendorSrcLink(); err != nil {
		return err
	}

	toolBuild := func(toolType string, tools ...string) error {
		toolTargets := []interface{}{}
		for _, toolImport := range tools {
			toolParts := strings.Split(toolImport, "/")
			toolBin := path.Join(toolsBinDir, toolParts[len(toolParts)-1])
			Log("Check for changes:", toolBin, toolsVendorDir)
			changed, terr := target.Dir(toolBin, toolsVendorDir)
			if terr != nil {
				if !os.IsNotExist(terr) {
					panic(terr)
				}
				changed = true
			}
			if changed {
				localToolImport := toolImport
				f := func() error { return sh.RunWith(toolsEnv, "go", "install", "-v", localToolImport) }
				toolTargets = append(toolTargets, f)
			}
		}

		Log("Build", toolType, "tools")
		if berr := concurrencyLimitedBuild(toolTargets...); berr != nil {
			return berr
		}
		return nil
	}

	if berr := toolBuild("static", getCoreTools()...); berr != nil {
		return berr
	}

	if berr := toolBuild("static", getMetalinters()...); berr != nil {
		return berr
	}

	return nil
}

// UpdateTools automatically updates tool dependencies to the latest version.
func UpdateTools() error {
	if err := ensureVendorSrcLink(); err != nil {
		return err
	}

	// Ensure govendor is up to date without doing anything
	govendorPkg := "github.com/kardianos/govendor"
	govendorParts := strings.Split(govendorPkg, "/")
	govendorBin := path.Join(toolsBinDir, govendorParts[len(govendorParts)-1])

	sh.RunWith(toolsEnv, "go", "get", "-v", "-u", govendorPkg)

	if changed, cerr := target.Dir(govendorBin, toolsSrcDir); changed || os.IsNotExist(cerr) {
		if err := sh.RunWith(toolsEnv, "go", "install", "-v", govendorPkg); err != nil {
			return err
		}
	} else if cerr != nil {
		panic(cerr)
	}

	// Set current directory so govendor has the right path
	previousPwd, wderr := os.Getwd()
	if wderr != nil {
		return wderr
	}
	if err := os.Chdir(toolDir); err != nil {
		return err
	}

	// govendor fetch core tools
	for _, toolImport := range append(getCoreTools(), getMetalinters()...) {
		sh.RunV("govendor", "fetch", "-v", toolImport)
	}

	// change back to original working directory
	if err := os.Chdir(previousPwd); err != nil {
		return err
	}
	return nil
}

// Assets builds binary assets to be bundled into the binary.
func Assets() error {
	mg.Deps(Tools)

	if err := os.MkdirAll("assets/generated", os.FileMode(0777)); err != nil {
		return err
	}

	return sh.RunV("go-bindata", "-pkg=assets", "-o", "assets/bindata.go", "-ignore=bindata.go",
		"-ignore=.*.map$", "-prefix=assets/generated", "assets/generated/...")
}

// Lint runs gometalinter for code quality. CI will run this before accepting PRs.
func Lint() error {
	mg.Deps(Tools)
	args := []string{"-j", fmt.Sprintf("%v", concurrency), fmt.Sprintf("--deadline=%s",
		linterDeadline.String()), "--enable-all", "--line-length=120",
		"--disable=gocyclo", "--disable=testify", "--disable=test", "--disable=lll", "--exclude=assets/bindata.go"}
	return sh.RunV("gometalinter", append(args, goDirs...)...)
}

// Style checks formatting of the file. CI will run this before acceptiing PRs.
func Style() error {
	mg.Deps(Tools)
	args := []string{"--disable-all", "--enable=gofmt", "--enable=goimports"}
	return sh.RunV("gometalinter", append(args, goSrc...)...)
}

// Fmt automatically formats all source code files
func Fmt() error {
	mg.Deps(Tools)
	fmtErr := sh.RunV("gofmt", append([]string{"-s", "-w"}, goSrc...)...)
	if fmtErr != nil {
		return fmtErr
	}
	impErr := sh.RunV("goimports", append([]string{"-w"}, goSrc...)...)
	if impErr != nil {
		return fmtErr
	}
	return nil
}

func listCoverageFiles() ([]string, error) {
	result := []string{}
	finfos, derr := ioutil.ReadDir(coverageDir)
	if derr != nil {
		return result, derr
	}
	for _, finfo := range finfos {
		result = append(result, path.Join(coverageDir, finfo.Name()))
	}
	return result, nil
}

// Test run test suite
func Test() error {
	mg.Deps(Tools)

	// Ensure coverage directory exists
	if err := os.MkdirAll(coverageDir, os.FileMode(0777)); err != nil {
		return err
	}

	// Clean up coverage directory
	coverFiles, derr := listCoverageFiles()
	if derr != nil {
		return derr
	}
	for _, coverFile := range coverFiles {
		if err := sh.Rm(coverFile); err != nil {
			return err
		}
	}

	// Run tests
	coverProfiles := []string{}
	for _, pkg := range goPkgs {
		coverProfile := path.Join(coverageDir, fmt.Sprintf("%s%s", strings.Replace(pkg, "/", "-", -1), ".out"))
		testErr := sh.Run("go", "test", "-v", "-covermode", "count", fmt.Sprintf("-coverprofile=%s", coverProfile),
			pkg)
		if testErr != nil {
			return testErr
		}
		coverProfiles = append(coverProfiles, coverProfile)
	}

	return nil
}

// Build the intgration test binary
func IntegrationTestBinary() error {
	changed, err := target.Path("postgres_exporter_integration_test", goSrc...)
	if (changed && (err == nil)) || os.IsNotExist(err) {
		return sh.RunWith(map[string]string{"CGO_ENABLED": "0"}, "go", "test", "./cmd/postgres_exporter",
			"-c", "-tags", "integration",
			"-a", "-ldflags", "-extldflags '-static'",
			"-X", fmt.Sprintf("main.Branch=%s", branch),
			"-X", fmt.Sprintf("main.BuildDate=%s", buildDate),
			"-X", fmt.Sprintf("main.Revision=%s", revision),
			"-X", fmt.Sprintf("main.VersionShort=%s", versionShort),
			"-o", "postgres_exporter_integration_test", "-cover", "-covermode", "count")
	}
	return err
}

// TestIntegration runs integration tests
func TestIntegration() error {
	mg.Deps(Binary, IntegrationTestBinary)

	exporterPath := mustStr(filepath.Abs("postgres_exporter"))
	testBinaryPath := mustStr(filepath.Abs("postgres_exporter_integration_test"))
	testScriptPath := mustStr(filepath.Abs("postgres_exporter_integration_test_script"))

	integrationCoverageProfile := path.Join(coverageDir, "cover.integration.out")

	return sh.RunV("cmd/postgres_exporter/tests/test-smoke", exporterPath,
		fmt.Sprintf("%s %s %s", testScriptPath, testBinaryPath, integrationCoverageProfile))
}

// Coverage sums up the coverage profiles in .coverage. It does not clean up after itself or before.
func Coverage() error {
	// Clean up coverage directory
	coverFiles, derr := listCoverageFiles()
	if derr != nil {
		return derr
	}

	mergedCoverage, err := sh.Output("gocovmerge", coverFiles...)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(constCoverFile, []byte(mergedCoverage), os.FileMode(0777))
}

// All runs a full suite suitable for CI
func All() error {
	mg.SerialDeps(Style, Lint, Test, TestIntegration, Coverage, Release)
	return nil
}

// Release builds release archives under the release/ directory
func Release() error {
	mg.Deps(ReleaseBin)

	for _, platform := range platforms {
		owd, wderr := os.Getwd()
		if wderr != nil {
			return wderr
		}
		os.Chdir(binDir)

		if platform.OS == "windows" {
			// build a zip binary as well
			err := archiver.Zip.Make(fmt.Sprintf("%s.zip", platform.ReleaseBase()), []string{platform.ArchiveDir()})
			if err != nil {
				return err
			}
		}
		// build tar gz
		err := archiver.TarGz.Make(fmt.Sprintf("%s.tar.gz", platform.ReleaseBase()), []string{platform.ArchiveDir()})
		if err != nil {
			return err
		}
		os.Chdir(owd)
	}

	return nil
}

func makeBuilder(cmd string, platform Platform) func() error {
	f := func() error {
		// Depend on assets
		mg.Deps(Assets)

		cmdSrc := fmt.Sprintf("./%s/%s", mustStr(filepath.Rel(curDir, cmdDir)), cmd)

		Log("Make platform binary directory:", platform.PlatformDir())
		if err := os.MkdirAll(platform.PlatformDir(), os.FileMode(0777)); err != nil {
			return err
		}

		Log("Checking for changes:", platform.PlatformBin(cmd))
		if changed, err := target.Path(platform.PlatformBin(cmd), goSrc...); !changed {
			if err != nil {
				if !os.IsNotExist(err) {
					return err
				}
			} else {
				return nil
			}
		}

		fmt.Println("Building", platform.PlatformBin(cmd))
		return sh.RunWith(map[string]string{"CGO_ENABLED": "0", "GOOS": platform.OS, "GOARCH": platform.Arch},
			"go", "build", "-a", "-ldflags", fmt.Sprintf("-extldflags '-static' -X main.Version=%s", version),
			"-o", platform.PlatformBin(cmd), cmdSrc)
	}
	return f
}

func getCurrentPlatform() *Platform {
	var curPlatform *Platform
	for _, p := range platforms {
		if p.OS == runtime.GOOS && p.Arch == runtime.GOARCH {
			storedP := p
			curPlatform = &storedP
		}
	}
	Log("Determined current platform:", curPlatform)
	return curPlatform
}

// Binary build a binary for the current platform
func Binary() error {
	curPlatform := getCurrentPlatform()
	if curPlatform == nil {
		return errors.New("current platform is not supported")
	}

	for _, cmd := range goCmds {
		err := makeBuilder(cmd, *curPlatform)()
		if err != nil {
			return err
		}
		// Make a root symlink to the build
		cmdPath := path.Join(curDir, cmd)
		os.Remove(cmdPath)
		if err := os.Symlink(curPlatform.PlatformBin(cmd), cmdPath); err != nil {
			return err
		}
	}

	return nil
}

// ReleaseBin builds cross-platform release binaries under the bin/ directory
func ReleaseBin() error {
	buildCmds := []interface{}{}

	for _, cmd := range goCmds {
		for _, platform := range platforms {
			buildCmds = append(buildCmds, makeBuilder(cmd, platform))
		}
	}

	resultsCh := make(chan error, len(buildCmds))
	concurrencyControl := make(chan struct{}, concurrency)
	for _, buildCmd := range buildCmds {
		go func(buildCmd interface{}) {
			concurrencyControl <- struct{}{}
			resultsCh <- buildCmd.(func() error)()
			<-concurrencyControl

		}(buildCmd)
	}
	// Doesn't work at the moment
	//	mg.Deps(buildCmds...)
	results := []error{}
	var resultErr error = nil
	for len(results) < len(buildCmds) {
		err := <-resultsCh
		results = append(results, err)
		if err != nil {
			fmt.Println(err)
			resultErr = errors.New("parallel build failed")
		}
		fmt.Printf("Finished %v of %v\n", len(results), len(buildCmds))
	}

	return resultErr
}

// Docker builds the docker image
func Docker() error {
	mg.Deps(Binary)
	p := getCurrentPlatform()
	if p == nil {
		return errors.New("current platform is not supported")
	}

	return sh.RunV("docker", "build",
		fmt.Sprintf("--build-arg=binary=%s",
			mustStr(filepath.Rel(curDir, p.PlatformBin("postgres_exporter")))),
		"-t", containerName, ".")
}

// Clean deletes build output and cleans up the working directory
func Clean() error {
	for _, name := range goCmds {
		if err := sh.Rm(path.Join(binDir, name)); err != nil {
			return err
		}
	}

	for _, name := range outputDirs {
		if err := sh.Rm(name); err != nil {
			return err
		}
	}
	return nil
}

// Debug prints the value of internal state variables
func Debug() error {
	fmt.Println("Source Files:", goSrc)
	fmt.Println("Packages:", goPkgs)
	fmt.Println("Directories:", goDirs)
	fmt.Println("Command Paths:", goCmds)
	fmt.Println("Output Dirs:", outputDirs)
	fmt.Println("Tool Src Dir:", toolsSrcDir)
	fmt.Println("Tool Vendor Dir:", toolsVendorDir)
	fmt.Println("Tool GOPATH:", toolsGoPath)
	fmt.Println("PATH:", os.Getenv("PATH"))
	return nil
}

// Autogen configure local git repository with commit hooks
func Autogen() error {
	fmt.Println("Installing git hooks in local repository...")
	return os.Link(path.Join(curDir, toolDir, "pre-commit"), ".git/hooks/pre-commit")
}
