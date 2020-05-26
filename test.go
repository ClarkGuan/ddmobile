package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

var cmdTest = &command{
	run:   runTest,
	Name:  "test",
	Usage: "[-target android] [-o output] [build flags] [package]",
	Short: "generate android or iOS test binary file",
	Long: fmt.Sprintf(`
Test compiles and generate android or iOS test binary file which named by the import path.

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

func runTest(cmd *command) (err error) {
	_, err = runTestImpl(cmd)
	return
}

// runTestImpl test a package for mobiles based on the given commands.
// runTestImpl returns a built package information and an error if exists.
func runTestImpl(cmd *command) (*packages.Package, error) {
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

	// 单元测试不需要 package name 必须为 main
	//if pkg.Name != "main" && buildO != "" {
	//	return nil, fmt.Errorf("cannot set -o when building non-main package")
	//}

	switch targetOS {
	case "android":
		_, err = goAndroidTest(pkg, targetArchs)
		if err != nil {
			return nil, err
		}
	case "darwin":
		return nil, fmt.Errorf("not support iOS now")
		//if !xcodeAvailable() {
		//	return nil, fmt.Errorf("-target=ios requires XCode")
		//}
		//_, err = goIOSTest(pkg, buildBundleID, targetArchs)
		//if err != nil {
		//	return nil, err
		//}
	}

	return pkg, nil
}

func goAndroidTest(pkg *packages.Package, androidArchs []string) (map[string]bool, error) {
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
		libPath = "test/" + toolchain.abi + "/" + libName
		libAbsPath := filepath.Join(buildO, "android", libPath)
		if err := mkdir(filepath.Dir(libAbsPath)); err != nil {
			return nil, err
		}
		args = append(args, "-c") // not run tests
		args = append(args, "-buildmode=pie")
		args = append(args, "-o", libAbsPath)
		err = goTest(
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

func goTest(src string, env []string, args ...string) error {
	return goCmd("test", []string{src}, env, args...)
}
