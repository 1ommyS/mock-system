import { createEvent, createStore } from "effector";

export type AuthState = {
  isAuthenticated: boolean;
  accessToken: string | null;
  refreshToken: string | null;
  userId: string | null;
  roles: string[];
};

export const authLoggedIn = createEvent<{
  accessToken: string;
  refreshToken: string;
  userId?: string | null;
  roles?: string[];
}>();

export const authLoggedOut = createEvent();

export const $auth = createStore<AuthState>({
  isAuthenticated: false,
  accessToken: null,
  refreshToken: null,
  userId: null,
  roles: [],
})
  .on(authLoggedIn, (_, payload) => ({
    isAuthenticated: true,
    accessToken: payload.accessToken,
    refreshToken: payload.refreshToken,
    userId: payload.userId ?? null,
    roles: payload.roles ?? [],
  }))
  .on(authLoggedOut, () => ({
    isAuthenticated: false,
    accessToken: null,
    refreshToken: null,
    userId: null,
    roles: [],
  }));
