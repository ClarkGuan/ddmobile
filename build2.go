package main

import (
	"fmt"
	"os"

	"golang.org/x/tools/go/packages"
)

var cmdBuild2 = &command{
	run:   runBuild2,
	Name:  "build2",
	Usage: "[-target android|ios] [-o output] [-bundleid bundleID] [build flags] [package]",
	Short: "compile android shared library and iOS static library",
	Long: fmt.Sprintf(`
Build2 compiles and encodes the app named by the import path.

The named package must define a main function.

The -target flag takes a target system name, either android (the
default) or ios.

For -target android, by default, this builds a fat APK for all supported
instruction sets (arm, 386, amd64, arm64). A subset of instruction sets can
be selected by specifying target type with the architecture name. E.g.
-target=android/arm,android/386.

For -target ios, %s must be run on an OS X machine with Xcode
installed.

Flag -iosversion sets the minimal version of the iOS SDK to compile against.
The default version is 7.0.

Flag -androidapi sets the Android API version to compile against.
The default and minimum is 15.

The -bundleid flag is required for -target ios and sets the bundle ID to use
with the app.

The -o flag specifies the output file name. If not specified, the
output file name depends on the package built.

The -v flag provides verbose output, including the list of packages built.

The build flags -a, -i, -n, -x, -gcflags, -ldflags, -tags, -trimpath, and -work are
shared with the build command. For documentation, see 'go help build'.
`, gomobileName),
}

func runBuild2(cmd *command) (err error) {
	_, err = runBuildImpl2(cmd)
	return
}

// runBuildImpl builds a package for mobiles based on the given commands.
// runBuildImpl returns a built package information and an error if exists.
func runBuildImpl2(cmd *command) (*packages.Package, error) {
	cleanup, err := buildEnvInit()
	if err != nil {
		return nil, err
	}
	defer cleanup()

	args := cmd.flag.Args()

	targetOS, targetArchs, err := parseBuildTarget(buildTarget)
	if err != nil {
		return nil, fmt.Errorf(`invalid -target=%q: %v`, buildTarget, err)
	}

	var buildPath string
	switch len(args) {
	case 0:
		buildPath = "."
	case 1:
		buildPath = args[0]
	default:
		cmd.usage()
		os.Exit(1)
	}
	pkgs, err := packages.Load(packagesConfig(targetOS), buildPath)
	if err != nil {
		return nil, err
	}
	// len(pkgs) can be more than 1 e.g., when the specified path includes `...`.
	if len(pkgs) != 1 {
		cmd.usage()
		os.Exit(1)
	}

	pkg := pkgs[0]

	if pkg.Name != "main" && buildO != "" {
		return nil, fmt.Errorf("cannot set -o when building non-main package")
	}

	var nmpkgs map[string]bool
	switch targetOS {
	case "android":
		if pkg.Name != "main" {
			for _, arch := range targetArchs {
				if err := goBuild(pkg.PkgPath, androidEnv[arch]); err != nil {
					return nil, err
				}
			}
			return pkg, nil
		}
		nmpkgs, err = goAndroidBuild(pkg, targetArchs)
		if err != nil {
			return nil, err
		}
	case "darwin":
		if !xcodeAvailable() {
			return nil, fmt.Errorf("-target=ios requires XCode")
		}
		if pkg.Name != "main" {
			for _, arch := range targetArchs {
				if err := goBuild(pkg.PkgPath, darwinEnv[arch]); err != nil {
					return nil, err
				}
			}
			return pkg, nil
		}
		if buildBundleID == "" {
			return nil, fmt.Errorf("-target=ios requires -bundleid set")
		}
		nmpkgs, err = goIOSBuild(pkg, buildBundleID, targetArchs)
		if err != nil {
			return nil, err
		}
	}

	if !nmpkgs["golang.org/x/mobile/app"] {
		return nil, fmt.Errorf(`%s does not import "golang.org/x/mobile/app"`, pkg.PkgPath)
	}

	return pkg, nil
}
