### Consumer from Release (Local LLM, Proxy-Router & MorpheusUI): 

This is the simplest way to get started with the Morpheus Lumerin Node as a Consumer.  This will allow you to interact with the Morpheus network and the models offered on the network as a consumer.

It will run 3 different pieces of software on your local machine:
* `llama.cpp` (llama-server) - a simple sample AI model that can be run on the same machine as the proxy-router and MorpheusUI to show how the components work together and run local (free) inference
* `proxy-router` - the same software as the provider side, but with different environment variables and a different role
* `MorpheusUI` - Electron GUI that enables the user to interact with the models (via the API) to browse offered bids, purchase and send prompts

## Installation Steps:
1. Obtain the software: 
    1. Package: Download latest release for your operating system: https://github.com/Lumerin-protocol/Morpheus-Lumerin-Node/releases
        * Mainnet releases will be prefixed with `main-*`
        * Testnet releases will be prefixed with `test-*` 

1. Extract the zip to a local folder (examples)
    * Windows: `(%USERPROFILE%)/Downloads/morpheus)` 
    * Linux & MacOS: `~/Downloads/morpheus`
    * On MacOS you may need to execute `xattr -c mor-launch proxy-router MorpheusUI.app llama-server` in a command window to remove the quarantine flag on MacOS

1. Launch the node - this should open a command window to see local LLM model server and proxy-router start and then should launch the user interface  
    * Windows: Double click the `mor-launch.exe` (You will need to tell Windows Defender this is ok to run) 
    * Linux & MacOS: Open a terminal and navigate to the folder and run `./mor-launch`

1. Startup User Interface: 
    1. Read & accept terms & Conditions 
    1. Set a strong password (this is for the MorpheusUI only)
    1. Follow the instructions for creating a new wallet (be sure to save the mnemonic in a safe place)
    1. **OPTIONAL to use existing Wallet** 
        - Instead of creating an new wallet and if you have the existing wallet's mnemonic, when prompted, select **`Recover your wallet Saved Mnemonic`** instead.

5. Startup CLI Interface (Optional): 
    * Linux & MacOS: Open a terminal and navigate to the folder and run `./mor-cli`

## Validation Steps:
1. Local Test: Once the UI is up and running,
    1. You should see tokens for saETH and saMOR that you sent to this wallet earlier. 
        * If this is a new wallet, you will need to send saMOR and saETH to this wallet to be able to interact with the blockchain 
        * This can be done externally via metamask or usual Arbitrum testnet faucets
    1. Once you have a funded Wallet, you can interact with the local model
    1. Click on the `Chat` icon on the left side of the screen
    1. Make sure the `Local Model` is selected
    1. Begin your conversation with the model by typing in the chat window and pressing `Enter`
        * You should see the model respond with the appropriate response to the prompt you entered, if not, there may be an issue with the local model service

1. Remote Test: Once you've verified that your wallet can access the blockchain and you can see the local model working, you can switch to a remote model and test that as well
    1. In the `Chat` window, select `Change Model `
        1. Select a different model from remote providers
        1. DropDown and select the contract address of the model you want to use 
        1. Click Change 
        1. Click Open Session 
        1. MUST Enter at least **5** MOR to open session 
    1. You can now chat with the remote model and see the responses in the chat window 

1. Cleanup/Closeout 
    * Manually End all Remote Sessions: 
        * In the Chat Window, click on the Time icon to the right of the Model line - this will expand and show current sessions, click the "X" next to each one to make sure it's closed 
    * Closing the MorpheusUI window should leave the CMD window open
        * You’ll have to ctrl-c in the window to kill the local model and proxy-router
    * To FULLY delete and force a clean startup of the UI (including forcing new password and mnemonic recovery), delete the MorpheusUI folder and start the UI again
        * Windows:  `%USERPROFILE%\AppData\Roaming\MorpheusUI`
        * Linux: `~/.config/MorpheusUI`
        * MacOS: `~/Library/Application Support/MorpheusUI`
