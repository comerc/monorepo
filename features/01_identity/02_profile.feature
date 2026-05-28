Feature: Профиль пользователя

  Как вошедший пользователь
  Я хочу указать nickname в своём профиле
  Чтобы другие части продукта могли показывать моё публичное имя

  @browser
  Scenario: 01_set_unique_nickname
    Given пользователь вошёл по email "user@example.com"
    When пользователь указывает nickname "aka"
    Then профиль пользователя содержит nickname "aka"

  @browser
  Scenario: 01A_set_nickname_rejects_taken_nickname
    Given пользователь вошёл по email "user@example.com"
    And nickname "aka" уже занят другим пользователем
    When пользователь указывает nickname "aka"
    Then пользователь видит ошибку nickname "nickname is already taken"

  @api
  Scenario: 01B_set_nickname_trims_outer_spaces
    Given пользователь вошёл по email "user@example.com"
    When пользователь указывает nickname " aka "
    Then профиль пользователя содержит nickname "aka"

  @api
  Scenario: 01C_set_nickname_rejects_too_short_nickname
    Given пользователь вошёл по email "user@example.com"
    When пользователь указывает nickname "a"
    Then пользователь видит ошибку nickname "nickname is too short"

  @api
  Scenario: 01D_set_nickname_allows_case_sensitive_variants
    Given пользователь вошёл по email "user@example.com"
    And nickname "Aka" уже занят другим пользователем
    When пользователь указывает nickname "aka"
    Then профиль пользователя содержит nickname "aka"

  @api
  Scenario: 01E_change_existing_nickname
    Given пользователь вошёл по email "user@example.com" с nickname "aka"
    When пользователь указывает nickname "level85"
    Then профиль пользователя содержит nickname "level85"

  @api
  Scenario: 01F_profile_requires_authenticated_user
    When пользователь открывает профиль без входа
    Then пользователь видит ошибку профиля "unauthenticated"

  @browser
  Scenario: 01G_browser_profile_without_login_opens_login_form
    When пользователь открывает профиль без входа
    Then пользователь видит форму входа без технической ошибки

  @browser
  Scenario: 01H_browser_set_nickname_shows_short_error
    Given пользователь вошёл по email "user@example.com"
    When пользователь указывает nickname "a"
    Then пользователь видит ошибку nickname "nickname слишком короткий"

  @browser
  Scenario: 01I_browser_stale_session_redirects_to_login
    Given пользователь вошёл по email "user@example.com"
    And данные пользователя удалены на сервере
    When пользователь открывает профиль
    Then пользователь видит форму входа без технической ошибки

  @browser
  Scenario: 02_show_saved_nickname_in_header
    Given пользователь вошёл по email "user@example.com" с nickname "aka"
    Then шапка профиля показывает nickname "aka"

  @browser
  Scenario: 03_check_nickname_availability_after_debounce
    Given пользователь вошёл по email "user@example.com"
    When пользователь вводит nickname "aka" без сохранения
    Then пользователь видит проверку доступности nickname

  @browser
  Scenario: 03A_check_nickname_availability_not_started_for_two_chars
    Given пользователь вошёл по email "user@example.com"
    When пользователь вводит nickname "ak" без сохранения
    Then проверка доступности nickname не запускается

  @browser
  Scenario: 03B_check_nickname_availability_shows_taken
    Given пользователь вошёл по email "user@example.com"
    And nickname "aka" уже занят другим пользователем
    When пользователь вводит nickname "aka" без сохранения
    Then пользователь видит, что nickname "aka" занят

  @browser
  Scenario: 03C_check_nickname_availability_shows_free
    Given пользователь вошёл по email "user@example.com"
    When пользователь вводит nickname "level85" без сохранения
    Then пользователь видит, что nickname "level85" свободен
