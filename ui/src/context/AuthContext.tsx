import {
  createContext,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";
import type { components } from "../api/types";
import { apiFetch } from "../api/client";

type User = components["schemas"]["UserResponse"];

// Cookie names — must match internal/domain/cookies.go
const COOKIE_USER_ID = "user_id";
const COOKIE_SESSION_EXPIRES_AT = "session_expires_at";

// Sessions are refreshed automatically when within this many ms of expiry.
const REAUTH_THRESHOLD_MS = 2 * 60 * 60 * 1000; // 2 hours

export type AuthState =
  | { status: "loading" }
  | { status: "unauthenticated" }
  | { status: "authenticated"; user: User };

type AuthContextValue = {
  auth: AuthState;
  login: (userId: string) => Promise<void>;
  register: (userId: string, displayName: string) => Promise<void>;
  logout: () => Promise<void>;
  updateUser: (user: User) => void;
  updateDisplayName: (displayName: string) => Promise<void>;
  checkUserIdFree: (userId: string) => Promise<boolean>;
};

const AuthContext = createContext<AuthContextValue | null>(null);

// ---------------------------------------------------------------------------
// Cookie helpers
// ---------------------------------------------------------------------------

function getCookie(name: string): string | null {
  const match = document.cookie
    .split("; ")
    .find((row) => row.startsWith(name + "="));
  return match ? decodeURIComponent(match.split("=")[1]) : null;
}

/**
 * Parses Go's time.Time.String() format:
 *   "2006-01-02 15:04:05.999999999 -0700 MST [m=+...]"
 * Returns null if the string cannot be parsed.
 */
function parseGoTime(s: string): Date | null {
  // Extract: date, time (with optional fractional seconds), numeric offset.
  // Everything after (timezone abbreviation, monotonic clock) is ignored.
  const match = s.match(
    /^(\d{4}-\d{2}-\d{2}) (\d{2}:\d{2}:\d{2}(?:\.\d+)?) ([+-]\d{4})/,
  );
  if (!match) return null;
  const d = new Date(`${match[1]}T${match[2]}${match[3]}`);
  return isNaN(d.getTime()) ? null : d;
}

/**
 * Returns true if the stored session is expired or within REAUTH_THRESHOLD_MS
 * of expiry. Returns false if the cookie is absent or unparseable (let the
 * normal GET /api/auth check decide in that case).
 */
function sessionNeedsRefresh(): boolean {
  const expiryStr = getCookie(COOKIE_SESSION_EXPIRES_AT);
  if (!expiryStr) return false;
  const expiry = parseGoTime(expiryStr);
  if (!expiry) return false;
  return expiry.getTime() - Date.now() <= REAUTH_THRESHOLD_MS;
}

// ---------------------------------------------------------------------------
// Provider
// ---------------------------------------------------------------------------

export function AuthProvider({ children }: { children: ReactNode }) {
  const [auth, setAuth] = useState<AuthState>({ status: "loading" });

  // POST /api/auth — creates a new session for userId.
  async function login(userId: string): Promise<void> {
    const r = await apiFetch("/api/auth", {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ user_id: userId }),
    });
    if (!r.ok) throw new Error("Login failed");
    const user: User = await r.json();
    setAuth({ status: "authenticated", user });
  }

  // POST /api/user then auto-login.
  async function register(userId: string, displayName: string): Promise<void> {
    const r = await apiFetch("/api/user", {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ user_id: userId, display_name: displayName }),
    });
    if (!r.ok) throw new Error("Registration failed");
    await login(userId);
  }

  // DELETE /api/auth — clears the server session and local state.
  async function logout(): Promise<void> {
    await apiFetch("/api/auth", { method: "DELETE", credentials: "include" });
    setAuth({ status: "unauthenticated" });
  }

  // Sync an updated User object into context (e.g. after display name change).
  function updateUser(user: User): void {
    setAuth({ status: "authenticated", user });
  }

  // Update display name with the server.
  async function updateDisplayName(displayName: string): Promise<void> {
    const r = await apiFetch("/api/user", {
      method: "PUT",
      credentials: "include",
      body: JSON.stringify({ display_name: displayName }),
    });
    if (!r.ok) throw new Error("Display name update failed");
    const user: User = await r.json();
    setAuth({ status: "authenticated", user });
  }

  async function checkUserIdFree(userId: string): Promise<boolean> {
    const r = await apiFetch(`/api/user?user_id=${encodeURIComponent(userId)}`);
    // 404 → free, 200 → taken, anything else → assume taken to be safe
    return r.status === 404;
  }

  // On mount: auto-reauth if the stored session is about to expire, otherwise
  // validate the existing session cookie with the server.
  useEffect(() => {
    async function checkSession() {
      if (sessionNeedsRefresh()) {
        const userId = getCookie(COOKIE_USER_ID);
        if (userId) {
          try {
            await login(userId);
            return;
          } catch {
            // Session refresh failed — fall through to the normal check.
          }
        }
      }

      // Normal path: ask the server whether the current cookie is valid.
      try {
        const r = await apiFetch("/api/auth", { credentials: "include" });
        if (r.ok) {
          const user: User = await r.json();
          setAuth({ status: "authenticated", user });
        } else {
          setAuth({ status: "unauthenticated" });
        }
      } catch {
        setAuth({ status: "unauthenticated" });
      }
    }

    checkSession();

    const interval = setInterval(checkSession, 60 * 60 * 1000); // every hour
    return () => clearInterval(interval);
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <AuthContext.Provider
      value={{
        auth,
        login,
        register,
        logout,
        updateUser,
        updateDisplayName,
        checkUserIdFree,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

// ---------------------------------------------------------------------------
// Hook
// ---------------------------------------------------------------------------

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within <AuthProvider>");
  return ctx;
}
