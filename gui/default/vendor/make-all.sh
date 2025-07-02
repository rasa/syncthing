#!/usr/bin/env bash

function pause() {
    read -n1 -r -p "Press any key to continue: " </dev/tty
    echo
}

paths=(
    angular
    bootstrap
    daterangepicker
    fancytree
    fork-awesome
    HumanizeDuration.js
    jquery
    moment
    ../index.html
)

set -e
set -v

make uninstall_all

make nuke

git restore -q "${paths[@]}"

set -a

ANGULAR_VER=1.3.20

. "v${ANGULAR_VER}.sh"

test -z "${SAVE_STASHES:-}" || \
    git stash clear

clear

make "$@" 2>&1 | tee "${ANGULAR_VER}.log"

test -z "${SAVE_STASHES:-}" || \
    git stash list

# ls -l tags diffs || true
