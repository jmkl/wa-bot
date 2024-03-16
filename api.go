package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

var startDir = "D:/FREETEXTURE"

func TodoLists(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	messages, err := GetAllMessages()
	if err != nil {
		log.Fatal(err)
	}

	json.NewEncoder(w).Encode(messages)

}

func CheckAndDoneTask(r *http.Request, w http.ResponseWriter, key string) {
	w.Header().Set("Content-type", "application/json")
	params := mux.Vars(r)
	id := params["id"]
	value := params["value"]
	isdone, err := strconv.ParseBool(value)
	if err == nil {
		err := UpdateStatus(id, key, isdone)
		if err == nil {

			chat := GetMessageById(id)
			if chat != nil {
				response := Result{
					Status: true,
					Data:   *chat,
				}
				json.NewEncoder(w).Encode(response)
				return
			}
		}
		fmt.Println(err)
	}

	json.NewEncoder(w).Encode(ErrorResult{Status: false, ErrorCode: 69})

}

func TodoListDone(w http.ResponseWriter, r *http.Request) {
	CheckAndDoneTask(r, w, "done")
}

func TodoListCheck(w http.ResponseWriter, r *http.Request) {
	CheckAndDoneTask(r, w, "check")
}

func TodoListDelete(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-type", "application/json")
	err := DeleteAll()
	if err != nil {

		json.NewEncoder(w).Encode(EmptyResult{Status: true})
		return
	}

	json.NewEncoder(w).Encode(EmptyResult{Status: false})
}

func TodoListDeleteAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	err := DeleteAll()
	if err != nil {
		fmt.Println("error", err)
		json.NewEncoder(w).Encode(EmptyResult{Status: true})
		return
	}
	json.NewEncoder(w).Encode(EmptyResult{Status: false})

}
func TexturesList(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-type", "application/json")
	collections := FetchTextures()
	json.NewEncoder(w).Encode(collections)

}
func InitTexturesList(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-type", "application/json")

	files, err := ListDirs(startDir)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, f := range files {
		InsertTextures(f)
	}
	json.NewEncoder(w).Encode(EmptyResult{Status: false})

}
func GetTextureCategories(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	categories := GetAllCategories()
	json.NewEncoder(w).Encode(TextureCategoriesResult{Status: true, Categories: categories})

}
func GetTexturesFor(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	params := mux.Vars(r)
	category := params["category"]
	categories := GetTextureByCategory(category)
	json.NewEncoder(w).Encode(TexturesResult{Status: true, Textures: categories})

}
func UpdateTexture(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	params := mux.Vars(r)
	filename := params["filename"]
	fav := params["favorite"]
	isfav, err := strconv.ParseBool(fav)
	if err != nil {

		json.NewEncoder(w).Encode(EmptyResult{Status: false})
		return
	}

	UpdateTextureFavorite(filename, isfav)
	json.NewEncoder(w).Encode(EmptyResult{Status: true})
}

func FavoriteTextures(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	textures := GetFavoriteTextures()
	json.NewEncoder(w).Encode(TexturesResult{Status: true, Textures: textures})

}

func PageNotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(NotFound{Status: false, Code: 404})

}

func NotAllowed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(NotFound{Status: false, Code: 405})

}
func SendMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	var data ChatData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if data.Channel == "rh" {
		imagepath := data.Message
		fb, err := os.ReadFile(imagepath)
		if err != nil {

			fmt.Println("error", err)
			return
		}

		uploaded, err := client.Upload(context.Background(), fb, whatsmeow.MediaImage)
		if err != nil {
			fmt.Println("error", err)
			return
		}
		msg := &waProto.Message{ImageMessage: &waProto.ImageMessage{
			Caption:       proto.String(filepath.Base(imagepath)),
			Url:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String(http.DetectContentType(fb)),
			FileEncSha256: uploaded.FileEncSHA256,
			FileSha256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(fb))),
		}}
		client.SendMessage(context.Background(), JID, msg)
	}

	json.NewEncoder(w).Encode(data)

}
