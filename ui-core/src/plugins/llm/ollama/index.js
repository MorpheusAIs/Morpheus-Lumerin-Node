const axios = require('axios')

let modelName = 'llama2:70b',
  modelUrl = ''

module.exports.init = function init(config) {
  axios.defaults.baseURL = modelUrl = config.modelUrl
  modelName = config.modelName || modelName
}

export const chat = {
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

      return response.data
    } catch (error) {
      console.log(error)
      throw error
    }
  },
}

export default {
  chat,
}
