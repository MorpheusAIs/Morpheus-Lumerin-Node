#!/bin/bash

# Change directory to the proxy group
cd ..

for $repo_name in contracts-js wallet-core
do
  in_setup_and_out on $repo_name
done

core_links()
desktop_links()

in_setup_and_out() {
  cd $1

  git checkout dev
  npm i
  npm link

  cd ..
}

core_links() {
  cd wallet-core

  npm link @lumerin/contracts

  cd ..
}

desktop_links() {
  cd wallet-desktop

  npm link @lumerin/wallet-core
}
