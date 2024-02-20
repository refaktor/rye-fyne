- [What is Rye language 🌾](#what-is-rye-language-)
- [What is Rye-Front](#what-is-rye-front)
- [Modules](#modules)
  - [Fyne - GUI](fyne-gui)
  - [Ebitengine - Game engine](Ebitengine-game-engine)
  - [Webview](Webview)

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

## Modules

### Fyne ⭐⭐

Fyne is crossplatform GUI framework with it's own OpenGL renderer inspired by material design.

![Simple Fyne example](https://ryelang.org/rye-fyne-1.png)

To run the example in rye-front directory run:

```
./buildfyne
bin/ryef examples/fyne/button_ctx.rye
```

[Fyne website](https://fyne.io)

### Ebitengine

[Ebitengine website](https://ebitengine.org)

### Webview

[Webview github page](https://github.com/webview/webview)


