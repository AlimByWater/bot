# High-Level Objective
- Добавить методы репозитория для emoji-pack.

# Implementation Notes
- Be sure to fulfill every detail of every task 
- Use another repository methods as reference for implementation

# Low-level tasks
> ordered from start to finish
> generate edits one at a time, so we don't overlap any changes

1. Cоздай файл internal/repository/postgres/elysium/emoji-pack.go.
2. В файле telegram/usecase.go есть интерфейс репозитория. нужно реализовать все методы связанные с emoji-pack в репозитории.
3. Реализовать тесты для новых методов на примере из файла user_test.go