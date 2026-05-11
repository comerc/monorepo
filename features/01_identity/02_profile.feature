Feature: Профиль пользователя

  Как вошедший пользователь
  Я хочу указать nickname в своём профиле
  Чтобы другие части продукта могли показывать моё публичное имя

  @api
  Scenario: 01_set_unique_nickname
    Given пользователь вошёл по email "user@example.com"
    When пользователь указывает nickname "aka"
    Then профиль пользователя содержит nickname "aka"
