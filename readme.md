# Transmisstion Desktop

* 使用Transmission-rpc接口，制作的桌面端
* 使用Go语言的fyne库 (跨平台支持Windows、安卓等)
* 主要解决：
* PT种子下载，那名字基本都是很长的英文且可读性很差
* 手机端，Transmission自带的网页端非常不好用

# 应用截图(Windows端，手机端差不多)

![Login](./img/welcome.png)

![main](./img/main.png)


# Build
## 安装Windwow平台的GCC编译器

如：MSYS2 from msys2.org (MSYS2 MinGW 64-bit)

Execute the following commands (if asked for install options be sure to choose “all”):

  $ pacman -Syu

  $ pacman -S git mingw-w64-x86_64-toolchain

You will need to add /c/Program\ Files/Go/bin and ~/Go/bin to your PATH, for MSYS2 you can paste the following command into your terminal:

  $ echo "export PATH=\$PATH:/c/Program\ Files/Go/bin:~/Go/bin" >> ~/.bashrc

## Go相关初始化

$ cd myapp

$ go mod init MODULE_NAME

$ go get fyne.io/fyne/v2@latest

$ go install fyne.io/fyne/v2/cmd/fyne@latest

最新版本还遇到过要设置这个

$ go env -w CGO_ENABLED=1

## Android Build

$ fyne package -os android -appID org.zhaobi.transmissionclient

