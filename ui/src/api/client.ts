// Base URL for API requests. Set VITE_BACKEND_URL in your environment to point
// at a different origin (e.g. http://localhost:8080 during development).
// Defaults to "" so that paths like "/api/auth" are relative to the current origin.
const BASE_URL: string = import.meta.env.VITE_BACKEND_URL ?? "";

export function apiFetch(path: string, init?: RequestInit): Promise<Response> {
  return fetch(`${BASE_URL}${path}`, init);
}
