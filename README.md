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

## Quick demo

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

## What is Rye-Front

Rye-Front is an external extension of Rye language focused on frontend technologies like: GUI, Game engine, Graphics, Browsers ...

### Why a separate repository

 * So Rye remains lighter on dependencies, easier to build, focused on backend and interactive shell
 * So that "frontend" related development is separated from language development
 * So that we test and improve on how users of Rye can externally extend it, add their own (private) bindings and write their own Go (private) builtin functions for hot-code optimization

## Status

Rye-front is in early development. We are focusing on Fyne GUI at first.

## Modules

### Fyne - GUI ⭐⭐

Fyne is crossplatform GUI framework with it's own OpenGL renderer inspired by material design.


#### Build and test

In **rye-front** directory run:

```
# build rye with fyne in bin/fyne/rye
./buildfyne

# Try the hello example
bin/fyne/rye examples/fyne/button.rye

# Try the feedback example
bin/fyne/rye examples/fyne/feedback.rye

# Try the Live GUI demo
bin/fyne/rye examples/fyne/live.rye
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

### Ebitengine

[Ebitengine website](https://ebitengine.org)

### Webview

[Webview github page](https://github.com/webview/webview)


