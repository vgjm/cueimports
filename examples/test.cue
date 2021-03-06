package stack

import (
	"dagger.io/dagger"
	"github.com/h8r-dev/stacks/chain/v3/component/deploy"
	"strings"
	"github.com/h8r-dev/stacks/chain/v3/component/repository"
	utilsKubeconfig "github.com/h8r-dev/stacks/chain/v3/internal/utils/kubeconfig"
	"universe.dagger.io/docker"
	"github.com/h8r-dev/stacks/chain/v3/internal/addon"
	"encoding/json"
	"universe.dagger.io/bash"
	"github.com/h8r-dev/stacks/chain/v3/internal/var"
)

#Install: {
	args: {
		name:           string
		domain:         string
		networkType:    string
		repoVisibility: string
		organization:   string
		githubToken:    dagger.#Secret
		kubeconfig:     dagger.#Secret
		frameworks: [...]
		addons: [...]
	}

	_var: var.#Generator & {
		input: {
			applicationName: args.name
			domain:          args.domain
			networkType:     args.networkType
			organization:    args.organization
			frameworks:      args.frameworks
			addons:          args.addons
		}
	}

	_transformKubeconfig: utilsKubeconfig.#TransformToInternal & {
		input: kubeconfig: args.kubeconfig
	}

	_infra: addon.#ReadInfraConfig & {
		input: kubeconfig: _transformKubeconfig.output.kubeconfig
	}

	_createRepos: repository.#Create & {
		input: {
			appName:         args.name
			scmOrganization: args.organization
			repoVisibility:  args.repoVisibility
			githubToken:     args.githubToken
			kubeconfig:      _transformKubeconfig.output.kubeconfig
			vars:            _var
			frameworks:      args.frameworks
		}
	}

	_deployApp: deploy.#Init & {
		input: {
			name:           args.name
			domain:         args.domain
			repoVisibility: args.repoVisibility
			organization:   args.organization
			githubToken:    args.githubToken
			kubeconfig:     _transformKubeconfig.output.kubeconfig
			frameworks:     args.frameworks
			vars:           _var
			cdVar:          _infra.argoCD
			waitFor:        _createRepos.output.success
		}
	}

	_config: {
		for f in args.frameworks {
			(f.name): framework.#Config & {
				name:   f.name
				addons: args.addons
			}
		}
	}
}
