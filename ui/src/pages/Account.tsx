import { useState, useEffect } from "react";
import { useAuth } from "../context/AuthContext";
import { EditIcon } from "../components/icons";

// ---------------------------------------------------------------------------
// Validation helpers (mirrors backend rules)
// ---------------------------------------------------------------------------

function validateUserId(id: string): string | null {
  if (id.length === 0) return null; // empty is handled by `required`
  if (id !== id.trim()) return "User ID must not start or end with a space";
  return null;
}

function validateDisplayName(name: string): string | null {
  if (name.length === 0) return null; // empty is handled by `required`
  if (name.trim() === "") return "Display name must contain at least one non-space character";
  return null;
}

// ---------------------------------------------------------------------------
// Shared input border helper
// ---------------------------------------------------------------------------

type FieldState = "empty" | "invalid" | "valid";

function borderClass(state: FieldState, overrideValid?: string): string {
  if (state === "invalid") return "border-red-500";
  if (state === "valid") return overrideValid ?? "border-green-500";
  return "border-gray-600";
}

// ---------------------------------------------------------------------------
// Login / Register form
// ---------------------------------------------------------------------------

type LoggedOutTab = "login" | "register";

// For user ID availability: invalid format takes priority over taken/free.
type UserIdState = "empty" | "invalid" | "checking" | "free" | "taken";

function LoginRegister() {
  const { login, register, checkUserIdFree } = useAuth();

  const [tab, setTab] = useState<LoggedOutTab>("login");
  const [error, setError] = useState<string | null>(null);

  // Login form
  const [loginUserId, setLoginUserId] = useState("");
  const [loginUserIdState, setLoginUserIdState] = useState<FieldState>("empty");

  // Register form
  const [regUserId, setRegUserId] = useState("");
  const [regUserIdState, setRegUserIdState] = useState<UserIdState>("empty");
  const [regDisplayName, setRegDisplayName] = useState("");
  const [regDisplayNameState, setRegDisplayNameState] = useState<FieldState>("empty");

  // Validate login user ID immediately on change
  function handleLoginUserIdChange(value: string) {
    setLoginUserId(value);
    if (!value) { setLoginUserIdState("empty"); return; }
    setLoginUserIdState(validateUserId(value) ? "invalid" : "valid");
  }

  // Validate register display name immediately on change
  function handleRegDisplayNameChange(value: string) {
    setRegDisplayName(value);
    if (!value) { setRegDisplayNameState("empty"); return; }
    setRegDisplayNameState(validateDisplayName(value) ? "invalid" : "valid");
  }

  // Register user ID: validate format immediately, then debounce availability check
  useEffect(() => {
    setRegUserIdState("empty");
    if (!regUserId) return;

    const formatError = validateUserId(regUserId);
    if (formatError) {
      setRegUserIdState("invalid");
      return;
    }

    setRegUserIdState("checking");
    const timeout = setTimeout(async () => {
      const free = await checkUserIdFree(regUserId);
      setRegUserIdState(free ? "free" : "taken");
    }, 400);

    return () => clearTimeout(timeout);
  }, [regUserId, checkUserIdFree]);

  async function handleLogin(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    try {
      await login(loginUserId);
    } catch {
      setError("Login failed — check your user ID.");
    }
  }

  async function handleRegister(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    try {
      await register(regUserId, regDisplayName);
    } catch {
      setError("Registration failed — user ID may already be taken.");
    }
  }

  // Map UserIdState → FieldState for borderClass
  const regUserIdFieldState: FieldState =
    regUserIdState === "free" ? "valid"
    : regUserIdState === "taken" || regUserIdState === "invalid" ? "invalid"
    : "empty";

  return (
    <div className="max-w-sm">
      <div className="flex gap-4 mb-6">
        <button
          onClick={() => { setTab("login"); setError(null); }}
          className={tab === "login" ? "font-bold text-orange-200 underline" : "text-orange-400"}
        >
          Log In
        </button>
        <button
          onClick={() => { setTab("register"); setError(null); }}
          className={tab === "register" ? "font-bold text-orange-200 underline" : "text-orange-400"}
        >
          Register
        </button>
      </div>

      {tab === "login" && (
        <form onSubmit={handleLogin} className="flex flex-col gap-3">
          <div>
            <input
              className={`w-full p-2 bg-gray-800 text-white rounded border transition-colors ${borderClass(loginUserIdState)}`}
              placeholder="User ID"
              value={loginUserId}
              onChange={(e) => handleLoginUserIdChange(e.target.value)}
              required
            />
            {loginUserIdState === "invalid" && (
              <p className="mt-1 text-xs text-red-400">{validateUserId(loginUserId)}</p>
            )}
          </div>
          <button className="p-2 bg-orange-600 rounded hover:bg-orange-500" type="submit">
            Log In
          </button>
        </form>
      )}

      {tab === "register" && (
        <form onSubmit={handleRegister} className="flex flex-col gap-3">
          <div>
            <input
              className={`w-full p-2 bg-gray-800 text-white rounded border transition-colors ${borderClass(regUserIdFieldState)}`}
              placeholder="User ID"
              value={regUserId}
              onChange={(e) => setRegUserId(e.target.value)}
              required
            />
            {regUserIdState === "invalid" && (
              <p className="mt-1 text-xs text-red-400">{validateUserId(regUserId)}</p>
            )}
            {regUserIdState === "taken" && (
              <p className="mt-1 text-xs text-red-400">User ID already taken</p>
            )}
            {regUserIdState === "free" && (
              <p className="mt-1 text-xs text-green-400">User ID available</p>
            )}
            {regUserIdState === "checking" && (
              <p className="mt-1 text-xs text-gray-400">Checking...</p>
            )}
          </div>
          <div>
            <input
              className={`w-full p-2 bg-gray-800 text-white rounded border transition-colors ${borderClass(regDisplayNameState)}`}
              placeholder="Display Name"
              value={regDisplayName}
              onChange={(e) => handleRegDisplayNameChange(e.target.value)}
              maxLength={50}
              required
            />
            {regDisplayNameState === "invalid" && (
              <p className="mt-1 text-xs text-red-400">{validateDisplayName(regDisplayName)}</p>
            )}
          </div>
          <button className="p-2 bg-orange-600 rounded hover:bg-orange-500" type="submit">
            Register
          </button>
        </form>
      )}

      {error && <p className="mt-3 text-red-400">{error}</p>}
    </div>
  );
}

// ---------------------------------------------------------------------------
// Account info
// ---------------------------------------------------------------------------

function AccountInfo() {
  const { auth, updateDisplayName, logout } = useAuth();
  if (auth.status !== "authenticated") return null;
  const { user } = auth;

  const [editing, setEditing] = useState(false);
  const [newDisplayName, setNewDisplayName] = useState(user.display_name);
  const [displayNameState, setDisplayNameState] = useState<FieldState>("valid");
  const [error, setError] = useState<string | null>(null);

  function handleDisplayNameChange(value: string) {
    setNewDisplayName(value);
    if (!value) { setDisplayNameState("empty"); return; }
    setDisplayNameState(validateDisplayName(value) ? "invalid" : "valid");
  }

  async function handleSave(e: React.FormEvent) {
    e.preventDefault();
    if (displayNameState === "invalid") return;
    setError(null);
    try {
      await updateDisplayName(newDisplayName);
      setEditing(false);
    } catch {
      setError("Failed to update display name.");
    }
  }

  function handleCancel() {
    setNewDisplayName(user.display_name);
    setDisplayNameState("valid");
    setEditing(false);
    setError(null);
  }

  const createdAt = new Date(user.created_at).toLocaleDateString(undefined, {
    year: "numeric",
    month: "long",
    day: "numeric",
  });

  return (
    <div className="max-w-sm flex flex-col gap-6">
      {/* User ID — read only */}
      <div>
        <label className="block text-xs text-gray-500 uppercase tracking-wide mb-1">
          User ID
        </label>
        <p className="text-gray-500 text-xl">{user.user_id}</p>
      </div>

      {/* Display name — editable */}
      <div>
        <label className="block text-xs text-gray-500 uppercase tracking-wide mb-1">
          Display Name
        </label>
        {editing ? (
          <form onSubmit={handleSave} className="flex flex-col gap-2">
            <input
              className={`p-2 bg-gray-800 text-white rounded border transition-colors focus:outline-none ${borderClass(displayNameState, "border-orange-500")}`}
              value={newDisplayName}
              onChange={(e) => handleDisplayNameChange(e.target.value)}
              maxLength={50}
              autoFocus
              required
            />
            {displayNameState === "invalid" && (
              <p className="text-xs text-red-400">{validateDisplayName(newDisplayName)}</p>
            )}
            <div className="flex gap-2">
              <button
                type="submit"
                disabled={displayNameState === "invalid"}
                className="px-3 py-1 bg-orange-600 hover:bg-orange-500 disabled:opacity-50 disabled:cursor-not-allowed rounded text-white text-sm"
              >
                Save
              </button>
              <button
                type="button"
                onClick={handleCancel}
                className="px-3 py-1 bg-gray-700 hover:bg-gray-600 rounded text-white text-sm"
              >
                Cancel
              </button>
            </div>
            {error && <p className="text-red-400 text-sm">{error}</p>}
          </form>
        ) : (
          <div className="flex items-center gap-2">
            <p className="text-orange-200 text-lg">{user.display_name}</p>
            <button
              onClick={() => setEditing(true)}
              className="text-gray-500 hover:text-orange-400 transition-colors"
              aria-label="Edit display name"
            >
              <EditIcon size={16} />
            </button>
          </div>
        )}
      </div>

      {/* Member since */}
      <div>
        <label className="block text-xs text-gray-500 uppercase tracking-wide mb-1">
          Member Since
        </label>
        <p className="text-gray-400 text-lg">{createdAt}</p>
      </div>

      {/* Log out */}
      <button
        onClick={logout}
        className="self-start px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded text-white text-sm transition-colors"
      >
        Log Out
      </button>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Page
// ---------------------------------------------------------------------------

export default function Account() {
  const { auth } = useAuth();

  return (
    <div>
      <h1 className="text-2xl font-bold text-orange-200 mb-8">Account</h1>
      {auth.status === "loading" && <p className="text-gray-400">Loading...</p>}
      {auth.status === "unauthenticated" && <LoginRegister />}
      {auth.status === "authenticated" && <AccountInfo />}
    </div>
  );
}
