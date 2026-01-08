export type AuthTokens = {
  accessToken: string;
  refreshToken: string;
  expiresIn?: number | null;
};

export type UserContext = {
  userId: string;
  roles: string[];
};

const ACCESS_KEY = "dip_auth_access";
const REFRESH_KEY = "dip_auth_refresh";
const EXPIRES_KEY = "dip_auth_expires";
const USER_ID_KEY = "dip_auth_user_id";
const ROLES_KEY = "dip_auth_roles";

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
    storage.removeItem(USER_ID_KEY);
    storage.removeItem(ROLES_KEY);
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

export function hasRememberedTokens() {
  if (typeof window === "undefined") return false;
  return Boolean(localStorage.getItem(ACCESS_KEY));
}

export function storeUserContext(context: UserContext, remember: boolean) {
  if (typeof window === "undefined") return;
  const storage = getStorage(remember);
  storage.setItem(USER_ID_KEY, context.userId);
  storage.setItem(ROLES_KEY, JSON.stringify(context.roles));
}

export function getUserContext(): UserContext | null {
  if (typeof window === "undefined") return null;
  const userId =
    localStorage.getItem(USER_ID_KEY) ?? sessionStorage.getItem(USER_ID_KEY);
  const rolesRaw =
    localStorage.getItem(ROLES_KEY) ?? sessionStorage.getItem(ROLES_KEY);

  if (!userId) return null;

  let roles: string[] = [];
  if (rolesRaw) {
    try {
      const parsed = JSON.parse(rolesRaw);
      if (Array.isArray(parsed)) {
        roles = parsed.filter((role) => typeof role === "string");
      }
    } catch {
      roles = [];
    }
  }

  return { userId, roles };
}
