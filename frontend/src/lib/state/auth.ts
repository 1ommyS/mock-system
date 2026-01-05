import { createEvent, createStore } from "effector";

export type AuthState = {
  isAuthenticated: boolean;
  accessToken: string | null;
  refreshToken: string | null;
};

export const authLoggedIn = createEvent<{
  accessToken: string;
  refreshToken: string;
}>();

export const authLoggedOut = createEvent();

export const $auth = createStore<AuthState>({
  isAuthenticated: false,
  accessToken: null,
  refreshToken: null,
})
  .on(authLoggedIn, (_, payload) => ({
    isAuthenticated: true,
    accessToken: payload.accessToken,
    refreshToken: payload.refreshToken,
  }))
  .on(authLoggedOut, () => ({
    isAuthenticated: false,
    accessToken: null,
    refreshToken: null,
  }));
