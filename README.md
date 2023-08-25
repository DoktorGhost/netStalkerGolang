Мой первый телеграм-бот под названием  
# Net Stalker Bot (for golang)

Код получает id пользователя vk.com и следит, когда тот появится online, после чего бот отправляет соответствующее сообщение.  
Следить можно за неопределенным количеством людей, а так же удалять из списка отслеживания тех, за кем мы следить не хотим.

Для корректной работы бота нужно создать файл *secrets.go* с таким содержимым:

```golang
package main

const (
	VkToken  = "ваш VK-токкен"
	BotToken = "ваш токкен для телеграм бота"
)
```
Либо же записать их в переменные окружения, в файле *main.go* это 18-19 строки

```golang
//код
  vkTocken := "ваш VK-токкен"
  botTocken := "ваш токкен для телеграм бота"
//код
```

*p.s. возможно, нужно вынести логику добавления/удаления в отдельный пакет, займусь этим в версии 2.0.0*
