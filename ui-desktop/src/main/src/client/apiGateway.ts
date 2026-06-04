export type ApiGateway = {
  getAuthHeaders: () => Promise<Record<string, string>>;
};
