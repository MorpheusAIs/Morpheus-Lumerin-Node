//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// DiamondCutFacet
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const diamondCutFacetAbi = [
  {
    type: 'error',
    inputs: [{ name: '_selector', internalType: 'bytes4', type: 'bytes4' }],
    name: 'CannotAddFunctionToDiamondThatAlreadyExists',
  },
  {
    type: 'error',
    inputs: [
      { name: '_selectors', internalType: 'bytes4[]', type: 'bytes4[]' },
    ],
    name: 'CannotAddSelectorsToZeroAddress',
  },
  {
    type: 'error',
    inputs: [{ name: '_selector', internalType: 'bytes4', type: 'bytes4' }],
    name: 'CannotRemoveFunctionThatDoesNotExist',
  },
  {
    type: 'error',
    inputs: [{ name: '_selector', internalType: 'bytes4', type: 'bytes4' }],
    name: 'CannotRemoveImmutableFunction',
  },
  {
    type: 'error',
    inputs: [{ name: '_selector', internalType: 'bytes4', type: 'bytes4' }],
    name: 'CannotReplaceFunctionThatDoesNotExists',
  },
  {
    type: 'error',
    inputs: [{ name: '_selector', internalType: 'bytes4', type: 'bytes4' }],
    name: 'CannotReplaceFunctionWithTheSameFunctionFromTheSameFacet',
  },
  {
    type: 'error',
    inputs: [
      { name: '_selectors', internalType: 'bytes4[]', type: 'bytes4[]' },
    ],
    name: 'CannotReplaceFunctionsFromFacetWithZeroAddress',
  },
  {
    type: 'error',
    inputs: [{ name: '_selector', internalType: 'bytes4', type: 'bytes4' }],
    name: 'CannotReplaceImmutableFunction',
  },
  {
    type: 'error',
    inputs: [{ name: '_action', internalType: 'uint8', type: 'uint8' }],
    name: 'IncorrectFacetCutAction',
  },
  {
    type: 'error',
    inputs: [
      {
        name: '_initializationContractAddress',
        internalType: 'address',
        type: 'address',
      },
      { name: '_calldata', internalType: 'bytes', type: 'bytes' },
    ],
    name: 'InitializationFunctionReverted',
  },
  {
    type: 'error',
    inputs: [
      { name: '_contractAddress', internalType: 'address', type: 'address' },
      { name: '_message', internalType: 'string', type: 'string' },
    ],
    name: 'NoBytecodeAtAddress',
  },
  {
    type: 'error',
    inputs: [
      { name: '_facetAddress', internalType: 'address', type: 'address' },
    ],
    name: 'NoSelectorsProvidedForFacetForCut',
  },
  {
    type: 'error',
    inputs: [
      { name: '_user', internalType: 'address', type: 'address' },
      { name: '_contractOwner', internalType: 'address', type: 'address' },
    ],
    name: 'NotContractOwner',
  },
  {
    type: 'error',
    inputs: [
      { name: '_facetAddress', internalType: 'address', type: 'address' },
    ],
    name: 'RemoveFacetAddressMustBeZeroAddress',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: '_diamondCut',
        internalType: 'struct IDiamond.FacetCut[]',
        type: 'tuple[]',
        components: [
          { name: 'facetAddress', internalType: 'address', type: 'address' },
          {
            name: 'action',
            internalType: 'enum IDiamond.FacetCutAction',
            type: 'uint8',
          },
          {
            name: 'functionSelectors',
            internalType: 'bytes4[]',
            type: 'bytes4[]',
          },
        ],
        indexed: false,
      },
      {
        name: '_init',
        internalType: 'address',
        type: 'address',
        indexed: false,
      },
      {
        name: '_calldata',
        internalType: 'bytes',
        type: 'bytes',
        indexed: false,
      },
    ],
    name: 'DiamondCut',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: '_diamondCut',
        internalType: 'struct IDiamond.FacetCut[]',
        type: 'tuple[]',
        components: [
          { name: 'facetAddress', internalType: 'address', type: 'address' },
          {
            name: 'action',
            internalType: 'enum IDiamond.FacetCutAction',
            type: 'uint8',
          },
          {
            name: 'functionSelectors',
            internalType: 'bytes4[]',
            type: 'bytes4[]',
          },
        ],
        indexed: false,
      },
      {
        name: '_init',
        internalType: 'address',
        type: 'address',
        indexed: false,
      },
      {
        name: '_calldata',
        internalType: 'bytes',
        type: 'bytes',
        indexed: false,
      },
    ],
    name: 'DiamondCut',
  },
  {
    type: 'function',
    inputs: [
      {
        name: '_diamondCut',
        internalType: 'struct IDiamond.FacetCut[]',
        type: 'tuple[]',
        components: [
          { name: 'facetAddress', internalType: 'address', type: 'address' },
          {
            name: 'action',
            internalType: 'enum IDiamond.FacetCutAction',
            type: 'uint8',
          },
          {
            name: 'functionSelectors',
            internalType: 'bytes4[]',
            type: 'bytes4[]',
          },
        ],
      },
      { name: '_init', internalType: 'address', type: 'address' },
      { name: '_calldata', internalType: 'bytes', type: 'bytes' },
    ],
    name: 'diamondCut',
    outputs: [],
    stateMutability: 'nonpayable',
  },
] as const

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// DiamondLoupeFacet
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const diamondLoupeFacetAbi = [
  {
    type: 'function',
    inputs: [
      { name: '_functionSelector', internalType: 'bytes4', type: 'bytes4' },
    ],
    name: 'facetAddress',
    outputs: [
      { name: 'facetAddress_', internalType: 'address', type: 'address' },
    ],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'facetAddresses',
    outputs: [
      { name: 'facetAddresses_', internalType: 'address[]', type: 'address[]' },
    ],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [{ name: '_facet', internalType: 'address', type: 'address' }],
    name: 'facetFunctionSelectors',
    outputs: [
      {
        name: '_facetFunctionSelectors',
        internalType: 'bytes4[]',
        type: 'bytes4[]',
      },
    ],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'facets',
    outputs: [
      {
        name: 'facets_',
        internalType: 'struct IDiamondLoupe.Facet[]',
        type: 'tuple[]',
        components: [
          { name: 'facetAddress', internalType: 'address', type: 'address' },
          {
            name: 'functionSelectors',
            internalType: 'bytes4[]',
            type: 'bytes4[]',
          },
        ],
      },
    ],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [{ name: '_interfaceId', internalType: 'bytes4', type: 'bytes4' }],
    name: 'supportsInterface',
    outputs: [{ name: '', internalType: 'bool', type: 'bool' }],
    stateMutability: 'view',
  },
] as const

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// ERC20
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const erc20Abi = [
  {
    type: 'error',
    inputs: [
      { name: 'spender', internalType: 'address', type: 'address' },
      { name: 'allowance', internalType: 'uint256', type: 'uint256' },
      { name: 'needed', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'ERC20InsufficientAllowance',
  },
  {
    type: 'error',
    inputs: [
      { name: 'sender', internalType: 'address', type: 'address' },
      { name: 'balance', internalType: 'uint256', type: 'uint256' },
      { name: 'needed', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'ERC20InsufficientBalance',
  },
  {
    type: 'error',
    inputs: [{ name: 'approver', internalType: 'address', type: 'address' }],
    name: 'ERC20InvalidApprover',
  },
  {
    type: 'error',
    inputs: [{ name: 'receiver', internalType: 'address', type: 'address' }],
    name: 'ERC20InvalidReceiver',
  },
  {
    type: 'error',
    inputs: [{ name: 'sender', internalType: 'address', type: 'address' }],
    name: 'ERC20InvalidSender',
  },
  {
    type: 'error',
    inputs: [{ name: 'spender', internalType: 'address', type: 'address' }],
    name: 'ERC20InvalidSpender',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'owner',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      {
        name: 'spender',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      {
        name: 'value',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
    ],
    name: 'Approval',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      { name: 'from', internalType: 'address', type: 'address', indexed: true },
      { name: 'to', internalType: 'address', type: 'address', indexed: true },
      {
        name: 'value',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
    ],
    name: 'Transfer',
  },
  {
    type: 'function',
    inputs: [
      { name: 'owner', internalType: 'address', type: 'address' },
      { name: 'spender', internalType: 'address', type: 'address' },
    ],
    name: 'allowance',
    outputs: [{ name: '', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: 'spender', internalType: 'address', type: 'address' },
      { name: 'value', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'approve',
    outputs: [{ name: '', internalType: 'bool', type: 'bool' }],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [{ name: 'account', internalType: 'address', type: 'address' }],
    name: 'balanceOf',
    outputs: [{ name: '', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'decimals',
    outputs: [{ name: '', internalType: 'uint8', type: 'uint8' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'name',
    outputs: [{ name: '', internalType: 'string', type: 'string' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'symbol',
    outputs: [{ name: '', internalType: 'string', type: 'string' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'totalSupply',
    outputs: [{ name: '', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: 'to', internalType: 'address', type: 'address' },
      { name: 'value', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'transfer',
    outputs: [{ name: '', internalType: 'bool', type: 'bool' }],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [
      { name: 'from', internalType: 'address', type: 'address' },
      { name: 'to', internalType: 'address', type: 'address' },
      { name: 'value', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'transferFrom',
    outputs: [{ name: '', internalType: 'bool', type: 'bool' }],
    stateMutability: 'nonpayable',
  },
] as const

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Marketplace
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const marketplaceAbi = [
  { type: 'error', inputs: [], name: 'ActiveBidNotFound' },
  { type: 'error', inputs: [], name: 'BidTaken' },
  { type: 'error', inputs: [], name: 'KeyExists' },
  { type: 'error', inputs: [], name: 'KeyNotFound' },
  { type: 'error', inputs: [], name: 'ModelOrAgentNotFound' },
  {
    type: 'error',
    inputs: [
      { name: '_user', internalType: 'address', type: 'address' },
      { name: '_contractOwner', internalType: 'address', type: 'address' },
    ],
    name: 'NotContractOwner',
  },
  { type: 'error', inputs: [], name: 'NotEnoughBalance' },
  { type: 'error', inputs: [], name: 'NotSenderOrOwner' },
  { type: 'error', inputs: [], name: 'ProviderNotFound' },
  { type: 'error', inputs: [], name: 'ZeroKey' },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'provider',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      {
        name: 'modelAgentId',
        internalType: 'bytes32',
        type: 'bytes32',
        indexed: true,
      },
      {
        name: 'nonce',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
    ],
    name: 'BidDeleted',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'provider',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      {
        name: 'modelAgentId',
        internalType: 'bytes32',
        type: 'bytes32',
        indexed: true,
      },
      {
        name: 'nonce',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
    ],
    name: 'BidPosted',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'bidFee',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
    ],
    name: 'FeeUpdated',
  },
  {
    type: 'function',
    inputs: [],
    name: 'bidFee',
    outputs: [{ name: '', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [{ name: 'bidId', internalType: 'bytes32', type: 'bytes32' }],
    name: 'bidMap',
    outputs: [
      {
        name: '',
        internalType: 'struct Bid',
        type: 'tuple',
        components: [
          { name: 'provider', internalType: 'address', type: 'address' },
          { name: 'modelAgentId', internalType: 'bytes32', type: 'bytes32' },
          { name: 'pricePerSecond', internalType: 'uint256', type: 'uint256' },
          { name: 'nonce', internalType: 'uint256', type: 'uint256' },
          { name: 'createdAt', internalType: 'uint128', type: 'uint128' },
          { name: 'deletedAt', internalType: 'uint128', type: 'uint128' },
        ],
      },
    ],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [{ name: 'bidId', internalType: 'bytes32', type: 'bytes32' }],
    name: 'deleteModelAgentBid',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [
      { name: 'modelAgentId', internalType: 'bytes32', type: 'bytes32' },
    ],
    name: 'getActiveBidsByModelAgent',
    outputs: [
      { name: '', internalType: 'bytes32[]', type: 'bytes32[]' },
      {
        name: '',
        internalType: 'struct Bid[]',
        type: 'tuple[]',
        components: [
          { name: 'provider', internalType: 'address', type: 'address' },
          { name: 'modelAgentId', internalType: 'bytes32', type: 'bytes32' },
          { name: 'pricePerSecond', internalType: 'uint256', type: 'uint256' },
          { name: 'nonce', internalType: 'uint256', type: 'uint256' },
          { name: 'createdAt', internalType: 'uint128', type: 'uint128' },
          { name: 'deletedAt', internalType: 'uint128', type: 'uint128' },
        ],
      },
    ],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [{ name: 'provider', internalType: 'address', type: 'address' }],
    name: 'getActiveBidsByProvider',
    outputs: [
      { name: '', internalType: 'bytes32[]', type: 'bytes32[]' },
      {
        name: '',
        internalType: 'struct Bid[]',
        type: 'tuple[]',
        components: [
          { name: 'provider', internalType: 'address', type: 'address' },
          { name: 'modelAgentId', internalType: 'bytes32', type: 'bytes32' },
          { name: 'pricePerSecond', internalType: 'uint256', type: 'uint256' },
          { name: 'nonce', internalType: 'uint256', type: 'uint256' },
          { name: 'createdAt', internalType: 'uint128', type: 'uint128' },
          { name: 'deletedAt', internalType: 'uint128', type: 'uint128' },
        ],
      },
    ],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: 'modelAgentId', internalType: 'bytes32', type: 'bytes32' },
      { name: 'offset', internalType: 'uint256', type: 'uint256' },
      { name: 'limit', internalType: 'uint8', type: 'uint8' },
    ],
    name: 'getActiveBidsRatingByModelAgent',
    outputs: [
      { name: '', internalType: 'bytes32[]', type: 'bytes32[]' },
      {
        name: '',
        internalType: 'struct Bid[]',
        type: 'tuple[]',
        components: [
          { name: 'provider', internalType: 'address', type: 'address' },
          { name: 'modelAgentId', internalType: 'bytes32', type: 'bytes32' },
          { name: 'pricePerSecond', internalType: 'uint256', type: 'uint256' },
          { name: 'nonce', internalType: 'uint256', type: 'uint256' },
          { name: 'createdAt', internalType: 'uint128', type: 'uint128' },
          { name: 'deletedAt', internalType: 'uint128', type: 'uint128' },
        ],
      },
      {
        name: '',
        internalType: 'struct ProviderModelStats[]',
        type: 'tuple[]',
        components: [
          {
            name: 'tpsScaled1000',
            internalType: 'struct LibSD.SD',
            type: 'tuple',
            components: [
              { name: 'mean', internalType: 'int64', type: 'int64' },
              { name: 'sqSum', internalType: 'int64', type: 'int64' },
            ],
          },
          {
            name: 'ttftMs',
            internalType: 'struct LibSD.SD',
            type: 'tuple',
            components: [
              { name: 'mean', internalType: 'int64', type: 'int64' },
              { name: 'sqSum', internalType: 'int64', type: 'int64' },
            ],
          },
          { name: 'totalDuration', internalType: 'uint32', type: 'uint32' },
          { name: 'successCount', internalType: 'uint32', type: 'uint32' },
          { name: 'totalCount', internalType: 'uint32', type: 'uint32' },
        ],
      },
    ],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: 'modelAgentId', internalType: 'bytes32', type: 'bytes32' },
      { name: 'offset', internalType: 'uint256', type: 'uint256' },
      { name: 'limit', internalType: 'uint8', type: 'uint8' },
    ],
    name: 'getBidsByModelAgent',
    outputs: [
      { name: '', internalType: 'bytes32[]', type: 'bytes32[]' },
      {
        name: '',
        internalType: 'struct Bid[]',
        type: 'tuple[]',
        components: [
          { name: 'provider', internalType: 'address', type: 'address' },
          { name: 'modelAgentId', internalType: 'bytes32', type: 'bytes32' },
          { name: 'pricePerSecond', internalType: 'uint256', type: 'uint256' },
          { name: 'nonce', internalType: 'uint256', type: 'uint256' },
          { name: 'createdAt', internalType: 'uint128', type: 'uint128' },
          { name: 'deletedAt', internalType: 'uint128', type: 'uint128' },
        ],
      },
    ],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: 'provider', internalType: 'address', type: 'address' },
      { name: 'offset', internalType: 'uint256', type: 'uint256' },
      { name: 'limit', internalType: 'uint8', type: 'uint8' },
    ],
    name: 'getBidsByProvider',
    outputs: [
      { name: '', internalType: 'bytes32[]', type: 'bytes32[]' },
      {
        name: '',
        internalType: 'struct Bid[]',
        type: 'tuple[]',
        components: [
          { name: 'provider', internalType: 'address', type: 'address' },
          { name: 'modelAgentId', internalType: 'bytes32', type: 'bytes32' },
          { name: 'pricePerSecond', internalType: 'uint256', type: 'uint256' },
          { name: 'nonce', internalType: 'uint256', type: 'uint256' },
          { name: 'createdAt', internalType: 'uint128', type: 'uint128' },
          { name: 'deletedAt', internalType: 'uint128', type: 'uint128' },
        ],
      },
    ],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [{ name: 'modelID', internalType: 'bytes32', type: 'bytes32' }],
    name: 'getModelStats',
    outputs: [
      {
        name: '',
        internalType: 'struct ModelStats',
        type: 'tuple',
        components: [
          {
            name: 'tpsScaled1000',
            internalType: 'struct LibSD.SD',
            type: 'tuple',
            components: [
              { name: 'mean', internalType: 'int64', type: 'int64' },
              { name: 'sqSum', internalType: 'int64', type: 'int64' },
            ],
          },
          {
            name: 'ttftMs',
            internalType: 'struct LibSD.SD',
            type: 'tuple',
            components: [
              { name: 'mean', internalType: 'int64', type: 'int64' },
              { name: 'sqSum', internalType: 'int64', type: 'int64' },
            ],
          },
          {
            name: 'totalDuration',
            internalType: 'struct LibSD.SD',
            type: 'tuple',
            components: [
              { name: 'mean', internalType: 'int64', type: 'int64' },
              { name: 'sqSum', internalType: 'int64', type: 'int64' },
            ],
          },
          { name: 'count', internalType: 'uint32', type: 'uint32' },
        ],
      },
    ],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: 'providerAddr', internalType: 'address', type: 'address' },
      { name: 'modelId', internalType: 'bytes32', type: 'bytes32' },
      { name: 'pricePerSecond', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'postModelBid',
    outputs: [{ name: 'bidId', internalType: 'bytes32', type: 'bytes32' }],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [{ name: '_bidFee', internalType: 'uint256', type: 'uint256' }],
    name: 'setBidFee',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [
      { name: 'addr', internalType: 'address', type: 'address' },
      { name: 'amount', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'withdraw',
    outputs: [],
    stateMutability: 'nonpayable',
  },
] as const

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// ModelRegistry
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const modelRegistryAbi = [
  { type: 'error', inputs: [], name: 'KeyExists' },
  { type: 'error', inputs: [], name: 'KeyNotFound' },
  { type: 'error', inputs: [], name: 'ModelHasActiveBids' },
  { type: 'error', inputs: [], name: 'ModelNotFound' },
  {
    type: 'error',
    inputs: [
      { name: '_user', internalType: 'address', type: 'address' },
      { name: '_contractOwner', internalType: 'address', type: 'address' },
    ],
    name: 'NotContractOwner',
  },
  { type: 'error', inputs: [], name: 'NotSenderOrOwner' },
  { type: 'error', inputs: [], name: 'StakeTooLow' },
  { type: 'error', inputs: [], name: 'ZeroKey' },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'owner',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      {
        name: 'modelId',
        internalType: 'bytes32',
        type: 'bytes32',
        indexed: true,
      },
    ],
    name: 'ModelDeregistered',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'newStake',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
    ],
    name: 'ModelMinStakeUpdated',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'owner',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      {
        name: 'modelId',
        internalType: 'bytes32',
        type: 'bytes32',
        indexed: true,
      },
    ],
    name: 'ModelRegisteredUpdated',
  },
  {
    type: 'function',
    inputs: [{ name: 'id', internalType: 'bytes32', type: 'bytes32' }],
    name: 'modelDeregister',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [{ name: 'id', internalType: 'bytes32', type: 'bytes32' }],
    name: 'modelExists',
    outputs: [{ name: '', internalType: 'bool', type: 'bool' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'modelGetAll',
    outputs: [
      { name: '', internalType: 'bytes32[]', type: 'bytes32[]' },
      {
        name: '',
        internalType: 'struct Model[]',
        type: 'tuple[]',
        components: [
          { name: 'ipfsCID', internalType: 'bytes32', type: 'bytes32' },
          { name: 'fee', internalType: 'uint256', type: 'uint256' },
          { name: 'stake', internalType: 'uint256', type: 'uint256' },
          { name: 'owner', internalType: 'address', type: 'address' },
          { name: 'name', internalType: 'string', type: 'string' },
          { name: 'tags', internalType: 'string[]', type: 'string[]' },
          { name: 'createdAt', internalType: 'uint128', type: 'uint128' },
          { name: 'isDeleted', internalType: 'bool', type: 'bool' },
        ],
      },
    ],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [{ name: 'index', internalType: 'uint256', type: 'uint256' }],
    name: 'modelGetByIndex',
    outputs: [
      { name: 'modelId', internalType: 'bytes32', type: 'bytes32' },
      {
        name: 'model',
        internalType: 'struct Model',
        type: 'tuple',
        components: [
          { name: 'ipfsCID', internalType: 'bytes32', type: 'bytes32' },
          { name: 'fee', internalType: 'uint256', type: 'uint256' },
          { name: 'stake', internalType: 'uint256', type: 'uint256' },
          { name: 'owner', internalType: 'address', type: 'address' },
          { name: 'name', internalType: 'string', type: 'string' },
          { name: 'tags', internalType: 'string[]', type: 'string[]' },
          { name: 'createdAt', internalType: 'uint128', type: 'uint128' },
          { name: 'isDeleted', internalType: 'bool', type: 'bool' },
        ],
      },
    ],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'modelGetCount',
    outputs: [{ name: 'count', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'modelGetIds',
    outputs: [{ name: '', internalType: 'bytes32[]', type: 'bytes32[]' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [{ name: 'id', internalType: 'bytes32', type: 'bytes32' }],
    name: 'modelMap',
    outputs: [
      {
        name: '',
        internalType: 'struct Model',
        type: 'tuple',
        components: [
          { name: 'ipfsCID', internalType: 'bytes32', type: 'bytes32' },
          { name: 'fee', internalType: 'uint256', type: 'uint256' },
          { name: 'stake', internalType: 'uint256', type: 'uint256' },
          { name: 'owner', internalType: 'address', type: 'address' },
          { name: 'name', internalType: 'string', type: 'string' },
          { name: 'tags', internalType: 'string[]', type: 'string[]' },
          { name: 'createdAt', internalType: 'uint128', type: 'uint128' },
          { name: 'isDeleted', internalType: 'bool', type: 'bool' },
        ],
      },
    ],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'modelMinStake',
    outputs: [{ name: '', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: 'modelId', internalType: 'bytes32', type: 'bytes32' },
      { name: 'ipfsCID', internalType: 'bytes32', type: 'bytes32' },
      { name: 'fee', internalType: 'uint256', type: 'uint256' },
      { name: 'addStake', internalType: 'uint256', type: 'uint256' },
      { name: 'owner', internalType: 'address', type: 'address' },
      { name: 'name', internalType: 'string', type: 'string' },
      { name: 'tags', internalType: 'string[]', type: 'string[]' },
    ],
    name: 'modelRegister',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [{ name: 'id', internalType: 'bytes32', type: 'bytes32' }],
    name: 'modelResetStats',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [{ name: '_minStake', internalType: 'uint256', type: 'uint256' }],
    name: 'modelSetMinStake',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [{ name: 'id', internalType: 'bytes32', type: 'bytes32' }],
    name: 'modelStats',
    outputs: [
      {
        name: '',
        internalType: 'struct ModelStats',
        type: 'tuple',
        components: [
          {
            name: 'tpsScaled1000',
            internalType: 'struct LibSD.SD',
            type: 'tuple',
            components: [
              { name: 'mean', internalType: 'int64', type: 'int64' },
              { name: 'sqSum', internalType: 'int64', type: 'int64' },
            ],
          },
          {
            name: 'ttftMs',
            internalType: 'struct LibSD.SD',
            type: 'tuple',
            components: [
              { name: 'mean', internalType: 'int64', type: 'int64' },
              { name: 'sqSum', internalType: 'int64', type: 'int64' },
            ],
          },
          {
            name: 'totalDuration',
            internalType: 'struct LibSD.SD',
            type: 'tuple',
            components: [
              { name: 'mean', internalType: 'int64', type: 'int64' },
              { name: 'sqSum', internalType: 'int64', type: 'int64' },
            ],
          },
          { name: 'count', internalType: 'uint32', type: 'uint32' },
        ],
      },
    ],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [{ name: 'index', internalType: 'uint256', type: 'uint256' }],
    name: 'models',
    outputs: [{ name: '', internalType: 'bytes32', type: 'bytes32' }],
    stateMutability: 'view',
  },
] as const

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// MorpheusToken
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const morpheusTokenAbi = [
  { type: 'constructor', inputs: [], stateMutability: 'nonpayable' },
  {
    type: 'error',
    inputs: [
      { name: 'spender', internalType: 'address', type: 'address' },
      { name: 'allowance', internalType: 'uint256', type: 'uint256' },
      { name: 'needed', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'ERC20InsufficientAllowance',
  },
  {
    type: 'error',
    inputs: [
      { name: 'sender', internalType: 'address', type: 'address' },
      { name: 'balance', internalType: 'uint256', type: 'uint256' },
      { name: 'needed', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'ERC20InsufficientBalance',
  },
  {
    type: 'error',
    inputs: [{ name: 'approver', internalType: 'address', type: 'address' }],
    name: 'ERC20InvalidApprover',
  },
  {
    type: 'error',
    inputs: [{ name: 'receiver', internalType: 'address', type: 'address' }],
    name: 'ERC20InvalidReceiver',
  },
  {
    type: 'error',
    inputs: [{ name: 'sender', internalType: 'address', type: 'address' }],
    name: 'ERC20InvalidSender',
  },
  {
    type: 'error',
    inputs: [{ name: 'spender', internalType: 'address', type: 'address' }],
    name: 'ERC20InvalidSpender',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'owner',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      {
        name: 'spender',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      {
        name: 'value',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
    ],
    name: 'Approval',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      { name: 'from', internalType: 'address', type: 'address', indexed: true },
      { name: 'to', internalType: 'address', type: 'address', indexed: true },
      {
        name: 'value',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
    ],
    name: 'Transfer',
  },
  {
    type: 'function',
    inputs: [
      { name: 'owner', internalType: 'address', type: 'address' },
      { name: 'spender', internalType: 'address', type: 'address' },
    ],
    name: 'allowance',
    outputs: [{ name: '', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: 'spender', internalType: 'address', type: 'address' },
      { name: 'value', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'approve',
    outputs: [{ name: '', internalType: 'bool', type: 'bool' }],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [{ name: 'account', internalType: 'address', type: 'address' }],
    name: 'balanceOf',
    outputs: [{ name: '', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'decimals',
    outputs: [{ name: '', internalType: 'uint8', type: 'uint8' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'name',
    outputs: [{ name: '', internalType: 'string', type: 'string' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'symbol',
    outputs: [{ name: '', internalType: 'string', type: 'string' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'totalSupply',
    outputs: [{ name: '', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: 'to', internalType: 'address', type: 'address' },
      { name: 'value', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'transfer',
    outputs: [{ name: '', internalType: 'bool', type: 'bool' }],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [
      { name: 'from', internalType: 'address', type: 'address' },
      { name: 'to', internalType: 'address', type: 'address' },
      { name: 'value', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'transferFrom',
    outputs: [{ name: '', internalType: 'bool', type: 'bool' }],
    stateMutability: 'nonpayable',
  },
] as const

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// OwnershipFacet
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const ownershipFacetAbi = [
  {
    type: 'error',
    inputs: [
      { name: '_user', internalType: 'address', type: 'address' },
      { name: '_contractOwner', internalType: 'address', type: 'address' },
    ],
    name: 'NotContractOwner',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'previousOwner',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      {
        name: 'newOwner',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
    ],
    name: 'OwnershipTransferred',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'previousOwner',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      {
        name: 'newOwner',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
    ],
    name: 'OwnershipTransferred',
  },
  {
    type: 'function',
    inputs: [],
    name: 'owner',
    outputs: [{ name: 'owner_', internalType: 'address', type: 'address' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [{ name: '_newOwner', internalType: 'address', type: 'address' }],
    name: 'transferOwnership',
    outputs: [],
    stateMutability: 'nonpayable',
  },
] as const

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// ProviderRegistry
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const providerRegistryAbi = [
  { type: 'error', inputs: [], name: 'ErrNoStake' },
  { type: 'error', inputs: [], name: 'ErrNoWithdrawableStake' },
  { type: 'error', inputs: [], name: 'ErrProviderNotDeleted' },
  { type: 'error', inputs: [], name: 'KeyExists' },
  { type: 'error', inputs: [], name: 'KeyNotFound' },
  {
    type: 'error',
    inputs: [
      { name: '_user', internalType: 'address', type: 'address' },
      { name: '_contractOwner', internalType: 'address', type: 'address' },
    ],
    name: 'NotContractOwner',
  },
  { type: 'error', inputs: [], name: 'NotSenderOrOwner' },
  { type: 'error', inputs: [], name: 'ProviderHasActiveBids' },
  { type: 'error', inputs: [], name: 'StakeTooLow' },
  { type: 'error', inputs: [], name: 'ZeroKey' },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'provider',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
    ],
    name: 'ProviderDeregistered',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'newStake',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
    ],
    name: 'ProviderMinStakeUpdated',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'provider',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
    ],
    name: 'ProviderRegisteredUpdated',
  },
  {
    type: 'function',
    inputs: [{ name: 'addr', internalType: 'address', type: 'address' }],
    name: 'providerDeregister',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [{ name: 'addr', internalType: 'address', type: 'address' }],
    name: 'providerExists',
    outputs: [{ name: '', internalType: 'bool', type: 'bool' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'providerGetAll',
    outputs: [
      { name: '', internalType: 'address[]', type: 'address[]' },
      {
        name: '',
        internalType: 'struct Provider[]',
        type: 'tuple[]',
        components: [
          { name: 'endpoint', internalType: 'string', type: 'string' },
          { name: 'stake', internalType: 'uint256', type: 'uint256' },
          { name: 'createdAt', internalType: 'uint128', type: 'uint128' },
          { name: 'limitPeriodEnd', internalType: 'uint128', type: 'uint128' },
          {
            name: 'limitPeriodEarned',
            internalType: 'uint256',
            type: 'uint256',
          },
          { name: 'isDeleted', internalType: 'bool', type: 'bool' },
        ],
      },
    ],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [{ name: 'index', internalType: 'uint256', type: 'uint256' }],
    name: 'providerGetByIndex',
    outputs: [
      { name: 'addr', internalType: 'address', type: 'address' },
      {
        name: 'provider',
        internalType: 'struct Provider',
        type: 'tuple',
        components: [
          { name: 'endpoint', internalType: 'string', type: 'string' },
          { name: 'stake', internalType: 'uint256', type: 'uint256' },
          { name: 'createdAt', internalType: 'uint128', type: 'uint128' },
          { name: 'limitPeriodEnd', internalType: 'uint128', type: 'uint128' },
          {
            name: 'limitPeriodEarned',
            internalType: 'uint256',
            type: 'uint256',
          },
          { name: 'isDeleted', internalType: 'bool', type: 'bool' },
        ],
      },
    ],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'providerGetCount',
    outputs: [{ name: 'count', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'providerGetIds',
    outputs: [{ name: '', internalType: 'address[]', type: 'address[]' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [{ name: 'addr', internalType: 'address', type: 'address' }],
    name: 'providerMap',
    outputs: [
      {
        name: '',
        internalType: 'struct Provider',
        type: 'tuple',
        components: [
          { name: 'endpoint', internalType: 'string', type: 'string' },
          { name: 'stake', internalType: 'uint256', type: 'uint256' },
          { name: 'createdAt', internalType: 'uint128', type: 'uint128' },
          { name: 'limitPeriodEnd', internalType: 'uint128', type: 'uint128' },
          {
            name: 'limitPeriodEarned',
            internalType: 'uint256',
            type: 'uint256',
          },
          { name: 'isDeleted', internalType: 'bool', type: 'bool' },
        ],
      },
    ],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'providerMinStake',
    outputs: [{ name: '', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: 'addr', internalType: 'address', type: 'address' },
      { name: 'addStake', internalType: 'uint256', type: 'uint256' },
      { name: 'endpoint', internalType: 'string', type: 'string' },
    ],
    name: 'providerRegister',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [{ name: '_minStake', internalType: 'uint256', type: 'uint256' }],
    name: 'providerSetMinStake',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [{ name: 'addr', internalType: 'address', type: 'address' }],
    name: 'providerWithdrawStake',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [{ name: 'index', internalType: 'uint256', type: 'uint256' }],
    name: 'providers',
    outputs: [{ name: '', internalType: 'address', type: 'address' }],
    stateMutability: 'view',
  },
] as const

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// SessionRouter
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const sessionRouterAbi = [
  { type: 'error', inputs: [], name: 'ApprovedForAnotherUser' },
  { type: 'error', inputs: [], name: 'BidNotFound' },
  { type: 'error', inputs: [], name: 'CannotDecodeAbi' },
  { type: 'error', inputs: [], name: 'DuplicateApproval' },
  { type: 'error', inputs: [], name: 'ECDSAInvalidSignature' },
  {
    type: 'error',
    inputs: [{ name: 'length', internalType: 'uint256', type: 'uint256' }],
    name: 'ECDSAInvalidSignatureLength',
  },
  {
    type: 'error',
    inputs: [{ name: 's', internalType: 'bytes32', type: 'bytes32' }],
    name: 'ECDSAInvalidSignatureS',
  },
  { type: 'error', inputs: [], name: 'KeyExists' },
  { type: 'error', inputs: [], name: 'KeyNotFound' },
  {
    type: 'error',
    inputs: [
      { name: '_user', internalType: 'address', type: 'address' },
      { name: '_contractOwner', internalType: 'address', type: 'address' },
    ],
    name: 'NotContractOwner',
  },
  { type: 'error', inputs: [], name: 'NotEnoughWithdrawableBalance' },
  { type: 'error', inputs: [], name: 'NotSenderOrOwner' },
  { type: 'error', inputs: [], name: 'ProviderSignatureMismatch' },
  { type: 'error', inputs: [], name: 'SessionAlreadyClosed' },
  { type: 'error', inputs: [], name: 'SessionNotClosed' },
  { type: 'error', inputs: [], name: 'SessionNotFound' },
  { type: 'error', inputs: [], name: 'SessionTooShort' },
  { type: 'error', inputs: [], name: 'SignatureExpired' },
  { type: 'error', inputs: [], name: 'WithdrawableBalanceLimitByStakeReached' },
  { type: 'error', inputs: [], name: 'WrongChaidId' },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'userAddress',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      {
        name: 'sessionId',
        internalType: 'bytes32',
        type: 'bytes32',
        indexed: true,
      },
      {
        name: 'providerId',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
    ],
    name: 'SessionClosed',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'userAddress',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      {
        name: 'sessionId',
        internalType: 'bytes32',
        type: 'bytes32',
        indexed: true,
      },
      {
        name: 'providerId',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
    ],
    name: 'SessionOpened',
  },
  {
    type: 'function',
    inputs: [],
    name: 'MAX_SESSION_DURATION',
    outputs: [{ name: '', internalType: 'uint32', type: 'uint32' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'MIN_SESSION_DURATION',
    outputs: [{ name: '', internalType: 'uint32', type: 'uint32' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'SIGNATURE_TTL',
    outputs: [{ name: '', internalType: 'uint32', type: 'uint32' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'activeSessionsCount',
    outputs: [{ name: '', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: 'sessionId', internalType: 'bytes32', type: 'bytes32' },
      { name: 'amountToWithdraw', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'claimProviderBalance',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [
      { name: 'receiptEncoded', internalType: 'bytes', type: 'bytes' },
      { name: 'signature', internalType: 'bytes', type: 'bytes' },
    ],
    name: 'closeSession',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [{ name: 'sessionId', internalType: 'bytes32', type: 'bytes32' }],
    name: 'deleteHistory',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [{ name: 'provider', internalType: 'address', type: 'address' }],
    name: 'getActiveSessionsByProvider',
    outputs: [
      {
        name: '',
        internalType: 'struct Session[]',
        type: 'tuple[]',
        components: [
          { name: 'id', internalType: 'bytes32', type: 'bytes32' },
          { name: 'user', internalType: 'address', type: 'address' },
          { name: 'provider', internalType: 'address', type: 'address' },
          { name: 'modelAgentId', internalType: 'bytes32', type: 'bytes32' },
          { name: 'bidID', internalType: 'bytes32', type: 'bytes32' },
          { name: 'stake', internalType: 'uint256', type: 'uint256' },
          { name: 'pricePerSecond', internalType: 'uint256', type: 'uint256' },
          { name: 'closeoutReceipt', internalType: 'bytes', type: 'bytes' },
          { name: 'closeoutType', internalType: 'uint256', type: 'uint256' },
          {
            name: 'providerWithdrawnAmount',
            internalType: 'uint256',
            type: 'uint256',
          },
          { name: 'openedAt', internalType: 'uint256', type: 'uint256' },
          { name: 'endsAt', internalType: 'uint256', type: 'uint256' },
          { name: 'closedAt', internalType: 'uint256', type: 'uint256' },
        ],
      },
    ],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [{ name: 'user', internalType: 'address', type: 'address' }],
    name: 'getActiveSessionsByUser',
    outputs: [
      {
        name: '',
        internalType: 'struct Session[]',
        type: 'tuple[]',
        components: [
          { name: 'id', internalType: 'bytes32', type: 'bytes32' },
          { name: 'user', internalType: 'address', type: 'address' },
          { name: 'provider', internalType: 'address', type: 'address' },
          { name: 'modelAgentId', internalType: 'bytes32', type: 'bytes32' },
          { name: 'bidID', internalType: 'bytes32', type: 'bytes32' },
          { name: 'stake', internalType: 'uint256', type: 'uint256' },
          { name: 'pricePerSecond', internalType: 'uint256', type: 'uint256' },
          { name: 'closeoutReceipt', internalType: 'bytes', type: 'bytes' },
          { name: 'closeoutType', internalType: 'uint256', type: 'uint256' },
          {
            name: 'providerWithdrawnAmount',
            internalType: 'uint256',
            type: 'uint256',
          },
          { name: 'openedAt', internalType: 'uint256', type: 'uint256' },
          { name: 'endsAt', internalType: 'uint256', type: 'uint256' },
          { name: 'closedAt', internalType: 'uint256', type: 'uint256' },
        ],
      },
    ],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [{ name: 'timestamp', internalType: 'uint256', type: 'uint256' }],
    name: 'getComputeBalance',
    outputs: [{ name: '', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [{ name: 'sessionId', internalType: 'bytes32', type: 'bytes32' }],
    name: 'getProviderClaimableBalance',
    outputs: [{ name: '', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [{ name: 'sessionId', internalType: 'bytes32', type: 'bytes32' }],
    name: 'getSession',
    outputs: [
      {
        name: '',
        internalType: 'struct Session',
        type: 'tuple',
        components: [
          { name: 'id', internalType: 'bytes32', type: 'bytes32' },
          { name: 'user', internalType: 'address', type: 'address' },
          { name: 'provider', internalType: 'address', type: 'address' },
          { name: 'modelAgentId', internalType: 'bytes32', type: 'bytes32' },
          { name: 'bidID', internalType: 'bytes32', type: 'bytes32' },
          { name: 'stake', internalType: 'uint256', type: 'uint256' },
          { name: 'pricePerSecond', internalType: 'uint256', type: 'uint256' },
          { name: 'closeoutReceipt', internalType: 'bytes', type: 'bytes' },
          { name: 'closeoutType', internalType: 'uint256', type: 'uint256' },
          {
            name: 'providerWithdrawnAmount',
            internalType: 'uint256',
            type: 'uint256',
          },
          { name: 'openedAt', internalType: 'uint256', type: 'uint256' },
          { name: 'endsAt', internalType: 'uint256', type: 'uint256' },
          { name: 'closedAt', internalType: 'uint256', type: 'uint256' },
        ],
      },
    ],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: 'modelId', internalType: 'bytes32', type: 'bytes32' },
      { name: 'offset', internalType: 'uint256', type: 'uint256' },
      { name: 'limit', internalType: 'uint8', type: 'uint8' },
    ],
    name: 'getSessionsByModel',
    outputs: [
      {
        name: '',
        internalType: 'struct Session[]',
        type: 'tuple[]',
        components: [
          { name: 'id', internalType: 'bytes32', type: 'bytes32' },
          { name: 'user', internalType: 'address', type: 'address' },
          { name: 'provider', internalType: 'address', type: 'address' },
          { name: 'modelAgentId', internalType: 'bytes32', type: 'bytes32' },
          { name: 'bidID', internalType: 'bytes32', type: 'bytes32' },
          { name: 'stake', internalType: 'uint256', type: 'uint256' },
          { name: 'pricePerSecond', internalType: 'uint256', type: 'uint256' },
          { name: 'closeoutReceipt', internalType: 'bytes', type: 'bytes' },
          { name: 'closeoutType', internalType: 'uint256', type: 'uint256' },
          {
            name: 'providerWithdrawnAmount',
            internalType: 'uint256',
            type: 'uint256',
          },
          { name: 'openedAt', internalType: 'uint256', type: 'uint256' },
          { name: 'endsAt', internalType: 'uint256', type: 'uint256' },
          { name: 'closedAt', internalType: 'uint256', type: 'uint256' },
        ],
      },
    ],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: 'provider', internalType: 'address', type: 'address' },
      { name: 'offset', internalType: 'uint256', type: 'uint256' },
      { name: 'limit', internalType: 'uint8', type: 'uint8' },
    ],
    name: 'getSessionsByProvider',
    outputs: [
      {
        name: '',
        internalType: 'struct Session[]',
        type: 'tuple[]',
        components: [
          { name: 'id', internalType: 'bytes32', type: 'bytes32' },
          { name: 'user', internalType: 'address', type: 'address' },
          { name: 'provider', internalType: 'address', type: 'address' },
          { name: 'modelAgentId', internalType: 'bytes32', type: 'bytes32' },
          { name: 'bidID', internalType: 'bytes32', type: 'bytes32' },
          { name: 'stake', internalType: 'uint256', type: 'uint256' },
          { name: 'pricePerSecond', internalType: 'uint256', type: 'uint256' },
          { name: 'closeoutReceipt', internalType: 'bytes', type: 'bytes' },
          { name: 'closeoutType', internalType: 'uint256', type: 'uint256' },
          {
            name: 'providerWithdrawnAmount',
            internalType: 'uint256',
            type: 'uint256',
          },
          { name: 'openedAt', internalType: 'uint256', type: 'uint256' },
          { name: 'endsAt', internalType: 'uint256', type: 'uint256' },
          { name: 'closedAt', internalType: 'uint256', type: 'uint256' },
        ],
      },
    ],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: 'user', internalType: 'address', type: 'address' },
      { name: 'offset', internalType: 'uint256', type: 'uint256' },
      { name: 'limit', internalType: 'uint8', type: 'uint8' },
    ],
    name: 'getSessionsByUser',
    outputs: [
      {
        name: '',
        internalType: 'struct Session[]',
        type: 'tuple[]',
        components: [
          { name: 'id', internalType: 'bytes32', type: 'bytes32' },
          { name: 'user', internalType: 'address', type: 'address' },
          { name: 'provider', internalType: 'address', type: 'address' },
          { name: 'modelAgentId', internalType: 'bytes32', type: 'bytes32' },
          { name: 'bidID', internalType: 'bytes32', type: 'bytes32' },
          { name: 'stake', internalType: 'uint256', type: 'uint256' },
          { name: 'pricePerSecond', internalType: 'uint256', type: 'uint256' },
          { name: 'closeoutReceipt', internalType: 'bytes', type: 'bytes' },
          { name: 'closeoutType', internalType: 'uint256', type: 'uint256' },
          {
            name: 'providerWithdrawnAmount',
            internalType: 'uint256',
            type: 'uint256',
          },
          { name: 'openedAt', internalType: 'uint256', type: 'uint256' },
          { name: 'endsAt', internalType: 'uint256', type: 'uint256' },
          { name: 'closedAt', internalType: 'uint256', type: 'uint256' },
        ],
      },
    ],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [{ name: 'timestamp', internalType: 'uint256', type: 'uint256' }],
    name: 'getTodaysBudget',
    outputs: [{ name: '', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: '_stake', internalType: 'uint256', type: 'uint256' },
      { name: 'providerApproval', internalType: 'bytes', type: 'bytes' },
      { name: 'signature', internalType: 'bytes', type: 'bytes' },
    ],
    name: 'openSession',
    outputs: [{ name: 'sessionId', internalType: 'bytes32', type: 'bytes32' }],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'sessionsCount',
    outputs: [{ name: '', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: 'index', internalType: 'uint256', type: 'uint256' },
      {
        name: 'pool',
        internalType: 'struct Pool',
        type: 'tuple',
        components: [
          { name: 'initialReward', internalType: 'uint256', type: 'uint256' },
          { name: 'rewardDecrease', internalType: 'uint256', type: 'uint256' },
          { name: 'payoutStart', internalType: 'uint128', type: 'uint128' },
          {
            name: 'decreaseInterval',
            internalType: 'uint128',
            type: 'uint128',
          },
        ],
      },
    ],
    name: 'setPoolConfig',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [
      { name: 'sessionStake', internalType: 'uint256', type: 'uint256' },
      { name: 'timestamp', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'stakeToStipend',
    outputs: [{ name: '', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: 'stipend', internalType: 'uint256', type: 'uint256' },
      { name: 'timestamp', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'stipendToStake',
    outputs: [{ name: '', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [{ name: 'timestamp', internalType: 'uint256', type: 'uint256' }],
    name: 'totalMORSupply',
    outputs: [{ name: '', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: 'sessionStake', internalType: 'uint256', type: 'uint256' },
      { name: 'pricePerSecond', internalType: 'uint256', type: 'uint256' },
      { name: 'openedAt', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'whenSessionEnds',
    outputs: [{ name: '', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: 'amountToWithdraw', internalType: 'uint256', type: 'uint256' },
      { name: 'iterations', internalType: 'uint8', type: 'uint8' },
    ],
    name: 'withdrawUserStake',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [
      { name: 'userAddr', internalType: 'address', type: 'address' },
      { name: 'iterations', internalType: 'uint8', type: 'uint8' },
    ],
    name: 'withdrawableUserStake',
    outputs: [
      { name: 'avail', internalType: 'uint256', type: 'uint256' },
      { name: 'hold', internalType: 'uint256', type: 'uint256' },
    ],
    stateMutability: 'view',
  },
] as const

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// StakingMasterChef
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const stakingMasterChefAbi = [
  {
    type: 'constructor',
    inputs: [
      {
        name: '_stakingToken',
        internalType: 'contract IERC20',
        type: 'address',
      },
      {
        name: '_rewardToken',
        internalType: 'contract IERC20',
        type: 'address',
      },
    ],
    stateMutability: 'nonpayable',
  },
  { type: 'error', inputs: [], name: 'LockNotEnded' },
  { type: 'error', inputs: [], name: 'LockReleaseTimePastPoolEndTime' },
  { type: 'error', inputs: [], name: 'NoRewardAvailable' },
  {
    type: 'error',
    inputs: [{ name: 'owner', internalType: 'address', type: 'address' }],
    name: 'OwnableInvalidOwner',
  },
  {
    type: 'error',
    inputs: [{ name: 'account', internalType: 'address', type: 'address' }],
    name: 'OwnableUnauthorizedAccount',
  },
  { type: 'error', inputs: [], name: 'PoolOrStakeNotExists' },
  { type: 'error', inputs: [], name: 'StakingFinished' },
  { type: 'error', inputs: [], name: 'StakingNotStarted' },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'previousOwner',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      {
        name: 'newOwner',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
    ],
    name: 'OwnershipTransferred',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'poolId',
        internalType: 'uint256',
        type: 'uint256',
        indexed: true,
      },
      {
        name: 'startTime',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
      {
        name: 'endTime',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
    ],
    name: 'PoolAdded',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'poolId',
        internalType: 'uint256',
        type: 'uint256',
        indexed: true,
      },
    ],
    name: 'PoolStopped',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'userAddress',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      {
        name: 'poolId',
        internalType: 'uint256',
        type: 'uint256',
        indexed: true,
      },
      {
        name: 'stakeId',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
      {
        name: 'amount',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
    ],
    name: 'RewardWithdrawal',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'userAddress',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      {
        name: 'poolId',
        internalType: 'uint256',
        type: 'uint256',
        indexed: true,
      },
      {
        name: 'stakeId',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
      {
        name: 'amount',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
    ],
    name: 'Stake',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'userAddress',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      {
        name: 'poolId',
        internalType: 'uint256',
        type: 'uint256',
        indexed: true,
      },
      {
        name: 'stakeId',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
      {
        name: 'amount',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
    ],
    name: 'Unstake',
  },
  {
    type: 'function',
    inputs: [],
    name: 'PRECISION',
    outputs: [{ name: '', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: '_startTime', internalType: 'uint256', type: 'uint256' },
      { name: '_duration', internalType: 'uint256', type: 'uint256' },
      { name: '_totalReward', internalType: 'uint256', type: 'uint256' },
      {
        name: '_lockDurations',
        internalType: 'struct StakingMasterChef.Lock[]',
        type: 'tuple[]',
        components: [
          { name: 'durationSeconds', internalType: 'uint256', type: 'uint256' },
          {
            name: 'multiplierScaled',
            internalType: 'uint256',
            type: 'uint256',
          },
        ],
      },
    ],
    name: 'addPool',
    outputs: [{ name: '', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [{ name: '_poolId', internalType: 'uint256', type: 'uint256' }],
    name: 'getLockDurations',
    outputs: [
      {
        name: '',
        internalType: 'struct StakingMasterChef.Lock[]',
        type: 'tuple[]',
        components: [
          { name: 'durationSeconds', internalType: 'uint256', type: 'uint256' },
          {
            name: 'multiplierScaled',
            internalType: 'uint256',
            type: 'uint256',
          },
        ],
      },
    ],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: '_user', internalType: 'address', type: 'address' },
      { name: '_poolId', internalType: 'uint256', type: 'uint256' },
      { name: '_stakeId', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'getReward',
    outputs: [{ name: '', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: '_addr', internalType: 'address', type: 'address' },
      { name: '_poolId', internalType: 'uint256', type: 'uint256' },
      { name: '_stakeId', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'getStake',
    outputs: [
      {
        name: '',
        internalType: 'struct StakingMasterChef.UserStake',
        type: 'tuple',
        components: [
          { name: 'stakeAmount', internalType: 'uint256', type: 'uint256' },
          { name: 'shareAmount', internalType: 'uint256', type: 'uint256' },
          { name: 'rewardDebt', internalType: 'uint256', type: 'uint256' },
          { name: 'lockEndsAt', internalType: 'uint256', type: 'uint256' },
        ],
      },
    ],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: '_addr', internalType: 'address', type: 'address' },
      { name: '_poolId', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'getStakes',
    outputs: [
      {
        name: '',
        internalType: 'struct StakingMasterChef.UserStake[]',
        type: 'tuple[]',
        components: [
          { name: 'stakeAmount', internalType: 'uint256', type: 'uint256' },
          { name: 'shareAmount', internalType: 'uint256', type: 'uint256' },
          { name: 'rewardDebt', internalType: 'uint256', type: 'uint256' },
          { name: 'lockEndsAt', internalType: 'uint256', type: 'uint256' },
        ],
      },
    ],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'owner',
    outputs: [{ name: '', internalType: 'address', type: 'address' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [{ name: '', internalType: 'uint256', type: 'uint256' }],
    name: 'pools',
    outputs: [
      {
        name: 'rewardPerSecondScaled',
        internalType: 'uint256',
        type: 'uint256',
      },
      { name: 'lastRewardTime', internalType: 'uint256', type: 'uint256' },
      {
        name: 'accRewardPerShareScaled',
        internalType: 'uint256',
        type: 'uint256',
      },
      { name: 'totalShares', internalType: 'uint256', type: 'uint256' },
      { name: 'startTime', internalType: 'uint256', type: 'uint256' },
      { name: 'endTime', internalType: 'uint256', type: 'uint256' },
    ],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [{ name: '_poolId', internalType: 'uint256', type: 'uint256' }],
    name: 'recalculatePoolReward',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'renounceOwnership',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'rewardToken',
    outputs: [{ name: '', internalType: 'contract IERC20', type: 'address' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: '_poolId', internalType: 'uint256', type: 'uint256' },
      { name: '_amount', internalType: 'uint256', type: 'uint256' },
      { name: '_lockId', internalType: 'uint8', type: 'uint8' },
    ],
    name: 'stake',
    outputs: [{ name: '', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'stakingToken',
    outputs: [{ name: '', internalType: 'contract IERC20', type: 'address' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [{ name: '_poolId', internalType: 'uint256', type: 'uint256' }],
    name: 'stopPool',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [{ name: 'newOwner', internalType: 'address', type: 'address' }],
    name: 'transferOwnership',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [
      { name: '_poolId', internalType: 'uint256', type: 'uint256' },
      { name: '_stakeId', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'unstake',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [
      { name: '_poolId', internalType: 'uint256', type: 'uint256' },
      { name: '_stakeId', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'withdrawReward',
    outputs: [],
    stateMutability: 'nonpayable',
  },
] as const

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Test1Facet
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const test1FacetAbi = [
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'something',
        internalType: 'address',
        type: 'address',
        indexed: false,
      },
    ],
    name: 'TestEvent',
  },
  {
    type: 'function',
    inputs: [{ name: '_interfaceID', internalType: 'bytes4', type: 'bytes4' }],
    name: 'supportsInterface',
    outputs: [{ name: '', internalType: 'bool', type: 'bool' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test1Func1',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test1Func10',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test1Func11',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test1Func12',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test1Func13',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test1Func14',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test1Func15',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test1Func16',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test1Func17',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test1Func18',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test1Func19',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test1Func2',
    outputs: [{ name: '', internalType: 'address', type: 'address' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test1Func20',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test1Func3',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test1Func4',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test1Func5',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test1Func6',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test1Func7',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test1Func8',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test1Func9',
    outputs: [],
    stateMutability: 'nonpayable',
  },
] as const

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Test2Facet
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const test2FacetAbi = [
  {
    type: 'function',
    inputs: [],
    name: 'test2Func1',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test2Func10',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test2Func11',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test2Func12',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test2Func13',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test2Func14',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test2Func15',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test2Func16',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test2Func17',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test2Func18',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test2Func19',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test2Func2',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test2Func20',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test2Func3',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test2Func4',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test2Func5',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test2Func6',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test2Func7',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test2Func8',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'test2Func9',
    outputs: [],
    stateMutability: 'nonpayable',
  },
] as const
