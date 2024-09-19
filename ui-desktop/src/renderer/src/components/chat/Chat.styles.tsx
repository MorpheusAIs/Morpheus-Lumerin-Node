import styled from 'styled-components';
import TextareaAutosize from 'react-textarea-autosize';

export const View = styled.div`
  height: 100vh;
  max-width: 100%;
  min-width: 600px;
  position: relative;
`;

export const Container = styled.div`
    max-width: 1120px;
    height: calc(100% - 200px);
    justify-content: space-between;
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    padding: 20px 2.4rem 0;
`

export const ChatBlock = styled.div`
    width: 100%;
    height: 100%;
    overflow-y: auto;
    margin-bottom: 20px;

    &.createSessionMode {
        display: flex;
        align-items: center;
        justify-content: center;
    }

    &.createSessionMode .session-container {
        width: 450px;
        padding: 1rem;
        background-color: rgba(138, 43, 226, 0.05);
        border: 1px solid rgba(255, 255, 255, 0.16);
    }

    &.createSessionMode .session-title {
        text-align: center;
        margin-bottom: 10px;
    }
`

export const Control = styled.div`
    height: fit-content;
    position: relative;
    display: flex;
    flex-direction: column;

    textarea {
        resize: none;
        padding-right: 6rem;
    }

    textarea:focus, input:focus{
        outline: none!important;
    }
`

export const SendBtn = styled.div`
    position: absolute;
    right: 16px;
    border-radius: 5px;
    width: fit-content;
    padding: 2px 5px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    text-align: center;
    bottom: 12px;
    background: ${p => p.theme.colors.morMain};
    cursor: pointer;

    &[disabled] {
        opacity: 0.5;
    }
`

export const Avatar = styled.div`
    height: 36px;
    min-width: 36px;
    width: 36px;
    display: flex;
    justify-content: center;
    align-items: center;
    /* border: 1px solid; */
    background: ${p => p.color};
    font-weight: 400;
    font-size: 15px;
    border-radius: 4px;
`

export const AvatarHeader = styled.div`
    color: ${p => p.theme.colors.morMain}
    font-weight: 900;
    padding: 0 8px;
    font-size: 18px;
    line-height: 18px;
    margin-bottom: 5px;
`

export const MessageBody = styled.div`
    font-weight: 400;
    padding: 0 8px;
    font-size: 18px;
    max-width: calc(100vw - 165px);

    code {
        color: ${p => p.theme.colors.morMain}
    }

    @media (min-width: 800px) {
        max-width: calc(100vw - 310px);
    }
`

export const ChatTitleContainer = styled.div`
    color: ${p => p.theme.colors.morMain}
    font-weight: 900;
    padding: 0 8px;
    font-size: 18px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 24px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.16);
`

export const ChatAvatar = styled.div`
    display: flex;
    align-items: center;
`

export const CustomTextArrea = styled(TextareaAutosize)`
    background: transparent;
    box-sizing: border-box;
    width: 100%;
    font-size: 18px;
    border-radius: 12px;
    color: white;
    padding: 12px 16px;

    &::focus {
        outline: none!important;
    }

    textarea:focus, input:focus{
        outline: none!important;
    }
`

export const ContainerTitle = styled.div`
  display: flex;
  flex-direction: row;
  justify-content: space-between;
  align-items: center;
  position: sticky;
  width: 100%;
  padding: 0 2.4rem;
  z-index: 2;
  right: 0;
  left: 0;
  top: 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.16);
`;

export const TitleRow = styled.div`
  width: 100%;
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
`;

export const Title = styled.label`
  font-size: 2.4rem;
  line-height: 3rem;
  white-space: nowrap;
  margin: 0;
  max-width: 1120px;
  font-weight: 600;
  color: ${p => p.theme.colors.morMain};
  margin-bottom: 4.8px;
  margin-right: 2.4rem;
  cursor: default;
  /* width: 100%; */

  @media (min-width: 1140px) {
  }

  @media (min-width: 1200px) {
  }
`;

export const LoadingCover = styled.div`
    width: 100%;
    height: 100%;
    position: absolute;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(0,0,0,0.4);

    z-index: 5;
`

export const ImageContainer = styled.img`
    cursor: pointer;
    padding: 0.25rem;
    background-color: ${p => p.theme.colors.morMain}B3;
    border: var(--bs-border-width) solid var(--bs-highlight-color);
    border-radius: var(--bs-border-radius);
    max-width: 100%;
    height: 256px;

    @media (min-height: 700px) { 
        height: 320px;
    } 
`

export const SubPriceLabel = styled.span`
  color: ${p => p.theme.colors.morMain};
`