import type { Page, Route } from '@playwright/test'

const userID = 'user-1'
const token = 'browser-bdd-token'

interface GraphQLRequest {
  operationName?: string
  query?: string
  variables?: Record<string, unknown>
}

interface BrowserUser {
  id: string
  email: string
  nickname: string | null
}

export async function mockGraphQL(
  page: Page,
  email = 'user@example.com',
  nickname: string | null = null,
) {
  const user: BrowserUser = { id: userID, email, nickname }

  await page.route('**/graphql', async (route) => {
    const request = route.request().postDataJSON() as GraphQLRequest
    const operationName = request.operationName ?? operationNameFromQuery(request.query)

    switch (operationName) {
      case 'RequestEmailCode':
        await respond(route, { requestEmailCode: { accepted: true } })
        return
      case 'LoginWithEmailCode':
        user.email = String(request.variables?.email ?? email)
        await respond(route, {
          loginWithEmailCode: {
            token,
            userID,
            email: user.email,
          },
        })
        return
      case 'MyProfile':
        await respond(route, {
          myProfile: {
            userID,
            nickname: user.nickname,
          },
        })
        return
      case 'SetNickname':
        user.nickname = String(request.variables?.nickname ?? '')
        await respond(route, {
          setNickname: {
            userID,
            nickname: user.nickname,
          },
        })
        return
      case 'Logout':
        await respond(route, { logout: { revoked: true } })
        return
      default:
        await route.fulfill({
          status: 400,
          contentType: 'application/json',
          body: JSON.stringify({ errors: [{ message: `unsupported operation: ${operationName}` }] }),
        })
    }
  })
}

export async function seedBrowserSession(page: Page, email: string) {
  await page.addInitScript(
    (session) => {
      window.localStorage.setItem(
        'auth',
        JSON.stringify({
          state: {
            token: session.token,
            user: session.user,
            isAuthenticated: true,
          },
          version: 0,
        }),
      )
    },
    {
      token,
      user: { id: userID, email, nickname: null },
    },
  )
}

function operationNameFromQuery(query = '') {
  return query.match(/\b(?:mutation|query)\s+([A-Za-z0-9_]+)/)?.[1]
}

async function respond(route: Route, data: unknown) {
  await route.fulfill({
    status: 200,
    contentType: 'application/json',
    body: JSON.stringify({ data }),
  })
}
