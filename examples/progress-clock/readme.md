# How to build

## Desktop

## Android


*in go.mod replace github/re../rye with local rye, so that the buildtemp folder in runner will be found*

export RYE_HOME=/home/jimez/Work/rye
$RYE_HOME/bin/rye $RYE_HOME/util/ryel.rye buildfyne embed_main do_main

