package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var client *whatsmeow.Client
var GroupChat string
var JID types.JID

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("error reading .env")
	}
	GroupChat = os.Getenv("TEMPGROUP")
	JID = types.NewJID(os.Getenv("JID"), "s.whatsapp.net")
	//Init Database
	if err := InitDB("./database/chats.db"); err != nil {
		log.Fatal("error", err)
	}
	defer CloseDB()

	//SERVER
	r := mux.NewRouter()

	r.HandleFunc("/", TodoLists).Methods("GET")
	r.HandleFunc("/todolist", TodoLists).Methods("GET")
	r.HandleFunc("/todolist/check/{id}/{value}", TodoListCheck).Methods("GET")
	r.HandleFunc("/todolist/done/{id}/{value}", TodoListDone).Methods("GET")
	r.HandleFunc("/todolist/delete", TodoListDeleteAll).Methods("GET")

	//Textures
	r.HandleFunc("/textures", TexturesList).Methods("GET")
	r.HandleFunc("/textures/list", GetTextureCategories).Methods("GET")
	r.HandleFunc("/textures/category/{category}", GetTexturesFor).Methods("GET")
	r.HandleFunc("/textures/favorite/{filename}/{favorite}", UpdateTexture).Methods("GET")
	r.HandleFunc("/textures/favorites", FavoriteTextures).Methods("GET")
	r.HandleFunc("/chat", SendMessage).Methods("POST")
	//Init Database
	r.HandleFunc("/textures/init", InitTexturesList).Methods("GET")
	r.NotFoundHandler = http.HandlerFunc(PageNotFound)
	r.MethodNotAllowedHandler = http.HandlerFunc(NotAllowed)
	go func() {
		fmt.Println("serving on http:/localhost:8811")
		err := http.ListenAndServe(":8811", r)
		if err != nil {
			fmt.Println("error", err)
		}
	}()
	//WHATSAPP
	dbLog := waLog.Stdout("Database", "ERROR", true)
	container, err := sqlstore.New("sqlite3", "file:database/log_store.db?_foreign_keys=on", dbLog)

	deviceStore, err := container.GetFirstDevice()
	if err != nil {

		panic(err)
	}
	clientLog := waLog.Stdout("Client", "ERROR", true)
	client = whatsmeow.NewClient(deviceStore, clientLog)

	client.AddEventHandler(eventHandler)

	if client.Store.ID == nil {
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				//base64code of it fmt.Println("QR code:", evt.Code)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		fmt.Println("Connected!")
		err = client.Connect()
		if err != nil {
			panic(err)
		}
	}

	// Listen to Ctrl+C (you can also do something else that prevents the program from exiting)
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	client.Disconnect()

}

func eventHandler(evt interface{}) {
	switch v := evt.(type) {

	case *events.Message:
		if v.Info.IsGroup {
			fmt.Println(GroupChat)
			if v.Info.Chat.String() == GroupChat+"@g.us" {

				if v.Message.ProtocolMessage != nil {

					msg := v.Message.ProtocolMessage
					//edit or delete
					switch *v.Message.ProtocolMessage.Type {
					case proto.ProtocolMessage_REVOKE:
						err := UpdateStatus(msg.GetKey().GetId(), "delete", true)
						if err != nil {
							fmt.Println("error updating", err)
						}
					case proto.ProtocolMessage_MESSAGE_EDIT:
						m := msg.GetEditedMessage().GetConversation()
						i := msg.GetKey().GetId()
						if err := UpdateMessage(i, m); err != nil {
							fmt.Println("error", err)
						}

					}
					return

				}

				var chat ChatMessage

				if v.Message.ExtendedTextMessage != nil {
					chat.IsChecked = false
					chat.IsDone = false
					chat.Message = v.Message.GetExtendedTextMessage().GetText()
					chat.TimeStamp = v.Info.Timestamp
					chat.MessageID = v.Info.ID
					err := InsertMessage(chat)
					if err != nil {
						log.Fatal(err)
					}
				}
			}

		}

	}
}
