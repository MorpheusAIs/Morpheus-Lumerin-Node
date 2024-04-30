import React, { useState, useEffect } from 'react';
import { List as RVList, AutoSizer } from 'react-virtualized';
import ScanningContractsPlaceholder from '../contracts/contracts-list/ScanningContractsPlaceholder';
import NoContractsPlaceholder from '../contracts/contracts-list/NoContractsPlaceholder';
import { ItemFilter } from '../common';
import Header from '../contracts/contracts-list/Header';
import ContractsRow from '../contracts/contracts-list/Row';
import {
  Container,
  ListContainer,
  Contracts,
  FooterLogo
} from '../contracts/contracts-list/ContractsList.styles';
import { ContractsRowContainer } from '../contracts/contracts-list/ContractsRow.styles';
import styled from 'styled-components';
import { Btn } from '../common';

function ModelsTable({
  contracts,
  syncStatus,
  cancel,
  deleteContract,
  address,
  contractsRefresh,
  noContractsMessage,
  allowSendTransaction,
  tabs,
  edit,
  setEditContractData,
  offset,
  underProfitContracts,
} : any) {
  const [selectedContracts, setSelectedContracts] = useState([]);
  const [search, setSearch] = useState('');
  const [sort, setSort] = useState(null);

  const [headerOptions, setHeaderOptions] = useState({});

  console.log("render");

  let contractsToShow = contracts || [];
   
  const hasContracts = contractsToShow.length;

  const defaultTabs = [
    { value: 'name', name: 'Name', ratio: 2 },
    { value: 'size', name: 'Size', ratio: 2 },
    { value: 'ipc', name: 'IPC', ratio: 2 },
  ];

  const tabsToShow = tabs || defaultTabs;
  const ratio = tabsToShow.map(x => x.ratio);


  const rowRenderer = (contractsList, ratio, converters) => ({
    key,
    index,
    style
  }) => (
    <ContractsRowContainer style={style} key={`${key}-${index}`}>
      <ContractsRow
        key={contractsList[index].id}
        data-testid="Contracts-row"
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
    <Container data-testid="Models-list">
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

              <ListContainer offset={200}>
                {!hasContracts &&
                  (syncStatus === 'syncing' ? (
                    <ScanningContractsPlaceholder />
                  ) : (
                    <NoContractsPlaceholder
                      message={
                        "No models yet!"
                      }
                    />
                  ))}
                <AutoSizer>
                  {({ width, height }) => (
                    <RVList
                      rowRenderer={
                        rowRenderer(filteredItems, ratio, headerOptions)
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

export default ModelsTable;
