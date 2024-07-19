import { useEffect, useState } from 'react';
import styled from 'styled-components';
import Select from "react-select";
import {
    RightBtn,
} from '../../contracts/modals/CreateContractModal.styles';
import { abbreviateAddress } from '../../../utils';
import { formatSmallNumber } from '../utils';
import { IconEdit, IconX } from '@tabler/icons-react';

const RowContainer = styled.div`
  padding: 1.2rem 0;
  display: grid;
  grid-template-columns: 2fr 4fr 160px;
  text-align: center;
  box-shadow: 0 -1px 0 0 ${p => p.theme.colors.morMain} inset;
  color: ${p => p.theme.colors.morMain};
`;

const FlexCenter = styled.div`
  display: flex;
  justify-content: center;
  align-items: center;
`;

const Buttons = styled.div`
display: flex;
justify-content: end;
min-width: 150px;
width: 150px;
align-items: center;
gap: 10px;
`

const PriceContainer = styled.div`
    display: flex;
    justify-content: ${p => p.hasLocal ? "space-evenly" : 'center'};
    align-items: center;
`

const UseLocalBlock = styled(FlexCenter)`
    text-decoration: underline;
    cursor: pointer;

    &:hover {
        opacity: 0.8;
    }
`

const selectorStyles = {
    control: (base, state) => ({ ...base, borderColor: '#20dc8e', width: '100%', background: 'transparent' }),
    option: (base, state) => ({
        ...base,
        backgroundColor: state.isSelected ? '#0e4353' : "#03160e",
        color: '#20dc8e',
        ':active': {
            ...base[':active'],
            backgroundColor: '#0e435380',
            color: '#20dc8e'
        },
        zIndex: 100
    }),
    placeholder: base => ({
        ...base,
        color: '#20dc8e',
        fontSize: '1.4rem',
        fontWeight: 700
    }),
    singleValue: base => ({
        ...base,
        color: '#20dc8e',
        fontWeight: 600,
        fontSize: '1.4rem'
    }),
    indicatorsContainer: base => ({
        ...base,
        color: '#20dc8e'
    }),
};

function ModelRow(props) {
    const bids = props?.model?.bids || [];
    const modelId = props?.model?.Id || '';
    const hasLocal = props?.model?.hasLocal;
    const hasBids = Boolean(bids.length);

    const [selected, changeSelected] = useState<any>();
    const [useSelect, setUseSelect] = useState<boolean>();
    const [targetBid, setTargetBid] = useState<any>();

    const options = bids.map(x => {
        return ({ value: x.Id, label: `${abbreviateAddress(x.Provider || "", 3)} ${formatSmallNumber(x.PricePerSecond / (10 ** 18))} MOR` })
    });

    const handleChangeModel = () => {
        props.onChangeModel({ modelId, bidId: targetBid?.Id })
    }

    const selectLocal = () => {
        props.onChangeModel({ modelId, isLocal: true })
    }

    const onChangeSelector = (data) => {
        const bid = bids.find(x => x.Id == data.value);
        setTargetBid({ ...bid, customSelection: true });
    }

    const formatPrice = () => {
        if (targetBid) {
            return `${formatSmallNumber(targetBid?.PricePerSecond / (10 ** 18))} MOR`;
        }

        const prices = bids.filter(x => x.Id).map(x => x.PricePerSecond);
        if (prices.length == 1) {
            return `${formatSmallNumber(prices[0] / (10 ** 18))} MOR`;
        }

        const minPrice = Math.min(prices);
        const maxPrice = Math.max(prices);

        return `${formatSmallNumber(minPrice / (10 ** 18))} - ${formatSmallNumber(maxPrice / (10 ** 18))} MOR`
    }

    return (
        <RowContainer useSelect={useSelect}>
            <FlexCenter>
                {props?.model?.Name}
            </FlexCenter>
            <PriceContainer hasLocal={hasLocal}>
                {
                    useSelect
                        ? <div style={{ width: '100%', display: 'flex', alignItems: 'center' }}>
                            <div style={{ marginRight: '10px', width: '95%' }}>
                                <Select
                                    menuPlacement='auto'
                                    onChange={onChangeSelector}
                                    styles={selectorStyles}
                                    value={selected}
                                    options={options}></Select>
                            </div>

                            <IconX width={'1.5rem'} style={{ cursor: 'pointer' }} onClick={() => {
                                setUseSelect(!useSelect);
                                setTargetBid(undefined);
                                changeSelected(undefined);
                            }}></IconX>
                        </div>
                        : <div>
                            {
                                hasBids ? <FlexCenter>
                                    <span
                                        data-rh-negative
                                        data-rh={`Bid ID: ${abbreviateAddress(targetBid?.Id || "", 5)}`}
                                        style={{ marginRight: '10px' }}>
                                        {formatPrice()}
                                    </span>
                                    
                                    {/* <IconEdit width={'1.5rem'} style={{ cursor: 'pointer' }} onClick={() => setUseSelect(!useSelect)}></IconEdit> */}
                                </FlexCenter>
                                    : <FlexCenter>-</FlexCenter>

                            }

                        </div>
                }
            </PriceContainer>
            <Buttons>
                {
                    hasLocal &&
                    <RightBtn block onClick={selectLocal}>Local</RightBtn>
                }
                {
                    hasBids &&
                    <RightBtn block onClick={handleChangeModel}>Select</RightBtn>
                }
            </Buttons>
        </RowContainer>
    );
}

export default ModelRow;
