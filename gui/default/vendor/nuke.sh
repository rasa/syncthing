#!/usr/bin/env bash

function pause() {
    read -n1 -r -p "Press any key to continue: " </dev/tty
    echo
}

# set -v

if [ ! -v NVM_DIR ]; then
    export NVM_DIR="${HOME}/.config/nvm"
fi

sed -i '/nvm/d; /NVM/d;' ~/.bashrc

# rm -fr ${NVM_DIR}/versions/node
# rm -fr ${NVM_DIR}/v0*

paths=(
    "${NVM_DIR}"
    "${HOME}/.config/configstore"
    "${HOME}/.node-gyp"
    "${HOME}/.npm"
    "${HOME}/.nvmrc"
)

find "${paths[@]}" -mindepth 1 -depth -exec rm -rf -- {} + 2>/dev/null

unset NVM_DIR
echo Remember to:
echo "  unset NVM_DIR ; make install_nvm ; . ~/.config/nvm/nvm.sh ; nvm install --default 0.10"
echo "-or-"
echo "  unset NVM_DIR ; make install_nvm ; . ~/.config/nvm/nvm.sh ; nvm install --default 4.2"
