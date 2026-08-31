import { useEffect, useState } from "react";
import type { components } from "../api/types";

type User = components["schemas"]["UserResponse"];

type View = "loading" | "logged-out" | "logged-in";
type LoggedOutTab = "login" | "register";

export default function Home() {
  const [view, setView] = useState<View>("loading");
  const [user, setUser] = useState<User | null>(null);
  const [tab, setTab] = useState<LoggedOutTab>("login");
  const [error, setError] = useState<string | null>(null);

  // Login form
  const [loginUserId, setLoginUserId] = useState("");

  // Register form
  const [regUserId, setRegUserId] = useState("");
  const [regDisplayName, setRegDisplayName] = useState("");

  // Change display name
  const [newDisplayName, setNewDisplayName] = useState("");

  useEffect(() => {
    fetch("/api/auth", { credentials: "include" })
      .then((r) => (r.ok ? r.json() : null))
      .then((u: User | null) => {
        if (u) {
          setUser(u);
          setView("logged-in");
        } else {
          setView("logged-out");
        }
      })
      .catch(() => setView("logged-out"));
  }, []);

  async function handleLogin(e: React.SubmitEvent) {
    e.preventDefault();
    setError(null);
    const r = await fetch("/api/auth", {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ user_id: loginUserId }),
    });
    if (r.ok) {
      const u: User = await r.json();
      setUser(u);
      setView("logged-in");
    } else {
      setError("Login failed — check your user ID.");
    }
  }

  async function handleRegister(e: React.SubmitEvent) {
    e.preventDefault();
    setError(null);
    const r = await fetch("/api/user", {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ user_id: regUserId, display_name: regDisplayName }),
    });
    if (r.ok) {
      // Auto-login after registration
      const loginR = await fetch("/api/auth", {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ user_id: regUserId }),
      });
      if (loginR.ok) {
        const u: User = await loginR.json();
        setUser(u);
        setView("logged-in");
      }
    } else {
      setError("Registration failed — user ID may already be taken.");
    }
  }

  async function handleLogout() {
    await fetch("/api/auth", { method: "DELETE", credentials: "include" });
    setUser(null);
    setView("logged-out");
  }

  async function handleUpdateDisplayName(e: React.SubmitEvent) {
    e.preventDefault();
    setError(null);
    const r = await fetch("/api/user", {
      method: "PUT",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ display_name: newDisplayName }),
    });
    if (r.ok) {
      const u: User = await r.json();
      setUser(u);
      setNewDisplayName("");
    } else {
      setError("Failed to update display name.");
    }
  }

  if (view === "loading") {
    return <p className="p-4 text-orange-200">Loading...</p>;
  }

  if (view === "logged-out") {
    return (
      <div className="p-6 max-w-sm">
        <div className="flex gap-4 mb-4">
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
            <input
              className="p-2 bg-gray-800 text-white border border-gray-600 rounded"
              placeholder="User ID"
              value={loginUserId}
              onChange={(e) => setLoginUserId(e.target.value)}
              required
            />
            <button className="p-2 bg-orange-600 rounded hover:bg-orange-500" type="submit">
              Log In
            </button>
          </form>
        )}

        {tab === "register" && (
          <form onSubmit={handleRegister} className="flex flex-col gap-3">
            <input
              className="p-2 bg-gray-800 text-white border border-gray-600 rounded"
              placeholder="User ID"
              value={regUserId}
              onChange={(e) => setRegUserId(e.target.value)}
              required
            />
            <input
              className="p-2 bg-gray-800 text-white border border-gray-600 rounded"
              placeholder="Display Name"
              value={regDisplayName}
              onChange={(e) => setRegDisplayName(e.target.value)}
              required
            />
            <button className="p-2 bg-orange-600 rounded hover:bg-orange-500" type="submit">
              Register
            </button>
          </form>
        )}

        {error && <p className="mt-3 text-red-400">{error}</p>}
      </div>
    );
  }

  return (
    <div className="p-6 max-w-sm">
      <p className="text-orange-200 mb-1">
        <span className="text-gray-400">User ID:</span> {user?.user_id}
      </p>
      <p className="text-orange-200 mb-4">
        <span className="text-gray-400">Display Name:</span> {user?.display_name}
      </p>

      <form onSubmit={handleUpdateDisplayName} className="flex gap-2 mb-4">
        <input
          className="flex-1 p-2 bg-gray-800 text-white border border-gray-600 rounded"
          placeholder="New display name"
          value={newDisplayName}
          onChange={(e) => setNewDisplayName(e.target.value)}
          required
        />
        <button className="p-2 bg-orange-600 rounded hover:bg-orange-500" type="submit">
          Change
        </button>
      </form>

      <button
        onClick={handleLogout}
        className="p-2 bg-gray-700 rounded hover:bg-gray-600 text-white"
      >
        Log Out
      </button>

      {error && <p className="mt-3 text-red-400">{error}</p>}
    </div>
  );
}
