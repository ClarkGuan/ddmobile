package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

var cmdBuild2 = &command{
	run:   runBuild2,
	Name:  "build2",
	Usage: "[-target android|ios] [-o output] [build flags] [package]",
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
		_, err = goAndroidBuild2(pkg, targetArchs)
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
		_, err = goIOSBuild2(pkg, buildBundleID, targetArchs)
		if err != nil {
			return nil, err
		}
	}

	return pkg, nil
}

func goAndroidBuild2(pkg *packages.Package, androidArchs []string) (map[string]bool, error) {
	_, err := ndkRoot()
	if err != nil {
		return nil, err
	}
	appName := path.Base(pkg.PkgPath)
	libName := androidPkgName(appName)

	if buildO == "" {
		buildO = "build"
	}
	if buildProName != "" {
		libName = buildProName
	}
	args := []string(nil)
	libPath := ""

	for _, arch := range androidArchs {
		toolchain := ndk.Toolchain(arch)
		if buildExe {
			libPath = "app/" + toolchain.abi + "/" + libName
		} else {
			libPath = "lib/" + toolchain.abi + "/lib" + libName + ".so"
		}
		libAbsPath := filepath.Join(buildO, "android", libPath)
		if err := mkdir(filepath.Dir(libAbsPath)); err != nil {
			return nil, err
		}
		args = nil
		if !buildExe {
			args = append(args, "-buildmode=c-shared")
		}
		args = append(args, "-o", libAbsPath)
		err = goBuild(
			pkg.PkgPath,
			androidEnv[arch],
			args...,
		)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func goIOSBuild2(pkg *packages.Package, bundleID string, archs []string) (map[string]bool, error) {
	src := pkg.PkgPath

	productName := rfc1034Label(path.Base(pkg.PkgPath))
	if productName == "" {
		productName = "ProductName" // like xcode.
	}

	if buildO == "" {
		buildO = "build"
	}
	if buildProName != "" {
		productName = buildProName
	}

	// We are using lipo tool to build multiarchitecture binaries.
	cmd := exec.Command(
		"xcrun", "lipo",
		"-o", filepath.Join(buildO, "iOS", "lib"+productName+".a"),
		"-create",
	)
	for _, arch := range archs {
		path := filepath.Join(tmpdir, arch)
		// Disable DWARF; see golang.org/issues/25148.
		if err := goBuild(src, darwinEnv[arch], "-ldflags=-w", "-buildmode=c-archive", "-o="+path); err != nil {
			return nil, err
		}
		cmd.Args = append(cmd.Args, path)
	}

	if err := runCmd(cmd); err != nil {
		return nil, err
	}

	return nil, nil
}
