# distributed_calc
Distributed Arithmetic Expression Calculator

# Распределенный Калькулятор Арифметических Выражений

Распределенная система для вычисления арифметических выражений с поддержкой базовых операций (+, -, *, /), которая распределяет вычисления между несколькими рабочими агентами.

## Архитектура

Система следует распределенной архитектуре:

1. **Клиент** отправляет выражение через API
2. **Оркестратор** разбирает выражение на отдельные операции
3. **Оркестратор** отправляет эти операции как задачи в очередь задач
4. **Рабочие агенты** берут задачи из очереди и обрабатывают их
5. **Рабочие агенты** отправляют результаты обратно **Оркестратору**
6. **Оркестратор** продолжает обработку выражения с полученными результатами
7. Когда все операции завершены, итоговый результат сохраняется

![image](https://github.com/user-attachments/assets/78450c9f-4b9a-43e6-a587-68beeacb3f0b)


## Установка

### Требования

- Go 1.23 или выше

### Настройка

1. Клонируйте репозиторий:
   ```
   git clone https://github.com/Oleg-Neevin/distributed_calc.git
   ```
2. Перейдите в корневую директорию:
   ```
   cd distributed_calc
   ```
3. Запустите приложение:
   ```
   go run cmd/main.go
   ```

## Конфигурация

Приложение можно настроить с помощью переменных среды:

| Переменная | Описание | Значение по умолчанию |
|------------|----------|----------------------|
| `COMPUTING_POWER` | Количество рабочих горутин на агента | 3 |
| `TIME_ADDITION_MS` | Время обработки операций сложения (мс) | 100 |
| `TIME_SUBTRACTION_MS` | Время обработки операций вычитания (мс) | 100 |
| `TIME_MULTIPLICATIONS_MS` | Время обработки операций умножения (мс) | 200 |
| `TIME_DIVISIONS_MS` | Время обработки операций деления (мс) | 300 |

Пример:
```
COMPUTING_POWER=5 TIME_ADDITION_MS=50 go run cmd/main.go
```

## Использование API

### Вычисление выражения

**Запрос:**
```
POST /api/v1/calculate
```

**Тело:**
```json
{
  "expression": "2+3*4"
}
```

**Ответ:**
```json
{
  "id": 1
}
```

**Пример:**
```bash
curl --location 'http://localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "2+3*4"
}'
```

### Получение всех выражений

**Запрос:**
```
GET /api/v1/expressions
```

**Ответ:**
```json
{
  "expressions": [
    {
      "id": 1,
      "expression": "2+3*4",
      "status": "completed",
      "result": 14
    },
    {
      "id": 2,
      "expression": "10/2+5",
      "status": "processing",
      "result": 0
    }
  ]
}
```

**Пример:**
```bash
curl --location 'http://localhost:8080/api/v1/expressions'
```

### Получение выражения по ID

**Запрос:**
```
GET /api/v1/expressions/{id}
```

**Ответ:**
```json
{
  "expression": {
    "id": 1,
    "expression": "2+3*4",
    "status": "completed",
    "result": 14
  }
}
```

**Пример:**
```bash
curl --location 'http://localhost:8080/api/v1/expressions/1'
```

## Примеры использования

### Успешный случай

Отправка выражения:
```bash
curl --location 'http://localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "2+3*4"
}'
```

Ответ:
```json
{
  "id": 1
}
```

Проверка результата:
```bash
curl --location 'http://localhost:8080/api/v1/expressions/1'
```

Ответ:
```json
{
  "expression": {
    "id": 1,
    "expression": "2+3*4",
    "status": "completed",
    "result": 14
  }
}
```

### Случай ошибки (деление на ноль)

Отправка выражения с делением на ноль:
```bash
curl --location 'http://localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "5/0"
}'
```

Ответ:
```json
{
  "id": 2
}
```

Проверка результата:
```bash
curl --location 'http://localhost:8080/api/v1/expressions/2'
```

Ответ:
```json
{
  "expression": {
    "id": 2,
    "expression": "5/0",
    "status": "error",
    "result": 0
  }
}
```

### Случай ошибки (неверное выражение)

Отправка неверного выражения:
```bash
curl --location 'http://localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "2++3"
}'
```

Ответ:
```json
{
  "id": 3
}
```

Проверка результата:
```bash
curl --location 'http://localhost:8080/api/v1/expressions/3'
```

Ответ:
```json
{
  "expression": {
    "id": 3,
    "expression": "2++3",
    "status": "error",
    "result": 0
  }
}
```

## Тестирование

Проект покрыт тестами. Для запуска тестов используйте:

```bash
go test ./...
```

Тесты включают:
- Unit-тесты агента и оркестратора
- Интеграционные тесты для проверки полного цикла обработки выражений
- Тесты производительности с несколькими одновременными запросами
