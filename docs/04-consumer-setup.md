### Consumer from Release (Local LLM, Proxy-Router & MorpheusUI): 

This is the simplest way to get started with the Morpheus Lumerin Node as a Consumer.  This will allow you to interact with the Morpheus network and the models offered on the network as a consumer.

It will run 3 different pieces of software on your local machine:
* `llama.cpp` (llama-server) - a simple sample AI model that can be run on the same machine as the proxy-router and MorpheusUI to show how the components work together and run local (free) inference
* `proxy-router` - the same software as the provider side, but with different environment variables and a different role
* `MorpheusUI` - Electron GUI that enables the user to interact with the models (via the API) to browse offered bids, purchase and send prompts

## Installation Steps:
1. Obtain the software: 
    * Package: Download latest release for your operating system: https://github.com/Lumerin-protocol/Morpheus-Lumerin-Node/releases
        - Mainnet releases will have no suffix (eg: `v2.3.0`)
        - Testnet releases will be suffixed with `*-test` (eg: `v2.2.56-test`) 

1. Extract the zip to a local folder (examples)
    * Windows: `(%USERPROFILE%)/Downloads/morpheus)` 
    * Linux & MacOS: `~/Downloads/morpheus`
    * On MacOS you will need to execute `xattr -c mor-launch mor-cli proxy-router MorpheusUI.app llama-server` in a command window to remove the quarantine flag on MacOS

1. [OPTIONAL] Examine the .env file included in the package and make any necessary updates for your setup 
    * See [proxy-router.all.env](./proxy-router.all.env) for a list of environment variables that can be set in the `.env` file

1. Launch the node - this should open a command window to see local LLM model server and proxy-router start and then should launch the user interface  
    * Windows: Double click the `mor-launch.exe` (You will need to tell Windows Defender this is ok to run) 
    * Linux & MacOS: Open a terminal and navigate to the folder and run `./mor-launch`

1. Startup User Interface: 
    * Read & accept terms & Conditions 
    * Set a strong password (this is for the MorpheusUI only)
    * Follow the instructions for creating a new wallet (be sure to save the mnemonic in a safe place)
    * **OPTIONAL to use existing Wallet** 
        - Instead of creating an new wallet and if you have the existing wallet's mnemonic, when prompted, select **`Recover your wallet Saved Mnemonic`** instead.

1. Startup CLI Interface (Optional): 
    * Linux & MacOS: Open a terminal and navigate to the folder and run `./mor-cli`

## Validation Steps:
1. Local Test: Once the UI is up and running,
    * You should see tokens for ETH and MOR that you sent to this wallet earlier. 
        - If this is a new wallet, you will need to send MOR and ETH to this wallet to be able to interact with the blockchain 
        - This can be done externally via metamask or usual Arbitrum testnet faucets
    * Once you have a funded Wallet, you can interact with the local model
    * Click on the `Chat` icon on the left side of the screen
    * Make sure the `Local Model` is selected
    * Begin your conversation with the model by typing in the chat window and pressing `Enter`
        - You should see the model respond with the appropriate response to the prompt you entered, if not, there may be an issue with the local model service

1. Remote Test: Once you've verified that your wallet can access the blockchain and you can see the local model working, you can switch to a remote model and test that as well
    * In the `Chat` window, select `Change Model `
        - Select a different model from remote providers
        - DropDown and select the contract address of the model you want to use 
        - Click Change 
        - Click Open Session 
        - MUST Enter at least **5** MOR to open session 
    * You can now chat with the remote model and see the responses in the chat window 

1. Closeout 
    * Manually End all Remote Sessions: 
        - In the Chat Window, click on the Time icon to the right of the Model line - this will expand and show current sessions, click the "X" next to each one to make sure it's closed 
    * Closing the MorpheusUI window should leave the CMD window open
        - You’ll have to ctrl-c in the window to kill the local model and proxy-router
1. Cleanup
    * When testing multiple versions or if you get into an inconsistent state, you may need to clean up the local environment
        - **Ensure that you have (or saved) your `wallet mnemonic or private key` before proceeding!**
        - Close the MorpheusUI window
        - Close the CMD window (or CTRL+C to stop the processes)
        - Delete the following files/folders from your downloaded, unzipped Morpheus folder
        ``` bash
        # Mac and Ubuntu Commands: 
        rm .cookie proxy.conf
        rm -rf ~/Library/Logs/morpheus-ui
        rm -rf ~/Library/Application\ Support/morpheus-ui
        rm -rf data
        xattr -c MorpheusUI.app llama-server mor-launch proxy-router
        ```
    * Check these other locations for Morpheus related files and delete them if necessary
        - Windows:  `%USERPROFILE%\AppData\Roaming\morpheus-ui`
        - Linux: `~/.config/morpheus-ui`
        - MacOS: `~/Library/Application Support/morpheus-ui`
    * At this point, all stored is removed and you can start fresh with a new wallet or recover an existing wallet per instructions above
