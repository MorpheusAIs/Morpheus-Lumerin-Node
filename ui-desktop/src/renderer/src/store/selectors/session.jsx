// Returns if user is logged id or not
export const getIsLoggedIn = state => state.session.isLoggedIn;

// Returns if session is active or not
// Possible improvements: expire session after timeout
export const isSessionActive = getIsLoggedIn;

// Returns if renderer has enough data to load the wallet UI.
// Renderer will display the "Gathering data..." screen until it does.
export const hasEnoughData = state => state.session.hasEnoughData;
