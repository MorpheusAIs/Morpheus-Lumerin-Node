import styled from 'styled-components'

import withLoadingState from '../store/hocs/withLoadingState'

import { LoadingBar, AltLayout, Flex } from './common'
import ChecklistItem from './common/ChecklistItem'

const ChecklistContainer = styled(Flex.Row)`
  margin: 4rem -20rem;
`

const Title = styled.div`
  display: none;
  font-weight: 600;
  cursor: default;
  font-size: 0.9rem;
  letter-spacing: 1.6px;
  opacity: 0.5;
  margin-bottom: 0.8rem;
  padding-left: 8rem;
`

const Checklist = styled.div`
  opacity: 0.5;
  color: ${(p) => p.theme.colors.dark}
  padding-left: 0;
`

function Loading({ chainStatus }) {
  return (
    <AltLayout title="Gathering Information..." data-testid="loading-scene">
      <LoadingBar />
      <ChecklistContainer justify="center">
        <div key={chainStatus.displayName}>
          <Title>{chainStatus.displayName}</Title>
          <Checklist>
            <ChecklistItem isActive={chainStatus.hasBlockHeight} text="Blockchain status" />
            <ChecklistItem
              isActive={chainStatus.hasCoinRate}
              text={`${chainStatus.symbol || "MOR"} exchange data`}
            />
            <ChecklistItem
              isActive={chainStatus.hasCoinBalance}
              text={`${chainStatus.symbol || "MOR"} balance`}
            />
            {/* <ChecklistItem isActive={chainStatus.hasLmrBalance} text={`${chainStatus.symbol || "MOR"} balance`} /> */}
          </Checklist>
        </div>
      </ChecklistContainer>
    </AltLayout>
  )
}

export default withLoadingState(Loading)
