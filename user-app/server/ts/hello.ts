export default async function handler({ body, query, headers }) {
  return {
    message: "Hello from TypeScript!",
    received: { body, query, headers }
  };
} 