package main

import (
	"database/sql"
	"fmt"
	"time"
)

type ChatMessage struct {
	TimeStamp time.Time
	Message   string
	MessageID string
	IsChecked bool
	IsDone    bool
	IsDeleted bool
}

type Result struct {
	Status bool
	Data   ChatMessage
}
type EmptyResult struct {
	Status bool
}
type ErrorResult struct {
	Status    bool
	ErrorCode int
}

type TextureFile struct {
	Parent   string
	Filename string
	Favorite bool
}

type TexturesCollection struct {
	Parent string
	Files  []string
}

type TextureCategoriesResult struct {
	Status     bool
	Categories []string
}

type TexturesResult struct {
	Status   bool
	Textures []TextureFile
}

type NotFound struct {
	Status bool
	Code   int
}

type ChatData struct {
	Message string
	Channel string
}

var db *sql.DB // Global database connection

func InitDB(databasePath string) error {
	var err error
	db, err = sql.Open("sqlite3", databasePath)
	if err != nil {
		return err
	}
	return createTableIfNotExists()
}

func createTableIfNotExists() error {
	createTable := `
        CREATE TABLE IF NOT EXISTS chat_message (
            timestamp TEXT,
            message TEXT,
            message_id TEXT PRIMARY KEY,
            is_checked INTEGER,
            is_done INTEGER,
			is_deleted INTEGER
        );

		CREATE TABLE IF NOT EXISTS textures (
			parent TEXT,
			filename TEXT,
			favorite INTEGER	
		);
    `
	_, err := db.Exec(createTable)
	return err
}

func CloseDB() error {
	return db.Close()
}

func (c *ChatMessage) String() string {
	return fmt.Sprintf("Timestamp: %s, Message: %s, MessageID: %s, IsChecked: %t, IsDone: %t, IsDeleted: %t",
		c.TimeStamp, c.Message, c.MessageID, c.IsChecked, c.IsDone, c.IsDeleted)
}
func InsertMessage(message ChatMessage) error {
	_, err := db.Exec("INSERT INTO chat_message (timestamp, message, message_id, is_checked, is_done, is_deleted) VALUES (?, ?, ?, ?, ?, ?)",
		message.TimeStamp.Format(time.RFC3339), message.Message, message.MessageID, message.IsChecked, message.IsDone, message.IsDeleted)
	return err
}

func UpdateStatus(messageId string, CheckDone string, value bool) error {
	var query string
	if CheckDone == "check" {
		query = "UPDATE chat_message SET is_checked = ? WHERE message_id=?"
	} else if CheckDone == "done" {
		query = "UPDATE chat_message SET is_done = ? WHERE message_id=?"
	} else {
		query = "UPDATE chat_message SET is_deleted = ? WHERE message_id=?"
	}
	_, err := db.Exec(query, value, messageId)
	return err
}

func UpdateMessage(messageId string, message string) error {

	_, err := db.Exec("UPDATE chat_message SET message = ? WHERE message_id=?", message, messageId)
	fmt.Println("result:", messageId)
	return err
}

func DeleteAll() error {
	_, err := db.Exec("DELETE FROM chat_message")
	return err
}
func GetMessageById(messageId string) *ChatMessage {
	var message ChatMessage
	var timestamp string
	err := db.QueryRow("SELECT * FROM chat_message WHERE message_id=?", messageId).Scan(&timestamp, &message.Message, &message.MessageID, &message.IsChecked, &message.IsDone, &message.IsDeleted)
	if err != nil {
		return nil
	}
	ts, error := time.Parse(time.RFC3339, timestamp)
	if error != nil {
		err = error
		return nil
	}
	message.TimeStamp = ts

	return &message
}

func GetAllMessages() ([]ChatMessage, error) {
	rows, err := db.Query("SELECT timestamp, message, message_id, is_checked, is_done, is_deleted FROM chat_message")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []ChatMessage
	for rows.Next() {
		var message ChatMessage
		var timestamp string
		if err := rows.Scan(&timestamp, &message.Message, &message.MessageID, &message.IsChecked, &message.IsDone, &message.IsDeleted); err != nil {
			return nil, err
		}
		message.TimeStamp, err = time.Parse(time.RFC3339, timestamp)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return messages, nil
}

func GetAllCategories() []string {
	var categories []string
	rows, err := db.Query("SELECT * FROM textures WHERE parent=?", "FREETEXTURE")
	if err != nil {
		fmt.Println("error", err)
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		var file TextureFile
		if err := rows.Scan(&file.Parent, &file.Filename, &file.Favorite); err != nil {
			fmt.Println("error", err)
			return nil

		}
		categories = append(categories, file.Filename)

	}
	return categories
}

func FetchTextures() []TexturesCollection {
	var texturefiles []TextureFile
	rows, err := db.Query("SELECT * FROM textures WHERE filename != ?", "FREETEXTURE")
	if err != nil {
		fmt.Println("error", err)
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		var file TextureFile
		if err := rows.Scan(&file.Parent, &file.Filename, &file.Favorite); err != nil {
			fmt.Println("error", err)
			return nil

		}
		texturefiles = append(texturefiles, file)

	}
	groupTextures := make(map[string][]string)
	for _, tf := range texturefiles {

		groupTextures[tf.Parent] = append(groupTextures[tf.Parent], tf.Filename)
	}
	var collection []TexturesCollection
	for parent, filename := range groupTextures {
		collection = append(collection, TexturesCollection{
			Parent: parent,
			Files:  filename})
	}
	return collection
}

func GetTextureByCategory(category string) []TextureFile {
	var categories []TextureFile
	rows, err := db.Query("SELECT * FROM textures WHERE parent=?", category)
	if err != nil {
		fmt.Println("error", err)
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		var file TextureFile
		if err := rows.Scan(&file.Parent, &file.Filename, &file.Favorite); err != nil {
			fmt.Println("error", err)
			return nil

		}
		categories = append(categories, file)

	}
	return categories
}

func GetFavoriteTextures() []TextureFile {
	var categories []TextureFile
	rows, err := db.Query("SELECT * FROM textures WHERE favorite=?", true)
	if err != nil {
		fmt.Println("error", err)
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		var file TextureFile
		if err := rows.Scan(&file.Parent, &file.Filename, &file.Favorite); err != nil {
			fmt.Println("error", err)
			return nil

		}
		categories = append(categories, file)

	}
	return categories
}
func UpdateTextureFavorite(filename string, isfav bool) error {
	_, err := db.Exec("UPDATE textures SET favorite=? WHERE filename=?", isfav, filename)
	return err
}

func InsertTextures(file TextureFile) {
	var f TextureFile
	err := db.QueryRow("SELECT * FROM textures where filename=?", file.Filename).Scan(&f.Parent, &f.Favorite, &f.Filename)
	if err != nil {

		_, e := db.Exec("INSERT INTO textures (filename, parent, favorite) VALUES (?, ?, ?)", file.Filename, file.Parent, file.Favorite)

		if e != nil {
			fmt.Println("error", e)
		}

	}
}
