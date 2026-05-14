import { expect, type Page } from "@playwright/test";

const availabilityChecks = new WeakMap<Page, Promise<void>>();
const unexpectedAvailabilityChecks = new WeakMap<Page, Promise<void>>();

export class ProfilePage {
  constructor(private readonly page: Page) {}

  async expectOpen() {
    await expect(this.page).toHaveURL(/\/profile$/);
    await expect(
      this.page.getByRole("heading", { name: "Профиль" }),
    ).toBeVisible();
  }

  async setNickname(nickname: string) {
    await this.page.getByLabel("Nickname").fill(nickname);
    if (nickname.trim().length >= 3) {
      await expect(
        this.page.getByText(/nickname (свободен|занят)/),
      ).toBeVisible();
      if (await this.page.getByText("nickname занят").isVisible()) {
        return;
      }
    }
    await this.page.getByRole("button", { name: "Сохранить" }).click();
  }

  async typeNickname(nickname: string) {
    const checkedNickname = nickname.trim();
    if (checkedNickname.length >= 3) {
      availabilityChecks.set(
        this.page,
        this.page
          .waitForResponse((response) =>
            isNicknameAvailabilityResponse(response, checkedNickname),
          )
          .then(async (response) => {
            const body = (await response.json()) as {
              data?: {
                nicknameAvailability?: {
                  nickname?: string;
                  available?: boolean;
                };
              };
            };
            expect(body.data?.nicknameAvailability?.nickname).toBe(
              checkedNickname,
            );
            expect(typeof body.data?.nicknameAvailability?.available).toBe(
              "boolean",
            );
          }),
      );
    } else {
      availabilityChecks.delete(this.page);
      unexpectedAvailabilityChecks.set(
        this.page,
        this.page
          .waitForResponse(
            (response) =>
              isNicknameAvailabilityResponse(response, checkedNickname),
            { timeout: 1200 },
          )
          .then(() => {
            throw new Error("nickname availability check was started");
          })
          .catch((error: Error) => {
            if (!error.message.includes("Timeout")) {
              throw error;
            }
          }),
      );
    }
    await this.page.getByLabel("Nickname").fill(nickname);
  }

  async expectNickname(nickname: string) {
    await expect(this.page.getByLabel("Nickname")).toHaveValue(nickname);
    await expect(this.page.getByText(nickname).first()).toBeVisible();
  }

  async expectHeaderNickname(nickname: string) {
    await expect(this.page.locator("header")).toContainText(nickname);
  }

  async expectNicknameError(message: string) {
    if (message === "nickname is already taken") {
      await expect(this.page.getByText("nickname занят")).toBeVisible();
      await expect(
        this.page.getByRole("button", { name: "Сохранить" }),
      ).toBeDisabled();
      return;
    }
    await expect(this.page.getByText(message)).toBeVisible();
  }

  async expectProfileError(message: string) {
    await expect(this.page.getByText(message)).toBeVisible();
  }

  async expectAvailabilityChecking() {
    await expect(this.page.getByText("Проверяем nickname")).toBeVisible();
    await availabilityChecks.get(this.page);
  }

  async expectAvailabilityNotStarted() {
    await expect(this.page.getByText("Проверяем nickname")).toBeHidden();
    await unexpectedAvailabilityChecks.get(this.page);
    await expect(this.page.getByText("Проверяем nickname")).toBeHidden();
  }

  async expectNicknameTaken(_nickname: string) {
    await expect(this.page.getByText("nickname занят")).toBeVisible();
  }

  async expectNicknameFree(_nickname: string) {
    await expect(this.page.getByText("nickname свободен")).toBeVisible();
  }

  async openWithoutSession() {
    await this.page.goto("/profile");
  }
}

function isNicknameAvailabilityResponse(
  response: {
    url(): string;
    request(): { postDataJSON(): unknown };
  },
  nickname: string,
) {
  if (!response.url().includes("/graphql")) {
    return false;
  }
  const request = response.request().postDataJSON() as {
    operationName?: string;
    variables?: { nickname?: string };
  };
  return (
    request.operationName === "NicknameAvailability" &&
    String(request.variables?.nickname ?? "").trim() === nickname
  );
}
