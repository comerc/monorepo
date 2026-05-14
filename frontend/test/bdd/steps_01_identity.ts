import { createBdd } from "playwright-bdd";
import { expect, type BrowserContext, type Page } from "@playwright/test";
import { LoginPage } from "./pages/login.page";
import { ProfilePage } from "./pages/profile.page";
import {
  createUserWithNickname,
  expectTokenRevoked,
  loginByActualEmail,
  loginByEmail,
  markCodeRequested,
  receivedCode,
  rememberReceivedCode,
  requestedEmail,
  resetLiveBackendState,
  resolveScenarioEmail,
  seedBrowserSession,
} from "./support/live-api";

const { Before, Given, When, Then } = createBdd();

const otherSessions = new WeakMap<
  Page,
  { context: BrowserContext; page: Page; token: string }
>();

Before(async () => {
  await resetLiveBackendState();
});

Given("пользователь находится на странице входа", async ({ page }) => {
  await new LoginPage(page).open();
});

When(
  "пользователь запрашивает код доступа через браузер для email {string}",
  async ({ page }, email: string) => {
    const actualEmail = resolveScenarioEmail(page, email);
    markCodeRequested(page, actualEmail);
    await new LoginPage(page).requestCode(actualEmail);
    if (actualEmail.includes("@")) {
      await rememberReceivedCode(page);
    }
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
    await new LoginPage(page).open();
    const actualEmail = resolveScenarioEmail(page, email);
    markCodeRequested(page, actualEmail);
    await new LoginPage(page).requestCode(actualEmail);
    if (actualEmail.includes("@")) {
      await rememberReceivedCode(page);
    }
  },
);

Given(
  "пользователь запросил код доступа для email {string}",
  async ({ page }, email: string) => {
    await new LoginPage(page).open();
    const actualEmail = resolveScenarioEmail(page, email);
    markCodeRequested(page, actualEmail);
    await new LoginPage(page).requestCode(actualEmail);
    if (actualEmail.includes("@")) {
      await rememberReceivedCode(page);
    }
  },
);

Given(
  "пользователь получил код доступа для email {string}",
  async ({ page }, email: string) => {
    await new LoginPage(page).open();
    const actualEmail = resolveScenarioEmail(page, email);
    markCodeRequested(page, actualEmail);
    await new LoginPage(page).requestCode(actualEmail);
    if (actualEmail.includes("@")) {
      await rememberReceivedCode(page);
    }
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
  "форма входа ожидает код доступа для запрошенного email",
  async ({ page }) => {
    await new LoginPage(page).expectCodeForm(requestedEmail(page));
  },
);

Then(
  "форма входа показывает только ввод кода для email {string}",
  async ({ page }, email: string) => {
    await new LoginPage(page).expectOnlyCodeForm(email);
  },
);

Then(
  "форма входа показывает только ввод кода для запрошенного email",
  async ({ page }) => {
    await new LoginPage(page).expectOnlyCodeForm(requestedEmail(page));
  },
);

Then(
  "таймер повторной отправки кода идёт обратным отсчётом",
  async ({ page }) => {
    await new LoginPage(page).expectResendCountdown();
  },
);

When("пользователь вводит полученный код доступа через браузер", async ({
  page,
}) => {
  await new LoginPage(page).submitCode(receivedCode(page));
});

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
    await loginByEmail(page, email);
    await page.goto("/profile");
  },
);

Given(
  "пользователь вошёл по email {string} с nickname {string}",
  async ({ page }, email: string, nickname: string) => {
    await loginByEmail(page, email, nickname);
    await page.goto("/profile");
  },
);

Given(
  "пользователь вошёл по email {string} в двух сессиях",
  async ({ browser, page }, email: string) => {
    const actualEmail = resolveScenarioEmail(page, email);
    const firstSession = await loginByActualEmail(page, actualEmail);

    const otherContext = await browser.newContext();
    const otherPage = await otherContext.newPage();
    const secondSession = await loginByActualEmail(otherPage, actualEmail);
    await seedBrowserSession(page, firstSession);
    await seedBrowserSession(otherPage, secondSession);
    otherSessions.set(page, {
      context: otherContext,
      page: otherPage,
      token: secondSession.token,
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
    await createUserWithNickname(page, nickname);
  },
);

Then(
  "пользователь видит ошибку nickname {string}",
  async ({ page }, message: string) => {
    await new ProfilePage(page).expectNicknameError(message);
  },
);

When("пользователь открывает профиль без входа", async ({ page }) => {
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
    const logoutResponse = page.waitForResponse(
      (response) =>
        response.url().includes("/graphql") &&
        (response.request().postData() ?? "").includes("Logout"),
    );
    await page.getByRole("button", { name: "Выйти" }).click();
    await page.getByLabel("на всех устройствах").check();
    await page.getByRole("button", { name: "Выйти" }).last().click();
    const requestBody = logoutResponse.then((response) =>
      response.request().postDataJSON(),
    );
    await expect(await requestBody).toMatchObject({
      variables: { allDevices: true },
    });
  },
);

Then("все сессии пользователя больше не действуют", async ({ page }) => {
  await expect(page).toHaveURL(/\/login(?:\?|$)/);
  const otherSession = otherSessions.get(page);
  expect(otherSession).toBeDefined();
  await expectTokenRevoked(otherSession!.token);
  await otherSession!.context.close();
  otherSessions.delete(page);
});
