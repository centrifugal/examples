package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/centrifugal/gocent/v3"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

func handleConnect(w http.ResponseWriter, r *http.Request) {
	resp := map[string]any{
		"result": map[string]any{
			"user": "test",
		},
	}

	jsonData, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func handleRPC(w http.ResponseWriter, r *http.Request) {
	// Note, we are using response namespace here to inherit required channel options.
	uniqueChannel := "response:" + uuid.NewString()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":     "test",
		"channel": uniqueChannel,
		"exp":     time.Now().Unix() + 60,
	})
	tokenString, err := token.SignedString([]byte("keep-it-secret"))
	if err != nil {
		log.Fatal(err)
	}

	resp := map[string]any{
		"result": map[string]any{
			"data": map[string]any{
				"channel": uniqueChannel,
				"token":   tokenString,
			},
		},
	}

	jsonData, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)

	go func() {
		// Sleep for random time emulating long work.
		time.Sleep(time.Duration(rand.Intn(3000)) * time.Millisecond)

		// Then publish to Centrifugo channel.
		// TODO: Centrifugo API client must be initialized once on start.
		cent := gocent.New(gocent.Config{
			Addr: "http://localhost:8000/api",
			Key:  "keep-it-secret",
		})
		_, err := cent.Publish(context.Background(), uniqueChannel, []byte(`{"result": "result from `+uniqueChannel+`"}`))
		if err != nil {
			log.Printf("Error calling publish: %v", err)
		}
	}()
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./")))
	http.HandleFunc("/rpc", handleRPC)
	http.HandleFunc("/connect", handleConnect)

	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatal(err)
	}
}
