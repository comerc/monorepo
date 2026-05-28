import { execFile } from "node:child_process";
import { promisify } from "node:util";
import type { Page } from "@playwright/test";

const execFileAsync = promisify(execFile);

const graphQLEndpoint =
  process.env.BDD_GRAPHQL_URL ?? "http://127.0.0.1:3002/graphql";
const mailpitEndpoint =
  process.env.BDD_MAILPIT_URL ?? "http://127.0.0.1:8025";
const runID = Date.now().toString(36);

const pageState = new WeakMap<
  Page,
  {
    emailByExample: Map<string, string>;
    lastRequestedEmail?: string;
    lastReceivedCode?: string;
  }
>();

interface GraphQLResponse<T> {
  data?: T;
  errors?: Array<{ message: string; extensions?: unknown }>;
}

interface RequestEmailCodeData {
  requestEmailCode: {
    accepted: boolean;
    retryAfterSeconds: number;
    nextAllowedAt?: string | null;
  };
}

interface LoginWithEmailCodeData {
  loginWithEmailCode: {
    token: string;
    userID: string;
    email: string;
  };
}

interface MailpitMessages {
  messages?: MailpitSummary[];
  Messages?: MailpitSummary[];
}

interface MailpitSummary {
  ID?: string;
  id?: string;
  To?: Array<{ Address?: string; address?: string }>;
  to?: Array<{ Address?: string; address?: string }>;
  Snippet?: string;
  snippet?: string;
}

interface MailpitMessage extends MailpitSummary {
  Text?: string;
  text?: string;
  HTML?: string;
  html?: string;
}

export async function resetLiveBackendState() {
  await Promise.all([
    deleteBackendUserData(),
    execFileAsync("docker", ["exec", "monorepo-redis", "redis-cli", "FLUSHDB"]),
    fetch(`${mailpitEndpoint}/api/v1/messages`, { method: "DELETE" }).catch(
      () => undefined,
    ),
  ]);
}

export async function deleteBackendUserData() {
  await execFileAsync("docker", [
    "exec",
    "monorepo-postgres",
    "psql",
    "-U",
    "monorepo",
    "-d",
    "monorepo",
    "-c",
    "TRUNCATE profiles, users;",
  ]);
}

export function requestedEmail(page: Page) {
  const email = stateFor(page).lastRequestedEmail;
  if (!email) {
    throw new Error("email code was not requested");
  }
  return email;
}

export function receivedCode(page: Page) {
  const code = stateFor(page).lastReceivedCode;
  if (!code) {
    throw new Error("email code was not received");
  }
  return code;
}

export async function rememberReceivedCode(page: Page) {
  const state = stateFor(page);
  if (!state.lastRequestedEmail) {
    throw new Error("email code was not requested");
  }
  state.lastReceivedCode = await waitEmailCode(state.lastRequestedEmail);
}

export function resolveScenarioEmail(page: Page, exampleEmail: string) {
  if (!exampleEmail.includes("@")) {
    return exampleEmail;
  }
  const state = stateFor(page);
  const existing = state.emailByExample.get(exampleEmail);
  if (existing) {
    return existing;
  }
  const [localPart, domain] = exampleEmail.toLowerCase().split("@");
  const resolved = `${localPart}+${runID}-${state.emailByExample.size + 1}-${Math.random().toString(36).slice(2, 8)}@${domain}`;
  state.emailByExample.set(exampleEmail, resolved);
  return resolved;
}

export function markCodeRequested(page: Page, email: string) {
  const state = stateFor(page);
  state.lastRequestedEmail = email;
  state.lastReceivedCode = undefined;
}

export async function loginByEmail(
  page: Page,
  exampleEmail: string,
  nickname?: string,
) {
  const email = resolveScenarioEmail(page, exampleEmail);
  return loginByActualEmail(page, email, nickname);
}

export async function loginByActualEmail(
  page: Page,
  email: string,
  nickname?: string,
) {
  const response = await requestEmailCode(email);
  if (!response.accepted) {
    throw new Error(`email code request was throttled for ${email}`);
  }
  const code = await waitEmailCode(email);
  const session = await loginWithEmailCode(email, code);
  if (nickname !== undefined) {
    await setNickname(session.token, nickname);
  }
  await seedBrowserSession(page, session);
  return session;
}

export async function createUserWithNickname(page: Page, nickname: string) {
  const email = resolveScenarioEmail(page, `other-${nickname}@example.com`);
  try {
    const response = await requestEmailCode(email);
    if (!response.accepted) {
      throw new Error(`email code request was throttled for ${email}`);
    }
    const code = await waitEmailCode(email);
    const session = await loginWithEmailCode(email, code);
    await setNickname(session.token, nickname);
  } catch (error) {
    if (
      error instanceof Error &&
      error.message.includes("nickname is already taken")
    ) {
      return;
    }
    throw error;
  }
}

export async function seedBrowserSession(
  page: Page,
  session: LoginWithEmailCodeData["loginWithEmailCode"],
) {
  const authState = {
    state: {
      token: session.token,
      user: {
        id: session.userID,
        email: session.email,
        nickname: null,
      },
      isAuthenticated: true,
    },
    version: 0,
  };
  await page.addInitScript((value) => {
    window.localStorage.setItem("auth", JSON.stringify(value));
  }, authState);
  if (!page.isClosed()) {
    await page
      .evaluate((value) => {
        window.localStorage.setItem("auth", JSON.stringify(value));
      }, authState)
      .catch(() => undefined);
  }
}

export async function expectTokenRevoked(token: string) {
  try {
    await graphQL(
      `mutation Logout($allDevices: Boolean) {
        logout(allDevices: $allDevices) {
          revoked
        }
      }`,
      { allDevices: false },
      token,
    );
  } catch (error) {
    if (
      error instanceof Error &&
      /revoked|invalid|unauthenticated/i.test(error.message)
    ) {
      return;
    }
    throw error;
  }
  throw new Error("revoked token was accepted");
}

async function requestEmailCode(email: string) {
  const data = await graphQL<RequestEmailCodeData>(
    `mutation RequestEmailCode($email: String!) {
      requestEmailCode(email: $email) {
        accepted
        retryAfterSeconds
        nextAllowedAt
      }
    }`,
    { email },
  );
  return data.requestEmailCode;
}

async function loginWithEmailCode(email: string, code: string) {
  const data = await graphQL<LoginWithEmailCodeData>(
    `mutation LoginWithEmailCode($email: String!, $code: String!) {
      loginWithEmailCode(email: $email, code: $code) {
        token
        userID
        email
      }
    }`,
    { email, code },
  );
  return data.loginWithEmailCode;
}

async function setNickname(token: string, nickname: string) {
  await graphQL(
    `mutation SetNickname($nickname: String!) {
      setNickname(nickname: $nickname) {
        userID
        nickname
      }
    }`,
    { nickname },
    token,
  );
}

async function graphQL<T>(
  query: string,
  variables?: Record<string, unknown>,
  token?: string,
) {
  const response = await fetch(graphQLEndpoint, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify({ query, variables }),
  });
  const payload = (await response.json()) as GraphQLResponse<T>;
  if (!response.ok || payload.errors?.length) {
    throw new Error(
      payload.errors?.map((error) => error.message).join("; ") ??
        `graphql status ${response.status}`,
    );
  }
  if (!payload.data) {
    throw new Error("graphql response does not contain data");
  }
  return payload.data;
}

async function waitEmailCode(email: string) {
  const deadline = Date.now() + 15_000;
  while (Date.now() < deadline) {
    const code = await findEmailCode(email);
    if (code) {
      return code;
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`email code for ${email} was not received`);
}

async function findEmailCode(email: string) {
  const response = await fetch(`${mailpitEndpoint}/api/v1/messages`);
  if (!response.ok) {
    throw new Error(`mailpit status ${response.status}`);
  }
  const payload = (await response.json()) as MailpitMessages;
  const messages = payload.messages ?? payload.Messages ?? [];
  for (const summary of messages) {
    if (!mailSentTo(summary, email)) {
      continue;
    }
    const fullMessage = await fetchMail(summary);
    const code = codeFromMail(fullMessage);
    if (code) {
      return code;
    }
  }
  return null;
}

async function fetchMail(summary: MailpitSummary) {
  const id = summary.ID ?? summary.id;
  if (!id) {
    return summary;
  }
  const response = await fetch(`${mailpitEndpoint}/api/v1/message/${id}`);
  if (!response.ok) {
    return summary;
  }
  return (await response.json()) as MailpitMessage;
}

function mailSentTo(message: MailpitSummary, email: string) {
  const recipients = message.To ?? message.to ?? [];
  return recipients.some(
    (recipient) =>
      (recipient.Address ?? recipient.address ?? "").toLowerCase() ===
      email.toLowerCase(),
  );
}

function codeFromMail(message: MailpitMessage) {
  const body = [
    message.Text,
    message.text,
    message.HTML,
    message.html,
    message.Snippet,
    message.snippet,
  ]
    .filter(Boolean)
    .join("\n");
  return body.match(/access code is (\d{5})/i)?.[1] ?? null;
}

function stateFor(page: Page) {
  let state = pageState.get(page);
  if (!state) {
    state = { emailByExample: new Map() };
    pageState.set(page, state);
  }
  return state;
}
