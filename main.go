package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

func establishConnection() *websocket.Conn {
	u := url.URL{Scheme: "ws", Host: "localhost:24050", Path: "/ws"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Print("dial:", err)
		log.Print("Retrying connection in 5 seconds...")

		time.Sleep(5 * time.Second)

		return establishConnection()
	}
	return c

}

func main() {
	c := establishConnection()
	defer c.Close()

	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			data := string(message[:])

			if (gjson.Get(data, "settings.menu.mainMenu")).Type.String() == "String" {
				SetId := gjson.Get(data, "menu.bm.set").String()
				OsuSongsDir := gjson.Get(data, "settings.folders.songs").String()
				r, err := http.Get(fmt.Sprintf("https://api.nerinyan.moe/d/%v", SetId))
				if err != nil {
					log.Fatal(err)
				}
				d := r.Header["Content-Disposition"][0]
				replacer := strings.NewReplacer("/", "_", `"`, "", "*", " ", "..", ".")
				filename := replacer.Replace(strings.Split(strings.Split(d, `filename="`)[1], `";`)[0])
				filepath := path.Join(OsuSongsDir, filename)
				f, err := os.Create(filepath)
				if err != nil {
					log.Fatal(err)
				}
				defer f.Close()
				io.Copy(f, r.Body)

			}
		}
	}()
	return
}
