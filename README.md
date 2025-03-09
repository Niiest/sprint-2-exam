# Итоговый проект второго модуля

##  Инфраструктура
Docker-compose файл лежит в файле [docker-compose.yaml](https://github.com/Niiest/sprint-2-exam/blob/master/infra/docker-compose.yaml)

Команды для создания топиков лежат в корневом файле [commands.txt](https://github.com/Niiest/sprint-2-exam/blob/master/commands.txt)

##  Приложение
Команда для запуска: `go run .`

### Структура
Модули:
1. main
    - Запуск компонент приложения
2. message
    - Фильтр сообщений: цензура (по словарю __masking-words-table__) + фильтрация по заблокированным пользователям (по словарю __blocked-users-table__)
    - Сообщения поступают в топик __messages__
    - Прошедшие фильтрацию и цензуру сообщения попадают в топик __filtered-messages__
3. user
    - Процессор заблокированных пользователей слушает стрим __users-to-block__
    - В таблице группы __blocked-users-table__ в значении хранится хэш-мапа __Blocks__ с ID заблокированных пользователей, т.е. черный список. Значение в хэш-мапе не используется, т.к. структура взята для простоты. По-хорошему, тут должен быть Set. Ключом таблицы группы является ID пользователя, чей черный список хранится
4. words
    - Процессор словаря слов для цензурирования слушает стрим __words-to-mask__
    - В таблице группы __masking-words-table__ хранятся слова, они же являются ключом. Внутри Value хранится флаг __IsActive__, активно ли слово для маскирования или нет
5. cmd
    - __block-user__ - утилита для отправки события о блокировке сообщений от указанного пользователя для другого указанного пользователя
        - Пример команды: `go run .\cmd\block-user\main.go -blockerUserId 1 -blockedUserId 0 -isBlocked 1`
    - __mask-word__  - утилита для отправки события для добавления или деактивации слова в словаре для маскирования (цензурирования)
        - Пример команды: `go run .\cmd\mask-word\main.go -word " Hey " -isActive 1`

## Тестирование

Адрес UI: http://localhost:8080/

### Топик цензурирования

Отправить ивент через UI в топик `words-to-mask`
```
Key: everything
Value: {
	"IsActive": true
}
```

### Топик заблокированных пользователей

Отправить ивент через UI в топик `users-to-block`
```
Key: 0
Value: {
	"BlockedUserId": 3,
	"IsBlocked": true
}
```

### Чат

Отправить ивенты через UI в топик `messages`
```
Key: 7e2b2ea1-4e73-4150-9c45-7d230d7bc45e
Value: {
	"SenderId": 1,
	"RecipientId": 0,
	"Text": "How are u",
	"CreatedAt": "2025-03-10T19:03:55Z"
}

Key: 8e2b2ea1-4e73-4150-9c45-7d230d7bc45e
Value: {
	"SenderId": 0,
	"RecipientId": 1,
	"Text": "I'm OK",
	"CreatedAt": "2025-03-10T21:05:55Z"
}

Key: 1e2b2ea1-4e73-4150-9c45-7d230d7bc45e
Value: {
	"SenderId": 1,
	"RecipientId": 0,
	"Text": "Fine then",
	"CreatedAt": "2025-03-10T22:05:55Z"
}

Key: 2e2b2ea1-4e73-4150-9c45-7d230d7bc45e
Value: {
	"SenderId": 2,
	"RecipientId": 0,
	"Text": "Show me everything",
	"CreatedAt": "2025-03-10T22:05:55Z"
}

Key: 3e2b2ea1-4e73-4150-9c45-7d230d7bc45e
Value: {
	"SenderId": 3,
	"RecipientId": 0,
	"Text": "I'm blocked! Unblock me!",
	"CreatedAt": "2025-03-10T22:05:55Z"
}
```

Ожидаемый вывод в stdout:
```
[proc] key: everything,  msg: &{IsActive:true}, data in group_table &{true} 
[proc] key: 0,  msg: &{BlockedUserId:3 IsBlocked:true}, data in group_table &{map[3:true]} 
Message from 1 to 0 is sending: 'How are u'
Message from 0 to 1 is sending: 'I'm OK'
Message from 1 to 0 is sending: 'Fine then'
Message from 2 to 0 is sending: 'Show me **********'
Message from 3 to 0 will NOT be sent: 'I'm blocked! Unblock me!'
```

### Разблокирование пользователей

Отправить ивент через UI в топик `users-to-block`
```
Key: 0
Value: {
	"BlockedUserId": 3,
	"IsBlocked": false
}
```

### Разблокирование пользователей

Отправить ивент через UI в топик `users-to-block`
```
Key: 0
Value: {
	"BlockedUserId": 3,
	"IsBlocked": false
}
```

Отправить ивент через UI в топик `messages`
```
Key: 4e2b2ea1-4e73-4150-9c45-7d230d7bc45e
Value: {
	"SenderId": 3,
	"RecipientId": 0,
	"Text": "I'm not blocked and everything is fine",
	"CreatedAt": "2025-03-10T23:05:55Z"
}
```

### Отмена маскирования

Отправить ивент через UI в топик `words-to-mask`
```
Key: everything
Value: {
	"IsActive": false
}
```

Отправить ивент через UI в топик `messages`
```
Key: 5e2b2ea1-4e73-4150-9c45-7d230d7bc45e
Value: {
	"SenderId": 3,
	"RecipientId": 0,
	"Text": "Now everything is really fine",
	"CreatedAt": "2025-03-10T23:06:55Z"
}
```

### Ожидаемый результат
1. В `filtered-messages` должно быть 4 из 5 сообщений (кроме последнего)
2. В сообщении 2e2b2ea1-4e73-4150-9c45-7d230d7bc45e слово `everything` маскировано
3. Сообщение 4e2b2ea1-4e73-4150-9c45-7d230d7bc45e от пользователя 3 дошло после разблокировки, и применено маскирование
3. Сообщение 5e2b2ea1-4e73-4150-9c45-7d230d7bc45e от пользователя 3 дошло до пользователя 0, и маскирование не применялось после его отмены
