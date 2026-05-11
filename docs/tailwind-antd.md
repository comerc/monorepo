# Tailwind CSS и Ant Design

Дата исследования: 2026-05-11. Цель: Tailwind CSS v4 + Ant Design v6.

## Решение

Для текущего `frontend` используем Tailwind CSS v4 как единственный слой продуктовой кастомизации Ant Design v6:

- Tailwind подключается через Vite-плагин `@tailwindcss/vite`.
- Дизайн-токены приложения живут в `frontend/src/styles.css` через `@theme`.
- AntD остаётся библиотекой поведения и готовых React-компонентов.
- Повторяемая кастомизация дизайн-системы живёт в `@layer components`; локальные utility-классы в JSX допустимы для разовой композиции, позиционирования и layout-контекста.
- `ConfigProvider.theme.token` и `ConfigProvider.theme.components` не используем для брендовой кастомизации, иначе появляется второй источник дизайна.

Исключение: `ConfigProvider` можно оставить для `locale`, встроенного алгоритма light/dark и `zeroRuntime: true`. Это не продуктовая кастомизация, а режим работы AntD.

## Подключение

В `frontend`:

```bash
npm install antd@^6 @ant-design/icons@^6
npm install -D tailwindcss @tailwindcss/vite
```

`frontend/vite.config.ts`:

```ts
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    proxy: {
      '/graphql': {
        target: 'http://localhost:3002',
        changeOrigin: true,
      },
    },
  },
})
```

В AntD v6 `zeroRuntime` включён по умолчанию: стили компонентов не генерируются на лету, поэтому нужно импортировать готовый `antd/dist/antd.css`. Чтобы Tailwind utilities были выше AntD в каскаде без постоянного `!`, лучше держать все CSS-импорты в `frontend/src/styles.css`.

`frontend/src/main.tsx`:

```ts
import './styles.css'
```

`frontend/src/styles.css`:

```css
@layer theme, base, antd, components, utilities;

@import "tailwindcss/theme.css" layer(theme);
@import "tailwindcss/preflight.css" layer(base);
@import "antd/dist/reset.css" layer(base);
@import "antd/dist/antd.css" layer(antd);
@import "tailwindcss/utilities.css" layer(utilities);

@custom-variant dark (&:where(.dark, .dark *));

@theme {
  --font-sans: Inter, ui-sans-serif, system-ui, sans-serif;

  --color-brand-50: oklch(0.97 0.02 252);
  --color-brand-100: oklch(0.93 0.04 252);
  --color-brand-500: oklch(0.62 0.18 252);
  --color-brand-600: oklch(0.54 0.2 252);
  --color-brand-700: oklch(0.47 0.18 252);

  --color-surface: oklch(0.99 0.003 247);
  --color-surface-dark: oklch(0.18 0.02 247);
}

html,
body,
#root {
  min-height: 100%;
  margin: 0;
}
```

Если Tailwind preflight даст регрессы с AntD, сначала проверяем конкретный конфликт. Как крайний вариант можно не импортировать preflight и подключить только theme/utilities:

```css
@layer theme, antd, utilities;
@import "tailwindcss/theme.css" layer(theme);
@import "antd/dist/reset.css" layer(antd);
@import "antd/dist/antd.css" layer(antd);
@import "tailwindcss/utilities.css" layer(utilities);
```

## ThemeProvider

Текущий `ThemeProvider` уже управляет `light | auto | dark`. Для Tailwind лучше вешать класс на `document.documentElement`, чтобы `dark:*` работал и для порталов AntD:

```ts
useEffect(() => {
  document.documentElement.classList.toggle('dark', effective === 'dark')
}, [effective])
```

`ConfigProvider` должен оставаться минимальным:

```tsx
<ConfigProvider
  locale={ruRU}
  theme={{
    algorithm: effective === 'dark' ? theme.darkAlgorithm : theme.defaultAlgorithm,
    zeroRuntime: true,
  }}
>
  {children}
</ConfigProvider>
```

AntD v6 включает CSS variables по умолчанию. `zeroRuntime: true` тоже значение по умолчанию, но его можно оставить явно, чтобы в коде было видно, почему импортируется `antd/dist/antd.css`.

Не добавляем сюда `token: { colorPrimary: ... }` или `components: { Button: ... }`, если цель — кастомизировать AntD только Tailwind.

## Как стилизовать AntD

### Граница ответственности

Кастомизация делится на два слоя:

- Дизайн-система: цвет, фон, радиус, тень, border, padding компонента, состояние hover/focus/disabled, размеры типовых кнопок и карточек. Это повторяемые правила, они живут в `@layer components`.
- Локальное применение: позиционирование конкретного экземпляра, внешний отступ в конкретном layout, ширина в конкретной сетке, `flex`/`grid`-композиция вокруг компонента. Это можно писать utility-классами прямо в JSX.

Пример: `auth-card` описывает вид карточки как часть дизайн-системы, а `mx-auto mt-8` остаётся локальным размещением.

```tsx
<Card className="auth-card mx-auto mt-8" />
```

```css
@layer components {
  .auth-card {
    @apply w-full max-w-[400px] rounded-[3px] border border-slate-200 shadow-sm;
  }
}
```

Если набор utility-классов начинает описывать внешний вид компонента, а не его место в конкретном layout, его нужно вынести в именованный класс внутри `@layer components`.

### Внутренний DOM AntD

Основной паттерн для кастомизации внутренних элементов AntD: дать компоненту свой root-класс и переопределять `.ant-*` только под этим классом через `@apply`.

```tsx
<Modal rootClassName="card-detail-window" ... />
```

```css
@layer components {
  .card-detail-window > .ant-modal-content {
    @apply overflow-auto rounded-[3px] bg-[var(--ds-surface-overlay,#f4f5f7)] p-0 !important;
  }
}
```

Правила:

- Селектор всегда начинается с проектного root-класса, например `.card-detail-window`.
- `.ant-*` без проектного scope запрещён.
- Для overlay-компонентов используем `rootClassName`, чтобы класс оказался на внешнем узле портала.
- Если AntD перебивает свойство, добавляем `!important` к `@apply`; без необходимости `!important` не ставим.
- CSS override кладём в `@layer components`, потому что этот слой в `styles.css` расположен после `antd` и до `utilities`.

### Простая обёртка компонента

```tsx
<Card className="profile-card">
  ...
</Card>
```

```css
@layer components {
  .profile-card {
    @apply max-w-[480px] rounded-[3px] border border-slate-200 shadow-sm dark:border-slate-800;
  }
}
```

### Кнопка

Для полной Tailwind-кастомизации лучше не полагаться на `type="primary"` как на источник цвета:

```tsx
<Button
  htmlType="submit"
  loading={loading}
  className="primary-action"
>
  Получить код
</Button>
```

```css
@layer components {
  .primary-action {
    @apply h-10 rounded-[3px] border-0 bg-brand-600 px-4 font-medium text-white hover:bg-brand-700 disabled:bg-slate-300 dark:disabled:bg-slate-700;
  }
}
```

`!important` у `@apply` нужен только там, где AntD перебивает свойство. Если обычный класс побеждает без `!important`, оставляем обычный класс.

### Слоты компонента

Если компонент поддерживает `classNames`, стилизуем внутренние части через него, а не через `.ant-*`:

```tsx
<Card
  className="auth-card"
  classNames={{
    body: 'auth-card-body',
  }}
>
  ...
</Card>
```

```css
@layer components {
  .auth-card {
    @apply w-full max-w-[400px] overflow-hidden rounded-[3px] border border-slate-200 dark:border-slate-800;
  }

  .auth-card-body {
    @apply p-6;
  }
}
```

Для dropdown/modal/popover проверяем документацию конкретного компонента. В v6 старые `dropdownClassName` и `popupClassName` заменены на `classNames.popup.root`, а `dropdownRender` заменён на `popupRender`.

## Миграция текущих CSS-классов

Существующие классы из `frontend/src/styles.css` нужно переписать на Tailwind `@apply` внутри `@layer components`, а в JSX оставить только имена проектных классов:

| Сейчас | После миграции |
|---|---|
| `.auth-page` | `@apply min-h-screen flex items-center justify-center p-6 bg-slate-50 dark:bg-black;` |
| `.auth-theme-switcher` | `@apply fixed right-4 top-4;` |
| `.auth-card` | `@apply w-full max-w-[400px];` |
| `.auth-title` | `@apply mb-6 text-center;` |
| `.app-shell` | `@apply min-h-screen;` |
| `.brand` | `@apply flex h-16 items-center justify-center gap-2 text-lg font-semibold text-white;` |
| `.app-header` | `@apply flex items-center justify-end px-6;` |
| `.app-content` | `@apply m-6 min-h-[280px];` |
| `.page-section` | `@apply max-w-[720px];` |
| `.profile-card` | `@apply max-w-[480px];` |
| `.loading-state` | `@apply flex justify-center p-12;` |
| `.form-alert` | `@apply mb-4;` |

После миграции в `styles.css` должны остаться Tailwind import, `@theme`, `@custom-variant`, базовые размеры `html/body/#root` и проектные component classes внутри `@layer components`.

## Запреты

- Не писать глобальные селекторы `.ant-btn`, `.ant-card`, `.ant-layout` и другие `.ant-*` без проектного root-класса.
- Не использовать Less-переменные AntD.
- Не добавлять брендовые значения в `ConfigProvider.theme.token` и `ConfigProvider.theme.components`.
- Не писать дизайн-системную кастомизацию длинными utility-строками в JSX; выносить повторяемый вид компонента в именованный класс внутри `@layer components`.
- Не выносить в `@layer components` разовую layout-композицию, которая относится только к месту использования компонента.
- Не использовать inline `style={{ ... }}` для статического внешнего вида; `@layer components` должен быть первым вариантом.
- Не возвращать `import 'antd/dist/reset.css'` и `import 'antd/dist/antd.css'` в `main.tsx`: без cascade layer AntD CSS будет сложнее переопределять Tailwind-утилитами.
- Не дублировать один и тот же паттерн строкой классов в десятках мест: для повторяемого AntD-компонента лучше сделать локальную React-обёртку, например `PrimaryButton`.

## Проверка перед merge

```bash
cd frontend
npm run build
rg "\\.ant-" src
rg "theme=\\{\\{" src
rg "style=\\{\\{" src
```

Наличие `.ant-*` допустимо только в scoped override вида `.project-root .ant-*` или `.project-root > .ant-*` внутри `@layer components`. Наличие `theme={{ ... }}` допустимо только в `ThemeProvider` для `algorithm` и `zeroRuntime`. Наличие `style={{ ... }}` допустимо для динамических вычислений, но не для постоянной визуальной кастомизации.

Для апгрейда на AntD v6 дополнительно:

```bash
npm ls antd @ant-design/icons
rg "dropdownClassName|popupClassName|dropdownRender|dropdownStyle|bordered" src
```

В v6 `@ant-design/icons` должен быть `>= 6.0.0`, а deprecated popup/dropdown API нужно заменить на новые `classNames`, `styles`, `popupRender`, `variant`.

## Источники

- Tailwind CSS, установка через Vite: https://tailwindcss.com/docs/installation/using-vite
- Tailwind CSS, theme variables и `@theme`: https://tailwindcss.com/docs/theme
- Tailwind CSS, ручной dark mode через `@custom-variant`: https://tailwindcss.com/docs/dark-mode
- Tailwind CSS, `@apply` и директивы: https://tailwindcss.com/docs/functions-and-directives
- Ant Design v6, `ConfigProvider.theme`, `zeroRuntime`, tokens и component tokens: https://ant.design/docs/react/customize-theme/
- Ant Design v6, cascade layers и `antd.css` при `zeroRuntime`: https://ant.design/docs/react/compatible-style/
- Ant Design v6 migration guide: https://ant.design/docs/react/migration-v6/
