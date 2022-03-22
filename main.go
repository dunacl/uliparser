package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gitlab.entel/jvalencia/uliparser/modelo"
)

func init() {
	envPtr := os.Getenv("RUN_AS")

	ambiente := ""
	switch envPtr {
	case "production":
		if err := godotenv.Load(".env.production"); err != nil {
			log.Print("No .env file found")
		}
		ambiente = envPtr

	case "development":
		if err := godotenv.Load(".env.development"); err != nil {
			log.Print("No .env file found")
		}
		ambiente = envPtr

	default:
		if err := godotenv.Load(".env"); err != nil {
			log.Print("No .env file found")
		}
		ambiente = "local"
	}
	version := os.Getenv("VERSION")
	log.Print("RUNNING ULI PARSER VERSION ", version)
	log.Print("Environment: ", ambiente)

}

func main() {
	r := gin.Default()

	// Logging to a file.
	f, _ := os.Create("gin.log")

	// MC
	Mc := memcache.New("192.168.0.6:11211")

	gin.DefaultWriter = io.MultiWriter(f)
	// By default gin.DefaultWriter = os.Stdout
	r.Use(gin.Logger())

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	r.Use(gin.Recovery())

	// test route
	r.GET("/ping", func(c *gin.Context) {
		start := time.Now()
		t := time.Now()
		elapsed := t.Unix() - start.Unix()
		c.JSON(http.StatusOK, gin.H{
			"message":    "pong",
			"time start": t,
			"duration":   elapsed,
		})
	})
	// test
	r.GET("/getlocation", func(c *gin.Context) {
		var jsonobject modelo.EntelNumberRequest
		if err := c.ShouldBindJSON(&jsonobject); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error empty json": err.Error()})
			return
		}

		msisdn := jsonobject.Msisdn
		fetchItem, err := Mc.Get(msisdn)
		if err != nil {
			log.Print("Error ", err)
		}
		data, err := modelo.DecodeData(fetchItem.Value)
		if err != nil {
			log.Print("error:", err)
		}
		c.JSON(http.StatusOK, gin.H{"status": "Tadaa!", "key": msisdn, "data": data})
	})

	// ruta Ulify
	r.POST("/ulify", func(c *gin.Context) {
		/*
			valido := controller.CheckHeaders(c)
			log.Print("funcion CheckHeaders ", valido)
		*/
		var jsonobject modelo.EntelRequest
		if err := c.ShouldBindJSON(&jsonobject); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error empty json": err.Error()})
			return
		}

		// apn := jsonobject.Apn
		msisdn := jsonobject.Msisdn
		uli := jsonobject.Uli
		// log.Print("MSISDN: ", msisdn, " ULI: ", uli, " APN: ", apn)
		// Si el APN contiene Roaming ignorar
		if strings.Contains(jsonobject.Apn, "roaming") {
			c.JSON(http.StatusOK, gin.H{"status": "OK", "detail": "roaming not processed"})
			return
		}
		var ResponseJson modelo.EntelResponseUlify
		if len(uli) == 16 {
			// log.Print("Es registro 3G/2G")
			cell, err := strconv.ParseInt(uli[12:], 16, 64)
			if err != nil {
				panic(err)
			}
			sector, err := strconv.ParseInt("0", 16, 64)
			if err != nil {
				panic(err)
			}
			log.Print("Celda: ", cell, " Sector: ", sector)
			ResponseJson.Celda = cell
			ResponseJson.Uli = uli
			ResponseJson.Sector = sector
			j, err := json.Marshal(ResponseJson)
			if err != nil {
				panic(err)
			}
			Mc.Set(&memcache.Item{Key: msisdn, Value: j})
			c.JSON(http.StatusOK, gin.H{"status": "Tadaa!", "data": ResponseJson})
			return
		}
		if len(uli) == 26 {
			// log.Print("Es registro 4G")
			// log.Print("Cell Hex: ", uli[19:24])
			// log.Print("Sector Hex: ", uli[24:26])
			cell, err := strconv.ParseInt(uli[20:24], 16, 64)
			if err != nil {
				panic(err)
			}
			sector, err := strconv.ParseInt(uli[25:26], 16, 64)
			if err != nil {
				panic(err)
			}
			log.Print("Celda: ", cell, " Sector: ", sector)
			ResponseJson.Celda = cell
			ResponseJson.Uli = uli
			ResponseJson.Sector = sector
			j, err := json.Marshal(ResponseJson)
			if err != nil {
				panic(err)
			}
			Mc.Set(&memcache.Item{Key: msisdn, Value: j})

			c.JSON(http.StatusOK, gin.H{"status": "Tadaa!", "data": ResponseJson})
			return

		}

		c.JSON(http.StatusOK, gin.H{"status": "Tadaa! why are we here?"})

	})
	// anything else
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "Didn't forget something?"})

	})

	port := os.Getenv("PORT")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		log.Printf("listen on: %s\n", srv.Addr)
		//router.RunTLS(":443", "./server.crt", "./server.key")
		//if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
		if err := srv.ListenAndServeTLS("./server.crt", "./server.key"); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Printf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
