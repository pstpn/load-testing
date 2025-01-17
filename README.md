# <span style="color:#C0BFEC">🦔 ***Нагрузочное тестирование СБЕР***</span>

## <span style="color:#C0BFEC">📑 ***Описание:*** </span>

Приложение позволяет автоматизированным производить нагрузочное тестирование.

## <span style="color:#C0BFEC">⚙️ ***Конфигурация:*** </span>

```yaml
app:               // Общая информация о приложении
   name: 'sberhl'
   version: 'v0.1'

logger:            // Настройки логгера
   level: 'info'   // Уровень логирования
   file: 'out.log' // Название файла для логов (для вывода в консоль - 'stdout')

worker:            // Настройки воркера, который осуществляет тестирование
   url: 'http://193.168.227.93'
   timeout: 10m
   threads: 5      // Кол-во потоков
```

## <span style="color:#C0BFEC">🏃🏻‍♂️ ***Запуск:*** </span>

1) Перед запуском можно поменять кол-во потоков (поле `threads`) в файле конфигурации `config/config.yaml`
2) Далее необходимо осуществить сборку:
```shell
go build cmd/app/main.go
```
3) Затем запустить исполняемый файл:
```shell
./main
```

Логирующие сообщения пишутся в файл, путь к которому указан в поле `file` конфига `config/config.yaml`