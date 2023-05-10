# Antibruteforce
Сервис предназначен для борьбы с подбором паролей при авторизации в какой-либо системе.
Сервис вызывается перед авторизацией пользователя и может либо разрешить, либо заблокировать попытку.
Cервис используется только для server-server, т.е. скрыт от конечного пользователя.

# Алгоритм работы
Сервис ограничивает частоту попыток авторизации для различных комбинаций параметров, например:

не более N = 10 попыток в минуту для данного логина.
не более M = 100 попыток в минуту для данного пароля (защита от обратного brute-force).
не более K = 1000 попыток в минуту для данного IP (число большое, т.к. NAT).

Параметры N,M и K задаются через переменные в файле конфигурации ./configs/dc/config.env:
LIMITFACTOR_LOGIN
LIMITFACTOR_PASSWORD
LIMITFACTOR_IP

Временной промежуток хранения частоты попыток в рамках одного измерения задается через параметр LIMIT_TIMECHECK.

Бакеты для подсчета количества попыток хранятся в Redis.

Частью микросервиса является БД MySQL,в ней хранятся подсети для black/white списков.
Если входящий IP находится в whitelist, то сервис безусловно разрешает авторизацию (ok=true).
Если - в blacklist, то отклоняет (ok=false).

# Развертывание
Развертывание микросервиса осуществляется командой make run (внутри docker compose up) в директории с проектом.
Для проекта предусмотрен CLI-интерфейс, он находится в папке ./bin
CLI-интерфейс не контейнеризирован, позволяет управлять сервисом внутри docker compose.
Настройки CLI-интерфейса находятся в файле конфигурации ./configs/config_cli.env

# Тестирование
Unit-тестирование производится с помощью команды make test.
Интеграционное тестирование производится с помощью команды  make integration-tests. 
При интеграционном тестировании поднимается тестовое окружение в docker compose, прогоняются интеграционные тесты,
в случае ошибки возвращается значение 1, иначе 0.

Релизный вывод интеграционных тестов:

2023-05-10T12:51:43.147Z	INFO	Integration tests up
=== RUN   TestAddToWhiteList
=== RUN   TestAddToWhiteList/AddToWhiteList_Positive
=== RUN   TestAddToWhiteList/AddToWhiteList_NegativeListCrossCheck
--- PASS: TestAddToWhiteList (0.69s)
    --- PASS: TestAddToWhiteList/AddToWhiteList_Positive (0.45s)
    --- PASS: TestAddToWhiteList/AddToWhiteList_NegativeListCrossCheck (0.24s)
=== RUN   TestRemoveFromWhiteList
=== RUN   TestRemoveFromWhiteList/RemoveFromWhiteList_Positive
=== RUN   TestRemoveFromWhiteList/RemoveFromWhiteList_NegativeNotInBase
--- PASS: TestRemoveFromWhiteList (1.25s)
    --- PASS: TestRemoveFromWhiteList/RemoveFromWhiteList_Positive (0.90s)
    --- PASS: TestRemoveFromWhiteList/RemoveFromWhiteList_NegativeNotInBase (0.35s)
=== RUN   TestIsIPInWhiteList
=== RUN   TestIsIPInWhiteList/IsIPInWhiteList_Positive
=== RUN   TestIsIPInWhiteList/IsIPInWhiteList_NegativeNotInBase
--- PASS: TestIsIPInWhiteList (0.60s)
    --- PASS: TestIsIPInWhiteList/IsIPInWhiteList_Positive (0.34s)
    --- PASS: TestIsIPInWhiteList/IsIPInWhiteList_NegativeNotInBase (0.26s)
=== RUN   TestGetAllIPInWhiteList
=== RUN   TestGetAllIPInWhiteList/GetAllIPInWhiteList_Positive
--- PASS: TestGetAllIPInWhiteList (0.24s)
    --- PASS: TestGetAllIPInWhiteList/GetAllIPInWhiteList_Positive (0.24s)
=== RUN   TestAddToBlackList
=== RUN   TestAddToBlackList/AddToBlackList_Positive
=== RUN   TestAddToBlackList/AddToBlackList_NegativeListCrossCheck
--- PASS: TestAddToBlackList (0.50s)
    --- PASS: TestAddToBlackList/AddToBlackList_Positive (0.25s)
    --- PASS: TestAddToBlackList/AddToBlackList_NegativeListCrossCheck (0.25s)
=== RUN   TestRemoveFromBlackList
=== RUN   TestRemoveFromBlackList/RemoveFromBlackList_Positive
=== RUN   TestRemoveFromBlackList/RemoveFromBlackList_NegativeNotInBase
--- PASS: TestRemoveFromBlackList (0.49s)
    --- PASS: TestRemoveFromBlackList/RemoveFromBlackList_Positive (0.26s)
    --- PASS: TestRemoveFromBlackList/RemoveFromBlackList_NegativeNotInBase (0.23s)
=== RUN   TestIsIPInBlackList
=== RUN   TestIsIPInBlackList/IsIPInBlackList_Positive
=== RUN   TestIsIPInBlackList/IsIPInBlackList_NegativeNotInBase
--- PASS: TestIsIPInBlackList (0.52s)
    --- PASS: TestIsIPInBlackList/IsIPInBlackList_Positive (0.24s)
    --- PASS: TestIsIPInBlackList/IsIPInBlackList_NegativeNotInBase (0.28s)
=== RUN   TestGetAllIPInBlackList
=== RUN   TestGetAllIPInBlackList/GetAllIPInBlackList_Positive
--- PASS: TestGetAllIPInBlackList (0.27s)
    --- PASS: TestGetAllIPInBlackList/GetAllIPInBlackList_Positive (0.27s)
=== RUN   TestClearBucketByLogin
=== RUN   TestClearBucketByLogin/ClearBucketByLogin_Positive
--- PASS: TestClearBucketByLogin (0.21s)
    --- PASS: TestClearBucketByLogin/ClearBucketByLogin_Positive (0.21s)
=== RUN   TestClearBucketByIP
=== RUN   TestClearBucketByIP/ClearBucketByIP_Positive
--- PASS: TestClearBucketByIP (0.22s)
    --- PASS: TestClearBucketByIP/ClearBucketByIP_Positive (0.22s)
=== RUN   TestAuthorizationRequest
=== RUN   TestAuthorizationRequest/AuthorizationRequestSimple_Positive
=== RUN   TestAuthorizationRequest/AuthorizationRequestComplexSynthetic_Positive
--- PASS: TestAuthorizationRequest (0.51s)
    --- PASS: TestAuthorizationRequest/AuthorizationRequestSimple_Positive (0.21s)
    --- PASS: TestAuthorizationRequest/AuthorizationRequestComplexSynthetic_Positive (0.30s)
PASS
2023-05-10T12:51:48.634Z	INFO	exitCode:0



