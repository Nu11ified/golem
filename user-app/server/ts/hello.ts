export default async function handler({
  body,
  query,
  headers,
}: {
  body: Record<string, unknown>;
  query: Record<string, unknown>;
  headers: Record<string, string | string[] | undefined>;
}) {
  try {
    return {
      message: "Hello from TypeScript!",
      received: { body, query, headers }
    };
  } catch (err: any) {
    return {
      error: err?.message || 'Unknown error',
      stack: err?.stack || null
    };
  }
} 