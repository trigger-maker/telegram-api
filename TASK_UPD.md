# 📋 Проектное задание — Patch v2.3
## telegram-session-api: пять точечных правок

**Репозиторий:** форк [github.com/ghmedinac123/telegram-session-api](https://github.com/ghmedinac123/telegram-session-api)
**Стек:** Go 1.23+, Fiber v2, gotd/td, pgx v5, zerolog
**Язык кода:** Go. Язык всех комментариев, логов, сообщений об ошибках — **английский**.
**Принцип:** минимальные хирургические правки. Архитектура, БД, существующие эндпоинты — не трогать.

***

## 1. Контекст

Оригинальный сервис работает как HTTP API-шлюз к Telegram через MTProto. Используется по модели **один Telegram-аккаунт = один пользователь API**. Пять правок устраняют критичные проблемы надёжности без изменения архитектуры.

***

## 2. Подготовительный шаг — изменения domain/session.go

> ⚠️ Все изменения `domain/session.go` вносятся **одним коммитом** до начала реализации правок — чтобы избежать конфликтов между правками 3, 4 и 5.

Добавить следующие константы к существующим (ничего не удалять):

**Новые `auth_state`:**
```
password_required   — аккаунт ожидает ввода 2FA пароля (Правка 3)
banned              — аккаунт заблокирован Telegram'ом (Правка 5)
frozen              — аккаунт временно заморожен Telegram'ом (Правка 5)
```

**Новый `auth_method`:**
```
tdata               — сессия импортирована из Telegram Desktop (Правка 4)
```

**Новые ошибки в `domain/errors.go`** (также одним коммитом):
```
ErrInvalidPassword      — неверный 2FA пароль (Правка 3)
ErrAlreadyAuthenticated — сессия не в состоянии password_required (Правка 3)
ErrTDataInvalid         — невалидные tdata файлы (Правка 4)
```

### Затронутые файлы
```
internal/domain/session.go  — UPDATE
internal/domain/errors.go   — UPDATE
```

***

## 3. Правка 1 — Персистентное хранилище сессий

### Проблема
Сессионные байты MTProto хранятся в оперативной памяти (`StorageMemory`). При любом рестарте сервиса они теряются. Telegram отвечает `AUTH_KEY_UNREGISTERED` и требует повторной авторизации всех аккаунтов.

### Требования
- Создать новый тип `PersistentSessionStorage` реализующий интерфейс `telegram.SessionStorage` из gotd/td
- При каждом вызове `StoreSession` — шифровать байты через существующий `pkg/crypto` (AES-256-GCM) и немедленно сохранять в Postgres
- При вызове `LoadSession` — читать из Postgres и расшифровывать
- Если в БД нет данных для сессии — возвращать пустой срез байт, не ошибку
- Добавить метод `UpdateSessionData(sessionID, data)` в интерфейс `domain.SessionRepository` и его реализацию в postgres-репозитории
- Заменить все вхождения `StorageMemory` в `session_pool.go` и `manager.go` на новый тип
- Для методов авторизации (`SendCode`, `SignIn`, `StartQRAuth`) — `session_id` передаётся как параметр, он известен до старта авторизации

### Затронутые файлы
```
internal/telegram/storage.go              — NEW
internal/telegram/session_pool.go         — UPDATE
internal/telegram/manager.go              — UPDATE
internal/domain/repository.go             — UPDATE
repository/postgres/session_repository.go — UPDATE
```

***

## 4. Правка 2 — Переиспользование TCP-соединения при отправке

### Проблема
`sender.go` создаёт новый `telegram.Client` (новое TCP-соединение) на каждый вызов `SendMessage`. Telegram регистрирует каждое подключение как отдельное устройство. При лимите ~10 устройств на аккаунт сервис получает `SESSION_REVOKED` и теряет авторизацию.

### Требования
- Метод `SendMessage` в `manager.go` принимает готовый `*tg.Client` из пула вместо создания нового
- `MessageService` получает зависимость `*SessionPool` и при отправке берёт `ActiveSession.API` через `GetActiveSession(sessionID)`
- Если сессия отсутствует в пуле — возвращать `ErrSessionNotActive`
- `SendBulk` берёт `ActiveSession.API` из пула **один раз до начала цикла** — переиспользует одно соединение для всех получателей в пачке
- Если сессия отсутствует в пуле при вызове `SendBulk` — возвращать `ErrSessionNotActive` до начала цикла
- Удалить из `sender.go` вспомогательный тип `memorySession` — он больше не нужен
- HTTP-клиент для скачивания медиафайлов в `downloadFile` должен иметь таймаут 30 секунд — оригинал использует `http.Get` без таймаута, что блокирует выполнение при недоступном URL

### Затронутые файлы
```
internal/telegram/sender.go         — UPDATE
internal/service/message_service.go — UPDATE
```

***

## 5. Правка 3 — Авторизация с 2FA паролем

### Проблема
`SignIn` в `manager.go` не обрабатывает ответ `SESSION_PASSWORD_NEEDED`. Аккаунты с включённой двухфакторной аутентификацией не могут авторизоваться — процесс обрывается с необработанной ошибкой.

### Требования

**Изменение `SignIn`:**
- При получении `SESSION_PASSWORD_NEEDED` — не возвращать ошибку, а вернуть флаг `needs_password: true` и поле `hint` из `account.getPassword`
- Если Telegram вернул пустой `hint` — передавать пустую строку `""`, не исключать поле из ответа
- `auth_state` сессии при этом обновляется до `password_required`

**Новый метод `SubmitPassword`:**
- Принимает `session_id` и пароль в открытом виде
- Выполняет полный SRP-flow: `account.getPassword` → вычисление `InputCheckPasswordSRP` → `auth.checkPassword`
- SRP-математика решается через встроенный хелпер gotd/td `telegram.SolvePassword` — реализовывать вручную не нужно
- При неверном пароле (`PASSWORD_HASH_INVALID`) — возвращать `ErrInvalidPassword`, `auth_state` не меняется
- При вызове для сессии с `auth_state` отличным от `password_required` — возвращать `ErrAlreadyAuthenticated`
- При успехе — `auth_state: authenticated`, сессия запускается в пул

**Новый endpoint:**
```
POST /api/v1/sessions/:id/submit-password
body:    {"password": "..."}
200:     {session} с auth_state: "authenticated"
400:     INVALID_PASSWORD — неверный пароль
404:     NOT_FOUND — сессия не существует
409:     ALREADY_AUTHENTICATED — сессия не в состоянии password_required
```

### Затронутые файлы
```
internal/telegram/manager.go        — UPDATE
internal/service/session_service.go — UPDATE
internal/handler/session_handler.go — UPDATE
```

***

## 6. Правка 4 — Импорт tdata

### Проблема
Нет возможности импортировать аккаунты из Telegram Desktop. Нужно для аккаунтов которые уже авторизованы в Telegram Desktop и не требуют SMS.

### Требования

**Новый метод `ImportTData` в `manager.go`:**
- Принимает `api_id`, `api_hash`, `session_name` и map файлов `filename → bytes`
- Использует `session.TDesktopSession` парсер из gotd/td
- Берёт первый аккаунт из найденных в tdata
- Сохраняет session bytes через `PersistentSessionStorage` (Правка 1)
- Верифицирует сессию через `client.Self()` — реальный вызов к Telegram
- При невалидных файлах — возвращать `ErrTDataInvalid` с описанием причины
- При отклонении сессии Telegram'ом — также `ErrTDataInvalid`
- Поддерживаемые версии Telegram Desktop: 3.x и 4.x (ограничение gotd/td)

**Запись в БД:**
- Сессия создаётся с `auth_method: "tdata"`
- `auth_state: authenticated`, `is_active: true` проставляются сразу при успешном импорте

**Новый endpoint:**
```
POST /api/v1/sessions/import-tdata
Content-Type: multipart/form-data
Fields: api_id (int, required), api_hash (string, required),
        session_name (string, optional, default: "TDesktop Import")
Files:  tdata files (map0, map1, D877F783D5D3EF8C, key_datas и др.)

201: {session_id, is_active: true, telegram_user_id, username,
      auth_method: "tdata", auth_state: "authenticated"}
400: VALIDATION — отсутствуют обязательные поля или не загружено ни одного файла
422: TDATA_INVALID — невалидные файлы или сессия отклонена Telegram'ом
```

### Затронутые файлы
```
internal/telegram/manager.go        — UPDATE
internal/service/session_service.go — UPDATE
internal/handler/session_handler.go — UPDATE
```

***

## 7. Правка 5 — Централизованная обработка MTProto-ошибок

### Проблема
Оригинал не обрабатывает специфические ошибки Telegram. При `FLOOD_WAIT`, `SESSION_REVOKED`, `USER_DEACTIVATED` и других — сервис либо зависает, либо возвращает `500 Internal Server Error` без понятного контекста и не меняет `auth_state` аккаунта.

### Требования

**Новый файл `internal/telegram/errors.go`:**
- Централизованный обработчик MTProto-ошибок
- Все вызовы Telegram API в `sender.go`, `manager.go`, `session_pool.go` проходят через него
- Все обработанные ошибки логируются через zerolog с контекстом `{session_id, error_code}`
- Неизвестные MTProto-ошибки логируются и возвращаются как `500 INTERNAL` — не паникуют

**Три категории обработки:**

**Категория A — блокирующие (остановка сессии):**

| MTProto ошибка | Новый `auth_state` | Действие |
|---|---|---|
| `AUTH_KEY_UNREGISTERED` | `banned` | Обновить Postgres, остановить `ActiveSession` в пуле, все pending задачи → `failed` |
| `SESSION_REVOKED` | `banned` | Обновить Postgres, остановить `ActiveSession` в пуле, все pending задачи → `failed` |
| `USER_DEACTIVATED_BAN` | `banned` | Обновить Postgres, остановить `ActiveSession` в пуле, все pending задачи → `failed` |
| `PHONE_NUMBER_BANNED` | `banned` | Обновить Postgres, остановить `ActiveSession` в пуле, все pending задачи → `failed` |
| `USER_DEACTIVATED` | `frozen` | Обновить Postgres, остановить `ActiveSession` в пуле, все pending задачи → `failed` |

**Категория B — временные (пауза, не остановка):**

| MTProto ошибка | Действие |
|---|---|
| `FLOOD_WAIT_X` | Извлечь X из текста ошибки, приостановить отправку на X секунд, затем продолжить автоматически |
| `SLOWMODE_WAIT_X` | Пауза только для данного peer на X секунд, остальные peer не затронуты |

**Категория C — ошибки конкретной задачи (задача failed, сессия продолжает):**

| MTProto ошибка | Действие |
|---|---|
| `PEER_ID_INVALID` | Задача → `failed`, следующая задача выполняется |
| `USERNAME_NOT_OCCUPIED` | Задача → `failed`, следующая задача выполняется |
| `INPUT_USER_DEACTIVATED` | Задача → `failed`, следующая задача выполняется |

**Авторизационные ошибки — обработка в существующих методах, не в `errors.go`:**

| MTProto ошибка | Где | Действие |
|---|---|---|
| `SESSION_PASSWORD_NEEDED` | `SignIn` | Покрыто Правкой 3 |
| `PHONE_CODE_INVALID` | `VerifyCode` | HTTP 400 `INVALID_CODE` |
| `PHONE_CODE_EXPIRED` | `VerifyCode` | HTTP 410 `CODE_EXPIRED` |
| `PASSWORD_HASH_INVALID` | `SubmitPassword` | Покрыто Правкой 3 |

### Затронутые файлы
```
internal/telegram/errors.go       — NEW
internal/telegram/sender.go       — UPDATE
internal/telegram/manager.go      — UPDATE
internal/telegram/session_pool.go — UPDATE
```

***

## 8. Удаление отладочного кода

- Удалить вызов `utils.PrintQRToTerminalWithName` из `manager.go` метода `StartQRAuth` — вывод QR в терминал сервера не для продакшна

***

## 9. Что не трогаем

| Компонент | Причина |
|---|---|
| Схема БД, миграции | Новых таблиц не требуется |
| `event_dispatche.go` | Работает корректно |
| `chat_resolver.go` | Работает корректно |
| Все существующие endpoints | Обратная совместимость |
| Bulk-отправка | Логика цикла не меняется, только источник API-клиента |

***

## 10. Порядок реализации

> ⚠️ Порядок реализации отличается от нумерации правок. Подготовительный шаг и Правка 5 выполняются первыми.

1. **Подготовительный шаг** — все изменения `domain/session.go` и `domain/errors.go` одним коммитом
2. **Правка 1** — `storage.go` + `UpdateSessionData` в репозитории + замена `StorageMemory`
3. **Правка 5** — `errors.go` + подключение во всех вызовах Telegram API
4. **Правка 2** — переработка `sender.go` + передача `pool` в `MessageService` + `SendBulk`
5. **Правка 3** — `SubmitPassword` в `manager.go` + handler + endpoint
6. **Правка 4** — `ImportTData` в `manager.go` + handler + endpoint
7. **Очистка** — удалить `PrintQRToTerminalWithName`

***

## 11. Требования к тестированию

### 11.1 Правка 1 — PersistentSessionStorage

| # | Сценарий | Ожидаемый результат |
|---|---|---|
| 1.1 | `StoreSession` с валидными байтами | Байты зашифрованы и сохранены в Postgres; повторный `LoadSession` возвращает исходные байты |
| 1.2 | `LoadSession` когда запись в БД отсутствует | Возвращает пустой срез, не ошибку |
| 1.3 | `StoreSession` при недоступной БД | Возвращает ошибку, не паникует |
| 1.4 | Рестарт сервиса после авторизации | Все `is_active: true` сессии поднимаются в пул автоматически без повторной авторизации |
| 1.5 | DC switch во время активной сессии | `StoreSession` вызывается gotd/td, новые байты сохраняются, сессия продолжает работу |
| 1.6 | `UpdateSessionData` при несуществующем `session_id` | Возвращает ошибку репозитория |

### 11.2 Правка 2 — Переиспользование TCP

| # | Сценарий | Ожидаемый результат |
|---|---|---|
| 2.1 | Отправка текста через активную сессию | Сообщение доставлено, новый TCP не создаётся |
| 2.2 | Отправка фото / видео / аудио / файла по `media_url` | Файл скачан, загружен через `uploader`, доставлен |
| 2.3 | `SendMessage` когда сессия не в пуле | Возвращает `ErrSessionNotActive` |
| 2.4 | `media_url` недоступен (таймаут сервера) | Через 30 секунд возвращает ошибку, не зависает |
| 2.5 | 20 последовательных отправок одним аккаунтом | Все доставлены, количество TCP-соединений = 1 |
| 2.6 | Отправка после DC switch аккаунта | Сообщение доставлено через обновлённое соединение |
| 2.7 | `SendBulk` на 10 получателей | API из пула берётся один раз, все 10 сообщений доставлены через одно соединение |
| 2.8 | `SendBulk` когда сессия не в пуле | Возвращает `ErrSessionNotActive` до начала цикла |

### 11.3 Правка 3 — 2FA

| # | Сценарий | Ожидаемый результат |
|---|---|---|
| 3.1 | `POST /sessions/:id/verify` с аккаунтом без 2FA | Поведение не изменилось, `auth_state: authenticated` |
| 3.2 | `POST /sessions/:id/verify` с аккаунтом с 2FA | HTTP 200, `auth_state: password_required`, поле `hint` присутствует |
| 3.3 | `POST /sessions/:id/verify` с аккаунтом с 2FA без hint | HTTP 200, `auth_state: password_required`, `hint: ""` |
| 3.4 | `POST /sessions/:id/submit-password` с верным паролем | HTTP 200, `auth_state: authenticated`, сессия в пуле |
| 3.5 | `POST /sessions/:id/submit-password` с неверным паролем | HTTP 400 `INVALID_PASSWORD`, `auth_state` не меняется |
| 3.6 | `POST /sessions/:id/submit-password` для несуществующей сессии | HTTP 404 |
| 3.7 | `POST /sessions/:id/submit-password` для сессии с `auth_state != password_required` | HTTP 409 `ALREADY_AUTHENTICATED` |
| 3.8 | Повторный `submit-password` после успешной авторизации | HTTP 409 `ALREADY_AUTHENTICATED` |

### 11.4 Правка 4 — tdata

| # | Сценарий | Ожидаемый результат |
|---|---|---|
| 4.1 | Импорт валидных tdata от Telegram Desktop 4.x | HTTP 201, `auth_method: "tdata"`, `auth_state: authenticated`, `telegram_user_id` заполнен |
| 4.2 | Импорт валидных tdata от Telegram Desktop 3.x | HTTP 201, `auth_method: "tdata"`, `auth_state: authenticated` |
| 4.3 | Загрузка повреждённых файлов | HTTP 422 `TDATA_INVALID` |
| 4.4 | Файлы от неподдерживаемой версии Telegram Desktop | HTTP 422 `TDATA_INVALID` с описанием причины |
| 4.5 | Отсутствуют обязательные поля `api_id` / `api_hash` | HTTP 400 `VALIDATION` |
| 4.6 | Не загружено ни одного файла | HTTP 400 `VALIDATION` |
| 4.7 | tdata от забаненного аккаунта | HTTP 422 `TDATA_INVALID` — Telegram отклоняет `client.Self()` |
| 4.8 | `session_name` не передан | `session_name` = `"TDesktop Import"` по умолчанию |
| 4.9 | После успешного импорта — отправка сообщения | Сообщение доставлено через импортированную сессию |

### 11.5 Правка 5 — MTProto Error Handling

| # | Сценарий | Ожидаемый результат |
|---|---|---|
| 5.1 | Telegram возвращает `FLOOD_WAIT_30` при отправке | Отправка приостанавливается на 30 сек, затем продолжается автоматически |
| 5.2 | Telegram возвращает `FLOOD_WAIT_X` с разными значениями X | Пауза всегда равна X из текста ошибки, не фиксированному значению |
| 5.3 | Telegram возвращает `SLOWMODE_WAIT_X` для конкретного чата | Пауза применяется только для данного peer, другие peer не затронуты |
| 5.4 | Telegram возвращает `SESSION_REVOKED` | `auth_state: banned` в Postgres, сессия остановлена, все pending задачи → `failed`, повторная отправка → `ErrSessionNotActive` |
| 5.5 | Telegram возвращает `AUTH_KEY_UNREGISTERED` | `auth_state: banned` в Postgres, сессия остановлена, все pending задачи → `failed` |
| 5.6 | Telegram возвращает `USER_DEACTIVATED` | `auth_state: frozen` в Postgres, сессия остановлена, все pending задачи → `failed` |
| 5.7 | Telegram возвращает `USER_DEACTIVATED_BAN` | `auth_state: banned` в Postgres, сессия остановлена, все pending задачи → `failed` |
| 5.8 | Telegram возвращает `PHONE_NUMBER_BANNED` | `auth_state: banned` в Postgres, сессия остановлена, все pending задачи → `failed` |
| 5.9 | Отправка на несуществующий `@username` (`USERNAME_NOT_OCCUPIED`) | Задача помечается `failed`, следующая задача выполняется нормально |
| 5.10 | Отправка на невалидный peer ID (`PEER_ID_INVALID`) | Задача помечается `failed`, следующая задача выполняется нормально |
| 5.11 | Отправка удалённому аккаунту (`INPUT_USER_DEACTIVATED`) | Задача помечается `failed`, следующая задача выполняется нормально |
| 5.12 | `PHONE_CODE_EXPIRED` при verify | HTTP 410 `CODE_EXPIRED` |
| 5.13 | `PHONE_CODE_INVALID` при verify | HTTP 400 `INVALID_CODE` |
| 5.14 | Неизвестная MTProto-ошибка | Логируется с `{session_id, error_code}`, возвращается `500 INTERNAL`, не паникует |

### 11.6 Регрессионные тесты

| # | Сценарий | Ожидаемый результат |
|---|---|---|
| R.1 | SMS авторизация без 2FA — полный flow | Без изменений |
| R.2 | QR авторизация — полный flow | Без изменений, QR не печатается в терминал |
| R.3 | QR regenerate | Без изменений |
| R.4 | Отправка text / photo / video / audio / file | Без изменений |
| R.5 | Bulk-отправка | Без изменений в поведении, одно TCP-соединение на пачку |
| R.6 | Webhook доставка событий | Без изменений |
| R.7 | `GET /messages/:jobId/status` | Без изменений |
| R.8 | `DELETE /sessions/:id` | Без изменений, logout корректен |
| R.9 | Входящие сообщения — webhook `message.new` | Без изменений |
| R.10 | Сервис после рестарта — все активные сессии в пуле | Без повторной авторизации |