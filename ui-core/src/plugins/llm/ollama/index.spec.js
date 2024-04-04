import sinon from 'sinon'
import ollama from './index'
import { expect } from 'chai'

describe('test ollama', () => {
  describe('api integration', () => {
    before('should init with config', () => {
      ollama.init({
        modelUrl: 'http://localhost:11434',
        modelName: null,
      });

      expect(ollama.chat.createChatCompletion).is.a('function')
    })

    it('should create chat completion', async () => {
      const chat = []
      const message = 'how are you?'
      const response = await ollama.chat.createChatCompletion(chat, message)
      expect(response.messages[0].content).toContain("I'm doing well")
    })
  })

  describe.skip('unit', () => {
    let request = null, response = null;

    before(async () => {
      //mock axios
      request = sinon.stub(axios, 'post')
      response = {
        data: {
          messages: [
            {
              role: 'bot',
              content: "I'm doing well",
            },
          ],
        },
      }
      request.returns(Promise.resolve(response))
    })

    after(async () => {
      request.restore()
    })

    it('should throw error on invalid config', async () => {
      await expect(() => ollama.init({})).toThrow()
    })

    it('should set default model name', async () => {
      ollama.init({
        modelUrl: 'http://localhost:11434',
      })

      expect(ollama.chat.createChatCompletion).toBeDefined()
      expect(ollama.modelName).toEqual('llama2:70b')
    })

    it('should create chat completion', async () => {
      const chat = []
      const message = 'how are you?'
      const response = await ollama.chat.createChatCompletion(chat, message)
      expect(response.messages[0].content).toContain("I'm doing well")
    })
  })
})
