// Copyright 2019 Red Hat, Inc. and/or its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package remove

import (
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/shared"
	"github.com/spf13/cobra"
)

type removeKeycloakFlags struct {
	namespace string
}

type removeKeycloakCommand struct {
	context.CommandContext
	flags   removeKeycloakFlags
	command *cobra.Command
	Parent  *cobra.Command
}

func newRemoveKeycloakCommand(ctx *context.CommandContext, parent *cobra.Command) context.KogitoCommand {
	command := removeKeycloakCommand{
		CommandContext: *ctx,
		Parent:         parent,
	}

	command.RegisterHook()
	command.InitHook()

	return &command
}

func (i *removeKeycloakCommand) Command() *cobra.Command {
	return i.command
}

func (i *removeKeycloakCommand) RegisterHook() {
	i.command = &cobra.Command{
		Use:     "keycloak",
		Short:   "removes installed keycloak instance into the OpenShift/Kubernetes cluster",
		Example: "remove keycloak -p my-project",
		Long:    `removes installed keycloak instance via custom Kubernetes resources.`,
		RunE:    i.Exec,
		PreRun:  i.CommonPreRun,
		PostRun: i.CommonPostRun,
		Args: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}

func (i *removeKeycloakCommand) InitHook() {
	i.flags = removeKeycloakFlags{}
	i.Parent.AddCommand(i.command)
	i.command.Flags().StringVarP(&i.flags.namespace, "project", "p", "", "The project name where the operator will be deployed")
}

func (i *removeKeycloakCommand) Exec(cmd *cobra.Command, args []string) error {
	var err error
	if i.flags.namespace, err = shared.EnsureProject(i.Client, i.flags.namespace); err != nil {
		return err
	}

	return shared.ServicesRemovalBuilder(i.Client, i.flags.namespace).RemoveKeycloak().GetError()
}
