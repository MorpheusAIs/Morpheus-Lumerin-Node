# User Compute Note Setup via AWS

### Purpose
This guide has been developed to help setup your Morpheus compute node within AWS using EC2. This guide will help you setup an LLM model within AWS EC2, and connect to the Morpehus Proxy-Router to provide compute through the Morpheus ecosystem. Please note that the Morpheus Compute Ecosystem is still in testnet, and this guide is serving as a first draft, which will be modified for optimization and to address feedback.

## Setting up your LLM Instance 

1.	Sign into AWS and Open AWS Console
2.	Go to EC2 Dashboard and Select, Launch Instances
3.	Name the instance and choose “Amazon Linux”

 ![Instance Setup](https://github.com/user-attachments/assets/d9dd288c-d594-4584-9185-b80628756ff7)

4.	Choose your instance type: THIS IS A KEY STEP. The instance type you choose will directly impact which model you can run and the associated performance of that model. For this testing, I will use the TinyLlama model with m5.xlarge, larger instances will be needed for larger models (this is for testing). 
 
 ![Instance Selection](https://github.com/user-attachments/assets/e3752a2f-88a1-4c72-81d7-6fd175bad893)

5.	Select your Key Pair, or create one if you haven’t
 
 ![Key Pair Selection](https://github.com/user-attachments/assets/700b2bdc-efe9-435a-8e1f-f022a1e3ec6e)


6.	Create your security group: This step is critical for security. You will need to specify the ports that will be opened and who they can be accessed by. For the communication between the LLM and Proxy Router to work properly, you need to enable inbound traffic from port range 8080-8082. Additionally, to connect to your instance you need to enable traffic on port 20 for your IP. Outbound traffic can remain open, depending on your personal security preferences. 
 
 <img width="1012" alt="Security Groups" src="https://github.com/user-attachments/assets/b105da85-9664-4f70-80ed-3b04b3bc8b33">

7.	Configure Storage: Choose the storage space needed for the model
 
![Configure Storage](https://github.com/user-attachments/assets/218cd1ed-c26e-4fd3-8187-5121387b5dc7)

8.	Click Launch Instance
9.	Go back to the dashboard and click on your new instance
10.	Click “Connect” and choose “Connect using EC2 Instance Connect

<img width="814" alt="Connect to Instance" src="https://github.com/user-attachments/assets/a3e4b50a-30da-4331-acd2-a6c16d843e23">

### Configuring your Instance

1.	Once the terminal windows opens, you will need to complete some configuration to run the model. Run the following commands, one at a time
2.	 `sudo yum install git -y`
3.	 `sudo yum install golang-go -y`
4.	`sudo yum groupinstall "Development Tools" -y`
5.	`git clone https://github.com/ggerganov/llama.cpp.git
cd llama.cpp 
make -j 8`

Download and host the model
Now you are ready to install your model and host it. You will need a few key pieces of information for this. First, you will need your Public IPv4 DNS name and the host you will be running on (we are using 8080 as default). Next, you will need the model_url, model_collection, and model_file_name, which you can find from a Huggingface directory (must be GGUF). We will be using Tinyllama for testing with the information below.

Public IPv4 DNS from EC2 Dashboard = model_host

![IPv4 DNS](https://github.com/user-attachments/assets/c2d0b3c3-8d85-4b2a-854a-179336f4dd10)

Model on Hugging Face: https://huggingface.co/TheBloke/TinyLlama-1.1B-Chat-v1.0-GGUF/blob/main/tinyllama-1.1b-chat-v1.0.Q4_K_M.gguf
Model URL, Collection and File Name extracted for below

1.	Within the terminal, ensure you are in the llama director with the following command:
2.	 `cp llama.cpp`
3.	Download and host your model. You need to edit the parameters from the code below with your own information mentioned above. This will take some time to run
4.	 `model_host=ec2-3-147-83-19.us-east-2.compute.amazonaws.com
model_port=8080
model_url=https://huggingface.co/TheBloke
model_collection=TinyLlama-1.1B-Chat-v1.0-GGUF
model_file_name=tinyllama-1.1b-chat-v1.0.Q4_K_M.gguf
wget -O models/${model_file_name} ${model_url}/${model_collection}/resolve/main/${model_file_name}  
./llama-server -m models/${model_file_name} --host ${model_host} --port ${model_port} --n-gpu-layers 4096`
5.	You will get the following response when complete
 
 ![LLM Code Compelte](https://github.com/user-attachments/assets/7bdd6d2a-159b-4c9c-9d13-b9ec098e42ad)

6.	To check that your host is running correctly, go to a web browser and type in http://<model_host>:<model_port>
7.	You will see the llama.cpp UI and can interact with your model through the web browser

 ![Llama UI](https://github.com/user-attachments/assets/d2d689dd-de33-4e99-95fb-012f354bd3c2)

## Setting up the Proxy-Router

1.	Go back to the EC2 dashboard and create a new instance
2.	Follow the same setup instructions as before, except you can use the smallest instance type available for the proxy router
3.	Launch the instance and connect through EC2 Instance Connect

### Configuring your Instance
Now that your instance is running, you will need to do some configuration to ensure it connects properly to the LLM instance, and can interact on behalf of your BASE wallet. From your instance, you will need the model_host and model_port. From your wallet you will need your private key. You will also need an ETH node WSS address, which can be obtained through Alchemy or a similar website (for testing we will use BASE Sepolia node in wss format). Lastly, you will need to choose the port for the proxy-router to run on. 

1.	Once the terminal windows opens, you will need to complete some configuration to run the model. Run the following commands, one at a time
2.	 `sudo yum install git -y`
3.	 `sudo yum install golang-go -y`
4.	`sudo yum groupinstall "Development Tools" -y`
5.	`git clone https://github.com/MorpheusAIs/Morpheus-Lumerin-Node.git`
6.	 `cd Morpheus-Lumerin-Node/proxy-router`
7.	 `cp .env.example .env`
8.	 `vi .env`
9.	 Within the env file, you will be modifying a few key fields:
 
WALLET_PRIVATE_KEY from the wallet you will be using
ETH_NODE_ADDRESS from Alchemy or similar
WEB_ADDRESS for where this will be hosted
WEB_PUBLIC_URL for where this will be hosted
OPENAI_BASE_URL from the LLM model instance DNS

10.	Type `:wq` to save and quit
11.	Run the last file to build and launch the proxy-router
12.	 `./build.sh 
make run`
13.	Go to a web browser and type: http://<web_public_url>/swagger/index.html and you will be brought to the proxy-router UI

Congratulations, you have now setup your LLM and Proxy Router. There are more steps to conduct transactions within proxy router and setup your LLM as a provider, which are within this guide https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/blob/consumer-readme/docs/proxy-router-api-direct.md

