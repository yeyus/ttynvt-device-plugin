package main

import (
	"fmt"
	"os"

	"github.com/kubevirt/device-plugin-manager/pkg/dpm"
	"github.com/urfave/cli/v2"
	"github.com/yeyus/ttynvt-device-plugin/internal/pkg/deviceplugin"
	"k8s.io/klog/v2"
)

// This variable gets injected at build time, see Dockerfile
var gitDescribe string

func printVersion() {
	versionStrings := [...]string{
		"ttyNVT device plugin for Kubernetes",
		fmt.Sprintln("Jesus Trujillo <elyeyus@gmail.com>"),
		fmt.Sprintf("%s version %s", os.Args[0], gitDescribe),
	}

	for _, v := range versionStrings {
		klog.Infoln(v)
	}
}

func main() {
	klog.Info("Starting cli app")

	app := &cli.App{
		Name:    "ttynvt-device-plugin",
		Usage:   "Device plugin for creating virtual RFC2217 backed TTY ports",
		Version: "v0.1",
	}
	app.Action = func(ctx *cli.Context) error {
		printVersion()
		return start(ctx, app.Flags)
	}

	if err := app.Run(os.Args); err != nil {
		klog.Fatal(err)
	}
}

func start(c *cli.Context, flags []cli.Flag) error {
	klog.Info("Starting ttynvt-device-plugin")
	l := deviceplugin.Lister{
		ResUpdateChan: make(chan dpm.PluginNameList),
		Heartbeat:     make(chan bool),
	}
	manager := dpm.NewManager(&l)

	manager.Run()

	return nil
}
