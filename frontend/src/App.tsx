import { useQuery } from "@tanstack/react-query";
import { fetchHealth, fetchMe } from "./api";

export function App() {
  const health = useQuery({
    queryKey: ["health"],
    queryFn: fetchHealth,
  });

  const me = useQuery({
    queryKey: ["me"],
    queryFn: fetchMe,
    retry: false,
  });

  return (
    <div style={{ maxWidth: 600, margin: "40px auto", fontFamily: "sans-serif" }}>
      <h1>Plat</h1>

      <section>
        <h2>API Health</h2>
        {health.isLoading && <p>Checking...</p>}
        {health.isError && <p>API unavailable</p>}
        {health.data && <p>Status: {health.data.status}</p>}
      </section>

      <section>
        <h2>User</h2>
        {me.isLoading && <p>Loading...</p>}
        {me.isError && <p>Not authenticated. Please log in via Kratos.</p>}
        {me.data && (
          <div>
            <p>ID: {me.data.id}</p>
            <p>Email: {me.data.email}</p>
          </div>
        )}
      </section>
    </div>
  );
}
