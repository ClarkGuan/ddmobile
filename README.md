# ddmobile

gomobile 的魔改版本。主要是不满意 gomobile 默认情况下直接构建 module 包这个功能，针对 Android 和 iOS 平台修改为构建产出 .so 和 .a 库文件。（Android 平台还支持直接生成可执行文件）

#### 安装

```
$ cd <your-work-space>
$ git clone github.com/ClarkGuan/ddmobile
$ cd ddmobile
$ go get
$ go install
```

#### 帮助

```
$ ddmobile help
Ddmobile is a tool for building and running mobile apps written in Go.

To install:

	$ go get github.com/ClarkGuan/ddmobile
	$ ddmobile init

At least Go 1.10 is required.
For detailed instructions, see https://golang.org/wiki/Mobile.

Usage:

	ddmobile command [arguments]

Commands:

	bind        build a library for Android and iOS
	build       compile android APK and iOS app
	build2      compile android shared library and iOS static library
	clean       remove object files and cached ddmobile files
	init        build OpenAL for Android
	install     compile android APK and install on device
	version     print version

Use 'ddmobile help [command]' for more information about that command.
```

#### 使用

##### 1、初始化

Android 环境初始化依赖 NDK 工具，默认情况下会搜索如下环境变量：

* ANDROID_HOME

实际上是指定 Android SDK 的位置，如果 SDK 中装有 NDK（ndk-bundle 目录）则使用该版本 NDK。

* NDK 或 ANDROID_NDK_HOME

因为 Android SDK 新版将 NDK 内置路径（ndk-bundle 目录）作出修改，为了便于 ddmobile 查找，可以定义 ANDROID_NDK_HOME 环境变量指向具体位置。

iOS 环境初始化依赖 Xcode 以及相关命令行工具（xcrun 等）。

##### 2、Android 构建

我们假设有一个 go 工程在 `<your-work-space>` 目录下，

```
$ cd <your-work-space>
$ ddmobile build2 -target android/arm,android/arm64
```

如果构建顺利，会在 $GOPATH/src/hello_world 目录中生成 build/android/lib 子目录，并列出 arm 32 位和 64 位的动态库。

如果我们编译的是可执行文件，则运行

```
$ cd <your-work-space>
$ ddmobile build2 -exe
```

这时生成的子目录为 build/android/app。另：我们并没有特殊指明 `-target android/arm,android/arm64`，默认会产出 Android 所有支持的平台产物。

##### 3、iOS 构建

```
$ cd <your-work-space>
$ ddmobile build2 -target ios
```

和 Android 构建类似，因为这里并没有指定使用何种架构编译，所以会生成所有支持的 iOS 架构产物，构建目录是 build/iOS 子目录。

当然，我们也可以指定目标架构：

```
$ cd <your-work-space>
$ ddmobile build2 -target ios/arm,ios/arm64,ios/386,ios/amd64
```

这个命令和上一个命令是等价的。

#### 举例

我们以可以在 Android 上运行的 HelloWorld 工程为例，创建目录

```
$ mkdir -p <your-work-space>
$ cd <your-work-space>
$ touch main.go
```

main.go 内容如下：

```go
package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello world")
}
```

此时运行命令

```
$ go mod init hello
$ ddmobile build2 -target android/arm -exe
```

使用我编写的另一个工具 arun（ https://github.com/ClarkGuan/arun ）：

```
$ GO111MODULE=off go get -u github.com/ClarkGuan/arun
$ arun -exe build/android/app/armeabi-v7a/hello
```

输出类似下面（我们假定您已经将 adb 命令加入到 $PATH 中）

```
prepare to push /Users/xxx/<your-work-space>/build/android/app/armeabi-v7a/hello to device
/Users/xxx/<your-work-space>/build/android/app/armeabi-v7a/hello: 1 file pushed. 9.9 MB/s (1963673 bytes in 0.189s)
[程序输出如下]
Hello world
[程序执行完毕]
```

#### 修改日志

* 2019-09-05 去掉 iOS 生成的目标文件中的 bitcode 段（一是增大了包体积；二是仍然无法满足 Apple 100% bitcode 覆盖的要求）
* 2019-10-10 创建分支 v1.0。该分支为 2018 年代码，支持 NDKr17 以及更早的版本
* 2019-10-11
    * 添加选项 -p，可以指定 Android 和 iOS 输出库文件的名称。例如 -p hello，对应 Android 动态库 libhello.so；对应 iOS 静态库 libhello.a
    * 【Android SDK 内 NDK 目录名称又变化了】添加对 $NDK 环境变量的识别，优先使用该环境变了定位 NDK 的位置
* 2020-04-15
    * 创建分支 v1.1。该分支为 2019 年代码，支持 NDKr17～NDKr20 或更高的版本（最高支持版本未知）
    * 合并最新 gomobile 代码；将生成命令修改为 build2 以兼容 gomobile 已有功能
    * 因为最大限度兼容 gomobile，所以恢复了生成 bitcode 的功能
