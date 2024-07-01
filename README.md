
# ELYSIUM MANAGER

## Токены

Arima DJ  7447550770:AAHaO6tRmqNtb53fD9cIXPJVjYi1mHN3i_0 
Demethra 7445477091:AAGOqZ_0_5vTkhHRNfK2iHWgk4ejM8UkL_8
Demethra Test 7486051673:AAGXMsNZ3ia99ljU48IErrA5PH4ZV-VncFo

## HOWTO
### Миграции

##### Накатить новую миграцию

1. Добавить новый файл миграции\
   `migrate create -ext sql -dir ./migrations -seq {name_of_migration_file} `
2. Описать миграцию внутри файла
3. Накатить миграцию\
   `migrate -database 'postgres://login:password@addr:port/db_name?sslmode=disable' -path ./migrations up 1`

Если возника ошибка, необходимо исправить ошибку в файле миграции и зафорсить предыдущую версию миграции и накатить новую:

Попробуйте откатить текущую миграцию:\
`migrate -database 'postgres://login:password@addr:port/db_name?sslmode=disable' -path ./migrations down 1`\

Вам вернется ошибка:\
`Dirty database version 2. Fix and force version.`\
*2* - в этом случае текущая версия базы.\

Необходимо откатить на прошлую. Следует выполнить сначала команду `force` с  текущей версией (в данном случае 2):\
`migrate -database 'postgres://login:password@addr:port/db_name?sslmode=disable' -path ./migrations force 2`


Далее откатить на 1 миграцию назад (не на версию 1 а на одну версию вниз):\
`migrate -database 'postgres://login:password@addr:port/db_name?sslmode=disable' -path ./migrations down 1`