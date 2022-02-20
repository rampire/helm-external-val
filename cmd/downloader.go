package cmd

import (
	"errors"
	"fmt"
	k8s "helm-external-val/util/kubernetes"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// downloaderCmd represents the downloader command
var downloaderCmd = &cobra.Command{
	Use:   "downloader certFile keyFile caFile URL",
	Short: "Get value from a remote source and output it to stdout",
	Long: `Get value from a remote source and output it to stdout.
URL is formatted like below
<protocol_required>://<namespace_optional>/<name_required>/<key_optional>

Helm will invoke this command with the url in the 4th parameter.
See https://helm.sh/docs/topics/plugins/#downloader-plugins.`,
	Args: cobra.ExactArgs(5),
	Run: func(cmd *cobra.Command, args []string) {
		protocol, ns, name, key, err := ParseUrl(args[4])
		if err != nil {
			cmd.PrintErrln(err)
			os.Exit(1)
		}
		switch protocol {
		case "cm":
			ComposeCM(ns, name, key, cmd)
		case "secret":
			ComposeSecret(ns, name, key, cmd)
		}

	},
}

func ComposeSecret(ns string, secretName string, dataKey string, cmd *cobra.Command) {
	client := k8s.GetK8sClient()
	secret, err := k8s.GetSecret(ns, secretName, client)
	if err != nil {
		cmd.PrintErrln(err)
		os.Exit(1)
	}
	values := k8s.ComposeSecretValues(secret, dataKey)
	fmt.Printf("%s\n", values)
}

func ComposeCM(ns string, cmName string, dataKey string, cmd *cobra.Command) {
	client := k8s.GetK8sClient()
	cm, err := k8s.GetConfigMap(ns, cmName, client)
	if err != nil {
		cmd.PrintErrln(err)
		os.Exit(1)
	}
	values := k8s.ComposeValues(cm, dataKey)
	fmt.Printf("%s\n", values)
}

func ParseUrl(url string) (protocol string, namespace string, configMapName string, dataKey string, err error) {
	parsedUrl := strings.Split(url, "://")
	protocol = parsedUrl[0]
	err = nil
	if len(parsedUrl) < 2 {
		err = errors.New(":// missing after protocol")
		return
	}
	config := strings.Split(parsedUrl[1], "/")
	if config[0] == "" {
		err = errors.New("no config provided after protocol")
		return
	} else if len(config) == 1 {
		namespace = "default"
		configMapName = config[0]
		dataKey = "values.yaml"
	} else if len(config) == 2 {
		namespace = config[0]
		configMapName = config[1]
		dataKey = "values.yaml"
	} else {
		namespace = config[0]
		configMapName = config[1]
		dataKey = config[2]
	}
	return
}

func init() {
	rootCmd.AddCommand(downloaderCmd)
}
