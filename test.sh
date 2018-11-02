#!/bin/bash

./gostd2joker -v --source tests/small/net > tests/small/net/TEST.out 2>&1
diff -u tests/small/net/TEST.gold tests/small/net/TEST.out

./gostd2joker -v --source tests/big/net > tests/big/net/TEST.out 2>&1
diff -u tests/big/net/TEST.gold tests/big/net/TEST.out

if [ -n "$GOSRC" -a -d "$GOSRC" ]; then
    ./gostd2joker -v --source $GOSRC > tests/GOSRC.out 2>&1
    diff -u tests/GOSRC.gold tests/GOSRC.out
fi
