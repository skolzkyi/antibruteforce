# Antibruteforce
![goworkflow](https://github.com/skolzkyi/antibruteforce/actions/workflows/goworkflow.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/skolzkyi/antibruteforce)](https://goreportcard.com/report/github.com/skolzkyi/antibruteforce)<br/>
Сервис предназначен для борьбы с подбором паролей при авторизации в какой-либо системе.<br/>
Сервис вызывается перед авторизацией пользователя и может либо разрешить, либо заблокировать попытку.<br/>
Cервис используется только для server-server, т.е. скрыт от конечного пользователя.<br/>

# Алгоритм работы
Сервис ограничивает частоту попыток авторизации для различных комбинаций параметров, например:<br/>
<br/>
не более N = 10 попыток в минуту для данного логина.<br/>
не более M = 100 попыток в минуту для данного пароля (защита от обратного brute-force).<br/>
не более K = 1000 попыток в минуту для данного IP (число большое, т.к. NAT).<br/>

Параметры N,M и K задаются через переменные в файле конфигурации ./configs/dc/config.env:<br/>
LIMITFACTOR_LOGIN<br/>
LIMITFACTOR_PASSWORD<br/>
LIMITFACTOR_IP<br/>
<br/>
Временной промежуток хранения частоты попыток в рамках одного измерения задается через параметр LIMIT_TIMECHECK.<br/>
<br/>
Бакеты для подсчета количества попыток хранятся в Redis.<br/>
<br/>
Частью микросервиса является БД MySQL,в ней хранятся подсети для black/white списков.<br/>
Если входящий IP находится в whitelist, то сервис безусловно разрешает авторизацию (ok=true).<br/>
Если - в blacklist, то отклоняет (ok=false).<br/>

# Развертывание
Развертывание микросервиса осуществляется командой make run (внутри docker compose up) в директории с проектом.<br/>
Для проекта предусмотрен CLI-интерфейс, он находится в папке ./bin<br/>
CLI-интерфейс не контейнеризирован, позволяет управлять сервисом внутри docker compose.<br/>
Настройки CLI-интерфейса находятся в файле конфигурации ./configs/config_cli.env<br/>

# Тестирование
Unit-тестирование производится с помощью команды make test.<br/>
Интеграционное тестирование производится с помощью команды  make integration-tests. <br/>
При интеграционном тестировании поднимается тестовое окружение в docker compose, прогоняются интеграционные тесты, в случае ошибки возвращается значение 1, иначе 0.<br/>
<br/>
Релизный вывод интеграционных тестов:<br/>
<br/>
2023-05-10T12:51:43.147Z	INFO	Integration tests up<br/>
=== RUN   TestAddToWhiteList<br/>
=== RUN   TestAddToWhiteList/AddToWhiteList_Positive<br/>
=== RUN   TestAddToWhiteList/AddToWhiteList_NegativeListCrossCheck<br/>
--- PASS: TestAddToWhiteList (0.69s)<br/>
    --- PASS: TestAddToWhiteList/AddToWhiteList_Positive (0.45s)<br/>
    --- PASS: TestAddToWhiteList/AddToWhiteList_NegativeListCrossCheck (0.24s)<br/>
=== RUN   TestRemoveFromWhiteList<br/>
=== RUN   TestRemoveFromWhiteList/RemoveFromWhiteList_Positive<br/>
=== RUN   TestRemoveFromWhiteList/RemoveFromWhiteList_NegativeNotInBase<br/>
--- PASS: TestRemoveFromWhiteList (1.25s)<br/>
    --- PASS: TestRemoveFromWhiteList/RemoveFromWhiteList_Positive (0.90s)<br/>
    --- PASS: TestRemoveFromWhiteList/RemoveFromWhiteList_NegativeNotInBase (0.35s)<br/>
=== RUN   TestIsIPInWhiteList<br/>
=== RUN   TestIsIPInWhiteList/IsIPInWhiteList_Positive<br/>
=== RUN   TestIsIPInWhiteList/IsIPInWhiteList_NegativeNotInBase<br/>
--- PASS: TestIsIPInWhiteList (0.60s)<br/>
    --- PASS: TestIsIPInWhiteList/IsIPInWhiteList_Positive (0.34s)<br/>
    --- PASS: TestIsIPInWhiteList/IsIPInWhiteList_NegativeNotInBase (0.26s)<br/>
=== RUN   TestGetAllIPInWhiteList<br/>
=== RUN   TestGetAllIPInWhiteList/GetAllIPInWhiteList_Positive<br/>
--- PASS: TestGetAllIPInWhiteList (0.24s)<br/>
    --- PASS: TestGetAllIPInWhiteList/GetAllIPInWhiteList_Positive (0.24s)<br/>
=== RUN   TestAddToBlackList<br/>
=== RUN   TestAddToBlackList/AddToBlackList_Positive<br/>
=== RUN   TestAddToBlackList/AddToBlackList_NegativeListCrossCheck<br/>
--- PASS: TestAddToBlackList (0.50s)<br/>
    --- PASS: TestAddToBlackList/AddToBlackList_Positive (0.25s)<br/>
    --- PASS: TestAddToBlackList/AddToBlackList_NegativeListCrossCheck (0.25s)<br/>
=== RUN   TestRemoveFromBlackList<br/>
=== RUN   TestRemoveFromBlackList/RemoveFromBlackList_Positive<br/>
=== RUN   TestRemoveFromBlackList/RemoveFromBlackList_NegativeNotInBase<br/>
--- PASS: TestRemoveFromBlackList (0.49s)<br/>
    --- PASS: TestRemoveFromBlackList/RemoveFromBlackList_Positive (0.26s)<br/>
    --- PASS: TestRemoveFromBlackList/RemoveFromBlackList_NegativeNotInBase (0.23s)<br/>
=== RUN   TestIsIPInBlackList<br/>
=== RUN   TestIsIPInBlackList/IsIPInBlackList_Positive<br/>
=== RUN   TestIsIPInBlackList/IsIPInBlackList_NegativeNotInBase<br/>
--- PASS: TestIsIPInBlackList (0.52s)<br/>
    --- PASS: TestIsIPInBlackList/IsIPInBlackList_Positive (0.24s)<br/>
    --- PASS: TestIsIPInBlackList/IsIPInBlackList_NegativeNotInBase (0.28s)<br/>
=== RUN   TestGetAllIPInBlackList<br/>
=== RUN   TestGetAllIPInBlackList/GetAllIPInBlackList_Positive<br/>
--- PASS: TestGetAllIPInBlackList (0.27s)<br/>
    --- PASS: TestGetAllIPInBlackList/GetAllIPInBlackList_Positive (0.27s)<br/>
=== RUN   TestClearBucketByLogin<br/>
=== RUN   TestClearBucketByLogin/ClearBucketByLogin_Positive<br/>
--- PASS: TestClearBucketByLogin (0.21s)<br/>
    --- PASS: TestClearBucketByLogin/ClearBucketByLogin_Positive (0.21s)<br/>
=== RUN   TestClearBucketByIP<br/>
=== RUN   TestClearBucketByIP/ClearBucketByIP_Positive<br/>
--- PASS: TestClearBucketByIP (0.22s)<br/>
    --- PASS: TestClearBucketByIP/ClearBucketByIP_Positive (0.22s)<br/>
=== RUN   TestAuthorizationRequest<br/>
=== RUN   TestAuthorizationRequest/AuthorizationRequestSimple_Positive<br/>
=== RUN   TestAuthorizationRequest/AuthorizationRequestComplexSynthetic_Positive<br/>
--- PASS: TestAuthorizationRequest (0.51s)<br/>
    --- PASS: TestAuthorizationRequest/AuthorizationRequestSimple_Positive (0.21s)<br/>
    --- PASS: TestAuthorizationRequest/AuthorizationRequestComplexSynthetic_Positive (0.30s)<br/>
PASS<br/>
2023-05-10T12:51:48.634Z	INFO	exitCode:0<br/>



