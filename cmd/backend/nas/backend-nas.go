/*
 * MinIO Cloud Storage, (C) 2018 MinIO, Inc.
 * PGG Obstor, (C) 2021-2026 PGG, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package nas

import (
	"context"

	obstor "github.com/obstor/obstor/cmd"
	"github.com/obstor/obstor/pkg/auth"
	"github.com/obstor/obstor/pkg/madmin"
	"github.com/urfave/cli"
)

func init() {
	const nasBackendTemplate = `NAME:
  {{.HelpName}} - {{.Usage}}

USAGE:
  {{.HelpName}} {{if .VisibleFlags}}[FLAGS]{{end}} PATH
{{if .VisibleFlags}}
FLAGS:
  {{range .VisibleFlags}}{{.}}
  {{end}}{{end}}
PATH:
  path to NAS mount point

EXAMPLES:
  1. Start obstor backend server for NAS backend
     {{.Prompt}} {{.EnvVarSetCommand}} OBSTOR_ROOT_USER{{.AssignmentOperator}}accesskey
     {{.Prompt}} {{.EnvVarSetCommand}} OBSTOR_ROOT_PASSWORD{{.AssignmentOperator}}secretkey
     {{.Prompt}} {{.HelpName}} /shared/nasvol

  2. Start obstor backend server for NAS with edge caching enabled
     {{.Prompt}} {{.EnvVarSetCommand}} OBSTOR_ROOT_USER{{.AssignmentOperator}}accesskey
     {{.Prompt}} {{.EnvVarSetCommand}} OBSTOR_ROOT_PASSWORD{{.AssignmentOperator}}secretkey
     {{.Prompt}} {{.EnvVarSetCommand}} OBSTOR_CACHE_DRIVES{{.AssignmentOperator}}"/mnt/drive1,/mnt/drive2,/mnt/drive3,/mnt/drive4"
     {{.Prompt}} {{.EnvVarSetCommand}} OBSTOR_CACHE_EXCLUDE{{.AssignmentOperator}}"bucket1/*,*.png"
     {{.Prompt}} {{.EnvVarSetCommand}} OBSTOR_CACHE_QUOTA{{.AssignmentOperator}}90
     {{.Prompt}} {{.EnvVarSetCommand}} OBSTOR_CACHE_AFTER{{.AssignmentOperator}}3
     {{.Prompt}} {{.EnvVarSetCommand}} OBSTOR_CACHE_WATERMARK_LOW{{.AssignmentOperator}}75
     {{.Prompt}} {{.EnvVarSetCommand}} OBSTOR_CACHE_WATERMARK_HIGH{{.AssignmentOperator}}85
     {{.Prompt}} {{.HelpName}} /shared/nasvol
`

	_ = obstor.RegisterBackendCommand(cli.Command{
		Name:               obstor.NASBackend,
		Usage:              "Network-attached storage (NAS)",
		Action:             nasBackendMain,
		CustomHelpTemplate: nasBackendTemplate,
		HideHelp:           true,
	})
}

// Handler for 'obstor backend nas' command line.
func nasBackendMain(ctx *cli.Context) {
	// Validate backend arguments.
	if !ctx.Args().Present() || ctx.Args().First() == "help" {
		cli.ShowCommandHelpAndExit(ctx, obstor.NASBackend, 1)
	}

	obstor.StartBackend(ctx, &NAS{ctx.Args().First()})
}

// NAS implements Backend.
type NAS struct {
	path string
}

// Name implements Backend interface.
func (g *NAS) Name() string {
	return obstor.NASBackend
}

// NewBackendLayer returns nas backendlayer.
func (g *NAS) NewBackendLayer(creds auth.Credentials) (obstor.ObjectLayer, error) {
	var err error
	newObject, err := obstor.NewFSObjectLayer(g.path)
	if err != nil {
		return nil, err
	}
	return &nasObjects{newObject}, nil
}

// NAS backend is production-ready
func (g *NAS) Production() bool {
	return true
}

// IsListenSupported returns whether listen bucket notification is applicable for this backend.
func (n *nasObjects) IsListenSupported() bool {
	return false
}

func (n *nasObjects) StorageInfo(ctx context.Context) (si obstor.StorageInfo, _ []error) {
	si, errs := n.ObjectLayer.StorageInfo(ctx)
	si.Backend.GatewayOnline = si.Backend.Type == madmin.FS
	si.Backend.Type = madmin.Gateway
	return si, errs
}

// nasObjects implements backend for Obstor and S3 compatible object storage servers.
type nasObjects struct {
	obstor.ObjectLayer
}

func (n *nasObjects) IsTaggingSupported() bool {
	return true
}
