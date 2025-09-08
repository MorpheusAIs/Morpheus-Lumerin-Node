## "Simple Run From Mac" - 4 terminal windows

### Overview
- This is a simple guide to get the Llama.cpp model, Lumerin proxy-router and MorpheusUI from source running on a Mac
- This is a guide for a developer or someone familiar with the codebase and build process
- Wallet: if you’re going to start with an existing wallet from something like MetaMask (recommended)…make sure it’s a tier 1, not a derived or secondary address. Desktop-ui recover from mnemonic won’t work properly if you try to recover a secondary address with the primary mnemonic.  
- Install & Configure OS-Specific Dependencies
  * git (https://git-scm.com/)
  * go (https://golang.org/) Version >= 1.22
  * node (https://nodejs.org/) Version >= 20
  * make (https://www.gnu.org/software/make/)
  * yarn (https://yarnpkg.com/)

- Four basic steps: 
    1. Clone, build select model and run local Llama.cpp model
    2. Clone the Morpheus-Lumerin-Node repo from Github 
    3. Configure, Build and run the proxy-router
    4. Configure, Build and run the MorpheusUI
    5. Configure, Build and run the cli

## **A. LLAMA.CPP**
* Open first terminal / cli window 
* You will need to know what port you want the local mode to listen on (8080 in this example)
* This work should only need to be done once as it provides a local model for validation of the local proxy-router

**1. Clone LLamaCPP Repo & Build**
```sh
git clone https://github.com/ggerganov/llama.cpp.git
cd llama.cpp
make -j 8
```

**2. Set General Variables:** 
* In same CLI window set the variables that the model command line will use (or you can build out a .env file with these and the others below)
  ```sh
  model_host=127.0.0.1
  model_port=8080
  ```

**3. Set Model Specific Variables:** (pick one of the variable blocks below)
* https://huggingface.co/TheBloke
  * Llama-2-7B-Chat-GGUF / llama-2-7b-chat.Q5_K_M.gguf (7.28GB RAM required)
    ```
    model_url=https://huggingface.co/TheBloke
    model_collection=Llama-2-7B-Chat-GGUF
    model_file_name=llama-2-7b-chat.Q5_K_M.gguf
    ```
  * CollectiveCognition-v1.1-Mistral-7B-GGUF / collectivecognition-v1.1-mistral-7b.Q5_K_M.gguf
    ```
    model_url=https://huggingface.co/TheBloke
    model_collection=CollectiveCognition-v1.1-Mistral-7B-GGUF
    model_file_name=collectivecognition-v1.1-mistral-7b.Q5_K_M.gguf
    ```
  * TinyLlama-1.1B-Chat-v1.0-GGUF / tinyllama-1.1b-chat-v1.0.Q4_K_M.gguf
    ```
    model_url=https://huggingface.co/TheBloke
    model_collection=TinyLlama-1.1B-Chat-v1.0-GGUF
    model_file_name=tinyllama-1.1b-chat-v1.0.Q4_K_M.gguf
    ```
  * CapybaraHermes-2.5-Mistral-7B-GGUF / capybarahermes-2.5-mistral-7b.Q5_K_M.gguf
    ```
    model_url=https://huggingface.co/TheBloke
    model_collection=CapybaraHermes-2.5-Mistral-7B-GGUF
    model_file_name=capybarahermes-2.5-mistral-7b.Q5_K_M.gguf
    ```


**4. Download and run the model using the variables set above:**
```sh
wget -O models/${model_file_name} ${model_url}/${model_collection}/resolve/main/${model_file_name}  
./llama-server -m models/${model_file_name} --host ${model_host} --port ${model_port} --n-gpu-layers 4096
```

* OPTIONAL To leave this running in the background (and tail -f nohup.out to monitor)
```sh 
wget -O models/${model_file_name} ${model_url}/${model_collection}/resolve/main/${model_file_name} 
nohup ./llama-server -m models/${model_file_name} --host ${model_host} --port ${model_port} --n-gpu-layers 4096 &
tail -f nohup.out
```

**5. Validate that the local model is running:** 
* Navigate to `http://127.0.0.1:8080` in a browser to see the model interface and test inferrence 
* You should also see (in the terminal window) the interaction with the model and the responses

## **B. PROXY-ROUTER**
* Open second terminal / cli window 
* You will need to know 
  * Your Wallet private-key **DON'T EVER SHARE WITH ANYONE** for use in the .env file 
  * Your ETH node wss address (eg: `wss://rinkeby.infura.io/ws/v3/your_infura_project_id`) for use in the .env file (we are currently working on a default, public leverage of http providers, but wss is most reliable at this point) 
  * TCP port you want the proxy-router to listen on (3333 in this example)
  * TCP port that you want the API to listen on (8082 in this example)

**1.  Clone repo from either the Lumerin Team fork or the Origin on Morpheus' Github**  
- MorpheusAIs Team Development Github: `git clone https://github.com/MorpheusAIs/Morpheus-Lumerin-Node.git`
- Origin from Morpheus Github: (`git clone https://github.com/MorpheusAIs/Morpheus-Lumerin-Node.git`)
- **NOTE:** if you have already cloned and configured your .env, and want to update to the lastest revision, you can `git pull` from within the Morpheus-Lumerin-Node directory and then skip to step 4 below to build and run the updated proxy-router

**2. Navigate to the proxy-router directory** `cd <your_path>/Morpheus-Lumerin-Node/proxy-router`

**3. Modify proxy-router environment variables**
- You will need WalletPrivateKey, WSS Eth Address and router and api listening ports
- Within the proxy-router directory, copy the `.env.example` to `.env` and edit to set the variables as needed (prompts are within the file)
```sh
cp .env.example .env
vi .env 
```
**4. Build and Run the proxy-router**
```sh
./build.sh 
make run
```
**- NOTE:** that when this launches, some anti-virus or network protection software may ask you to allow the ports to be open as well as MAC-OS firewall proteciton (watch for the pop-up windows and allow)

**5. Validate that the proxy-router is running:**
- In the terminal window where you started the proxy-router the following messaging ensures that you're running and watching events on the blockchain 
```sh
2024-08-07T11:35:49.116184	INFO	proxy state: running
2024-08-07T11:35:49.116534	INFO	Wallet address: <your wallet address 0x.....>
2024-08-07T11:35:49.116652	INFO	started watching events, address 0xb8C55cD613af947E73E262F0d3C54b7211Af16CF
2024-08-07T11:35:49.116924	INFO	HTTP	http server is listening: 0.0.0.0:8082
2024-08-07T11:35:49.116962	INFO	TCP	tcp server is listening: 0.0.0.0:3333
```
- Navigate to `http://localhost:8082/swagger/index.html` in a browser to see the proxy-router interface and test the Swagger API

**- NOTE** if you would like to interact directly with your proxy-router without the UI, see the instructions in [/docs/proxy-router-api-direct.md](/docs/proxy-router-api-direct.md)

## **C. MorpheusUI**
* Open third terminal / cli window 
* You will need to know 
  * TCP port that your proxy-router API interface is listening on (8082 in this example)

**1. Navigate to MorpheusUI**
`cd <your_path>/Morpheus-Lumerin-Node/MorpheusUI`

**2. Check Environment Variables**
- Within the MorpheusUI directory, copy the `.env.example` to `.env` and check the variables as needed 
- At the current time, the defaults in the .env file shold be sufficient to operate as long as none of the LLAMA.cpp or proxy-router variables for ports have changed. 
- Double check that your PROXY_WEB_URL is set to `8082` or what you used for the ASwagger API interface on the proxy-router. 
- This is what enables the UI to communicate to the proxy-router environment 

```sh
cp .env.example .env
vi .env 
```

**3. Install dependicies, compile and Run the MorpheusUI**
```sh
yarn install
yarn dev
```

**4. Validate that the MorpheusUI is running:**
- At this point, the electon app should start and if this is the first time walking through, should run through onboarding 
- Check the following: 
  - Lower left corner - is this your correct ERC20 Wallet Address? 
  - On the Wallet tab, do you show MOR and ETH balances? 
    - if not, you'll need to convert and transfer some ETH to MOR to get started
  - On the Chat tab, your should see your `Provider: (local)` selected by default...use the "Ask me anything..." prompt to run inference - this verifies that all three layers have been setup correctly?
  - Click the Change Model, select a remote model, click Start and now you should be able to interact with the model through the UI. 

**Cleaning & Troubleshooting**
- Sometimes, due to development changes or dependency issues, you may need to clean and start fresh.  
- Make sure you know what your WalletPrivate key is and have it saved in a secure location. before cleaning up and restarting the ui or proxy-router
  - `rm -rf  ./node_modules` from within MorpheusUI 
  - `rm -rf '~/Library/Application Support/MorpheusUI'` to clean old MorpheusUI cache and start new wallet
- The proxy-router or MorpheusUI may not start cleanly because of existing processes or open files from previous runs of the software  
  - If you need to exit either proxy-router or MorpheusUI, you can use `ctrl-c` from the terminal window you started them to kill the processes…
  - **Locked Processes** Doing this may leave "dangling" processes or open files to find them: 
    - Run `ps -ax | grep electron` or `ps -ax | grep proxy-router` which will show you the processes and particualrly the procexss id
    - Run `kill -9 xxxxx` where `xxxxx` is the process ID of the un-cleaned process  
  - **Locked Files** In certain cases the proxy-router will not let go of the `./data/` logging files… to clean them: 
    - Run `lsof | grep /proxy-router/data/` to list the current open files in that directory, they should all have a common process ID (Second column on the screen)
    - Run `kill -9 xxxxxx` the process id that matches to clean it all up

## **D. CLI**
* Open 4th terminal / cli window 
* You will need to know 
  * TCP port that your proxy-router API interface is listening on (8082 in this example)

**1. Navigate to MorpheusUI**
`cd <your_path>/Morpheus-Lumerin-Node/cli`

**2. Check Environment Variables**
- Within the cli directory, copy the `.env.example` to `.env` and check the variables as needed 
- At the current time, the defaults in the .env file shold be sufficient to operate as long as proxy-router variables for ports have changed. 
- Double check that your API_HOST is set to `http://localhost:8082` or what you used for the proxy-router. 
- This is what enables the CLI to communicate to the proxy-router environment 

```sh
cp .env.example .env
vi .env 
```

**3. Build CLI**
```sh
make build
```

**4. Validate that the cli is working:**
```sh
./mor-cli -h
```

**5. Validate that the cli is connected to proxy-router:**
```sh
./mor-cli healthcheck
```
```sh
{"Status":"healthy","Uptime":"18s","Version":"TO BE SET AT BUILD TIME"}
```