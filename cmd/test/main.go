package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	ttynvt "github.com/yeyus/ttynvt-device-plugin/internal/pkg/wrapper"
	"k8s.io/klog/v2"
)

// This variable gets injected at build time, see Dockerfile
var gitDescribe string

const MAX_INSTANCES = 16

func printInstanceInfo(c chan os.Signal, manager *ttynvt.TTYNVTManager) {
	for {
		<-c
		manager.Print()
	}
}

func printVersion() {
	versionStrings := [...]string{
		"ttyNVT device plugin for Kubernetes",
		fmt.Sprint("Jesus Trujillo <elyeyus@gmail.com>"),
		fmt.Sprintf("%s version %s", os.Args[0], gitDescribe),
	}

	for _, v := range versionStrings {
		klog.Infoln(v)
	}
}

func main() {
	printVersion()
	manager := ttynvt.NewTTYNVTManager(MAX_INSTANCES)
	for i := 0; i < 1; i++ {
		err := manager.Create(fmt.Sprintf("127.0.0.%d", i))
		if err != nil {
			klog.Error(err)
		}
	}

	// SIGUSR1
	printInfo := make(chan os.Signal, 1)
	signal.Notify(printInfo, syscall.SIGUSR1)
	go printInstanceInfo(printInfo, manager)

	// SIGTERM
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	klog.Infoln("Blocking, press ctrl+c to continue...")
	<-done // Will block here until user hits ctrl+c
}
