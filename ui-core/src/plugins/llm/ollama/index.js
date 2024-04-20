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
      // const response = await axios.post('/v1/chat/completions', {
      //   model: modelName,
      //   messages: [
      //     ...chat,
      //     {
      //       role: 'user',
      //       content: message,
      //     },
      //   ],
      // })

      // console.log('response.data: ', response.data)
      // return response.data.choices
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
    console.log('chat completion: ', chatCompletion) // {id: "…", choices: […], …}
    return chatCompletion.choices
  } catch (error) {
    console.log('chat completion error: ', error)
    throw error
  }
}

function jsonStringToArray(jsonString) {
  // Split the input string by newlines to get an array of strings, each representing a JSON object
  const lines = jsonString.trim().split('\n')
  // Map over each line, parsing it as JSON, and return the resulting array of objects
  const jsonArray = lines.map((line) => JSON.parse(line))

  return jsonArray
}
module.exports = {
  init,
  chat,
  modelName,
}
