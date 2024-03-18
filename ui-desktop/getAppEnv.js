// Require the axios module to make HTTP requests
import axios from "axios";
import CryptoJS from 'crypto-js';

// replace with your GitHub username, repository, and personal access token
const github_username = "srt0422";
const github_org = "Lumerin-protocol";
const github_repository = "WalletDesktop";
const personal_access_token = "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";


// Set the base URL for the github api
const baseURL = "https://api.github.com";

// Set the authorization header with your personal access token
const authHeader = { Authorization: "token " + personal_access_token };

// Set the organization and repository names
const org = github_org;
const repo = github_repository;
// Set the environment name
const env = 'dev';
// Define a function to get the repository public key
async function getRepoPublicKey() {
  try {
    // Make a GET request to the repository public key endpoint
    const response = await axios.get(`${baseURL}/repos/${org}/${repo}/actions/secrets/public-key`, {
      headers: authHeader,
    });

    // Extract the public key and key id from the response data
    const publicKey = response.data.key;
    const keyId = response.data.key_id;

    // Return the public key and key id as an object
    return {publicKey, keyId};
  } catch (error) {
    // Handle any errors
    console.error(error);
  }
}

// Define a function to get and decrypt a secret value by its name and environment name
async function getAndDecryptSecret(secretName, envName) {
  try {
    // Get the repository public key using the previous function
    const {publicKey, keyId} = await getRepoPublicKey();

    // Make a GET request to the environment secret endpoint
    const response = await axios.get(`${baseURL}/repos/${org}/${repo}/environments/${envName}/secrets/${secretName}`, {
      headers: authHeader,
    });

    // Extract the encrypted value from the response data
    const encryptedValue = response.data.encrypted_value;

    // Convert the public key and the encrypted value from base64 to WordArray
    const publicKeyArray = CryptoJS.enc.Base64.parse(publicKey);
    const encryptedValueArray = CryptoJS.enc.Base64.parse(encryptedValue);

    // Decrypt the encrypted value using the public key and the crypto-js library
    const valueArray = CryptoJS.AES.decrypt(encryptedValueArray, publicKeyArray);

    // Convert the decrypted value from WordArray to string
    const value = valueArray.toString(CryptoJS.enc.Utf8);

    // Return the value as a string
    return value;
  } catch (error) {
    // Handle any errors
    console.error(error);
  }
}

// Define a function to get and decrypt all secrets in an array of objects
async function getAndDecryptAllSecrets(secrets) {
  try {
    // Create an empty array to store the decrypted secrets
    const decryptedSecrets = [];

    // Loop through each secret object in the secrets array
    for (const secret of secrets) {
      // Get the secret name from the object
      const name = secret.name;

      // Get and decrypt the secret value using the previous function
      const value = await getAndDecryptSecret(name);

      // Create a new object with the name and value of the secret
      const decryptedSecret = {name, value};

      // Push the new object to the decrypted secrets array
      decryptedSecrets.push(decryptedSecret);
    }

    // Return the decrypted secrets array
    return decryptedSecrets;
  } catch (error) {
    // Handle any errors
    console.error(error);
  }
}

// Define a function to merge an array of secrets with an array of variables
function mergeSecretsAndVariables(secrets, variables) {
  try {
    // Merge the two arrays using the spread operator
    const merged = [...secrets, ...variables];

    // Return the merged array
    return merged;
  } catch (error) {
    // Handle any errors
    console.error(error);
  }
}

// Define a function to get and merge all secrets and variables for a given environment name
async function getAndMergeAllSecretsAndVariables(envName) {
  try {
    // Make a GET request to the environment secrets endpoint
    const secretsResponse = await axios.get(`${baseURL}/repos/${org}/${repo}/environments/${envName}/secrets`, {
      headers: authHeader,
    });

    // Extract the secrets from the response data
    const secrets = secretsResponse.data.secrets;

    // Make a GET request to the environment variables endpoint
    const variablesResponse = await axios.get(`${baseURL}/repos/${org}/${repo}/environments/${envName}/variables`, {
      headers: authHeader,
    });

    // Extract the variables from the response data
    const variables = variablesResponse.data.variables;

    // Get and decrypt all secrets using the previous function
    const decryptedSecrets = await getAndDecryptAllSecrets(secrets);

    // Merge the decrypted secrets and the variables using the previous function
    const merged = mergeSecretsAndVariables(decryptedSecrets, variables);

    // Return the merged array
    return merged;
  } catch (error) {
    // Handle any errors
    console.error(error);
  }
}

// Call the function with the dev environment name and log the result
getAndMergeAllSecretsAndVariables('dev').then((result) => console.log(result));