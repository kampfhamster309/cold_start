// Server Component: fetches app-backend's health endpoint server-side, over
// the docker-compose internal network (architecture §2.1 — app-frontend
// talks to app-backend only over its REST API, never a direct DB/git
// connection). This is the round-trip INFRA-2 exists to prove.

const BACKEND_URL = process.env.BACKEND_INTERNAL_URL ?? "http://localhost:8080";

type HealthResponse = {
  status: string;
};

async function getBackendHealth(): Promise<
  { ok: true; data: HealthResponse } | { ok: false; error: string }
> {
  try {
    const res = await fetch(`${BACKEND_URL}/healthz`, { cache: "no-store" });
    if (!res.ok) {
      return { ok: false, error: `backend responded with ${res.status}` };
    }
    return { ok: true, data: (await res.json()) as HealthResponse };
  } catch (err) {
    return { ok: false, error: err instanceof Error ? err.message : String(err) };
  }
}

export default async function Home() {
  const health = await getBackendHealth();

  return (
    <main style={{ fontFamily: "system-ui, sans-serif", padding: "2rem" }}>
      <h1>cold_start</h1>
      <p>app-frontend → app-backend round trip (INFRA-2):</p>
      {health.ok ? (
        <p>
          ✅ <code>{BACKEND_URL}/healthz</code> →{" "}
          <code>{JSON.stringify(health.data)}</code>
        </p>
      ) : (
        <p>
          ❌ could not reach <code>{BACKEND_URL}/healthz</code>: {health.error}
        </p>
      )}
    </main>
  );
}
