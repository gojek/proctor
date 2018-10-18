package configuration

import (
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/gojektech/proctor/daemon"
	"github.com/gojektech/proctor/io"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewCmd(printer io.Printer, proctorEngineClient daemon.Client) *cobra.Command {
	return &cobra.Command{
		Use:     "config",
		Short:   "configure proctor with arguments given",
		Long:    "To configure a proctor, this command helps configuring proctor by storing emailId and accessToken locally",
		Example: "proctor config set PROCTOR_HOST=example.proctor.com EMAIL_ID=example@proctor.com ACCESS_TOKEN=XXXXX",

		Run: func(cmd *cobra.Command, args []string) {
			proctorHost := args[0]
			emailId := args[1]
			accessToken := args[2]

			viper.AutomaticEnv()
			var configFileDir string

			if viper.GetString("ENVIRONMENT") == "test" {
				configFileDir = "/tmp"
			} else {
				configFileDir = "$HOME/.proctor"
			}

			configFile := filepath.Join(configFileDir, "proctor.yaml")
			if _, err := os.Stat(configFile); os.IsNotExist(err) {
				CreateDirIfNotExist(configFileDir)

				configFileContent := []byte("PROCTOR_HOST:" + proctorHost + "\n" + "EMAIL_ID:" + emailId + "\n" + "ACCESS_TOKEN:" + accessToken)

				f, err := os.Create(configFile)
				if err != nil {
					printer.Println(err.Error(), color.FgRed)
				}
				_, err = f.Write(configFileContent)
				if err != nil {
					printer.Println(err.Error(), color.FgRed)
				}
				defer f.Close()

				printer.Println("Proctor Successfully Configured!!!", color.FgGreen)
			}
		},
	}
}

func CreateDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
}
