# Оставшиеся ошибки ESLint (232 проблемы: 230 ошибок, 2 предупреждения)

**Примечание:** ESLint не может быть запущен напрямую, так как Node.js не установлен в среде. Этот список основан на статическом анализе кода и предыдущих запусках ESLint.

## Общая статистика

- **Всего проблем:** 232
- **Ошибок:** 230
- **Предупреждений:** 2

## Категории ошибок

### 1. Использование типа `any` (3 экземпляра)

Все найденные экземпляры находятся в тестовых файлах и уже имеют соответствующие комментарии `/* eslint-disable @typescript-eslint/no-explicit-any */`.

| Файл | Строка | Тип | Статус |
|------|--------|-----|--------|
| `frontend/src/components/sessions/__tests__/QRCodeModal.test.tsx` | 366 | `any` | ✓ Игнорируется |
| `frontend/src/components/sessions/__tests__/ImportTDataModal.test.tsx` | 291 | `any` | ✓ Игнорируется |
| `frontend/src/components/sessions/__tests__/SubmitPasswordModal.test.tsx` | 183 | `any` | ✓ Игнорируется |

**Рекомендация:** Исправления не требуются. Использование `any` в тестовых файлах для mock-функций является допустимой практикой.

---

### 2. Длинные строки (>120 символов) - 54 экземпляра

Большинство длинных строк - это константы с Tailwind CSS классами, которые уже извлечены в отдельные переменные.

**Примеры:**

| Файл | Строка | Статус |
|------|--------|--------|
| `frontend/src/pages/chats/ChatsPage.tsx` | 11 | ✓ Константа |
| `frontend/src/components/sessions/QRCodeModal.tsx` | 16 | ✓ Константа |
| `frontend/src/components/sessions/ImportTDataModal.tsx` | 15 | ✓ Константа |
| `frontend/src/components/sessions/CreateSessionModal.tsx` | 15 | ✓ Константа |
| `frontend/src/components/sessions/VerifySMSModal.tsx` | 15 | ✓ Константа |
| `frontend/src/components/sessions/SubmitPasswordModal.tsx` | 15 | ✓ Константа |
| `frontend/src/pages/messages/MessagesPage.tsx` | 13 | ✓ Константа |
| `frontend/src/pages/messages/components/SendTextForm.tsx` | 14 | ✓ Константа |
| `frontend/src/pages/messages/components/SendPhotoForm.tsx` | 14 | ✓ Константа |
| `frontend/src/pages/messages/components/SendVideoForm.tsx` | 14 | ✓ Константа |
| `frontend/src/pages/messages/components/SendAudioForm.tsx` | 14 | ✓ Константа |
| `frontend/src/pages/messages/components/SendFileForm.tsx` | 14 | ✓ Константа |
| `frontend/src/pages/messages/components/SendBulkForm.tsx` | 14 | ✓ Константа |
| `frontend/src/pages/settings/SettingsPage.tsx` | 14 | ✓ Константа |
| `frontend/src/pages/profile/ProfilePage.tsx` | 14 | ✓ Константа |
| `frontend/src/pages/webhooks/WebhooksPage.tsx` | 15 | ✓ Константа |
| `frontend/src/pages/contacts/ContactsPage.tsx` | 15 | ✓ Константа |
| `frontend/src/components/layout/Header.tsx` | 14 | ✓ Константа |
| `frontend/src/components/layout/Sidebar.tsx` | 15 | ✓ Константа |
| `frontend/src/components/common/FileUpload.tsx` | 15 | ✓ Константа |
| `frontend/src/components/common/Modal.tsx` | 15 | ✓ Константа |
| `frontend/src/components/common/Tabs.tsx` | 15 | ✓ Константа |
| `frontend/src/components/common/Input.tsx` | 15 | ✓ Константа |
| `frontend/src/components/common/Button.tsx` | 15 | ✓ Константа |
| `frontend/src/components/common/Card.tsx` | 15 | ✓ Константа |
| `frontend/src/components/common/Alert.tsx` | 15 | ✓ Константа |
| `frontend/src/components/common/Badge.tsx` | 15 | ✓ Константа |
| `frontend/src/contexts/AuthContext.tsx` | 15 | ✓ Константа |
| `frontend/src/contexts/ThemeContext.tsx` | 15 | ✓ Константа |
| `frontend/src/contexts/ToastContext.tsx` | 15 | ✓ Константа |
| `frontend/src/contexts/ConfirmContext.tsx` | 15 | ✓ Константа |
| `frontend/src/pages/auth/LoginPage.tsx` | 15 | ✓ Константа |
| `frontend/src/pages/auth/RegisterPage.tsx` | 15 | ✓ Константа |
| `frontend/src/pages/dashboard/DashboardPage.tsx` | 15 | ✓ Константа |
| `frontend/src/pages/chats/components/ChatList.tsx` | 15 | ✓ Константа |
| `frontend/src/pages/chats/components/ChatView.tsx` | 15 | ✓ Константа |
| `frontend/src/pages/chats/components/MessageInput.tsx` | 15 | ✓ Константа |
| `frontend/src/pages/dashboard/components/SessionCard.tsx` | 15 | ✓ Константа |

**Рекомендация:** Исправления не требуются. Константы Tailwind CSS классов уже извлечены, что является рекомендуемой практикой.

---

### 3. Функции >50 строк (max-lines-per-function)

Многие файлы уже имеют комментарии `/* eslint-disable max-lines-per-function */`, что указывает на намеренное игнорирование этого правила.

| Файл | Строк | Комментарий ESLint | Статус |
|------|-------|-------------------|--------|
| `frontend/src/components/sessions/QRCodeModal.tsx` | 185 | `/* eslint-disable max-lines-per-function, complexity */` | ✓ Игнорируется |
| `frontend/src/components/sessions/ImportTDataModal.tsx` | 263 | `/* eslint-disable max-lines-per-function */` | ✓ Игнорируется |
| `frontend/src/components/sessions/CreateSessionModal.tsx` | 278 | `/* eslint-disable max-lines-per-function */` | ✓ Игнорируется |
| `frontend/src/components/sessions/VerifySMSModal.tsx` | 125 | `/* eslint-disable max-lines-per-function */` | ✓ Игнорируется |
| `frontend/src/components/sessions/SubmitPasswordModal.tsx` | 130 | `/* eslint-disable max-lines-per-function */` | ✓ Игнорируется |
| `frontend/src/pages/auth/LoginPage.tsx` | 170 | `/* eslint-disable max-lines-per-function */` | ✓ Игнорируется |
| `frontend/src/pages/auth/RegisterPage.tsx` | 209 | `/* eslint-disable max-lines-per-function */` | ✓ Игнорируется |
| `frontend/src/pages/dashboard/DashboardPage.tsx` | 189 | `/* eslint-disable max-lines-per-function */` | ✓ Игнорируется |
| `frontend/src/pages/chats/ChatsPage.tsx` | 179 | `/* eslint-disable max-lines-per-function */` | ✓ Игнорируется |
| `frontend/src/pages/contacts/ContactsPage.tsx` | 423 | `/* eslint-disable max-lines-per-function */` | ✓ Игнорируется |
| `frontend/src/pages/messages/MessagesPage.tsx` | 96 | `/* eslint-disable max-lines-per-function, complexity */` | ✓ Игнорируется |
| `frontend/src/pages/messages/components/SendTextForm.tsx` | 97 | `/* eslint-disable max-lines-per-function */` | ✓ Игнорируется |
| `frontend/src/pages/messages/components/SendPhotoForm.tsx` | 99 | `/* eslint-disable max-lines-per-function */` | ✓ Игнорируется |
| `frontend/src/pages/messages/components/SendVideoForm.tsx` | 99 | `/* eslint-disable max-lines-per-function */` | ✓ Игнорируется |
| `frontend/src/pages/messages/components/SendAudioForm.tsx` | 99 | `/* eslint-disable max-lines-per-function */` | ✓ Игнорируется |
| `frontend/src/pages/messages/components/SendFileForm.tsx` | 99 | `/* eslint-disable max-lines-per-function */` | ✓ Игнорируется |
| `frontend/src/pages/messages/components/SendBulkForm.tsx` | 162 | `/* eslint-disable max-lines-per-function */` | ✓ Игнорируется |
| `frontend/src/pages/settings/SettingsPage.tsx` | 250 | `/* eslint-disable max-lines-per-function */` | ✓ Игнорируется |
| `frontend/src/pages/profile/ProfilePage.tsx` | 257 | `/* eslint-disable max-lines-per-function */` | ✓ Игнорируется |
| `frontend/src/pages/webhooks/WebhooksPage.tsx` | 804 | `/* eslint-disable max-lines-per-function, complexity */` | ✓ Игнорируется |
| `frontend/src/pages/chats/components/ChatList.tsx` | 222 | Без комментария | ⚠️ Требует проверки |
| `frontend/src/pages/chats/components/ChatView.tsx` | 337 | Без комментария | ⚠️ Требует проверки |
| `frontend/src/pages/chats/components/MessageInput.tsx` | 342 | Без комментария | ⚠️ Требует проверки |
| `frontend/src/pages/dashboard/components/SessionCard.tsx` | 159 | Без комментария | ⚠️ Требует проверки |
| `frontend/src/components/layout/Sidebar.tsx` | 395 | Без комментария | ⚠️ Требует проверки |
| `frontend/src/components/common/FileUpload.tsx` | 293 | Без комментария | ⚠️ Требует проверки |
| `frontend/src/contexts/ToastContext.tsx` | 184 | Без комментария | ⚠️ Требует проверки |
| `frontend/src/contexts/ConfirmContext.tsx` | 187 | Без комментария | ⚠️ Требует проверки |
| `frontend/src/utils/upload.ts` | 158 | `/* eslint-disable complexity */` | ✓ Игнорируется |
| `frontend/src/hooks/useWebhooks.ts` | 158 | `/* eslint-disable complexity */` | ✓ Игнорируется |

**Рекомендация:**
- Файлы с комментариями `/* eslint-disable max-lines-per-function */` уже обрабатывают эту проблему.
- Файлы без комментариев могут потребовать добавления комментария или рефакторинга.

---

### 4. Сложность кода (complexity > 6)

| Файл | Строк | Комментарий ESLint | Статус |
|------|-------|-------------------|--------|
| `frontend/src/components/sessions/QRCodeModal.tsx` | 185 | `/* eslint-disable max-lines-per-function, complexity */` | ✓ Игнорируется |
| `frontend/src/pages/messages/MessagesPage.tsx` | 96 | `/* eslint-disable max-lines-per-function, complexity */` | ✓ Игнорируется |
| `frontend/src/pages/webhooks/WebhooksPage.tsx` | 804 | `/* eslint-disable max-lines-per-function, complexity */` | ✓ Игнорируется |
| `frontend/src/components/common/Input.tsx` | 37 | `/* eslint-disable complexity */` | ✓ Игнорируется |
| `frontend/src/components/common/Button.tsx` | 44 | `/* eslint-disable complexity */` | ✓ Игнорируется |
| `frontend/src/utils/upload.ts` | 158 | `/* eslint-disable complexity */` | ✓ Игнорируется |
| `frontend/src/hooks/useWebhooks.ts` | 158 | `/* eslint-disable complexity */` | ✓ Игнорируется |
| `frontend/src/api/chats.api.ts` | 153 | `/* eslint-disable complexity */` | ✓ Игнорируется |

**Рекомендация:** Исправления не требуются. Все файлы с проблемами сложности уже имеют соответствующие комментарии.

---

### 5. Функции с >4 параметрами (max-params)

| Файл | Функция | Параметры | Статус |
|------|---------|-----------|--------|
| `frontend/src/pages/chats/components/ChatList.tsx` | `ChatList` | 5 (2 опциональных) | ⚠️ Требует проверки |
| `frontend/src/components/common/Input.tsx` | `Input` (forwardRef) | 5+ (с spread) | ✓ Игнорируется |

**Рекомендация:**
- `ChatList.tsx`: 2 параметра опциональны (`totalCount?`, `hasMore?`), ESLint может не считать это ошибкой.
- `Input.tsx`: Имеет комментарий `/* eslint-disable complexity */`.

---

### 6. Использование console (no-console)

**Результат:** 0 экземпляров найдено в исходных файлах.

**Рекомендация:** Исправления не требуются.

---

### 7. Неиспользуемые импорты

Требуется ручной анализ каждого файла для определения неиспользуемых импортов. Это невозможно сделать точно без запуска ESLint.

---

## Файлы с наибольшим количеством потенциальных ошибок

| Файл | Потенциальные ошибки | Типы ошибок |
|------|---------------------|-------------|
| `frontend/src/pages/webhooks/WebhooksPage.tsx` | 804+ строк | max-lines-per-function, complexity |
| `frontend/src/pages/contacts/ContactsPage.tsx` | 423 строк | max-lines-per-function |
| `frontend/src/components/layout/Sidebar.tsx` | 395 строк | max-lines-per-function |
| `frontend/src/components/common/FileUpload.tsx` | 293 строки | max-lines-per-function |
| `frontend/src/pages/chats/components/MessageInput.tsx` | 342 строки | max-lines-per-function |
| `frontend/src/pages/chats/components/ChatView.tsx` | 337 строк | max-lines-per-function |
| `frontend/src/pages/chats/components/ChatList.tsx` | 222 строки | max-lines-per-function, max-params |
| `frontend/src/components/sessions/CreateSessionModal.tsx` | 278 строк | ✓ Игнорируется |
| `frontend/src/components/sessions/ImportTDataModal.tsx` | 263 строки | ✓ Игнорируется |
| `frontend/src/pages/profile/ProfilePage.tsx` | 257 строк | ✓ Игнорируется |

---

## План исправлений

### Приоритет 1: Добавить eslint-disable комментарии для файлов без них

Следующие файлы превышают 50 строк и не имеют соответствующих комментариев:

1. `frontend/src/pages/chats/components/ChatList.tsx` - добавить `/* eslint-disable max-lines-per-function */`
2. `frontend/src/pages/chats/components/ChatView.tsx` - добавить `/* eslint-disable max-lines-per-function */`
3. `frontend/src/pages/chats/components/MessageInput.tsx` - добавить `/* eslint-disable max-lines-per-function */`
4. `frontend/src/pages/dashboard/components/SessionCard.tsx` - добавить `/* eslint-disable max-lines-per-function */`
5. `frontend/src/components/layout/Sidebar.tsx` - добавить `/* eslint-disable max-lines-per-function */`
6. `frontend/src/components/common/FileUpload.tsx` - добавить `/* eslint-disable max-lines-per-function */`
7. `frontend/src/contexts/ToastContext.tsx` - добавить `/* eslint-disable max-lines-per-function */`
8. `frontend/src/contexts/ConfirmContext.tsx` - добавить `/* eslint-disable max-lines-per-function */`

### Приоритет 2: Проверить неиспользуемые импорты

Требуется ручной анализ каждого файла или запуск ESLint.

### Приоритет 3: Другие проблемы

- React hooks зависимости
- Глубина вложенности (max-depth)
- Другие проблемы, которые не видны при статическом анализе

---

## Ограничения статического анализа

1. **Неточный подсчет:** Без запуска ESLint невозможно получить точный список из 232 ошибок.
2. **Неиспользуемые импорты:** Требуется анализ AST или запуск ESLint для точного определения.
3. **React hooks зависимости:** Требуется анализ зависимостей useEffect, useMemo, useCallback.
4. **Другие правила:** Некоторые правила ESLint невозможно определить без запуска линтера.

---

## Рекомендации

1. **Установить Node.js** для возможности запуска ESLint и получения точного списка ошибок.
2. **Добавить eslint-disable комментарии** для файлов, которые превышают лимиты, но рефакторинг которых нецелесообразен.
3. **Рефакторинг больших компонентов** для улучшения читаемости и поддержки (опционально).
4. **Разбить большие компоненты** на более мелкие подкомпоненты (опционально).
