# Море Парк — система управления аквапарком

Full-stack учебный проект: Go backend + React frontend + PostgreSQL.

## Требования

- Go 1.22+
- Node.js 18+
- PostgreSQL 14+

## Запуск

### 1. База данных

```sql
CREATE DATABASE morepark;
```

### 2. Backend

```bash
cd backend
cp .env.example .env
# Отредактируйте .env — укажите пароль PostgreSQL и JWT_SECRET

go run ./cmd/api
```

Сервер: http://localhost:8080

### 3. Frontend

```bash
cd frontend
npm install
npm run dev
```

Приложение: http://127.0.0.1:5173

## Тестовые аккаунты

Пароль для всех: `test123`

| Email | Роль |
|-------|------|
| director@morepark.ru | Директор |
| cashier@morepark.ru | Кассир |
| lifeguard@morepark.ru | Спасатель |
| technician@morepark.ru | Тех. служба |
| barman@morepark.ru | Бармен |

## Страницы

- `/login` — вход для персонала
- `/admin` — панель управления
- `/buy` — онлайн-покупка билетов (публичная)
