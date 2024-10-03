- [Quick demo](#quick-demo)
- [What is Rye language 🌾](#what-is-rye-language)
- [What is Rye-Front](#what-is-rye-front)
- [Modules](#modules)
  - [Fyne - GUI](#fyne---gui-)
    - [Build and test](#build-and-test)
    - [Example](#example)
    - [More about Fyne](#more-about-fyne)
  - [Ebitengine - Game engine](Ebitengine-game-engine)
  - [Webview](Webview)

## Current status

Most widgets work. We just created a (CookBook with plenty of examples](https://ryelang.org/cookbook/rye-fyne/examples/). Next step will be to update this 
README and repository in general. To provide prebuild binaries, etc ... stay tuned.

## A Cookbook

I'm writing a Cookbook page full of simple GUI example. See them here:
https://ryelang.org/cookbook/rye-fyne/examples/

## Live use video

https://www.youtube.com/watch?v=YmYQRPvkSpM

[![Live GUI over console demo](http://img.youtube.com/vi/YmYQRPvkSpM/0.jpg)](http://www.youtube.com/watch?v=QtK8hUPjo5Y "Video Title")

## What is Rye language

Rye is a high level, dynamic **programming language** based on ideas from **Rebol**, flavored by
Factor, Linux shells and Golang. It's still an experiment in language design, but it should slowly become more and
more useful in real world.

It features a Golang based interpreter and console and could also be seen as (modest) **Go's scripting companion** as
Go's libraries are quite easy to integrate, and Rye can be embedded into Go programs as a scripting or config language.

I believe that as language becomes higher level it starts touching the user interface boundary, besides being a language
we have great emphasis on **interactive use** (Rye shell) where we will explore that.

**[Rye language repository](https://github.com/refaktor/rye)** | **[Rye website](https://ryelang.org/)** | **[Reddit group](https://reddit.com/r/ryelang/)**

## Download Rye-Fyne

Download the binary: 
* Linux: [rye-fyne-linux-amd64.tar.gz](https://github.com/refaktor/rye-fyne/releases/download/23/rye-fyne-linux-amd64.tar.gz)
* MacOS: [rye-fyne-macos-amd64.tar.gz](https://github.com/refaktor/rye-fyne/releases/download/23/rye-fyne-macos-amd64.tar.gz)
* Windows: [rye-fyne.exe](https://github.com/refaktor/rye-fyne/releases/download/23/rye-fyne.exe)

This includes just the binary (executable), to download examples you should still need to download ZIP from the github. In future
we will make the binary download all required things itself if raw with `rye install`.

## Building Rye-Fyne

You need [Go](https://go.dev/) installed. Please follow Go's installation instructions for your opearating system. 

```
# To get rye-fyne use "Download ZIP" from "<> Code" button
# on the github page: https://github.com/refaktor/rye-fyne
# or use Git:
git clone https://github.com/refaktor/rye-fyne.git

# enter directory
cd rye-fyne

# build rye with fyne in bin/fyne/rye
./build

# Try the hello example
bin/rye-fyne examples/fyne/button.rye

# Try the feedback example
bin/rye-fyne examples/fyne/feedback.rye

# Try the Live GUI demo
bin/rye-fyne examples/fyne/live.rye

# Or enter the Rye console
bin/rye-fyne
x> cc fyne
x> app .window "Hello" |set-content label "world" |show-and-run
(Ctrl-c to exit)
```

#### Example

![Fyne Feedback example](https://ryelang.org/rye-fyne-2.png)

```
rye .needs { fyne }

do\in fyne {

	cont: container 'vbox vals {
		label "Send us feedback:"
		multiline-entry :ent
		button "Send" { ent .get-text |printv "Sending: {}" }
	}
	
	app .new-window "Feedback"
	|set-content cont
	|show-and-run
}
```

#### More about Fyne

[Fyne website](https://fyne.io)

