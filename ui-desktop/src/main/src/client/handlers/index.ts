import singleCore from './single-core'
import noCore from './no-core'

const handlers = Object.assign({}, singleCore, noCore)

export default handlers
export type Handlers = typeof handlers
