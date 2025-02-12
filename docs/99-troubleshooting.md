# Troubleshooting guides for proxy-router and MorpheusUI desktop applications
* These are some common scenarios and failure modes to watch out for when running the proxy-router and MorpheusUI desktop applications.
* One of the most critical things to remember, especially about the desktop application (MorpheusUI) or the Swagger API (http://localhost:8082/swagger/index.html) is that you **must** have a good startup of the proxy-router.
    * If the proxy-router is not running or is not able to connect to the blockchain node, the desktop application will not be able to talk to the blockchain, manage your wallet or send prompts to the provider.
    * The proxy-router is the bridge between the blockchain node and the desktop application. It is responsible for routing prompts and responses between the consumer and provider.
* One of the best ways to observe the health of the proxy-router is to start it (either `./proxy-router` or `./mor-launch` for the release) from a terminal or command line session where you can see the output in real-time.  
* There are also many options for adding log destination and details of log entries (please see [proxy-router.all.env](./proxy-router.all.env) #Logging Configuration section for more information.

## Proxy-Router
### Proxy-Router is not starting
* **Expected Result or "How do I know it's good?":** 
    * The proxy-router should start successfully and be able to connect to the blockchain node and listening on both 8082 and 3333 ports (these are the default ports)
    * In the terminal or logs after startup, you should see the following messages with regard to the proxy-router (there will be other messages as well): 
    ```
    INFO	HTTP	http server is listening: 0.0.0.0:8082
    INFO	TCP	    tcp server is listening: 0.0.0.0:3333
    ```
    * The other way to verify that the proxy-router has started is to query the Swagger API at http://localhost:8082/swagger/index.html (or whatever url you've set in the .env file for your proxy-router IP or DNS address) and see the API documentation.
    
* **Symptoms:** The proxy-router is not starting or is crashing immediately after starting.
* **Possible Causes:**
    * **.env file misconfiguration** (e.g. missing or incorrect values)
        * These four items MUST be accurate to the chain and your OS., Use the example files `/proxy-router/env.main.example` to make sure you have the correct values: 
            * `DIAMOND_CONTRACT_ADDRESS=`
            * `MOR_TOKEN_ADDRESS=`
            * `BLOCKSCOUT_API_URL=`
            * `ETH_NODE_CHAIN_ID=`
            * `PROXY_STORAGE_PATH=`
        * If you are running the proxy-router by itself (without the UI), you will need to set the private key of your provider wallet in the  `WALLET_PRIVATE_KEY=`
        * **IF** you use your own ETH node (Alchemy, Infura, etc.) to communicate with the blockchain, you will need to make sure that the `ETH_NODE_URL=` entry in the .env file is correct for your chain.  
            * We recommend https:// instead of wss:// for the ETH_NODE_URL ...which also means that `ETH_NODE_USE_SUBSCRIPTIONS=false` should be set in the .env file
   
    * **models-config.json misconfiguration** (e.g. missing or incorrect values)
        * The `models-config.json` file is used to direct the proxy to find the models that the proxy-router will use to route prompts and responses between consumers and providers.
        * Ensure that: 
            * the `MODELS_CONFIG_PATH=` in your .env file is correct and points to the correct directory and file and has the right permissions 
            * the `models-config.json` should follow the formatting shown in the example file [models-config.json.md](./models-config.json.md)
* **Resolution:**
    * Check the .env and models-config.json file configuration and ensure that the blockchain node is accessible.
    * Restart the proxy-router and check the logs for any errors related to the connection.
    