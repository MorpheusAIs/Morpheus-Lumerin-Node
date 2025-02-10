export type AgentUserRes = {
  agents: AgentUser[]
}

export type AgentUser = {
  allowances: Record<'ETH' | string, string>
  isConfirmed: boolean
  password: string
  perms: string[]
  username: string
}

export type AgentTxRes = {
  txHashes: string[]
  nextCursor: string
}

export type AgentAllowanceRequestsRes = {
  requests: AgentAllowanceRequest[]
}

export type AgentAllowanceRequest = {
  username: string
  token: string
  allowance: string
}

export type ChatHistoryTitle = {
  title: string
  chatId: string[]
}

export type ChatHistory = {
  title: string
  modelId: string
  sessionId: string
  messages: ChatMessage[]
}

export type ChatMessage = {
  response: string
  prompt: string
  promptAt: number
  responseAt: number
}
