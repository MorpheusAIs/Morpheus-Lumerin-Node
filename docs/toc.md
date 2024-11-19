
Document outline:
1. Design & Expectations 
    * How is the proxy-router supposed to be leveraged? 
    * Where does it fall in the larger design for Morpheus? 
    * Pictures to illustrate flow of interaction and setting expecations
    * How do you know it's working? (from which perspective?)

1. Installation 
    1. From Binaries, most simplistic to get up and running [install_from_binaries.md](install_from_binaries.md)
        * Need ETH node solution (Http round robin) from Shev 
        * Create new Wallet within MorpheusUI 
        * Send aETH and aMOR to wallet (from separate wallet or mechanism)
        * Variations for Mac ARM, Mac Intel, Linux, Windows
        * Needs to be as easy and as simple as possible (minimal clicks/configuration)
        * Both "private provider" and "consumer" 

    2. From Source, more complex but more control [install_from_source.md](install_from_source.md)
        * Variants include 
            * Use existing "on-box" model
            * Separate Model hosted elsewhere (with accessible endpoint) 
            * proxy-router node with private API access and publicly accessible router port 
            * directions on how to use the Swagger API / CLI / CURL 
            * Asssume separate eth node WSS subscription for proxy-router 
            * Assume using existing Wallet (with access to private key)
            * Full control of proxy-router as core (tying blockchain-contracts, provider, model, bid, etc between AI Compute and Consumers)
1. Utilization 
    * "Done-Consumer" looks like 
        - download, 
        - unzip, 
        - allow binaries, 
        - run, 
        - accept, 
        - create, 
        - receive (*may need separate doc for getting tokens/eth..especailly on sa), 
        - chat local, 
        - chat remote (existing providers)
    * "Done-Personal-Provider" looks like 
        - "Done-Consumer" plus 
        - authorize contract, 
        - create provider, 
        - create model, 
        - create bid
        - validate that model is working (available on chain and accessible via proxy-router)
    * "Done-Enterprise-Provider" looks like
        - Assume existing Wallet (with privatekey)
        - Assume existing ETH Node WSS subscription 
        - Assume familiarity working with API/CLI/CURL and existing blockchain stuff 
        - Assume existing AI Compute Resource/Model that wants to be provided

