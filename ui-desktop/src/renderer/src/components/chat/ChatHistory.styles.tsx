import styled from 'styled-components';
import { Btn } from '../common/Btn'

export const Container = styled.div`
    .dropdown-toggle::after {
        display: none !important; 
    }
    .history-block {
        height: calc(100vh - 100px);
    }
    .history-scroll-block {
        overflow-y: auto;
        
        .nav-link {
            color: ${p => p.theme.colors.morMain}
        }

        .nav-link.active {
            color: ${p => p.theme.colors.morMain}
            border-color: ${p => p.theme.colors.morMain}
            background-color: rgba(0,0,0,0.4);
        }
    }
`

export const Title = styled.div`
    text-align: center;
    margin-bottom: 2.4rem;

    span {
        cursor: pointer;
    }
`
export const HistoryItem = styled.div`
    color: ${p => p.theme.colors.morMain}
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 5px 0 0 0;
`
export const HistoryEntryContainer = styled.div`
    background: rgba(255,255,255, 0.04);
    border-width: 1px;
    border: 1px solid rgba(255, 255, 255, 0.04);
    color: white;
    margin-bottom: 15px;
    cursor: pointer;
    padding: 10px;

    &:hover {
        background: rgba(255,255,255, 0.10);
    }
`

export const FlexSpaceBetween = styled.div`
    display: flex;
    justify-content: space-between;
    align-items: center;
`

export const HistoryEntryTitle = styled(FlexSpaceBetween)`
    text-align: justify;
    color: ${p => p.theme.colors.morMain}
    margin: 0.5rem 0;
    padding: 1rem 1.5rem;
    opacity: 0.8

    &:hover, &[data-active] {
        cursor: pointer;
        opacity: 1;
        border-radius: 10px;
        background: rgba(255, 255, 255, 0.05);
    }

    .title {
        text-overflow: ellipsis;
        overflow: hidden;
        white-space: nowrap;
    }
`

export const ModelName = styled.div`
    text-overflow: ellipsis;
    width: 250px;
    height: 24px;
    overflow: hidden;
    text-wrap: nowrap;
`

export const CloseBtn = styled(Btn)`
    font-size: 1.4rem;
    padding: 0 1rem;
`

export const Duration = styled.div`
    color: white;
`

export const IconsContainer = styled.div`
    svg:hover {
        opacity: 0.8
    }
`
export const ChangeTitleContainer = styled(IconsContainer)`
    display: flex;
    align-items: center;
    width: 100%;

    input {
        background: transparent;
        color: white;
        border: none;
        border-bottom: 1px solid ${p => p.theme.colors.morMain}40;
    }

    input:focus {
        background: transparent;
        outline: none;
        box-shadow: none;
        color: white;
        border-bottom: 1px solid ${p => p.theme.colors.morMain};
    }
`