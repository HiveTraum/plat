const API_BASE = "/api";

export async function fetchHealth(): Promise<{ status: string }> {
  const res = await fetch(`${API_BASE}/health`);
  if (!res.ok) throw new Error("Health check failed");
  return res.json();
}

export interface User {
  id: string;
  email: string;
}

export async function fetchMe(): Promise<User> {
  const res = await fetch(`${API_BASE}/me`, { credentials: "include" });
  if (!res.ok) throw new Error("Not authenticated");
  return res.json();
}
