# ddmobile

gomobile 的魔改版本。主要是不满意 gomobile 默认情况下直接构建 module 包的功能，针对 Android 和 iOS 平台修改为构建产出 .so 和 .a 库文件。

#### 安装

```bash
go get github.com/ClarkGuan/ddmobile
```

#### 使用

##### 1、初始化

和 gomobile 类似，使用前我们运行命令

```bash
ddmobile init
```

这一过程本质上是预编译 Go 在 Android 和 iOS 平台上对应的标准库。

我们假设有一个 go 工程在 $GOPATH/src/hello_world 目录下，

Android 环境初始化依赖 NDK 工具，默认情况下会搜索如下环境变量：

* ANDROID_HOME

实际上是指定 Android SDK 的位置，如果 SDK 中装有 NDK（ndk-bundle 目录）则使用该版本 NDK。

iOS 环境初始化依赖 Xcode 以及相关命令行工具（xcrun 等）。

##### 2、Android 构建

```bash
cd $GOPATH/src/hello_world
ddmobile build -target android/arm,android/arm64
```

如果构建顺利，会在 $GOPATH/src/hello_world 目录中生成 build/ 子目录，并列出 arm 32 位和 64 位的动态库。

如果我们编译的是可执行文件，则运行

```bash
cd $GOPATH/src/hello_world
ddmobile build -exe
```

这时生成的子目录为 build/android/app/。另：我们并没有特殊指明 `-target android/arm,android/arm64`，默认会产出 Android 所有支持的平台产物。

#### 举例

我们以可以在 Android 上运行的 HelloWorld 工程为例，创建目录

```bash
mkdir -p $GOPATH/src/hello
cd $GOPATH/src/hello
touch main.go
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

```bash
ddmobile build -target android/arm -exe
```

使用我编写的另一个工具 arun（ https://github.com/ClarkGuan/arun ）：

```bash
arun -exe build/android/app/armeabi-v7a/hello
```

输出类似下面（我们假定您已经将 adb 命令加入到 $PATH 中）

```
prepare to push /Users/xxx/gopath/src/hello/build/android/app/armeabi-v7a/hello to device
/Users/xxx/gopath/src/hello/build/android/app/armeabi-v7a/hello: 1 file pushed. 9.9 MB/s (1963673 bytes in 0.189s)
[程序输出如下]
Hello world
[程序执行完毕]
```
