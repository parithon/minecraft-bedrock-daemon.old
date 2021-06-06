package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/parithon/minecraft-bedrock-daemon/minecraft.service/minecraftd/docker"
	"github.com/parithon/minecraft-bedrock-daemon/minecraft.service/minecraftd/utils"
)

type logWriter struct{}

func (writer logWriter) Write(bytes []byte) (int, error) {
	date := time.Now().UTC().Format(time.RFC3339Nano)
	return fmt.Printf("%s %s", date, string(bytes))
}

func main() {

	debug := flag.Bool("debug", false, "Used to setup docker container in debug mode (faster termination)")
	flag.Parse()

	log.SetFlags(0)
	log.SetOutput(new(logWriter))
	log.SetPrefix("INFO ")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)
	signal.Notify(sigs, syscall.SIGTERM)
	signal.Notify(sigs, syscall.SIGQUIT)

	lockfile := fmt.Sprintf("%s.lock", os.Args[0])
	if _, err := utils.CreateLock(lockfile); err != nil {
		log.Fatal(err)
	}

	go func() {
		s := <-sigs
		docker.Shutdown(&s)
	}()

	utils.CheckForUpdates()
	docker.Init(debug)
	docker.Wait()
	docker.Cleanup()
	utils.ReleaseLock()

}
