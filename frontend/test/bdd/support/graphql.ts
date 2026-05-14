import type { Page, Route } from "@playwright/test";

const userID = "user-1";
const token = "browser-bdd-token";

interface GraphQLRequest {
  operationName?: string;
  query?: string;
  variables?: Record<string, unknown>;
}

interface BrowserUser {
  id: string;
  email: string;
  nickname: string | null;
}

export interface BrowserSessionState {
  revokedEverywhere: boolean;
}

const takenNicknames = new WeakMap<Page, Set<string>>();
const requestCounts = new WeakMap<Page, number>();
const unavailableEmailCodeRequests = new WeakSet<Page>();

export async function mockGraphQL(
  page: Page,
  email = "user@example.com",
  nickname: string | null = null,
  sessionState: BrowserSessionState = { revokedEverywhere: false },
) {
  const user: BrowserUser = { id: userID, email, nickname };
  takenNicknames.set(page, new Set());
  requestCounts.set(page, 0);

  await page.route("**/graphql", async (route) => {
    const request = route.request().postDataJSON() as GraphQLRequest;
    const operationName =
      request.operationName ?? operationNameFromQuery(request.query);

    switch (operationName) {
      case "RequestEmailCode":
        if (unavailableEmailCodeRequests.has(page)) {
          await respondError(route, "request email code failed");
          return;
        }
        requestCounts.set(page, (requestCounts.get(page) ?? 0) + 1);
        await respond(route, {
          requestEmailCode:
            (requestCounts.get(page) ?? 0) > 1
              ? { accepted: false, retryAfterSeconds: 2, nextAllowedAt: null }
              : { accepted: true, retryAfterSeconds: 2, nextAllowedAt: null },
        });
        return;
      case "LoginWithEmailCode":
        if (request.variables?.code === "00000") {
          await respondError(route, "invalid code");
          return;
        }
        user.email = String(request.variables?.email ?? email);
        await respond(route, {
          loginWithEmailCode: {
            token,
            userID,
            email: user.email,
          },
        });
        return;
      case "MyProfile":
        if (sessionState.revokedEverywhere) {
          await respondError(route, "unauthenticated");
          return;
        }
        await respond(route, {
          myProfile: {
            userID,
            nickname: user.nickname,
          },
        });
        return;
      case "NicknameAvailability": {
        const requestedNickname = String(
          request.variables?.nickname ?? "",
        ).trim();
        await respond(route, {
          nicknameAvailability: {
            nickname: requestedNickname,
            available: !(takenNicknames.get(page) ?? new Set()).has(
              requestedNickname,
            ),
          },
        });
        return;
      }
      case "SetNickname":
        if (sessionState.revokedEverywhere) {
          await respondError(route, "unauthenticated");
          return;
        }
        user.nickname = String(request.variables?.nickname ?? "").trim();
        if (user.nickname.length < 2) {
          await respondError(route, "nickname is too short");
          return;
        }
        if ((takenNicknames.get(page) ?? new Set()).has(user.nickname)) {
          await respondError(route, "nickname is already taken");
          return;
        }
        await respond(route, {
          setNickname: {
            userID,
            nickname: user.nickname,
          },
        });
        return;
      case "Logout":
        if (request.variables?.allDevices === true) {
          sessionState.revokedEverywhere = true;
        }
        await respond(route, { logout: { revoked: true } });
        return;
      default:
        await route.fulfill({
          status: 400,
          contentType: "application/json",
          body: JSON.stringify({
            errors: [{ message: `unsupported operation: ${operationName}` }],
          }),
        });
    }
  });
}

export function markNicknameTaken(page: Page, nickname: string) {
  const nicknames = takenNicknames.get(page) ?? new Set<string>();
  nicknames.add(nickname);
  takenNicknames.set(page, nicknames);
}

export function markEmailCodeRequestUnavailable(page: Page) {
  unavailableEmailCodeRequests.add(page);
}

export async function seedBrowserSession(page: Page, email: string) {
  await page.addInitScript(
    (session) => {
      window.localStorage.setItem(
        "auth",
        JSON.stringify({
          state: {
            token: session.token,
            user: session.user,
            isAuthenticated: true,
          },
          version: 0,
        }),
      );
    },
    {
      token,
      user: { id: userID, email, nickname: null },
    },
  );
}

function operationNameFromQuery(query = "") {
  return query.match(/\b(?:mutation|query)\s+([A-Za-z0-9_]+)/)?.[1];
}

async function respond(route: Route, data: unknown) {
  await route.fulfill({
    status: 200,
    contentType: "application/json",
    body: JSON.stringify({ data }),
  });
}

async function respondError(route: Route, message: string) {
  await route.fulfill({
    status: 200,
    contentType: "application/json",
    body: JSON.stringify({ errors: [{ message }] }),
  });
}
