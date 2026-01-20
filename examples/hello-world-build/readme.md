
# util/ryel.rye is the script 

build can create a binary of the main if you use embed_main otherwise it creates a binary interpreter

build\fyne creates a tar.xz file that includes some folders and Make you can install app then, on linux at least

build\fyne\apk doesn't work yet, because we don't have bindings generated for that os / arch ... 

# plan

ryel becomes lrye localRYE

ryelc becomes rye-build (for example) ... and flags get more standard and polished improved

there is ryel.rye improved with some ideas 


for how to install ryel we can take some ideas from fyne xz tool or look at the best way in generl ... tool also requires RYE_HOME defined[D[D[D[D[D[D[D[D[D[D

# also needs

RYE_HOME defined
rye and rye-fyne must exist locally and go.mod for rye-fyne musl direct to local rye so embed_main can copy main.rye file inthere

we don't have solution for multiple files probably yet

