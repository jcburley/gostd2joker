#!/bin/bash

./gostd2joker -v --source tests/small/net > tests/small/net/TEST.gold 2>&1
git diff -u tests/small/net/TEST.gold

./gostd2joker -v --source tests/big/net > tests/big/net/TEST.gold 2>&1
git diff -u tests/big/net/TEST.gold

if [ -n "$GOSRC" -a -d "$GOSRC" ]; then
    ./gostd2joker -v --source $GOSRC > tests/GOSRC.gold 2>&1
    git diff -u tests/GOSRC.gold
fi
