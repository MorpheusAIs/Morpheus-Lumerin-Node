import React, { useEffect, useState } from 'react'
import { createRoot } from 'react-dom/client'
import { HashRouter, useLocation, useNavigate } from 'react-router-dom'
import { ThemeProvider } from 'styled-components'
import ReactModal from 'react-modal'
import 'bootstrap/dist/css/bootstrap.min.css'
import '../../src/renderer/src/ui/interface.css'
import './preview.css'
import theme from '../../src/renderer/src/ui/theme'
import Cowork from '../../src/renderer/src/components/cowork/Cowork'
import { Sidebar } from '../../src/renderer/src/components/sidebar/Sidebar'
import { Provider as ClientProvider } from '../../src/renderer/src/store/hocs/clientContext'
import { ToastsContext } from '../../src/renderer/src/components/toasts'
import QuickStartGuide, {
  openQuickStartGuide
} from '../../src/renderer/src/components/onboarding/QuickStartGuide'
import { ReceiveForm } from '../../src/renderer/src/components/dashboard/tx-modal/ReceiveForm'
import Modal from '../../src/renderer/src/components/common/Modal'
import { Btn } from '../../src/renderer/src/components/common/Btn'
import ModelSelectionModal from '../../src/renderer/src/components/chat/modals/ModelSelectionModal'
import { Agents } from '../../src/renderer/src/components/agents/Agents'
import type { TxModal } from '../../src/renderer/src/store/hocs/withAgentsState'

// Everything below is synthetic. Deliberately do not load App, its store, the
// Electron preload, .env, router credentials, filesystem tools, or real chats.
const params = new URLSearchParams(location.search)
let emptyProject = params.get('empty') === '1'
const now = Date.now()
const sampleAddress = '0x0000000000000000000000000000000000000000'
const policy = { schemaVersion: 1, id: 'workspace', mode: 'manual', revision: 1, updatedAt: now }
const project = {
  id: 'preview-project',
  name: 'September launch',
  folderName: 'sample-project',
  instructions: 'Keep the report concise and cite source files.',
  approvalMode: 'manual',
  createdAt: now,
  updatedAt: now
}
const replacement = {
  modelId: 'preview-model',
  modelName: 'Sample model with a deliberately long display name',
  sessionId: 'synthetic-session',
  sessionEndsAt: now + 86_400_000,
  isLocal: false,
  source: 'marketplace',
  dataBoundary: 'independent-provider',
  toolsSupported: true
}
const historicalModel = {
  ...replacement,
  modelId: 'previous-model',
  modelName: 'Previous sample model',
  sessionId: 'expired-synthetic-session',
  sessionEndsAt: now - 600_000
}
const longTitle =
  'Prepare the launch report, update the figures, and create a presentation for the team'
const messages = Array.from({ length: 14 }, (_, index) => ({
  id: `message-${index}`,
  role: index % 2 ? 'assistant' : 'user',
  content:
    index % 2
      ? `### ${index === 13 ? 'Ready for your review' : 'Progress update'}\n\nI reviewed the sample notes and organized the launch report around the main decisions. The project context stays with this task.\n\n- Reviewed the draft and source notes.\n- Added a concise summary and next steps.\n- Saved the output as \`launch-report.md\`.\n\nThis is synthetic preview text, not an actual model response.`
      : 'Review the launch notes, summarize the key decisions, and prepare a concise report for the team.',
  createdAt: now - (14 - index) * 60_000,
  sequence: index + 1,
  ...(index % 2
    ? {
        author: {
          kind: 'model',
          modelId: historicalModel.modelId,
          modelName: historicalModel.modelName,
          sessionId: historicalModel.sessionId
        }
      }
    : {})
}))
const artifact = {
  path: 'reports/launch-report.md',
  name: 'launch-report.md',
  kind: 'file',
  createdAt: now,
  updatedAt: now
}
const tasks = [
  {
    id: 'preview-task',
    projectId: project.id,
    title: longTitle,
    status: 'completed',
    createdAt: now - 900_000,
    updatedAt: now
  },
  {
    id: 'preview-task-2',
    projectId: project.id,
    title: 'Compare the Q3 research notes',
    status: 'paused',
    createdAt: now - 3_600_000,
    updatedAt: now - 3_600_000
  },
  {
    id: 'preview-task-3',
    projectId: project.id,
    title: 'Organize reference documents',
    status: 'completed',
    createdAt: now - 86_400_000,
    updatedAt: now - 86_400_000
  }
]
const records = Object.fromEntries(
  tasks.map((task) => [
    task.id,
    {
      ...task,
      goal: task.title,
      model: historicalModel,
      messages,
      plan: [
        { id: 'plan-1', title: 'Review the project notes and source files', status: 'completed' },
        { id: 'plan-2', title: 'Draft the launch report', status: 'completed' },
        { id: 'plan-3', title: 'Check figures and prepare final deliverables', status: 'completed' }
      ],
      activities: [],
      artifacts: [artifact],
      summary: 'Report prepared. Review the saved file before sharing.'
    }
  ])
)
const extensions = {
  skills: [],
  connectors: [],
  issues: [],
  executionAvailable: false,
  networkAccessPerformed: false
}
const fixtureApi = {
  getApprovalPolicy: async () => policy,
  updateApprovalPolicy: async (mode) =>
    Object.assign(policy, { mode, revision: policy.revision + 1 }),
  listProjects: async () => (emptyProject ? [] : [project]),
  listTasks: async () => tasks,
  getTask: async (id) => records[id] ?? records['preview-task'],
  listTaskMessages: async () => ({ messages: [], hasMore: false }),
  listModelOptions: async () => [replacement],
  listSchedules: async () => [],
  listExtensions: async () => extensions,
  configureExtensions: async () => extensions,
  onTaskEvent: () => () => undefined,
  previewArtifact: async () => ({
    ...artifact,
    content:
      '# Launch report\n\nSynthetic preview content.\n\nEverything in this preview is safe sample data. No local files were read.',
    truncated: false
  }),
  revealArtifact: async () => undefined,
  createProject: async ({ name, instructions }) => {
    emptyProject = false
    return Object.assign(project, { name, instructions })
  },
  updateProject: async (value) => Object.assign(project, value),
  archiveProject: async () => {
    emptyProject = true
  },
  createTask: async ({ title, goal, model }) => {
    const task = {
      ...records['preview-task'],
      id: `preview-${Date.now()}`,
      title,
      goal,
      model,
      messages: [{ id: 'new', role: 'user', content: goal, createdAt: Date.now() }],
      summary: '',
      status: 'paused'
    }
    records[task.id] = task
    tasks.unshift(task)
    return task
  },
  steerTask: async (id, content) => {
    records[id].messages = [
      ...records[id].messages,
      { id: `draft-${Date.now()}`, role: 'user', content, createdAt: Date.now() }
    ]
    return records[id]
  },
  rebindTask: async (id, model) => {
    records[id].model = model
    return records[id]
  },
  pauseTask: async (id) => records[id],
  resumeTask: async (id) => records[id],
  cancelTask: async (id) => records[id],
  deleteTask: async () => undefined
}
Object.defineProperty(window, 'cowork', { value: fixtureApi, configurable: true })
window.openLink = () => undefined
window.open = () => null
document.addEventListener(
  'click',
  (event) => {
    const link = (event.target as Element | null)?.closest('a[href]') as HTMLAnchorElement | null
    if (link && new URL(link.href, location.href).origin !== location.origin) event.preventDefault()
  },
  true
)
// Defense in depth: no application fetch/XHR can contact a real service here.
window.fetch = async () => {
  throw new Error('Network access is disabled in the synthetic UI preview.')
}
const deniedRequest = class {
  open() {
    throw new Error('Network access is disabled in the synthetic UI preview.')
  }
}
window.XMLHttpRequest = deniedRequest as unknown as typeof XMLHttpRequest

const catalog = [
  {
    Id: 'sample-chat',
    Name: 'Sample language model',
    Tags: ['llm'],
    ModelType: 'llm',
    isOnline: true,
    bids: [{ Id: 'sample-bid', PricePerSecond: '1000000000000000' }]
  },
  {
    Id: 'sample-vision',
    Name: 'Sample vision model with a long name',
    Tags: ['llm', 'vision'],
    ModelType: 'llm',
    isOnline: true,
    bids: [{ Id: 'sample-bid-2', PricePerSecond: '2000000000000000' }]
  },
  {
    Id: 'sample-offline',
    Name: 'Offline sample model',
    Tags: ['llm'],
    ModelType: 'llm',
    isOnline: false,
    bids: []
  }
]

function AgentsPreview() {
  const [txModal, setTxModal] = useState<TxModal>({ state: 'pending' })
  useEffect(() => {
    if (txModal.state !== 'loading') return
    const timeout = window.setTimeout(() => setTxModal({
      state: 'success',
      agentName: txModal.agentName,
      data: [`0x${'0'.repeat(64)}`]
    }), 400)
    return () => window.clearTimeout(timeout)
  }, [txModal])

  return <Agents {...({
    client: {}, config: {}, syncStatus: true, address: sampleAddress,
    symbol: 'SAMPLE MOR', symbolEth: 'SAMPLE ETH', morTokenAddress: sampleAddress,
    txUrlResolver: () => 'https://example.invalid/synthetic-transaction',
    agentsLoading: false, agentsError: null, retryAgents: () => undefined,
    pendingAgents: [], allowanceRequests: [],
    activeAgents: [{
      username: 'Research assistant with a deliberately long descriptive name',
      isConfirmed: true,
      perms: ['chat', 'wallet.read', 'models.list', 'sessions.read', 'long-permission-name-for-responsive-layout'],
      allowances: {
        'Sample token': '12.345',
        'Second sample token': '98765432109876543210.123456789',
        'Third sample token': '3.1415926535',
        'Fourth sample token': '42.00'
      }
    }, {
      username: 'Compact helper', isConfirmed: true, perms: ['chat'], allowances: { 'Sample token': '2.50' }
    }],
    txModal, setTxModal,
    handleApproveAccess: async () => undefined,
    handleApproveAllowance: async () => undefined,
    handleDeleteAgent: async () => undefined
  } as any)} />
}

function Preview() {
  const [compact, setCompact] = useState(params.get('compact') === '1')
  const [empty, setEmpty] = useState(emptyProject)
  const [modal, setModal] = useState(params.get('modal') || '')
  const [notice, setNotice] = useState(
    'Synthetic fixtures only · no live wallet, sessions, or files'
  )
  const navigate = useNavigate()
  const route = useLocation()
  const copy = async () => {
    setNotice('Sample address copy acknowledged. System clipboard unchanged.')
  }
  return (
    <ClientProvider
      value={{
        onHelpLinkClick: () => setNotice('Help click acknowledged; no external page opened.')
      }}
    >
      <ToastsContext.Provider value={{ toast: (_type, message) => setNotice(message) }}>
        <ThemeProvider theme={theme}>
          <>
            <div className="preview-toolbar">
              <strong>Synthetic UI preview</strong>
              <label>
                View{' '}
                <select
                  aria-label="Preview view"
                  value={route.pathname}
                  onChange={(e) => navigate(e.target.value)}
                >
                  <option value="/workspace">Workspace</option>
                  <option value="/wallet">Wallet controls</option>
                  <option value="/chat">Model selection</option>
                  <option value="/agents">Agents</option>
                  <option value="/settings">Guide / settings</option>
                </select>
              </label>
              <button onClick={() => setCompact(!compact)}>
                {compact ? 'Wide layout' : 'Compact layout'}
              </button>
              <button
                onClick={() => {
                  emptyProject = !empty
                  setEmpty(!empty)
                  navigate('/workspace')
                }}
              >
                {empty ? 'Saved project' : 'Empty project'}
              </button>
              <button onClick={openQuickStartGuide}>Replay guide</button>
            </div>
            <div className="preview-shell" data-compact={compact}>
              <Sidebar
                address={sampleAddress}
                copyToClipboard={copy}
                onRouteIntent={async () => undefined}
              />
              <div className="preview-surface">
                {route.pathname === '/workspace' ? (
                  <Cowork key={String(empty)} />
                ) : route.pathname === '/agents' ? (
                  <AgentsPreview />
                ) : (
                  <section className="preview-controls">
                    <h1>
                      {route.pathname === '/wallet'
                        ? 'Wallet controls'
                        : route.pathname === '/chat' || route.pathname === '/models'
                          ? 'Find a model'
                          : 'Explore the app'}
                    </h1>
                    <p>
                      These are the real components with isolated synthetic data. No wallet or
                      provider is connected.
                    </p>
                    <div className="preview-control-row">
                      <Btn onClick={() => setModal('receive')}>Receive (sample only)</Btn>
                      <Btn onClick={() => setModal('models')}>Choose model</Btn>
                      <Btn onClick={openQuickStartGuide}>Open quick start</Btn>
                    </div>
                    <label>
                      Sample public address
                      <input readOnly value={sampleAddress} />
                    </label>
                    <label>
                      Paste test field
                      <input placeholder="Paste a non-sensitive test value" />
                    </label>
                    <p role="status">{notice}</p>
                  </section>
                )}
              </div>
            </div>
            <QuickStartGuide />
            <Modal
              isOpen={modal === 'receive'}
              onRequestClose={() => setModal('')}
              variant="primary"
              title="Synthetic receive preview"
            >
              <ReceiveForm
                activeTab="receive"
                address={sampleAddress}
                copyToClipboard={copy}
                onRequestClose={() => setModal('')}
                explorerUrl="https://example.invalid"
                eth={{ symbol: 'SAMPLE ETH', value: 0 }}
                mor={{ symbol: 'SAMPLE MOR', value: 0 }}
              />
            </Modal>
            <ModelSelectionModal
              isActive={modal === 'models'}
              handleClose={() => setModal('')}
              onChangeModel={() => {
                setModal('')
                setNotice('Sample model selected; no session was opened.')
              }}
              models={catalog}
              symbol="SAMPLE MOR"
            />
          </>
        </ThemeProvider>
      </ToastsContext.Provider>
    </ClientProvider>
  )
}

if (!location.hash) location.hash = '/workspace'
ReactModal.setAppElement('#root')
createRoot(document.getElementById('root')!).render(
  <HashRouter>
    <Preview />
  </HashRouter>
)
