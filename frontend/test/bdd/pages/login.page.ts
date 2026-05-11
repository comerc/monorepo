import { expect, type Page } from '@playwright/test'

export class LoginPage {
  constructor(private readonly page: Page) {}

  async open() {
    await this.page.goto('/login')
  }

  async requestCode(email: string) {
    await this.page.getByLabel('Email').fill(email)
    await this.page.getByRole('button', { name: 'Получить код' }).click()
  }

  async expectCodeForm(email: string) {
    await expect(this.page.getByText(`Введите код для ${email}`)).toBeVisible()
  }

  async submitCode(code: string) {
    const digits = this.page.locator('.ant-otp input')

    for (const [index, digit] of [...code].entries()) {
      await digits.nth(index).fill(digit)
    }

    await this.page.getByRole('button', { name: 'Войти' }).click()
  }
}
