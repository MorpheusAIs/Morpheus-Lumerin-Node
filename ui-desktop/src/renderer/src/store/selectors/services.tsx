import { createSelector } from 'reselect';
import type { LoadingState } from 'src/main/orchestrator.types';

export const getServices = (state: any): LoadingState => state.services;

export const getServicesState = createSelector(
  getServices,
  (services) => services,
);
