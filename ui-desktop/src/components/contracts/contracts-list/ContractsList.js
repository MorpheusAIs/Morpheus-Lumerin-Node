import React, { useState, useEffect } from 'react';
import { List as RVList, AutoSizer } from 'react-virtualized';
import withContractsListState from '../../../store/hocs/withContractsListState';
import ScanningContractsPlaceholder from './ScanningContractsPlaceholder';
import NoContractsPlaceholder from './NoContractsPlaceholder';
import { ItemFilter, Flex } from '../../common';
import Header from './Header';
import ContractsRow from './Row';
import {
  Container,
  ListContainer,
  Contracts,
  FooterLogo
} from './ContractsList.styles';
import { ContractsRowContainer } from './ContractsRow.styles';
import StatusHeader from './StatusHeader';
import Search from './Search';
import styled from 'styled-components';
import Sort from './Sort';
import { Btn } from '../../common';

const Stats = styled.div`
  color: #0e4353;
  display: flex;
  justify-content: space-between;
  width: 100%;
  background: white;
  border-radius: 8px;
  gap: 10px;
  flex-wrap: wrap;
  margin-bottom: 4rem;
  margin-top: 1rem;
  padding: 1rem 3rem;
`;

const StatValue = styled.div`
  flex-grow: 1;
  flex-shrink: 0;
  flex-basis: 0;
  background-color: white;
  display: flex;
  flex-direction: column;
  justify-content: center;
  font-weight: 100;
  font-size: 1.5rem;
  padding: 0 4rem;
  :not(:last-child) {
    border-right: 1.75px solid #a8a8a8;
  }

  h3 {
    font-size: 1.2rem;
    font-weight: 400;
    margin: 0;
    margin-bottom: 1rem;
  }

  p {
    font-size: 3rem;
    font-weight: 500;
    margin: 0;
  }

  p.smaller {
    font-size: 95%;
  }

  span.unit {
    font-size: 1.2rem;
  }
`;

const ContractBtn = styled(Btn)`
  font-size: 1.3rem;
  padding: 0.6rem 1.4rem;

  @media (min-width: 1040px) {
    margin-left: 0;
  }
`;

const SearchSortWrapper = styled.div`
  display: flex;
`;

const sorting = (contracts, sortBy, underProfitContracts) => {
  switch (sortBy?.value) {
    case 'AscPrice':
      return contracts.sort((a, b) => a.price - b.price);
    case 'DescPrice':
      return contracts.sort((a, b) => b.price - a.price);
    case 'AscDuration':
      return contracts.sort((a, b) => a.length - b.length);
    case 'DescDuration':
      return contracts.sort((a, b) => b.length - a.length);
    case 'AscSpeed':
      return contracts.sort((a, b) => a.speed - b.speed);
    case 'DescSpeed':
      return contracts.sort((a, b) => b.speed - a.speed);
    case 'AvailableFirst':
      return contracts.sort((a, b) => (+b.state > +a.state ? -1 : 1));
    case 'RunningFirst':
      return contracts.sort((a, b) => (+b.state > +a.state ? 1 : -1));
    case 'UnderProfit':
      return contracts.sort((a, b) =>
        underProfitContracts.find(x => x.id == a.id) ? -1 : 1
      );
    default:
      return contracts.sort((a, b) => (+b.state > +a.state ? -1 : 1));
  }
};

function ContractsList({
  contracts,
  syncStatus,
  cancel,
  deleteContract,
  createContract,
  address,
  contractsRefresh,
  noContractsMessage,
  customRowRenderer,
  allowSendTransaction,
  tabs,
  isSellerTab,
  stats,
  edit,
  setEditContractData,
  sellerStats,
  offset,
  underProfitContracts,
  onAdjustFormOpen
}) {
  const [selectedContracts, setSelectedContracts] = useState([]);
  const [search, setSearch] = useState('');
  const [sort, setSort] = useState(null);

  const [headerOptions, setHeaderOptions] = useState({});

  let contractsToShow = search
    ? contracts.filter(c => c.id.toLowerCase().includes(search.toLowerCase()))
    : contracts;

  contractsToShow = sorting(contractsToShow, sort, underProfitContracts);

  const hasContracts = contractsToShow.length;
  const defaultTabs = [
    { value: 'timestamp', name: 'Started', ratio: 2 },
    { name: 'Status', ratio: 1 },
    { value: 'price', name: 'Price', ratio: 2 },
    { value: 'length', name: 'Duration', ratio: 2 },
    { value: 'speed', name: 'Speed', ratio: 2 },
    { value: 'action', name: '', ratio: 3 }
  ];

  const tabsToShow = tabs || defaultTabs;
  const ratio = tabsToShow.map(x => x.ratio);

  useEffect(() => {
    contractsRefresh();
  }, []);

  const onContractsClicked = ({ currentTarget }) => {
    setSelectedContracts(currentTarget.dataset.hash);
  };

  const rowRenderer = (contractsList, ratio, converters) => ({
    key,
    index,
    style
  }) => (
    <ContractsRowContainer style={style} key={`${key}-${index}`}>
      <ContractsRow
        key={contractsList[index].id}
        data-testid="Contracts-row"
        onClick={onContractsClicked}
        contract={contractsList[index]}
        cancel={cancel}
        deleteContract={deleteContract}
        converters={converters}
        address={address}
        ratio={ratio}
        edit={edit}
        setEditContractData={setEditContractData}
        allowSendTransaction={allowSendTransaction}
        underProfitContracts={underProfitContracts}
      />
    </ContractsRowContainer>
  );

  const filterExtractValue = ({ status }) => status;
  return (
    <Container data-testid="Contracts-list">
      <Flex.Row grow="1" style={{ flexDirection: 'column' }}>
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            flexWrap: 'wrap'
          }}
        >
          {stats && (
            <Stats>
              <StatValue>
                <h3>Total Contracts</h3>
                <p>{stats.count}</p>
              </StatValue>
              <StatValue>
                <h3>Rented Contracts</h3>
                <p>{stats.rented}</p>
              </StatValue>
              <StatValue>
                <h3>Expires in 1 hour</h3>
                <p>{stats.expiresInHour}</p>
              </StatValue>
            </Stats>
          )}
          {sellerStats && (
            <Stats>
              <StatValue>
                <h3>Contracts</h3>
                <p>{sellerStats.count}</p>
              </StatValue>
              <StatValue>
                <h3>Posted</h3>
                <p>
                  {sellerStats.totalPosted}
                  <span className="unit"> TH/s</span>
                </p>
              </StatValue>
              <StatValue>
                <h3>Rented</h3>
                <p>
                  {sellerStats.rented}
                  <span className="unit"> TH/s</span>
                </p>
              </StatValue>
              <StatValue
              // data-rh={
              //   sellerStats.networkReward
              //     ? `${formatExpNumber(
              //         fromMicro(sellerStats.networkReward)
              //       )} BTC/TH/day`
              //     : 'Calculating...'
              // }
              >
                <h3>Est. Network Profitability</h3>
                <p className="smaller">
                  {sellerStats.networkReward
                    ? `${sellerStats.networkReward}`
                    : 'Calculating...'}
                  <span className="unit"> BTC/TH/day</span>
                </p>
              </StatValue>
            </Stats>
          )}
        </div>
      </Flex.Row>
      <Flex.Row style={{ justifyContent: 'space-between', margin: '10px 0' }}>
        <div style={{ display: 'flex', gap: '10px' }}>
          {isSellerTab ? (
            <ContractBtn
              data-disabled={!allowSendTransaction}
              onClick={allowSendTransaction ? createContract : () => {}}
            >
              Create Contract
            </ContractBtn>
          ) : (
            <></>
          )}
          {/* <Sort sort={sort} setSort={setSort} /> */}
          <StatusHeader refresh={contractsRefresh} syncStatus={syncStatus} />
          {isSellerTab && underProfitContracts?.length ? (
            <ContractBtn
              onClick={() => {
                onAdjustFormOpen();
              }}
            >
              Adjust prices
            </ContractBtn>
          ) : (
            <></>
          )}
        </div>
        <SearchSortWrapper>
          <Sort sort={sort} setSort={setSort} />
          <Search onSearch={setSearch} />
        </SearchSortWrapper>
        {/* <StatusHeader refresh={contractsRefresh} syncStatus={syncStatus} /> */}
      </Flex.Row>
      <Contracts>
        <ItemFilter extractValue={filterExtractValue} items={contractsToShow}>
          {({ filteredItems }) => (
            <React.Fragment>
              <Header
                onFilterChange={() => {}}
                onColumnOptionChange={e =>
                  setHeaderOptions({
                    ...headerOptions,
                    [e.type]: e.value
                  })
                }
                activeFilter={null}
                tabs={tabsToShow}
              />

              <ListContainer offset={offset}>
                {!hasContracts &&
                  (syncStatus === 'syncing' ? (
                    <ScanningContractsPlaceholder />
                  ) : (
                    <NoContractsPlaceholder
                      message={
                        syncStatus === 'failed'
                          ? 'Failed to retrieve contracts'
                          : noContractsMessage
                      }
                    />
                  ))}
                <AutoSizer>
                  {({ width, height }) => (
                    <RVList
                      rowRenderer={
                        customRowRenderer
                          ? customRowRenderer(
                              filteredItems,
                              ratio,
                              headerOptions
                            )
                          : rowRenderer(filteredItems, ratio, headerOptions)
                      }
                      rowHeight={66}
                      rowCount={contractsToShow.length}
                      height={height || 500} // defaults for tests
                      width={width || 500} // defaults for tests
                    />
                  )}
                </AutoSizer>
                <FooterLogo></FooterLogo>
              </ListContainer>
            </React.Fragment>
          )}
        </ItemFilter>
      </Contracts>
    </Container>
  );
}

export default withContractsListState(ContractsList);
