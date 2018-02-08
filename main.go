package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	semver "gopkg.in/blang/semver.v1"
)

func commitMessage(message, version string) string {
	if strings.Contains(message, "%s") {
		return fmt.Sprintf(message, version)
	}
	return message
}

func getCurrentVersion(path string) (*semver.Version, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &semver.Version{}, nil
	}
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return semver.New(strings.TrimSpace(string(contents)))
}

const versionFileName = "VERSION"

func exitWithError(message string) {
	fmt.Fprintf(os.Stderr, message+"\n\n")
	flag.Usage()
	os.Exit(1)
}

func bump(old *semver.Version, part string) *semver.Version {
	// We don't want to mutate the input, but there's no Clone or Copy method on a semver.Version,
	// so we make a new one by parsing the string version of the old one.
	// We ignore any errors because we know it's valid semver.
	new, _ := semver.New(old.String())
	switch part {
	case "major":
		new.Major++
		new.Minor = 0
		new.Patch = 0
	case "minor":
		new.Minor++
		new.Patch = 0
	case "patch":
		new.Patch++
	}
	return new
}

type context struct {
	versionFile      string
	oldVersion       *semver.Version
	newVersion       *semver.Version
	shouldCleanCheck bool
	message          string
	shouldTag        bool
	preview          bool
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s: [options] version\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "version can be one of: <literal semver> | patch | minor | major\n\n")
		fmt.Fprintf(os.Stderr, "options:\n")
		flag.PrintDefaults()
	}
	var ctx context
	flag.StringVar(&ctx.message, "m", "%s", "commit message for version commit")
	help := flag.Bool("h", false, "print usage and exit")
	flag.BoolVar(&ctx.shouldTag, "tag", true, "whether or not to make a tag at the version commit")
	flag.BoolVar(&ctx.preview, "preview", false, "whether to show a preview of the version without modifying anything")
	noCleanCheck := flag.Bool("n", false, "whether to check git repo's cleanliness before proceeding")
	flag.Parse()

	if ctx.preview {
		*noCleanCheck = true
	}
	ctx.shouldCleanCheck = !*noCleanCheck

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if ctx.message == "" {
		exitWithError("missing message")
	}

	if len(flag.Args()) != 1 {
		exitWithError("gitsem takes exactly one non-flag argument: version")
	}

	root, err := repoRoot()
	if err != nil {
		log.Fatalf("unable to find git repo root: %+v", err)
	}

	if ctx.shouldCleanCheck {
		if clean, err := isRepoClean(); err != nil {
			log.Fatalf("unable to determine git repo clean status: %+v", err)
		} else if !clean {
			log.Fatal("repo isn't clean")
		}
	}

	ctx.versionFile = filepath.Join(root, versionFileName)
	version, err := getCurrentVersion(ctx.versionFile)
	if err != nil {
		log.Fatalf("unable to get current version: %+v", err)
	}
	ctx.oldVersion = version

	newVersionAction := flag.Args()[0]
	switch newVersionAction {
	case "patch", "minor", "major":
		version = bump(version, newVersionAction)
	default: // newVersionAction is an actual version string
		if strings.HasPrefix(newVersionAction, "v") {
			newVersionAction = strings.TrimPrefix(newVersionAction, "v")
		}
		if version, err = semver.New(newVersionAction); err != nil {
			log.Fatalf("failed to parse %s as semver: %s", newVersionAction, err.Error())
		}
	}
	ctx.newVersion = version

	if ctx.preview {
		printPreview(ctx)
	} else {
		commitNewVersion(ctx)
	}
}

func commitNewVersion(ctx context) {
	if err := ioutil.WriteFile(ctx.versionFile, []byte(ctx.newVersion.String()), 0666); err != nil {
		log.Fatalf("unable to write version file: %+v", err)
	}
	if err := addFile(ctx.versionFile); err != nil {
		log.Fatalf("unable to git add version file: %+v", err)
	}
	versionString := "v" + ctx.newVersion.String()
	message := commitMessage(ctx.message, versionString)
	if err := commit(message); err != nil {
		log.Fatalf("unable to git commit updated version: %+v", err)
	}
	if ctx.shouldTag {
		if err := tag(versionString); err != nil {
			log.Fatalf("unable to git tag updated version: %+v", err)
		}
	}
	fmt.Println(versionString)
}

func printPreview(ctx context) {
	clean, err := isRepoClean()
	if err != nil {
		log.Fatalf("unable to determine git repo clean status: %+v", err)
	}

	if !clean {
		fmt.Println("git repo isn't clean")
	}

	fmt.Printf("current version: v%s\n", ctx.oldVersion)
	fmt.Printf("    new version: v%s\n", ctx.newVersion)
}
