const axios = require('axios')
const OpenAI = require('openai')

let openai
let modelName = 'llama2:latest',
  modelUrl = ''

function init(config) {
  modelUrl = config.modelUrl
  modelName = config.modelName || modelName

  openai = new OpenAI({ baseUrl: modelUrl, apiKey: config.apiKey || '' })
}

const chat = {
  //TODO: map images between ollama api and openai api
  async createChatCompletion(chat, message) {
    try {
      return streamChatCompletions(chat, { role: 'user', content: message })
    } catch (error) {
      console.log(error)
      throw error
    }
  },
}

async function streamChatCompletions(chat, message) {
  try {
    const stream = await openai.beta.chat.completions.stream({
      model: modelName,
      messages: [...chat, message]
    });

    const chatCompletion = await stream.finalChatCompletion()
    
    return chatCompletion.choices
  } catch (error) {
    console.log('chat completion error: ', error)
    throw error
  }
}

module.exports = {
  init,
  chat,
  modelName,
}
