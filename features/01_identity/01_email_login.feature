Feature: Вход по email

  Как пользователь продукта
  Я хочу получить код доступа на email и войти по нему
  Чтобы пользоваться системой без пароля

  @api
  Scenario: 01_request_email_code
    When пользователь запрашивает код доступа для email "user@example.com"
    Then на email "user@example.com" отправлен код доступа из 5 цифр

  @api
  Scenario: 01A_request_email_code_rejects_empty_email
    When пользователь запрашивает код доступа для email ""
    Then пользователь видит ошибку email "email is required"

  @api
  Scenario: 01B_request_email_code_rejects_invalid_email
    When пользователь запрашивает код доступа для email "user"
    Then пользователь видит ошибку email "email is invalid"

  @api
  Scenario: 01C_request_email_code_normalizes_email
    When пользователь запрашивает код доступа для email " User@Example.Com "
    Then на email "user@example.com" отправлен код доступа из 5 цифр

  @api
  Scenario: 01D_request_email_code_respects_cooldown
    Given пользователь запросил код доступа для email "user@example.com"
    When пользователь повторно запрашивает код доступа для email "user@example.com"
    Then следующий код доступа можно запросить через 30 секунд

  @api
  Scenario: 01E_request_email_code_escalates_cooldown
    Given пользователь исчерпал быстрые повторные запросы кода для email "user@example.com"
    Then задержки повторной отправки кода для email "user@example.com" равны 30, 60, 120 и 86400 секунд

  @api
  Scenario: 02_login_with_email_code
    Given пользователь получил код доступа для email "user@example.com"
    When пользователь вводит полученный код доступа для email "user@example.com"
    Then пользователь получает JWT-токен

  @api
  Scenario: 02A_login_rejects_wrong_email_code
    Given пользователь получил код доступа для email "user@example.com"
    When пользователь вводит код доступа "00000" для email "user@example.com"
    Then пользователь видит ошибку входа "invalid code"

  @api
  Scenario: 02B_login_rejects_expired_email_code
    Given пользователь получил код доступа для email "user@example.com"
    And код доступа для email "user@example.com" истёк
    When пользователь вводит полученный код доступа для email "user@example.com"
    Then пользователь видит ошибку входа "invalid code"

  @api
  Scenario: 02C_login_accepts_any_active_requested_code
    Given пользователь получил два кода доступа для email "user@example.com"
    When пользователь вводит второй полученный код доступа для email "user@example.com"
    Then пользователь получает JWT-токен

  @api
  Scenario: 02D_login_consumes_all_active_email_codes
    Given пользователь получил два кода доступа для email "user@example.com"
    When пользователь вводит первый полученный код доступа для email "user@example.com"
    Then второй код доступа для email "user@example.com" больше не действует

  @api
  Scenario: 02E_login_rejects_reused_email_code
    Given пользователь получил код доступа для email "user@example.com"
    When пользователь вводит полученный код доступа для email "user@example.com"
    Then повторное использование этого кода для email "user@example.com" отклоняется

  @api
  Scenario: 02F_login_resets_email_code_cooldown
    Given пользователь получил код доступа для email "user@example.com"
    When пользователь вводит полученный код доступа для email "user@example.com"
    Then новый код доступа для email "user@example.com" можно запросить сразу

  @api
  Scenario: 03_logout_revokes_token
    Given пользователь вошёл по email "user@example.com"
    When пользователь выходит из системы
    Then выданный JWT-токен больше не действует

  @api
  Scenario: 03A_logout_current_session_keeps_other_sessions
    Given пользователь вошёл по email "user@example.com" в двух сессиях
    When пользователь выходит из текущей сессии
    Then текущая сессия больше не действует
    And другая сессия пользователя продолжает действовать

  @browser
  Scenario: 03B_logout_everywhere_revokes_all_sessions
    Given пользователь вошёл по email "user@example.com" в двух сессиях
    When пользователь выходит из системы на всех устройствах
    Then все сессии пользователя больше не действуют

  @browser
  Scenario: 04_login_with_email_code_in_browser
    Given пользователь находится на странице входа
    When пользователь запрашивает код доступа через браузер для email "user@example.com"
    Then форма входа ожидает код доступа для запрошенного email
    When пользователь вводит полученный код доступа через браузер
    Then пользователь видит страницу профиля

  @browser
  Scenario: 04A_browser_login_shows_invalid_code_error
    Given пользователь находится на странице входа
    When пользователь запрашивает код доступа через браузер для email "user@example.com"
    And пользователь вводит неверный код доступа "00000" через браузер
    Then пользователь видит ошибку входа "invalid code"

  @browser
  Scenario: 04B_browser_login_shows_email_validation_error
    Given пользователь находится на странице входа
    When пользователь запрашивает код доступа через браузер для email "user"
    Then пользователь видит ошибку email "email is invalid"

  @browser
  Scenario: 04C_browser_login_hides_email_after_code_request
    Given пользователь находится на странице входа
    When пользователь запрашивает код доступа через браузер для email "user@example.com"
    Then форма входа показывает только ввод кода для запрошенного email

  @browser
  Scenario: 04D_browser_resend_code_uses_countdown
    Given пользователь запросил код доступа для email "user@example.com"
    Then таймер повторной отправки кода идёт обратным отсчётом

  @browser
  Scenario: 04F_browser_login_validates_email_after_submit
    Given пользователь находится на странице входа
    When пользователь вводит email "user" без запроса кода
    Then пользователь не видит ошибку email "email is invalid"
    When пользователь нажимает Получить код
    Then пользователь видит ошибку email "email is invalid"

  @browser
  Scenario: 04G_browser_login_requires_email_after_submit
    Given пользователь находится на странице входа
    When пользователь нажимает Получить код
    Then пользователь видит ошибку email "Введите email"
