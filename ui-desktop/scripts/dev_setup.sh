#!/bin/bash

# Change directory to the proxy group
cd ..

for $repo_name in contracts-js wallet-ui-logic wallet-core
do
  in_setup_and_out on $repo_name
done

wallet-core_links
wallet-desktop_links

in_setup_and_out() {
  cd $1

  git clone git@gitlab.com:TitanInd/proxy/$1.git
  git checkout dev
  npm i
  npm link

  cd ..
}

wallet-core_links() {
  cd wallet-core

  npm link @lumerin/contracts

  cd ..
}

wallet-desktop_links() {
  cd wallet-desktop

  npm link @lumerin/wallet-core
}
