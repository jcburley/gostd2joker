#!/bin/bash

EXIT="exit 99"
if [ "$1" = "--on-error" ]; then
    EXIT="$2"
fi

./gostd2joker -v --source tests/small | grep -v '^Default context:' > tests/small/TEST.gold 2>&1
git diff --quiet -u tests/small/TEST.gold || { echo >&2 "FAILED: small test"; $EXIT; }

./gostd2joker -v --source tests/big --replace --populate joker/go | grep -v '^Default context:' > tests/big/TEST.gold 2>&1
git diff --quiet -u tests/big/TEST.gold || { echo >&2 "FAILED: big test"; $EXIT; }

if [ -n "$GOSRC" -a -d "$GOSRC" ]; then
    ./gostd2joker -v --source "$GOSRC" | grep -v '^Default context:' > tests/GOSRC.gold 2>&1
    git diff --quiet -u tests/GOSRC.gold || { echo >&2 "FAILED: \$GOSRC test"; $EXIT; }
fi
