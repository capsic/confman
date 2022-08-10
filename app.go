package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/tidwall/gjson"
	"golang.org/x/term"
)

var encryptionPassword string
var basePath string
var config string            // Confman config
var configData string        // The actual served config
var configurationFile string // Configuration JSON file

const encryptionKey = "ThWmZq4t7w!z$C&F)J@NcRfUjXn2r5u8"

type App struct {
	Router *mux.Router
}

func (a *App) Initialize() {
	bts, err := ioutil.ReadFile(filepath.Join(os.Getenv("CONFMANHOME"), "config.json"))
	if err != nil {
		panic(err)
	}
	config = string(bts)

	// Load config data
	configurationFile = gjson.Get(config, "configurationFile").String()
	encryptionPassword = ""
	if gjson.Get(config, "encrypt").Bool() {
		fmt.Println("Enter Encryption Passphrase: ")
		byteKey, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			panic("Missing/invalid encryption passphrase.")
		}
		encryptionPassword = string(byteKey)
	}
	_LoadConfiguration()

	// Register routes
	router := mux.NewRouter()
	router.Handle("/get", GetConfHandler()).Methods("GET")
	a.Router = router
}

func (a *App) Run() {
	port := gjson.Get(config, "port").String()

	addr := ":" + port
	var err error

	startMessage := "confman started on [::]:" + port

	if gjson.Get(config, "ssl").Bool() {
		startMessage = "confman (secure) started on [::]:" + port
		fmt.Println(startMessage)

		certFile := filepath.Join(os.Getenv("CONFMANHOME"), "cert", gjson.Get(config, "certFile").String())
		keyFile := filepath.Join(os.Getenv("CONFMANHOME"), "cert", gjson.Get(config, "keyFile").String())
		err = http.ListenAndServeTLS(addr, certFile, keyFile, a.Router)
	} else {
		fmt.Println(startMessage)
		err = http.ListenAndServe(addr, a.Router)
	}

	if err != nil {
		panic(err)
	}
}

func GetConfHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		remoteAddr := r.RemoteAddr

		var response interface{}
		if _RemoteAddrInWhitelist(remoteAddr) {
			key := r.FormValue("key")

			response = gjson.Get(configData, key).Value()
		} else {
			response = nil
		}

		_JSONResponse(w, 200, response)
	})
}

func _JSONResponse(w http.ResponseWriter, code int, output interface{}) {
	response, _ := json.Marshal(output)

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(code)

	_, err := w.Write(response)

	if err != nil {
		panic(err)
	}
}

func _RemoteAddrInWhitelist(remoteAddr string) bool {
	whitelist := gjson.Get(config, "ipWhitelist").Array()
	remoteAddrs := strings.Split(remoteAddr, ":")
	r := strings.Join(remoteAddrs[:len(remoteAddrs)-1], ":")

	if len(whitelist) > 0 {
		for _, wl := range whitelist {
			if wl.String() == r {
				return true
			}
		}
		return false
	} else {
		return true
	}
}

func _IsJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

func _ReadConfiguration() string {
	bts, err := ioutil.ReadFile(filepath.Join(os.Getenv("CONFMANHOME"), "data", configurationFile))
	if err != nil {
		panic(err)
	}
	return string(bts)
}

func _LoadConfiguration() {
	fileData := []byte(_ReadConfiguration())

	if _IsJSON(string(fileData)) {
		configData = string(fileData)

		if gjson.Get(config, "encrypt").Bool() {
			if encryptionPassword != "" {
				// Salt with passphrase
				configData = encryptionPassword + ":" + configData + ":" + encryptionPassword

				// Encrypt file
				block, err := aes.NewCipher([]byte(encryptionKey))
				if err != nil {
					panic(err)
				}
				gcm, err := cipher.NewGCM(block)
				if err != nil {
					panic(err)
				}

				nonce := make([]byte, gcm.NonceSize())
				if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
					panic(err)
				}

				ciphertext := gcm.Seal(nonce, nonce, []byte(configData), nil)

				err = ioutil.WriteFile(filepath.Join(basePath, "data", configurationFile), ciphertext, 0777)
				if err != nil {
					panic(err)
				}
			} else {
				panic("Missing/invalid encryption passphrase.")
			}
		}
	} else {
		if encryptionPassword == "" {
			fmt.Println("Configuration encrypted, encryption passphrase needed.")
			fmt.Println("Enter Encryption Passphrase: ")
			byteKey, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				panic("Missing/invalid encryption passphrase.")
			}
			encryptionPassword = string(byteKey)
		}

		if encryptionPassword != "" {
			block, err := aes.NewCipher([]byte(encryptionKey))
			if err != nil {
				fmt.Println("Restore configuration file manually.")
				panic(err)
			}

			gcm, err := cipher.NewGCM(block)
			if err != nil {
				fmt.Println("Restore configuration file manually.")
				panic(err)
			}

			nonce := fileData[:gcm.NonceSize()]
			fileData = fileData[gcm.NonceSize():]
			fileData, err = gcm.Open(nil, nonce, fileData, nil)
			if err != nil {
				fmt.Println("Restore configuration file manually.")
				panic(err)
			}
			configData = string(fileData)

			// Remove salt
			configData = strings.Replace(configData, encryptionPassword+":", "", 1)
			configData = strings.Replace(configData, ":"+encryptionPassword, "", 1)

			if !_IsJSON(configData) {
				panic("Missing/invalid encryption passphrase.")
			} else {
				if !gjson.Get(config, "encrypt").Bool() {
					// Decrypt file
					err = ioutil.WriteFile(filepath.Join(basePath, "data", configurationFile), []byte(configData), 0777)
					if err != nil {
						fmt.Println("Restore configuration file manually.")
						panic(err)
					}
				}
			}
		} else {
			panic("Missing/invalid encryption passphrase, restore configuration file manually.")
		}
	}
}
