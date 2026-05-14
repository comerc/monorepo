import { expect, type Page } from "@playwright/test";

export class LoginPage {
  constructor(private readonly page: Page) {}

  async open() {
    await this.page.goto("/login");
  }

  async requestCode(email: string) {
    await this.page.getByLabel("Email").fill(email);
    await this.page.getByRole("button", { name: "Получить код" }).click();
  }

  async typeEmail(email: string) {
    await this.page.getByLabel("Email").fill(email);
  }

  async submitEmail() {
    await this.page.getByRole("button", { name: "Получить код" }).click();
  }

  async requestCodeAgain() {
    await this.page
      .getByRole("button", { name: "Отправить код ещё раз" })
      .click();
  }

  async expectCodeForm(email: string) {
    await expect(this.page.getByText(`Введите код для ${email}`)).toBeVisible();
  }

  async expectOnlyCodeForm(email: string) {
    await this.expectCodeForm(email);
    await expect(this.page.getByLabel("Email")).toBeHidden();
    await expect(this.page.locator(".ant-otp input").first()).toBeVisible();
  }

  async submitCode(code: string) {
    const digits = this.page.locator(".ant-otp input");
    await digits.first().fill("");
    await digits.first().pressSequentially(code);
  }

  async expectEmailError(message: string) {
    await expect(this.page.getByText(message)).toBeVisible();
  }

  async expectNoEmailError(message: string) {
    await expect(this.page.getByText(message)).toBeHidden();
  }

  async expectLoginError(message: string) {
    await expect(this.page.getByText(message)).toBeVisible();
  }

  async expectRetryAfter(seconds: number) {
    await expect(
      this.page
        .getByText(`Следующий код можно запросить через ${seconds} сек.`)
        .first(),
    ).toBeVisible();
  }

  async expectResendCountdown() {
    const resend = this.page.getByRole("button", {
      name: "Отправить код ещё раз",
    });
    await expect(resend).toBeHidden();
    await expect(
      this.page.getByText(/Следующий код можно запросить через \d+ сек\./),
    ).toBeVisible();
    await expect(
      this.page.getByText(/Следующий код можно запросить через \d+ сек\./),
    ).toBeHidden({ timeout: 3500 });
    await expect(resend).toBeEnabled();
  }

  async expectLoginFormWithoutTechnicalError() {
    await expect(
      this.page.getByRole("heading", { name: "Вход" }),
    ).toBeVisible();
    await expect(this.page.getByText("unauthenticated")).toBeHidden();
  }
}
