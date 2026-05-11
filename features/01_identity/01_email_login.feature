Feature: Вход по email

  Как пользователь продукта
  Я хочу получить код доступа на email и войти по нему
  Чтобы пользоваться системой без пароля

  @api
  Scenario: 01_request_email_code
    When пользователь запрашивает код доступа для email "user@example.com"
    Then на email "user@example.com" отправлен код доступа из 4 цифр

  @api
  Scenario: 02_login_with_email_code
    Given пользователь получил код доступа для email "user@example.com"
    When пользователь вводит полученный код доступа для email "user@example.com"
    Then пользователь получает JWT-токен

  @api
  Scenario: 03_logout_revokes_token
    Given пользователь вошёл по email "user@example.com"
    When пользователь выходит из системы
    Then выданный JWT-токен больше не действует
