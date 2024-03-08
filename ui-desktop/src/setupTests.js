global.window.require = function() {
  return {
    ipcRenderer: {
      sendSync: jest.fn(),
      send: jest.fn(),
      on: jest.fn()
    },
    clipboard: {
      writeText: jest.fn()
    },
    shell: {
      openExternal: jest.fn()
    }
  };
};

/**
 * react-modal uses Portals to append the modal node to the end of the document
 * instead of the place where the <Modal /> component is placed, so we need to
 * mock the package behavior and make it render the modal children in-place.
 */
jest.mock('react-modal', () => props => (props.isOpen ? props.children : null));

/**
 * mock qrcode as Canvas is not implemented in jsdom
 */
jest.mock('qrcode.react', () => props => null);

Object.defineProperty(window.HTMLElement.prototype, 'dataset', {
  writable: true,
  value: {}
});
