//go:build windows

// Command handling for configuration "service" command
package main

import (
	"errors"
	"fmt"
	"io"

	"github.com/urfave/cli/v2"
	"golang.org/x/sys/windows"
)

func cliFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "service",
			Usage: "operate on the service (windows only)",
		},
		&cli.StringFlag{
			Name:  "service-name",
			Value: "Dana2",
			Usage: "service name (windows only)",
		},
		&cli.StringFlag{
			Name:  "service-display-name",
			Value: "Dana2 Data Collector Service",
			Usage: "service display name (windows only)",
		},
		&cli.StringFlag{
			Name:  "service-restart-delay",
			Value: "5m",
		},
		&cli.BoolFlag{
			Name:  "service-auto-restart",
			Usage: "auto restart service on failure (windows only)",
		},
		&cli.BoolFlag{
			Name:  "console",
			Usage: "run as console application (windows only)",
		},
	}
}

func getServiceCommands(outputBuffer io.Writer) []*cli.Command {
	return []*cli.Command{
		{
			Name:  "service",
			Usage: "commands for operate on the Windows service",
			Flags: nil,
			Subcommands: []*cli.Command{
				{
					Name:  "install",
					Usage: "install Dana2 as a Windows service",
					Description: `
The 'install' command with create a Windows service for automatically starting
Dana2 with the specified configuration and service parameters. If no
configuration(s) is specified the service will use the file in
"C:\Program Files\Dana2\Dana2.conf".

To install Dana2 as a service use

> Dana2 service install

In case you are planning to start multiple Dana2 instances as a service,
you must use distinctive service-names for each instance. To install two
services with different configurations use

> Dana2 --config "C:\Program Files\Dana2\Dana2-machine.conf" --service-name Dana2-machine service install
> Dana2 --config "C:\Program Files\Dana2\Dana2-service.conf" --service-name Dana2-service service install
`,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "display-name",
							Value: "Dana2 Data Collector Service",
							Usage: "service name as displayed in the service manager",
						},
						&cli.StringFlag{
							Name:  "restart-delay",
							Value: "5m",
							Usage: "duration for delaying the service restart on failure",
						},
						&cli.BoolFlag{
							Name:  "auto-restart",
							Usage: "enable automatic service restart on failure",
						},
					},
					Action: func(cCtx *cli.Context) error {
						cfg := &serviceConfig{
							displayName:  cCtx.String("display-name"),
							restartDelay: cCtx.String("restart-delay"),
							autoRestart:  cCtx.Bool("auto-restart"),

							configs:    cCtx.StringSlice("config"),
							configDirs: cCtx.StringSlice("config-directory"),
						}
						name := cCtx.String("service-name")
						if err := installService(name, cfg); err != nil {
							return err
						}
						fmt.Fprintf(outputBuffer, "Successfully installed service %q\n", name)
						return nil
					},
				},
				{
					Name:  "uninstall",
					Usage: "remove the Dana2 Windows service",
					Description: `
The 'uninstall' command removes the Dana2 service with the given name. To
remove a service use

> Dana2 service uninstall

In case you specified a custom service-name during install use

> Dana2 --service-name Dana2-machine service uninstall
`,
					Action: func(cCtx *cli.Context) error {
						name := cCtx.String("service-name")
						if err := uninstallService(name); err != nil {
							return err
						}
						fmt.Fprintf(outputBuffer, "Successfully uninstalled service %q\n", name)
						return nil
					},
				},
				{
					Name:  "start",
					Usage: "start the Dana2 Windows service",
					Description: `
The 'start' command triggers the start of the Windows service with the given
name. To start the service either use the Windows service manager or run

> Dana2 service start

In case you specified a custom service-name during install use

> Dana2 --service-name Dana2-machine service start
`,
					Action: func(cCtx *cli.Context) error {
						name := cCtx.String("service-name")
						if err := startService(name); err != nil {
							return err
						}
						fmt.Fprintf(outputBuffer, "Successfully started service %q\n", name)
						return nil
					},
				},
				{
					Name:  "stop",
					Usage: "stop the Dana2 Windows service",
					Description: `
The 'stop' command triggers the stop of the Windows service with the given
name and will wait until the service is actually stopped. To stop the service
either use the Windows service manager or run

> Dana2 service stop

In case you specified a custom service-name during install use

> Dana2 --service-name Dana2-machine service stop
`,
					Action: func(cCtx *cli.Context) error {
						name := cCtx.String("service-name")
						if err := stopService(name); err != nil {
							if errors.Is(err, windows.ERROR_SERVICE_NOT_ACTIVE) {
								fmt.Fprintf(outputBuffer, "Service %q not started\n", name)
								return nil
							}
							return err
						}
						fmt.Fprintf(outputBuffer, "Successfully stopped service %q\n", name)
						return nil
					},
				},
				{
					Name:  "status",
					Usage: "query the Dana2 Windows service status",
					Description: `
The 'status' command queries the current state of the Windows service with the
given name. To query the service either check the Windows service manager or run

> Dana2 service status

In case you specified a custom service-name during install use

> Dana2 --service-name Dana2-machine service status
`,
					Action: func(cCtx *cli.Context) error {
						name := cCtx.String("service-name")
						status, err := queryService(name)
						if err != nil {
							return err
						}
						fmt.Fprintf(outputBuffer, "Service %q is in %q state\n", name, status)
						return nil
					},
				},
			},
		},
	}
}
