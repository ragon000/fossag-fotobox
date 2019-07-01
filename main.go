package main

import (
	"flag"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
        "mime"
	"os"
        "encoding/json"
        "fmt"
        "crypto/rand"
        "path/filepath"
        b64 "encoding/base64"

	"github.com/gorilla/websocket"
        "github.com/skip2/go-qrcode"
)

const (

        // path where the uploads go
        uploadPath = "./uploads"
        staticPath = "./static"
        homeHTML = "./templates/index.html"

        maxUploadSize = 16 * 1024 * 1024 // 16MB

)

var (
	addr      = flag.String("addr", ":8080", "http service address")
	homeTempl = template.Must(template.ParseFiles(homeHTML))
        serverChan = make(chan chan string, 4)
        messageChan = make(chan string, 1)
	upgrader  = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)


func messageServer(serverChan chan chan string) {
    var clients []chan string
    // And now we listen to new clients and new messages:
    for {
        select {
        case client, _ := <-serverChan:
            clients = append(clients, client)
        case msg, _ := <-messageChan:
            // Send the uptime to all connected clients:
            for _, c := range clients {
                c <-  msg
            }
        }
    }
}


func server(serverChan chan chan string) {
    var clients []chan string
    for {
        select {
        case client, _ := <-serverChan:
            clients = append(clients, client)
            // Broadcast the number of clients to all clients:
            for _, c := range clients {
                c <- fmt.Sprintf("%d client(s) connected.", len(clients))
            }
        }
    }
}

func uploadFileHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// validate file size
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			renderError(w, "FILE_TOO_BIG", http.StatusBadRequest)
			return
		}

		// parse and validate file and post parameters
		file, _, err := r.FormFile("uploadFile")
		if err != nil {
			renderError(w, "INVALID_FILE", http.StatusBadRequest)
			return
		}
		defer file.Close()
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			renderError(w, "INVALID_FILE", http.StatusBadRequest)
			return
		}

		// check file type, detectcontenttype only needs the first 512 bytes
		fileType := http.DetectContentType(fileBytes)
		switch fileType {
		case "image/jpeg", "image/jpg":
		case "image/gif", "image/png":
			break
		default:
			renderError(w, "INVALID_FILE_TYPE", http.StatusBadRequest)
			return
		}
		fileName := randToken(12)
		fileEndings, err := mime.ExtensionsByType(fileType)
		if err != nil {
			renderError(w, "CANT_READ_FILE_TYPE", http.StatusInternalServerError)
			return
		}
		newPath := filepath.Join(uploadPath, fileName+fileEndings[0])
		fmt.Printf("FileType: %s, File: %s\n", fileType, newPath)

		// write file
		newFile, err := os.Create(newPath)
		if err != nil {
			renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
			return
		}
		defer newFile.Close() // idempotent, okay to call twice
		if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
			renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
			return
		}
		w.Write([]byte("SUCCESS"))
                fileUploaded(r.Host, fileName+fileEndings[0])
	})
}

func fileUploaded(Host string, filename string){
  var png []byte
  var httpstring string = "http://"+Host+"/files/"+filename
  png, err := qrcode.Encode(httpstring, qrcode.Medium, 256)
  if err != nil {
    log.Fatal(err)
  }
  pngenc := b64.StdEncoding.EncodeToString(png)

  resmap := map[string]string{"qr": pngenc, "img": httpstring}
  resjson, _ := json.Marshal(resmap)
  messageChan <- string(resjson)

}

func renderError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(message))
}

func randToken(len int) string {
	b := make([]byte, len)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}
        client := make(chan string, 1)
        serverChan <- client // i have no idea what this go magic is

        for {
          select {
            case text, _ := <-client:
             writer, _ := ws.NextWriter(websocket.TextMessage)
             writer.Write([]byte(text))
             writer.Close()
        }
    }

}

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var v = struct {
		Host    string
	}{
		r.Host,
	}
	homeTempl.Execute(w, &v)
}

func main() {
        // Start the server and keep track of the channel that it receives
        // new clients on:
        os.Mkdir(uploadPath, 0755)
        go messageServer(serverChan)

	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", serveWs)
        http.HandleFunc("/upload", uploadFileHandler())
        uploadfs := http.FileServer(http.Dir(uploadPath)) // File Server for uploads
        staticfs := http.FileServer(http.Dir(staticPath)) // File Server for staic files (js,css)
        http.Handle("/files/", http.StripPrefix("/files", uploadfs))
        http.Handle("/static/", http.StripPrefix("/static", staticfs))
        http.Handle("/favicon.ico", staticfs) // link favicon.ico to the static file server
        log.Printf("Server starting at %s, have fun", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal(err)
	}
}
