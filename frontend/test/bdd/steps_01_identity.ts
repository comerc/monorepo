import { createBdd } from "playwright-bdd";
import { expect, type BrowserContext, type Page } from "@playwright/test";
import { LoginPage } from "./pages/login.page";
import { ProfilePage } from "./pages/profile.page";
import {
  type BrowserSessionState,
  markEmailCodeRequestUnavailable,
  markNicknameTaken,
  mockGraphQL,
  seedBrowserSession,
} from "./support/graphql";

const { Given, When, Then } = createBdd();

const otherSessions = new WeakMap<
  Page,
  { context: BrowserContext; page: Page; sessionState: BrowserSessionState }
>();

Given("пользователь находится на странице входа", async ({ page }) => {
  await mockGraphQL(page);
  await new LoginPage(page).open();
});

Given("отправка кода доступа временно недоступна", async ({ page }) => {
  markEmailCodeRequestUnavailable(page);
});

When(
  "пользователь запрашивает код доступа через браузер для email {string}",
  async ({ page }, email: string) => {
    await new LoginPage(page).requestCode(email);
  },
);

When(
  "пользователь вводит email {string} без запроса кода",
  async ({ page }, email: string) => {
    await new LoginPage(page).typeEmail(email);
  },
);

When("пользователь нажимает Получить код", async ({ page }) => {
  await new LoginPage(page).submitEmail();
});

When(
  "пользователь запрашивает код доступа для email {string}",
  async ({ page }, email: string) => {
    await mockGraphQL(page);
    await new LoginPage(page).open();
    await new LoginPage(page).requestCode(email);
  },
);

Given(
  "пользователь запросил код доступа для email {string}",
  async ({ page }, email: string) => {
    await mockGraphQL(page);
    await new LoginPage(page).open();
    await new LoginPage(page).requestCode(email);
  },
);

Given(
  "пользователь получил код доступа для email {string}",
  async ({ page }, email: string) => {
    await mockGraphQL(page);
    await new LoginPage(page).open();
    await new LoginPage(page).requestCode(email);
  },
);

When(
  "пользователь повторно запрашивает код доступа для email {string}",
  async ({ page }, _email: string) => {
    await new LoginPage(page).requestCodeAgain();
  },
);

Then(
  "форма входа ожидает код доступа для email {string}",
  async ({ page }, email: string) => {
    await new LoginPage(page).expectCodeForm(email);
  },
);

Then(
  "форма входа показывает только ввод кода для email {string}",
  async ({ page }, email: string) => {
    await new LoginPage(page).expectOnlyCodeForm(email);
  },
);

Then(
  "таймер повторной отправки кода идёт обратным отсчётом",
  async ({ page }) => {
    await new LoginPage(page).expectResendCountdown();
  },
);

When(
  "пользователь вводит код доступа {string} через браузер",
  async ({ page }, code: string) => {
    await new LoginPage(page).submitCode(code);
  },
);

When(
  "пользователь вводит код доступа {string} для email {string}",
  async ({ page }, code: string, _email: string) => {
    await new LoginPage(page).submitCode(code);
  },
);

When(
  "пользователь вводит неверный код доступа {string} через браузер",
  async ({ page }, code: string) => {
    await new LoginPage(page).submitCode(code);
  },
);

Then("пользователь видит страницу профиля", async ({ page }) => {
  await new ProfilePage(page).expectOpen();
});

Then(
  "пользователь видит ошибку email {string}",
  async ({ page }, message: string) => {
    await new LoginPage(page).expectEmailError(message);
  },
);

Then(
  "пользователь не видит ошибку email {string}",
  async ({ page }, message: string) => {
    await new LoginPage(page).expectNoEmailError(message);
  },
);

Then(
  "пользователь видит ошибку входа {string}",
  async ({ page }, message: string) => {
    await new LoginPage(page).expectLoginError(message);
  },
);

Then(
  "следующий код доступа можно запросить через {int} секунд",
  async ({ page }, seconds: number) => {
    await new LoginPage(page).expectRetryAfter(seconds);
  },
);

Given(
  "пользователь вошёл по email {string}",
  async ({ page }, email: string) => {
    await mockGraphQL(page, email);
    await seedBrowserSession(page, email);
    await page.goto("/profile");
  },
);

Given(
  "пользователь вошёл по email {string} с nickname {string}",
  async ({ page }, email: string, nickname: string) => {
    await mockGraphQL(page, email, nickname);
    await seedBrowserSession(page, email);
    await page.goto("/profile");
  },
);

Given(
  "пользователь вошёл по email {string} в двух сессиях",
  async ({ browser, page }, email: string) => {
    const sessionState = { revokedEverywhere: false };
    await mockGraphQL(page, email, null, sessionState);
    await seedBrowserSession(page, email);

    const otherContext = await browser.newContext();
    const otherPage = await otherContext.newPage();
    await mockGraphQL(otherPage, email, null, sessionState);
    await seedBrowserSession(otherPage, email);
    otherSessions.set(page, {
      context: otherContext,
      page: otherPage,
      sessionState,
    });

    await page.goto("/profile");
    await otherPage.goto(page.url());
  },
);

When(
  "пользователь указывает nickname {string}",
  async ({ page }, nickname: string) => {
    await new ProfilePage(page).setNickname(nickname);
  },
);

When(
  "пользователь вводит nickname {string} без сохранения",
  async ({ page }, nickname: string) => {
    await new ProfilePage(page).typeNickname(nickname);
  },
);

Then(
  "профиль пользователя содержит nickname {string}",
  async ({ page }, nickname: string) => {
    await new ProfilePage(page).expectNickname(nickname);
  },
);

Then(
  "шапка профиля показывает nickname {string}",
  async ({ page }, nickname: string) => {
    await new ProfilePage(page).expectHeaderNickname(nickname);
  },
);

Given(
  "nickname {string} уже занят другим пользователем",
  async ({ page }, nickname: string) => {
    markNicknameTaken(page, nickname);
  },
);

Then(
  "пользователь видит ошибку nickname {string}",
  async ({ page }, message: string) => {
    await new ProfilePage(page).expectNicknameError(message);
  },
);

When("пользователь открывает профиль без входа", async ({ page }) => {
  await mockGraphQL(page);
  await new ProfilePage(page).openWithoutSession();
});

Then(
  "пользователь видит ошибку профиля {string}",
  async ({ page }, message: string) => {
    await new ProfilePage(page).expectProfileError(message);
  },
);

Then(
  "пользователь видит форму входа без технической ошибки",
  async ({ page }) => {
    await new LoginPage(page).expectLoginFormWithoutTechnicalError();
  },
);

Then("пользователь видит проверку доступности nickname", async ({ page }) => {
  await new ProfilePage(page).expectAvailabilityChecking();
});

Then("проверка доступности nickname не запускается", async ({ page }) => {
  await new ProfilePage(page).expectAvailabilityNotStarted();
});

Then(
  "пользователь видит, что nickname {string} занят",
  async ({ page }, nickname: string) => {
    await new ProfilePage(page).expectNicknameTaken(nickname);
  },
);

Then(
  "пользователь видит, что nickname {string} свободен",
  async ({ page }, nickname: string) => {
    await new ProfilePage(page).expectNicknameFree(nickname);
  },
);

When(
  "пользователь выходит из системы на всех устройствах",
  async ({ page }) => {
    await page.getByRole("button", { name: "Выйти" }).click();
    await page.getByLabel("на всех устройствах").check();
    await page.getByRole("button", { name: "Выйти" }).last().click();
  },
);

Then("все сессии пользователя больше не действуют", async ({ page }) => {
  await expect(page).toHaveURL(/\/login(?:\?|$)/);
  const otherSession = otherSessions.get(page);
  expect(otherSession).toBeDefined();
  const myProfileResponse = otherSession!.page
    .waitForResponse(
      (response) =>
        response.url().includes("/graphql") &&
        (response.request().postData() ?? "").includes("MyProfile"),
      { timeout: 5000 },
    )
    .then((response) => response.text())
    .catch(() => null);
  await otherSession!.page.goto(new URL("/profile", page.url()).toString(), {
    waitUntil: "networkidle",
  });
  await otherSession!.page.reload({ waitUntil: "networkidle" });
  const profileResponseBody = await myProfileResponse;
  if (profileResponseBody) {
    expect(profileResponseBody).toContain("unauthenticated");
    await otherSession!.context.close();
    otherSessions.delete(page);
    return;
  }
  if (otherSession!.page.url().includes("/login")) {
    await new LoginPage(
      otherSession!.page,
    ).expectLoginFormWithoutTechnicalError();
    await otherSession!.context.close();
    otherSessions.delete(page);
    return;
  }
  const otherSessionOnLogin = await otherSession!.page
    .waitForURL(/\/login(?:\?|$)/, { timeout: 2000 })
    .then(() => true)
    .catch(() => false);
  if (otherSessionOnLogin) {
    await new LoginPage(
      otherSession!.page,
    ).expectLoginFormWithoutTechnicalError();
    await otherSession!.context.close();
    otherSessions.delete(page);
    return;
  }
  const nicknameInput = otherSession!.page.getByLabel("Nickname");
  if (await nicknameInput.isVisible({ timeout: 1000 }).catch(() => false)) {
    const setNicknameResponse = otherSession!.page
      .waitForResponse(
        (response) =>
          response.url().includes("/graphql") &&
          (response.request().postData() ?? "").includes("SetNickname"),
      )
      .then((response) => response.text());
    await new ProfilePage(otherSession!.page).setNickname("revoked");
    expect(await setNicknameResponse).toContain("unauthenticated");
    await otherSession!.context.close();
    otherSessions.delete(page);
    return;
  }
  if (
    await otherSession!.page
      .getByRole("heading", { name: "Вход" })
      .isVisible()
      .catch(() => false)
  ) {
    await new LoginPage(
      otherSession!.page,
    ).expectLoginFormWithoutTechnicalError();
    await otherSession!.context.close();
    otherSessions.delete(page);
    return;
  }
  await expect(otherSession!.page.getByText("unauthenticated")).toBeVisible();
  await otherSession!.context.close();
  otherSessions.delete(page);
});
