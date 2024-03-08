import React from 'react';

const ClientContext = React.createContext();

export const Provider = ClientContext.Provider;

export const withClient = Component => {
  const WrappedComponent = props => {
    const displayName = `withClient(${Component.displayName ||
      Component.name})`;

    const injectClient = client => <Component {...props} client={client} />;

    return <ClientContext.Consumer>{injectClient}</ClientContext.Consumer>;
  };

  return WrappedComponent;
};
