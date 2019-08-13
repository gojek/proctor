package config

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"proctor/internal/app/cli/config"
	"proctor/internal/app/cli/utility/io"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func CreateDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
}

func NewCmd(printer io.Printer) *cobra.Command {
	return &cobra.Command{
		Use:     "config",
		Short:   "Configure proctor client",
		Long:    "This command helps configure client with proctord host, email id and access token",
		Example: fmt.Sprintf("proctor config %s=example.proctor.com %s=example@proctor.com %s=XXXXX", config.ProctorHost, config.EmailId, config.AccessToken),
		Args:    cobra.MinimumNArgs(1),

		Run: func(cmd *cobra.Command, args []string) {
			configFile := filepath.Join(config.ConfigFileDir(), "proctor.yaml")
			if _, err := os.Stat(configFile); err == nil {
				printer.Println("[Warning] This will overwrite current config:", color.FgYellow)
				existingProctorConfig, err := ioutil.ReadFile(configFile)
				if err != nil {
					printer.Println(fmt.Sprintf("Error reading config file: %s", configFile), color.FgRed)
					return
				}

				printer.Println(string(existingProctorConfig))
				printer.Println("\nDo you want to continue (Y/n)?\t", color.FgYellow)

				in := bufio.NewReader(os.Stdin)
				userPermission, err := in.ReadString('\n')

				if err != nil {
					printer.Println("Error getting user permission for overwriting config", color.FgRed)
					return
				}

				if userPermission != "y\n" && userPermission != "Y\n" {
					printer.Println("Skipped configuring proctor client", color.FgYellow)
					return
				}
			}

			CreateDirIfNotExist(config.ConfigFileDir())
			var configFileContent string
			for _, v := range args {
				arg := strings.Split(v, "=")

				if len(arg) != 2 {
					printer.Println(fmt.Sprintf("\nIncorrect config key-value pair format: %s. Correct format: CONFIG_KEY=VALUE\n", v), color.FgRed)
					return
				}

				switch arg[0] {
				case config.ProctorHost:
					configFileContent += fmt.Sprintf("%s: %s\n", config.ProctorHost, arg[1])
				case config.EmailId:
					configFileContent += fmt.Sprintf("%s: %s\n", config.EmailId, arg[1])
				case config.AccessToken:
					configFileContent += fmt.Sprintf("%s: %s\n", config.AccessToken, arg[1])
				case config.ConnectionTimeoutSecs:
					configFileContent += fmt.Sprintf("%s: %s\n", config.ConnectionTimeoutSecs, arg[1])
				case config.ProcExecutionStatusPollCount:
					configFileContent += fmt.Sprintf("%s: %s\n", config.ProcExecutionStatusPollCount, arg[1])
				default:
					printer.Println(fmt.Sprintf("Proctor doesn't support config key: %s", arg[0]), color.FgYellow)
				}
			}

			configFileContentBytes := []byte(configFileContent)
			f, err := os.Create(configFile)
			if err != nil {
				printer.Println(fmt.Sprintf("Error creating config file %s: %s", configFile, err.Error()), color.FgRed)
				return
			}
			_, err = f.Write(configFileContentBytes)
			if err != nil {
				printer.Println(fmt.Sprintf("Error writing content %v \n to config file %s: %s", configFileContentBytes, configFile, err.Error()), color.FgRed)
				return
			}
			defer f.Close()
			printer.Println("Proctor client configured successfully", color.FgGreen)
		},
	}
}
