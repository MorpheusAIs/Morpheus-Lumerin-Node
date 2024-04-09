const sinon = require('sinon')
const axios = require('axios')
const ollama = require('./index')
const chai = require('chai')

const { expect } = chai
const modelName = "llama2:latest"
// TODO: fix other tests so they can be run reliably
describe.only('test ollama', function () {
  this.timeout(15000)
  describe('api integration', () => {
    before('should init with config', () => {
      ollama.init({
        modelUrl: 'http://localhost:11435',
        modelName: modelName,
      })

      expect(ollama.chat.createChatCompletion).is.a('function')
    })

    it('should create chat completion', async () => {
      const chat = []
      const message = 'how are you?'
      const response = await ollama.chat.createChatCompletion(chat, message)

      // concatenate the values of all of the content fields in all of the response elements
      const contents = response.map(({ message }) => message.content)

      expect(contents.join('')).to.contain("I'm just an AI")
    })
  })

  describe('unit', () => {
    let request = null,
      response = null

    before(async () => {
      //mock axios
      request = sinon.stub(axios, 'post');
      response = {
        data: '{"model":"llama2:latest","created_at":"2024-04-08T02:05:27.47630621Z","message":{"role":"assistant","content":"I"},"done":false}\n' +
        `{"model":"llama2:latest","created_at":"2024-04-08T02:05:27.587310627Z","message":{"role":"assistant","content":"'"},"done":false}\n` +
        '{"model":"llama2:latest","created_at":"2024-04-08T02:05:27.698277877Z","message":{"role":"assistant","content":"m"},"done":false}\n' +
        '{"model":"llama2:latest","created_at":"2024-04-08T02:05:27.813684419Z","message":{"role":"assistant","content":" just"},"done":false}\n' +
        '{"model":"llama2:latest","created_at":"2024-04-08T02:05:27.925306794Z","message":{"role":"assistant","content":" an"},"done":false}\n' +
        '{"model":"llama2:latest","created_at":"2024-04-08T02:05:28.037869294Z","message":{"role":"assistant","content":" A"},"done":false}\n' +
        '{"model":"llama2:latest","created_at":"2024-04-08T02:05:28.151919127Z","message":{"role":"assistant","content":"I"},"done":false}\n'
      };

      request.returns(Promise.resolve(response))
    })

    after(async () => {
      request.restore()
    })

    it.skip('should throw error on invalid config', async () => {
      await expect(() => ollama.init({})).throws()
    })

    it('should set default model name', async () => {
      ollama.init({
        modelUrl: 'http://localhost:11434',
        modelName
      })

      expect(ollama.chat.createChatCompletion).is.a('function')
      expect(ollama.modelName).to.equal('llama2:latest')
    })

    it('should create chat completion', async () => {
      const chat = []
      const message = 'how are you?'
      const response = await ollama.chat.createChatCompletion(chat, message)

      const contents = response.map(({ message }) => message.content)

      expect(contents.join('')).to.contain("I'm just an AI")
    })
  })
})
