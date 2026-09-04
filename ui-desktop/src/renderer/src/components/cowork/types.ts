export type CoworkApprovalMode = 'manual' | 'auto' | 'skip';

export type CoworkTaskStatus =
  | 'draft'
  | 'queued'
  | 'running'
  | 'waiting_approval'
  | 'paused'
  | 'completed'
  | 'failed'
  | 'cancelled';

export type CoworkPlanStepStatus = 'pending' | 'in_progress' | 'completed';

export type CoworkDataBoundary =
  | 'on-device'
  | 'configured-endpoint'
  | 'independent-provider';

export interface CoworkProject {
  id: string;
  name: string;
  folderName?: string;
  instructions: string;
  approvalMode: CoworkApprovalMode;
  extensionSettings?: {
    folderInstructionsEnabled: boolean;
    enabledSkillIds: string[];
  };
  createdAt: number;
  updatedAt: number;
}

export interface CoworkModelTarget {
  modelId: string;
  modelName: string;
  isLocal: boolean;
  dataBoundary: CoworkDataBoundary;
  sessionId?: string;
  sessionEndsAt?: number;
}

export interface CoworkModelOption extends CoworkModelTarget {
  source: 'local' | 'marketplace';
  visionCapability: 'declared' | 'detected' | 'none';
}

export interface CoworkPlanStep {
  id: string;
  title: string;
  status: CoworkPlanStepStatus;
  note?: string;
}

export interface CoworkArtifact {
  path: string;
  name: string;
  kind: 'file' | 'folder';
  createdAt: number;
  updatedAt: number;
}

export interface CoworkDisplayMessage {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  createdAt: number;
}

export interface CoworkActivity {
  id: string;
  type: 'plan' | 'tool' | 'approval' | 'system';
  label: string;
  detail?: string;
  status: 'running' | 'success' | 'error' | 'waiting';
  createdAt: number;
}

export interface CoworkToolCall {
  id: string;
  type: 'function';
  function: {
    name: string;
    arguments: string;
  };
}

export interface CoworkPendingApproval {
  id: string;
  toolCall: CoworkToolCall;
  reason: string;
  risk: 'read' | 'write' | 'overwrite' | 'delete' | 'network';
  createdAt: number;
}

export interface CoworkTask {
  id: string;
  projectId: string;
  title: string;
  goal: string;
  status: CoworkTaskStatus;
  model: CoworkModelTarget;
  plan: CoworkPlanStep[];
  messages: CoworkDisplayMessage[];
  activities: CoworkActivity[];
  artifacts: CoworkArtifact[];
  pendingApproval?: CoworkPendingApproval;
  summary?: string;
  error?: string;
  createdAt: number;
  updatedAt: number;
  startedAt?: number;
  completedAt?: number;
}

export type CoworkTaskSummary = Pick<
  CoworkTask,
  | 'id'
  | 'projectId'
  | 'title'
  | 'status'
  | 'createdAt'
  | 'updatedAt'
  | 'startedAt'
  | 'completedAt'
>;

export type CoworkScheduleCadence =
  | { kind: 'manual' }
  | { kind: 'hourly'; minute: number }
  | { kind: 'daily'; hour: number; minute: number }
  | { kind: 'weekly'; dayOfWeek: number; hour: number; minute: number }
  | { kind: 'weekdays'; hour: number; minute: number };

export interface CoworkSchedule {
  id: string;
  projectId: string;
  name: string;
  task: {
    title: string;
    goal: string;
    model: CoworkModelTarget;
  };
  cadence: CoworkScheduleCadence;
  timeZone: string;
  status: 'active' | 'paused';
  createdAt: number;
  updatedAt: number;
  nextRunAt?: number;
  lastScheduledFor?: number;
  lastRunAt?: number;
  lastTaskId?: string;
  lastError?: string;
  runningSince?: number;
}

export interface CoworkExtensionCatalog {
  projectInstructions?: {
    kind: 'project-instructions';
    contentHash: string;
    source: string;
    bytes: number;
    enabled: boolean;
    trust: 'untrusted-project-content';
  };
  skills: Array<{
    kind: 'skill';
    id: string;
    name: string;
    description: string;
    source: string;
    declaredCapabilities: string[];
    activationRequested: boolean;
    enabled: boolean;
    executionMode: 'instructions-only';
  }>;
  connectors: Array<{
    kind: 'remote-mcp-connector';
    id: string;
    name: string;
    description: string;
    transport: 'streamable-http' | 'sse';
    authentication: 'none' | 'oauth2';
    declaredCapabilities: string[];
    activationRequested: boolean;
    enabled: false;
    connectionState: 'not-connected';
    source: string;
  }>;
  issues: Array<{
    severity: 'warning' | 'error';
    code: string;
    source: string;
    message: string;
  }>;
  executionAvailable: false;
  networkAccessPerformed: false;
}

export interface CoworkTaskEvent {
  taskId: string;
  projectId: string;
  title: string;
  status: CoworkTaskStatus;
  createdAt: number;
  updatedAt: number;
  startedAt?: number;
  completedAt?: number;
}

export interface CoworkArtifactPreview {
  path: string;
  name: string;
  kind: 'file' | 'folder';
  createdAt: number;
  updatedAt: number;
  content: string;
  truncated: boolean;
}
