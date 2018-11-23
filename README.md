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

##### 2、Android 构建

```bash
cd $GOPATH/src/hello_world
ddmobile build -target android/arm,android/arm64
```

如果构建顺利，会在 $GOPATH/src/hello_world 目录中生成 build/android/lib/ 子目录，并列出 arm 32 位和 64 位的动态库。

如果我们编译的是可执行文件，则运行

```bash
cd $GOPATH/src/hello_world
ddmobile build -pie
```

这时生成的子目录为 build/android/app/。另：我们并没有特殊指明 `-target android/arm,android/arm64`，默认会产出 Android 所有支持的平台产物。

##### 3、iOS 构建

```bash
cd $GOPATH/src/hello_world
ddmobile build -target ios
```

生成子目录 build/ios/。注意：与 Android 最大的不同是生成的是静态库文件（.a，经过 lipo 命令合并过的），并且在不同平台目录下还会有各自的静态库文件。
