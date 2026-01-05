export type AuthTokens = {
  accessToken: string;
  refreshToken: string;
  expiresIn?: number | null;
};

const ACCESS_KEY = "dip_auth_access";
const REFRESH_KEY = "dip_auth_refresh";
const EXPIRES_KEY = "dip_auth_expires";

function getStorage(remember: boolean) {
  return remember ? localStorage : sessionStorage;
}

export function storeTokens(tokens: AuthTokens, remember: boolean) {
  if (typeof window === "undefined") return;
  const storage = getStorage(remember);
  storage.setItem(ACCESS_KEY, tokens.accessToken);
  storage.setItem(REFRESH_KEY, tokens.refreshToken);
  if (tokens.expiresIn) {
    storage.setItem(EXPIRES_KEY, String(tokens.expiresIn));
  }
}

export function clearTokens() {
  if (typeof window === "undefined") return;
  for (const storage of [localStorage, sessionStorage]) {
    storage.removeItem(ACCESS_KEY);
    storage.removeItem(REFRESH_KEY);
    storage.removeItem(EXPIRES_KEY);
  }
}

export function getTokens(): AuthTokens | null {
  if (typeof window === "undefined") return null;
  const accessToken =
    localStorage.getItem(ACCESS_KEY) ?? sessionStorage.getItem(ACCESS_KEY);
  const refreshToken =
    localStorage.getItem(REFRESH_KEY) ?? sessionStorage.getItem(REFRESH_KEY);
  const expiresRaw =
    localStorage.getItem(EXPIRES_KEY) ?? sessionStorage.getItem(EXPIRES_KEY);

  if (!accessToken || !refreshToken) return null;

  return {
    accessToken,
    refreshToken,
    expiresIn: expiresRaw ? Number(expiresRaw) : null,
  };
}

export function isAuthenticated() {
  return Boolean(getTokens()?.accessToken);
}

export function getAccessToken() {
  return getTokens()?.accessToken ?? null;
}
