//@ts-check
import logger from '../../logger'

/*** Returns BTC network difficulty */
const getNetworkDifficulty = async () => {
  try {
    const baseUrl = 'https://blockchain.info'
    const response = await fetch(`${baseUrl}/q/getdifficulty`)

    if (!response.ok) {
      throw new Error(`HTTP error! Status: ${response.status}`)
    }

    const data = await response.text()
    return parseFloat(data)
  } catch (err) {
    logger.error('Failed to get network difficulty:', err)
  }
}

export { getNetworkDifficulty }
