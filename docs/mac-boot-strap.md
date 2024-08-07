## "Simple Run From Mac" - 3 terminal windows

### Overview
- This is a simple guide to get the Llama.cpp model, Lumerin proxy-router and ui-desktop from source running on a Mac
- This is a guide for a developer or someone familiar with the codebase and build process
- Wallet: if you’re going to start with an existing wallet from something like MetaMask (recommended)…make sure it’s a tier 1, not a derived or secondary address. Desktop-ui recover from mnemonic won’t work properly if you try to recover a secondary address with the primary mnemonic.  
- Four basic steps: 
    1. Clone, build select model and run local Llama.cpp model
    2. Clone the Morpheus-Lumerin-Node repo from Github 
    3. Configure, Build and run the proxy-router
    4. Configure, Build and run the ui-desktop

## **A. LLAMA.CPP**
* Open first terminal / cli window 
* You will need to know what port you want the local mode to listen on (8080 in this example)

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
```
wget -O models/${model_file_name} ${model_url}/${model_collection}/resolve/main/${model_file_name}  
./llama-server -m models/${model_file_name} --host ${model_host} --port ${model_port} --n-gpu-layers 4096

# OPTIONAL: 
# To leave this running in the background (and tail -f nohup.out to monitor)
# nohup ./llama-server -m models/${model_file_name} --host ${model_host} --port ${model_port} --n-gpu-layers 4096 &
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
- Lumerin Team Development Github: `git clone https://github.com/Lumerin-protocol/Morpheus-Lumerin-Node.git`
- Origin from Morpheus Github: (`git clone https://github.com/MorpheusAIs/Morpheus-Lumerin-Node.git`)
- **NOTE:** if you have already cloned and configured your .env, and want to update to the lastest revision, you can `git pull` from within the Morpheus-Lumerin-Node directory and then skip to step 4 below to build and run the updated proxy-router

**2. Navigate to the proxy-router directory** `cd <your_path>/Morpheus-Lumerin-Node/proxy-router`

**3. Modify proxy-router environment variables**
- You will need WalletPrivateKey, WSS Eth Address and router and api listening ports
- Copy the `.env.example` to `.env` and edit to set the variables as needed (prompts are within the file)
```sh
cp .env.example .env
vi .env # or nano, or whatever editor you're comfortable with
```
**4. Build and Run the proxy-router**
```sh
./build.sh 
go run cmd/main.go
```
**- NOTE:** that when this launches, some anti-virus or network protection software may ask you to allow the ports to be open as well as MAC-OS firewall proteciton (watch for the pop-up windows and allow)

**5. Validate that the proxy-router is running:**
- In the terminal window where you started the proxy-router the following messaging ensures that you're running and watching events on the blockchain 
```sh
2024-08-07T11:35:49.116184	INFO	proxy state: running
2024-08-07T11:35:49.116534	INFO	Wallet address: <your wallet address 0x.....>
2024-08-07T11:35:49.116652	INFO	started watching events, address 0x8e19288d908b2d9F8D7C539c74C899808AC3dE45
2024-08-07T11:35:49.116924	INFO	HTTP	http server is listening: 0.0.0.0:8082
2024-08-07T11:35:49.116962	INFO	TCP	tcp server is listening: 0.0.0.0:3333
```
- Navigate to `http://localhost:8082` in a browser to see the proxy-router interface and test the Swagger API

**- NOTE** if you would like to interact directly with your proxy-router without the UI, see the instructions in [/docs/proxy-router-api-direct.md](/docs/proxy-router-api-direct.md)

## **C. UI-DESKTOP**
* Open third terminal / cli window 
* You will need to know 
  * TCP port that your proxy-router API interface is listening on (8082 in this example)

**1. Navigate to ui-desktop**
`cd <your_path>/Morpheus-Lumerin-Node/ui-desktop`

**2. Check Environment Variables**
- Copy the `.env.example` to `.env` and check the variables as needed 
- At the current time, the defaults in the .env file shold be sufficient to operate as long as none of the LLAMA.cpp or proxy-router variables for ports have changed. 
- Double check that your PROXY_WEB_URL is set to `8082` or what you used for the ASwagger API interface on the proxy-router. 
- This is what enables the UI to communicate to the proxy-router environment 

```sh
cp .env.example .env
vi .env # or nano, or whatever editor you're comfortable with
```

**3. Install dependicies, compile and Run the ui-desktop**
```sh
yarn install
yarn build
yarn dev
```

**4. Validate that the ui-desktop is running:**
- At this point, the electon app should start and if this is the first time walking through, should run through onboarding 
- Check the following: 
  - Lower left corner - is this your correct ERC20 Wallet Address? 
  - On the Wallet tab, do you show saMOR and saETH balances? 
    - if not, you'll need to convert and transfer some saETH to saMOR to get started
  - On the Chat tab, your should see your `Provider: (local)` selected by default...use the "Ask me anything..." prompt to run inference - this verifies that all three layers have been setup correctly?
  - Click the Change Model, select a remote model, click Start and now you should be able to interact with the model through the UI. 

**Cleaning**
* Sometimes, due to development changes or dependency issues, you may need to clean and start fresh.  Make sure you know what your WalletPrivate key is and have it saved in a secure location. before cleaningup and restarting the ui or proxy-router.
- `rm -rf  ./node_modules` from within ui-desktop 
- `rm -rf '~/Library/Application Support/ui-desktop'` to clean old ui-desktop cache and start new wallet


