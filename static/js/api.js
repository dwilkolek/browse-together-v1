async function apiCreateSession() {
  const response = await fetch("/api/v1/sessions", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      name: "aaa",
      baseLocation: "aaa",
      creator: "aaa",
    }),
  });
  return await response.json();
}

async function apiListSessions() {
  const response = await fetch("/api/v1/sessions", {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
    },
  });
  return await response.json();
}
