CHART_DIR_PATH="charts/akm"
CHART_NAME="akm"
CHART_NAMESPACE="auth"
FORWARD_PORT=8079
DOCS_PORT=8078
MINIKUBE_KUBE_VERSION=1.27.1
PRIVATE_REGISTRY="registry.gitlab.com"
DOCKER_AUTH_FILE="${HOME}/.docker/config.json"
# https://docs.podman.io/en/latest/markdown/podman-login.1.html#authfile-path
REGISTRY_AUTH_FILE=${DOCKER_AUTH_FILE}
# CONTAINER_IMAGE=$(cat charts/akm/values.yaml | grep -P -o '(?<=image:\s\").*(?=\")')
# CONTAINER_TAG=registry.gitlab.com/georgeraven/authentik-manager:ldev
LOCAL_TAG=localhost/controller:local

SRC_VERSION=$(shell git describe --abbrev=0)
APP_VERSION=$(shell cat charts/ak/values.yaml | grep -P -o '(?<=ghcr.io/goauthentik/server:).*(?=\")')

# Docs arguments
TAG=akm/docs
CONTAINER_NAME=authentik-manager-docs
DOCS_DOCKERFILE=Dockerfile

.PHONY: help
help: ## display this auto generated help message
	@echo "Please provide a make target:"
	@grep -F -h "##" $(MAKEFILE_LIST) | grep -F -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'


.PHONY: all
all: lint minikube install ingress ## Create minikube cluster and apply operator to it

.PHONY: lint
lint: deps ## Lint the helm chart
	helm lint ${CHART_DIR_PATH}/.

.PHONY: deps
deps:	## Update all helm chart dependencies
	helm dependency update ${CHART_DIR_PATH}/.

.PHONY: minikube
minikube: ## Create a local minikube testing cluster
	minikube delete
	minikube start --driver=podman --kubernetes-version=${MINIKUBE_KUBE_VERSION}
	# minikube addons enable ingress

.PHONY: ingress
ingress: ## Enable minikube ingress addon
	minikube addons enable ingress

# .PHONY: login
# login: login.lock
#
# login.lock:
# 	docker login ${PRIVATE_REGISTRY}
# 	kubectl create -n ${CHART_NAMESPACE} secret generic regcred --from-file=.dockerconfigjson=${HOME}/.docker/config.json --type=kubernetes.io/dockerconfigjson --dry-run=client -o yaml > login.creds
# 	docker logout {PRIVATE_REGISTRY}
# 	touch login.lock

.PHONY: login
login: login.lock

login.lock:
	# please use your username and a token with sufficient permissions to access the repo / registry
	sudo podman login ${PRIVATE_REGISTRY} --authfile ${REGISTRY_AUTH_FILE}
	sudo kubectl create -n ${REGCRED_NAMESPACE} secret generic ${REGCRED_NAME} --from-file=.dockerconfigjson=${REGISTRY_AUTH_FILE} --type=kubernetes.io/dockerconfigjson --dry-run=client -o yaml > login.creds
	# eval $(minikube docker-env)
	sudo podman logout ${PRIVATE_REGISTRY} --authfile ${REGISTRY_AUTH_FILE}
	# podman protects the registry file unlike docker. If it exists it will throw a permission error for other apps that expect it unpermed.
	sudo rm ${REGISTRY_AUTH_FILE}
	touch login.lock

.PHONY: test
test: lint minikube install ## Test application (current does not)

.PHONY: template
template: templates.yaml ## Generate a concrete template for inspection

templates.yaml:
	helm template --set namespace.name=${CHART_NAMESPACE} --set namespace.create=true ${CHART_DIR_PATH}/. > templates.yaml

.PHONY: akm-build
akm-build: ## Build the operator dockerfile
	# just in case things change get the specific image that would have been pulled and build it
	@cd operator && podman build -t ${CONTAINER_TAG} -f Dockerfile .

.PHONY: install-full
install-full: ## Install helm chart to default cluster with registry images
	helm dependency build ${CHART_DIR_PATH}
	helm upgrade --install --create-namespace --namespace ${CHART_NAMESPACE} ${CHART_NAME} ${CHART_DIR_PATH}/.
.PHONY: upgrade-full
upgrade-full: install-full ## Upgrade the operator helm chart using registry

.PHONY: build
build: ## Build the container image
	# https://stackoverflow.com/questions/42564058/how-to-use-local-docker-images-with-minikube
	@cd operator && go mod tidy
	@make -C operator generate manifests
	@echo "Packaging authentik ${APP_VERSION} in authentik-manager ${SRC_VERSION}"
	@helm package --dependency-update --app-version ${APP_VERSION} --version ${SRC_VERSION} --destination operator/helm-charts/. charts/ak
	@cd operator && podman build --build-arg AK_VERSION=${APP_VERSION} --build-arg AKM_VERSION=${SRC_VERSION} -t ${LOCAL_TAG} -f Dockerfile .
	@rm -f controller.tar
	@podman inspect ${LOCAL_TAG}
	@podman save ${LOCAL_TAG} -o controller.tar
	@minikube image load controller.tar
	@rm -f controller.tar

.PHONY: install
install: build ## Install helm chart to default cluster with local images
	helm dependency build ${CHART_DIR_PATH}
	helm upgrade --install --create-namespace --namespace ${CHART_NAMESPACE} --set operator.deployment.imagePullPolicy=Never --set operator.deployment.image=${LOCAL_TAG} ${CHART_NAME} ${CHART_DIR_PATH}/.

.PHONY: upgrade
upgrade: install
	kubectl rollout restart -n auth deployment/authentik-manager

.PHONY: forward
forward: ## Forward authentik worker
	kubectl wait --timeout=600s --for=condition=Available=True -n ${CHART_NAMESPACE} deployment authentik-worker
	kubectl wait --timeout=600s --for=condition=Available=True -n ${CHART_NAMESPACE} deployment authentik-server
	@echo NOTE: the full domain is the .global.domain.full value in the auth chart that you probably set to something else
	@echo Please ensure you have added the full domain to authentik as the following line in your /etc/hosts file:
	@echo
	@echo ...
	@echo 127.0.0.1	auth.example.org
	@echo ...
	@echo
	@echo Please run in a new terminal:
	@echo
	@echo sudo socat TCP-LISTEN:443,fork TCP:127.0.0.1:${FORWARD_PORT}
	@echo xdg-open "https://auth.example.org:${FORWARD_PORT}/if/flow/initial-setup/"
	@echo
	kubectl port-forward svc/authentik-server -n ${CHART_NAMESPACE} ${FORWARD_PORT}:443

.PHONY: proxy
proxy: ## Proxy ingress for local testing through ingress
	kubectl wait --timeout=600s --for=condition=Available=True -n ${CHART_NAMESPACE} deployment authentik-worker
	kubectl wait --timeout=600s --for=condition=Available=True -n ${CHART_NAMESPACE} deployment authentik-server
	minikube -n ingress-nginx service ingress-nginx-controller --url
	#sudo socat TCP-LISTEN:443,fork TCP:192.168.49.2:30312

.PHONY: users
users: ## Defunkt
	@echo "Admin pass in secret:"
	@kubectl get secret -n ${CHART_NAMESPACE} auth -o=jsonpath='{.data.ldapAdminPassword}' | base64 -d
	@echo ""
	@echo "enter above password to login and to see all users:"
	# if you are not using the default domain name (which you should have changed in production) then this will fail as it is hardcoded to example.org so copy, paste, and modify to use
	@kubectl exec deploy/openldap --stdin --tty -n ${CHART_NAMESPACE} -- ldapsearch -H ldap://127.0.0.1:1389 -x -b "dc=example,dc=org" -D "cn=admin,dc=example,dc=org" -W
	# @kubectl get pods -n ${CAHRT_NAMESPACE} -l=app=authelia -o jsonpath='{.metadata.name}' | xargs -I {} kubectl exec --stdin --tty -n ${CHART_NAMESPACE} {} -- ldapsearch -H ldap://127.0.0.1:1389 -x -b "dc=example,dc=org" -D "cn=admin,dc=example,dc=org" -W

.PHONY: pgadmin
pgadmin: ## Defunkt
	@kubectl wait --timeout=600s --for=condition=Available=True -n ${CHART_NAMESPACE} deployment pgadmin
	@echo
	@echo "admin username:"
	@kubectl -n ${CHART_NAMESPACE} get deployment pgadmin -o jsonpath="{.spec.template.spec.containers[0].env[0].value}" && echo
	@echo
	@echo "admin password:"
	@kubectl -n ${CHART_NAMESPACE} get secret auth -o jsonpath="{.data.pgAdminPassword}" | base64 -d && echo
	@echo
	@echo "postgres password:"
	@kubectl -n ${CHART_NAMESPACE} get secret auth -o jsonpath="{.data.postgresPassword}" | base64 -d && echo
	@echo
	@xdg-open "http://localhost:${FORWARD_PORT}" &
	@kubectl port-forward svc/pgadmin -n ${CHART_NAMESPACE} ${FORWARD_PORT}:http-port

.PHONY: pla
pla: ## Defunkt
	@kubectl wait --timeout=600s --for=condition=Available=True -n ${CHART_NAMESPACE} deployment pla-deployment
	@echo "admin username:"
	@echo $(kubectl -n ${CHART_NAMESPACE} get deployment pla-deployment -o jsonpath="{.spec.template.spec.containers[0].env[0].value}") && echo
	@echo "admin password:"
	@kubectl -n ${CHART_NAMESPACE} get secret auth -o jsonpath="{.data.pgAdminPassword}" | base64 -d && echo
	@xdg-open "http://localhost:${FORWARD_PORT}" &
	@kubectl port-forward svc/pla -n ${CHART_NAMESPACE} ${FORWARD_PORT}:http

.PHONY: uninstall
uninstall: ## uninstall the operator helm chart
	helm uninstall --namespace ${CHART_NAMESPACE} ${CHART_NAME}
	# kubectl delete namespace ${CHART_NAMESPACE}

.PHONY: getBlueprint
getBlueprint: ## Fetch currently running blueprints in authentik-worker
	kubectl exec --namespace ${CHART_NAMESPACE} -it deployment/authentik-worker -- ak export_blueprint > export_blueprint.yaml

.PHONY: perm
perm: ## Change users groups to add docker (Defunkt)
	sudo usermod -aG docker ${USER}

.PHONY: unperm
unperm: ## Remove user from docker group (Defunkt)
	sudo gpasswd -d ${USER} docker

.PHONY: clean
clean: ## Wipe most residuals and clean up minikube
	rm -f login.lock login.creds templates.yaml
	minikube delete
	# podman logout {PRIVATE_REGISTRY}

.PHONY: docs
docs: doc-build doc-test doc-run ## Build, test, and run the docs

.PHONY: doc-build
doc-build: ## Build the docs in a container
	sudo podman build -t ${TAG} -f ${DOCS_DOCKERFILE} .

.PHONY: doc-test
doc-test: ## Test the docs
	cd docs/server && go test -short $(go list ./... | grep -v /vendor/)

.PHONY: doc-run
doc-run: doc-build ## Run the docs in a container and open a connection to it
	xdg-open "http://127.0.0.1:${DOCS_PORT}" &
	sudo podman run -p 127.0.0.1:${DOCS_PORT}:8080 -it ${TAG}

.PHONY: dbg-sa
dbg-sa: ## Debug the service account
	@echo
	@echo "NAMESPACED permissions (${CHART_NAMESPACE})"
	@echo "*******************************************"
	@kubectl auth can-i --as=system:serviceaccount:${CHART_NAMESPACE}:authentik-manager --namespace=${CHART_NAMESPACE} --list
	@echo
	@echo "CLUSTER-WIDE permissions"
	@echo "*******************************************"
	@kubectl auth can-i --as=system:serviceaccount:${CHART_NAMESPACE}:authentik-manager --list

.PHONY: stuck
stuck: ## Find any resources with finalizers that are blocking deletion
	kubectl api-resources --verbs=list --namespaced -o name | xargs -n 1 kubectl get --show-kind --ignore-not-found -n ${CHART_NAMESPACE}
