import React, { useState } from 'react';
import { IconTrashOff } from '@tabler/icons-react';
import styled from 'styled-components';
import Select from "react-select";
import {
    TitleWrapper,
    Title,
    Subtitle,
    Form,
    InputGroup,
    Row,
    Input,
    Label,
    Sublabel,
    RightBtn,
} from '../../contracts/modals/CreateContractModal.styles';
import { abbreviateAddress } from '../../../utils';

const RowContainer = styled.div`
  padding: 1.2rem 0;
  display: grid;
  grid-template-columns: 2fr 4fr 2fr;
  text-align: center;
  box-shadow: 0 -1px 0 0 ${p => p.theme.colors.morMain} inset;
  color: ${p => p.theme.colors.morMain};
`;

const FlexCenter = styled.div`
  display: flex;
  justify-content: center;
  align-items: center;
`;

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
        }
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
    const options = props?.model?.bids.map(x => ({ value: x.Id, label: `${abbreviateAddress(x.Provider || "", 3)} ${x.PricePerSecond / (10 ** 18)} MOR` }))
    const [selected, changeSelected] = useState<any>();

    return (
        <RowContainer>
            <FlexCenter>
                {props?.model?.Name}
            </FlexCenter>
            <FlexCenter>
                <div style={{ width: '100%' }}>
                    <Select
                        menuPlacement="auto"
                        onChange={changeSelected}
                        styles={selectorStyles}
                        value={selected}
                        options={options}
                    />
                </div>
            </FlexCenter>
            <FlexCenter>
                <RightBtn block onClick={() => props.onChangeModel(selected?.value)}>Change</RightBtn>
            </FlexCenter>
        </RowContainer>
    );
}

export default ModelRow;
