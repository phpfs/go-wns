# go-wns
Go library for interacting with the WindowsNotificationSystem to send notifications to Windows10 users.


##Example
```golang
package main

import (
	"fmt"
	"github.com/phpfs/go-wns"
)

func main() {

	appID := "ms-app://d-2-34-32-482847284-4884738434-483847264824-47374734774"
	clientSecret := "f82jdja82HDDSJ28SDI83q"

	wnsClient := wns.NewConn(appID, clientSecret)
	fmt.Println(wnsClient.Auth())

	uri := "https://db2.notify.windows.com/?token=27dnandu82hudnaduDNNA829DJAaSJDu24"

	toast := wns.NewToast()
	toast.SetSound("NotificationDefault")
	toast.SetTemplate("ToastText02")
	toast.SetText("This is a heading...", "...and this is a text field :)")
	toast.Build()
	fmt.Println(wnsClient.SendToast(uri, toast))

	badge := wns.NewBadge()
	badge.SetField("newMessage")
	badge.Build()
	fmt.Println(wnsClient.SendBadge(uri, badge))

}

```
