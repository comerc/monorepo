import { expect, type Page } from '@playwright/test'

export class ProfilePage {
  constructor(private readonly page: Page) {}

  async expectOpen() {
    await expect(this.page).toHaveURL(/\/profile$/)
    await expect(this.page.getByRole('heading', { name: 'Профиль' })).toBeVisible()
  }

  async setNickname(nickname: string) {
    await this.page.getByLabel('Nickname').fill(nickname)
    await this.page.getByRole('button', { name: 'Сохранить' }).click()
  }

  async expectNickname(nickname: string) {
    await expect(this.page.getByLabel('Nickname')).toHaveValue(nickname)
    await expect(this.page.getByText(nickname).first()).toBeVisible()
  }

  async expectHeaderNickname(nickname: string) {
    await expect(this.page.locator('header')).toContainText(nickname)
  }
}
