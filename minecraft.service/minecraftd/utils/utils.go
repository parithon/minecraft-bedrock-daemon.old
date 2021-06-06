package utils

import (
	"errors"
	"os"
	"time"

	"github.com/parithon/minecraft-bedrock-daemon/minecraft.service/minecraftd/docker"
)

var (
	lockfilePath string
	lockfile     *os.File
)

func CreateLock(filename string) (*os.File, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		lockfilePath = filename
		lockfile, err = os.OpenFile(filename, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0666)
		return lockfile, err
	}
	return nil, errors.New("an instance is already running")
}

func ReleaseLock() {
	lockfile.Close()
	os.Remove(lockfilePath)
}

func update() {}

func CheckForUpdates() {
	go func() {
		time.Sleep(time.Minute * time.Duration(5))
		if docker.UpdateAvailable() {
			update()
		}
	}()
}
