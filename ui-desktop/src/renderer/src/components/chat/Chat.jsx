import React, { useContext, useState } from 'react';
import styled from 'styled-components';
import { LayoutHeader } from '../common/LayoutHeader';
// import component 👇
import Drawer from 'react-modern-drawer'
import { IconHistory } from '@tabler/icons-react';

//import styles 👇
import 'react-modern-drawer/dist/index.css'

const View = styled.div`
  height: 100vh;
  max-width: 100%;
  min-width: 600px;
  position: relative;
  padding-top: 2rem;
`;

const Container = styled.div`
    display: flex;
    flex-direction: column;
    padding-left: 2.4rem;
`

const ChatBlock = styled.div`
    width: 100%;
    display: flex;
    flex-direction: column;
    justify-content: end;
    height: calc(100vh - 240px);
    margin-bottom: 20px;
`

const Control = styled.div`
    height: 100%;
    display: flex;
    flex-direction: column;

    textarea {
        resize: none;
    }
`

const Avatar = styled.div`
    height: 36px;
    min-width: 36px;
    width: 36px;
    display: flex;
    justify-content: center;
    align-items: center;
    /* border: 1px solid; */
    background: green;
    font-weight: 400;
    font-size: 15px;
    border-radius: 4px;
`

const AvatarHeader = styled.div`
    color: ${p => p.theme.colors.morMain}
    font-weight: 900;
    padding: 0 8px;
    font-size: 18px;
`

const MessageBody = styled.div`
    font-weight: 400;
    padding: 0 8px;
    font-size: 18px;
`

const ChatTitleContainer = styled.div`
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

const ChatAvatar = styled.div`
    display: flex;
    align-items: center;
`

const Loeer = "Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum."

const Chat = (props) => {
    const [messages, setMessages] = useState([]);
    const [isOpen, setIsOpen] = useState(false);
    const toggleDrawer = () => {
        setIsOpen((prevState) => !prevState)
    }

    return (
        <>
            <Drawer
                open={isOpen}
                onClose={toggleDrawer}
                direction='right'
                className='test'
            >
                <div>History</div>
            </Drawer>
            <View>
                <ContainerTitle style={{ padding: '0 2.4rem' }}>
                    <TitleRow>
                        <Title>Chat</Title>
                    </TitleRow>
                </ContainerTitle>
                <ChatTitleContainer>
                    <ChatAvatar>
                        <Avatar style={{ color: 'white'}}>
                            L
                        </Avatar>
                        <div style={{ marginLeft: '10px'}}>Llama GPT</div>
                    </ChatAvatar>
                    <div>Provider: 0x123...234</div>
                    <div>
                        <div onClick={toggleDrawer}>
                            <IconHistory size={"2.4rem"}></IconHistory>
                        </div>
                    </div>
                </ChatTitleContainer>

                <Container>
                    <ChatBlock>
                        <Message message={{ user: 'Me', text: "What is Lorem Ipsum?", icon: "M" }}></Message>
                        <Message message={{ user: 'GPT', text: Loeer, icon: "GPT" }}></Message>
                    </ChatBlock>
                    <Control>
                        <div>
                            <textarea>Text</textarea>
                        </div>
                    </Control>
                </Container>
            </View>
        </>
    )
}

const Message = ({ message }) => {
    return (
        <div style={{ display: 'flex', margin: '12px 0' }}>
            <Avatar>
                {message.icon}
            </Avatar>
            <div>
                <AvatarHeader>{message.user}</AvatarHeader>
                <MessageBody>{message.text}</MessageBody>
            </div>
        </div>)
}

export default Chat;

const ContainerTitle = styled.div`
  display: flex;
  flex-direction: row;
  justify-content: space-between;
  align-items: center;
  position: sticky;
  width: 100%;
  padding: 1.5rem 0;
  z-index: 2;
  right: 0;
  left: 0;
  top: 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.16);
  padding-bottom: 32px!important;
`;

const TitleRow = styled.div`
  width: 100%;
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
`;

const Title = styled.label`
  font-size: 2.4rem;
  line-height: 3rem;
  white-space: nowrap;
  margin: 0;
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