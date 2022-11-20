## In-memory key-value storage (like Redis)

### Необходимый функционал:
* Клиент и сервер tcp(telnet)/REST API
* Key-value хранилище строк, списков, словарей
* Возможность установить TTL на каждый ключ
* Реализовать операторы: GET, SET, DEL, KEYS
* Реализовать покрытие несколькими тестами функционала

### Детали реализации
* REST клиент/сервер, через Swagger пишу openapi.yaml
* Внутреннее хранилище: map\[key\]value + map[key]ttl, защищённые RWMutex
* По истечении TTL не делаем ничего, реально удаляем только при обращении к ключу

### Запуск
`docker-compose up -d`
По умолчанию запускается на localhost:8080

### Клиент 
Генерируем по openapi.yaml тут https://editor-next.swagger.io/