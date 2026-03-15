# 📋 Проектное задание — Patch v1.0
## telegram-session-api: четыре точечные правки

**Репозиторий:** форк [github.com/ghmedinac123/telegram-session-api](https://github.com/ghmedinac123/telegram-session-api)
**Стек:** Go 1.23+, Fiber v2, gotd/td, pgx v5, zerolog
**Язык кода:** Go. Язык всех комментариев, логов, сообщений об ошибках — **английский**.
**Принцип:** минимальные хирургические правки. Архитектура, БД, существующие эндпоинты — не трогать.

***

## 1. Контекст

Оригинальный сервис работает как HTTP API-шлюз к Telegram через MTProto. Используется по модели **один Telegram-аккаунт = один пользователь API**. Четыре правки устраняют критичные проблемы надёжности без изменения архитектуры.

***

## 2. Правка 1 — Персистентное хранилище сессий

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
internal/telegram/storage.go           — NEW
internal/telegram/session_pool.go      — UPDATE
internal/telegram/manager.go           — UPDATE
internal/domain/repository.go          — UPDATE
repository/postgres/session_repository.go — UPDATE
```

***

## 3. Правка 2 — Переиспользование TCP-соединения при отправке

### Проблема
`sender.go` создаёт новый `telegram.Client` (новое TCP-соединение) на каждый вызов `SendMessage`. Telegram регистрирует каждое подключение как отдельное устройство. При лимите ~10 устройств на аккаунт сервис получает `SESSION_REVOKED` и теряет авторизацию.

### Требования
- Метод `SendMessage` в `manager.go` должен принимать готовый `*tg.Client` из пула вместо создания нового
- `MessageService` получает зависимость `*SessionPool` и при отправке берёт `ActiveSession.API` через `GetActiveSession(sessionID)`
- Если сессия отсутствует в пуле — возвращать `ErrSessionNotActive`
- Удалить из `sender.go` вспомогательный тип `memorySession` — он больше не нужен
- HTTP-клиент для скачивания медиафайлов (`downloadFile`) должен иметь таймаут 30 секунд. Оригинал использует `http.Get` без таймаута — это блокирует воркер при недоступном URL

### Затронутые файлы
```
internal/telegram/sender.go         — UPDATE
internal/service/message_service.go — UPDATE
```

***

## 4. Правка 3 — Авторизация с 2FA паролем

### Проблема
`SignIn` в `manager.go` не обрабатывает ответ `SESSION_PASSWORD_NEEDED`. Аккаунты с включённой двухфакторной аутентификацией не могут авторизоваться — процесс обрывается с необработанной ошибкой.

### Требования

**Изменение `SignIn`:**
- При получении ошибки `SESSION_PASSWORD_NEEDED` — не возвращать ошибку, а вернуть флаг `needs_password: true` и `hint` из `account.getPassword`
- `auth_state` сессии при этом обновляется до `password_required`

**Новый метод `SubmitPassword`:**
- Принимает `session_id` и пароль в открытом виде
- Выполняет полный SRP-flow: `account.getPassword` → вычисление `InputCheckPasswordSRP` → `auth.checkPassword`
- SRP-математика решается через встроенный хелпер gotd/td `telegram.SolvePassword` — реализовывать вручную не нужно
- При неверном пароле (`PASSWORD_HASH_INVALID`) — возвращать `ErrInvalidPassword`, `auth_state` не меняется
- При успехе — `auth_state: authenticated`, сессия запускается в пул

**Новый endpoint:**
```
POST /api/v1/sessions/:id/submit-password
body: {"password": "..."}
200: {session} с auth_state: "authenticated"
400: INVALID_PASSWORD
404: NOT_FOUND
```

**Новая константа состояния:**
- Добавить `password_required` в `domain/session.go` к существующим `auth_state` константам

**Новая ошибка:**
- Добавить `ErrInvalidPassword` в `domain/errors.go`
- Добавить обработку в `handleSessionError` в `session_handler.go`

### Затронутые файлы
```
internal/telegram/manager.go        — UPDATE
internal/service/session_service.go — UPDATE
internal/handler/session_handler.go — UPDATE
internal/domain/session.go          — UPDATE
internal/domain/errors.go           — UPDATE
```

***

## 5. Правка 4 — Импорт tdata

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

**Новый endpoint:**
```
POST /api/v1/sessions/import-tdata
Content-Type: multipart/form-data
Fields: api_id (int, required), api_hash (string, required), session_name (string, optional)
Files:  tdata files (map0, map1, D877F783D5D3EF8C, key_datas и др.)

201: {session_id, is_active: true, telegram_user_id, username, auth_state: "authenticated"}
400: VALIDATION — отсутствуют обязательные поля или нет файлов
422: TDATA_INVALID — невалидные файлы или сессия отклонена Telegram'ом
```

**Новая ошибка:**
- Добавить `ErrTDataInvalid` в `domain/errors.go`
- Добавить обработку в `handleSessionError`

### Затронутые файлы
```
internal/telegram/manager.go        — UPDATE
internal/service/session_service.go — UPDATE
internal/handler/session_handler.go — UPDATE
internal/domain/errors.go           — UPDATE
```

***

## 6. Удаление отладочного кода

- Удалить вызов `utils.PrintQRToTerminalWithName` из `manager.go` метода `StartQRAuth` — это вывод QR в терминал сервера, не для продакшна

***

## 7. Что не трогаем

| Компонент | Причина |
|---|---|
| Схема БД, миграции | Новых таблиц не требуется |
| `event_dispatche.go` | Работает корректно |
| `chat_resolver.go` | Работает корректно |
| Все существующие endpoints | Обратная совместимость |
| Bulk-отправка | Остаётся как есть |
| `domain/session.go` структуры | Только добавляем константу |

***

## 8. Требования к тестированию

### 8.1 Правка 1 — PersistentSessionStorage

| # | Сценарий | Ожидаемый результат |
|---|---|---|
| 1.1 | `StoreSession` с валидными байтами | Байты зашифрованы и сохранены в Postgres; повторный `LoadSession` возвращает исходные байты |
| 1.2 | `LoadSession` когда запись в БД отсутствует | Возвращает пустой срез, не ошибку |
| 1.3 | `StoreSession` при недоступной БД | Возвращает ошибку, не паникует |
| 1.4 | Рестарт сервиса после авторизации | Все `is_active: true` сессии поднимаются в пул автоматически без повторной авторизации |
| 1.5 | DC switch во время активной сессии | `StoreSession` вызывается gotd/td, новые байты сохраняются, сессия продолжает работу |
| 1.6 | `UpdateSessionData` при несуществующем `session_id` | Возвращает ошибку репозитория |

### 8.2 Правка 2 — Переиспользование TCP

| # | Сценарий | Ожидаемый результат |
|---|---|---|
| 2.1 | Отправка текста через активную сессию | Сообщение доставлено, новый TCP не создаётся |
| 2.2 | Отправка фото / видео / аудио / файла по `media_url` | Файл скачан, загружен через `uploader`, доставлен |
| 2.3 | `SendMessage` когда сессия не в пуле | Возвращает `ErrSessionNotActive` |
| 2.4 | `media_url` недоступен (таймаут сервера) | Через 30 секунд возвращает ошибку, не зависает |
| 2.5 | 20 последовательных отправок одним аккаунтом | Все доставлены, количество TCP-соединений = 1 |
| 2.6 | Отправка после DC switch аккаунта | Сообщение доставлено через обновлённое соединение |

### 8.3 Правка 3 — 2FA

| # | Сценарий | Ожидаемый результат |
|---|---|---|
| 3.1 | `POST /sessions` с аккаунтом без 2FA | Поведение не изменилось, `auth_state: authenticated` |
| 3.2 | `POST /sessions/:id/verify` с аккаунтом с 2FA | HTTP 200, `auth_state: password_required`, `hint` заполнен |
| 3.3 | `POST /sessions/:id/submit-password` с верным паролем | HTTP 200, `auth_state: authenticated`, сессия в пуле |
| 3.4 | `POST /sessions/:id/submit-password` с неверным паролем | HTTP 400, `INVALID_PASSWORD`, `auth_state` не меняется |
| 3.5 | `POST /sessions/:id/submit-password` для несуществующей сессии | HTTP 404 |
| 3.6 | `POST /sessions/:id/submit-password` для сессии без `password_required` | HTTP 400 или 409 |
| 3.7 | Повторный `submit-password` после успешной авторизации | HTTP 409 `already authenticated` |

### 8.4 Правка 4 — tdata

| # | Сценарий | Ожидаемый результат |
|---|---|---|
| 4.1 | Импорт валидных tdata от Telegram Desktop 4.x | HTTP 201, `auth_state: authenticated`, `telegram_user_id` заполнен |
| 4.2 | Импорт валидных tdata от Telegram Desktop 3.x | HTTP 201, `auth_state: authenticated` |
| 4.3 | Загрузка повреждённых файлов | HTTP 422 `TDATA_INVALID` |
| 4.4 | Загрузка файлов от другой версии Telegram Desktop (не 3.x/4.x) | HTTP 422 `TDATA_INVALID` с описанием |
| 4.5 | Отсутствуют обязательные поля `api_id` / `api_hash` | HTTP 400 `VALIDATION` |
| 4.6 | Не загружено ни одного файла | HTTP 400 `VALIDATION` |
| 4.7 | tdata от забаненного аккаунта | HTTP 422 `TDATA_INVALID` — Telegram отклоняет `client.Self()` |
| 4.8 | После импорта — отправка сообщения | Сообщение доставлено через импортированную сессию |

### 8.5 Регрессионные тесты

После всех четырёх правок — обязательная проверка что существующий функционал не сломан:

| # | Сценарий | Ожидаемый результат |
|---|---|---|
| R.1 | SMS авторизация без 2FA — полный flow | Без изменений |
| R.2 | QR авторизация — полный flow | Без изменений, QR не печатается в терминал |
| R.3 | QR regenerate | Без изменений |
| R.4 | Отправка text / photo / video / audio / file | Без изменений |
| R.5 | Bulk-отправка | Без изменений |
| R.6 | Webhook доставка событий | Без изменений |
| R.7 | `GET /messages/:jobId/status` | Без изменений |
| R.8 | `DELETE /sessions/:id` | Без изменений, logout корректен |
| R.9 | Входящие сообщения — webhook `message.new` | Без изменений |