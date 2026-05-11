import { createBdd } from 'playwright-bdd'
import { LoginPage } from '../pages/login.page'
import { ProfilePage } from '../pages/profile.page'
import { mockGraphQL, seedBrowserSession } from '../support/graphql'

const { Given, When, Then } = createBdd()

Given('пользователь находится на странице входа', async ({ page }) => {
  await mockGraphQL(page)
  await new LoginPage(page).open()
})

When(
  'пользователь запрашивает код доступа через браузер для email {string}',
  async ({ page }, email: string) => {
    await new LoginPage(page).requestCode(email)
  },
)

Then(
  'форма входа ожидает код доступа для email {string}',
  async ({ page }, email: string) => {
    await new LoginPage(page).expectCodeForm(email)
  },
)

When('пользователь вводит код доступа {string} через браузер', async ({ page }, code: string) => {
  await new LoginPage(page).submitCode(code)
})

Then('пользователь видит страницу профиля', async ({ page }) => {
  await new ProfilePage(page).expectOpen()
})

Given('пользователь вошёл по email {string}', async ({ page }, email: string) => {
  await mockGraphQL(page, email)
  await seedBrowserSession(page, email)
  await page.goto('/profile')
})

When('пользователь указывает nickname {string}', async ({ page }, nickname: string) => {
  await new ProfilePage(page).setNickname(nickname)
})

Then('профиль пользователя содержит nickname {string}', async ({ page }, nickname: string) => {
  await new ProfilePage(page).expectNickname(nickname)
})
