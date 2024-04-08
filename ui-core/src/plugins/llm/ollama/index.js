const axios = require('axios')

let modelName = 'llama2:latest',
  modelUrl = ''

function init(config) {
  axios.defaults.baseURL = modelUrl = config.modelUrl
  modelName = config.modelName || modelName
}

const chat = {
  //TODO: map images between ollama api and openai api
  async createChatCompletion(chat, message) {
    try {
      const response = await axios.post('/api/chat', {
        model: modelName,
        messages: [
          ...chat,
          {
            role: 'user',
            content: message,
          },
        ],
      })
      
      return jsonStringToArray(response.data)
    } catch (error) {
      console.log(error)
      throw error
    }
  },
}
function jsonStringToArray(jsonString) {
  // Split the input string by newlines to get an array of strings, each representing a JSON object
  const lines = jsonString.trim().split('\n');
  // Map over each line, parsing it as JSON, and return the resulting array of objects
  const jsonArray = lines.map(line => JSON.parse(line));
  
  return jsonArray;
}
module.exports = {
  init,
  chat,
  modelName
}
