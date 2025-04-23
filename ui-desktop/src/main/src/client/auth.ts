import { setSessionPassword, getSessionPassword } from './secure.js'
import logger from '../../logger'
import { getPasswordHash, setPasswordHash } from './settings'
import { pbkdf2 } from './crypto'

const { hash, verify } = pbkdf2

export async function setPassword(password) {
  await setSessionPassword(password)

  const passwordHash = getPasswordHash()
  if (!passwordHash) {
    logger.info('No password set, using current as default')
  }
  await hash(password).then(setPasswordHash)
}

export function isValidPassword(password) {
  const passwordHash = getPasswordHash()

  return verify(passwordHash, password)
    .then(async function (isValid) {
      if (isValid) {
        await setSessionPassword(password)
        logger.verbose('Supplied password is valid')
      } else {
        logger.warn('Supplied password is invalid')
      }
      return isValid
    })
    .catch(function (err) {
      logger.warn('Could not verify password', err)

      return false
    })
}

export async function getHashedPassword() {
  const hashString = await getPasswordHash()

  const hash = JSON.parse(hashString, (key, val) => (key == 'hash' ? JSON.parse(val) : val))

  return hash.hash.secret
}

const exportedMembers = {
  isValidPassword,
  setPassword,
  setSessionPassword,
  getSessionPassword,
  getHashedPassword
}

export default exportedMembers
