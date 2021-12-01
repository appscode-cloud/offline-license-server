/*
Copyright AppsCode Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmds

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gomodules.xyz/blobfs"
	"gomodules.xyz/cert"
	"gomodules.xyz/cert/certstore"
	"k8s.io/klog/v2"
)

func NewCmdCreateClient(certDir string) *cobra.Command {
	var (
		org       []string
		prefix    string
		overwrite bool
	)
	cmd := &cobra.Command{
		Use:               "client-cert",
		Short:             "Generate client certificate pair",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				klog.Fatalln("Missing client name.")
			}
			if len(args) > 1 {
				klog.Fatalln("Multiple client name found.")
			}

			cfg := cert.Config{
				AltNames: cert.AltNames{
					DNSNames: []string{args[0]},
				},
				Organization: org,
			}

			store, err := certstore.New(blobfs.New("file:///"), certDir)
			if err != nil {
				fmt.Printf("Failed to create certificate store. Reason: %v.", err)
				os.Exit(1)
			}

			var p []string
			if prefix != "" {
				p = append(p, prefix)
			}
			if store.IsExists(Filename(cfg), p...) && overwrite {
				fmt.Printf("Client certificate found at %s. Do you want to overwrite?", store.Location())
				os.Exit(1)
			}

			if err := store.LoadCA(p...); err != nil {
				fmt.Printf("Failed to load ca certificate. Reason: %v.", err)
				os.Exit(1)
			}

			crt, key, err := store.NewClientCertPair(cfg.AltNames, cfg.Organization...)
			if err != nil {
				fmt.Printf("Failed to generate client certificate pair. Reason: %v.", err)
				os.Exit(1)
			}
			err = store.Write(Filename(cfg), crt, key)
			if err != nil {
				fmt.Printf("Failed to init client certificate pair. Reason: %v.", err)
				os.Exit(1)
			}
			fmt.Println("Wrote client certificates in ", store.Location())
			os.Exit(0)
		},
	}

	cmd.Flags().StringVar(&certDir, "cert-dir", certDir, "Path to directory where pki files are stored.")
	cmd.Flags().StringSliceVarP(&org, "organization", "o", org, "Name of client organizations.")
	cmd.Flags().StringVarP(&prefix, "prefix", "p", prefix, "Prefix added to certificate files")
	cmd.Flags().BoolVar(&overwrite, "overwrite", overwrite, "Overwrite existing cert/key pair")
	return cmd
}
