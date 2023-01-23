CHART_DIR_PATH="charts/auth"
CHART_NAME="auth"
CHART_NAMESPACE="auth"
FORWARD_PORT="8079"
PRIVATE_REGISTRY="registry.gitlab.com"
DOCKER_AUTH_FILE="${HOME}/.docker/config.json"
# https://docs.podman.io/en/latest/markdown/podman-login.1.html#authfile-path
REGISTRY_AUTH_FILE=${DOCKER_AUTH_FILE}

.PHONY: all
all: lint minikube install ingress

.PHONY: lint
lint: deps
	helm lint ${CHART_DIR_PATH}/.

.PHONY: deps
deps:
	helm dependency update ${CHART_DIR_PATH}/.

.PHONY: minikube
minikube:
	minikube delete
	minikube start
	# minikube addons enable ingress

.PHONY: ingress
ingress:
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
	sudo podman logout ${PRIVATE_REGISTRY} --authfile ${REGISTRY_AUTH_FILE}
	# podman protects the registry file unlike docker. If it exists it will throw a permission error for other apps that expect it unpermed.
	sudo rm ${REGISTRY_AUTH_FILE}
	touch login.lock

.PHONY: test
test: lint minikube install

.PHONY: template
template: templates.yaml

templates.yaml:
	helm template --set namespace.name=${CHART_NAMESPACE} --set namespace.create=true ${CHART_DIR_PATH}/. > templates.yaml

.PHONY: install
install: # login.lock
	helm dependency build ${CHART_DIR_PATH}
	kubectl create namespace ${CHART_NAMESPACE}
	# kubectl apply -f login.creds
	# kubectl get -n ${CHART_NAMESPACE} secret regcred --output="jsonpath={.data.\.dockerconfigjson}" | base64 --decode
	helm install --namespace ${CHART_NAMESPACE} ${CHART_NAME} ${CHART_DIR_PATH}/.

.PHONY: users
users:
	@echo "Admin pass in secret:"
	@kubectl get secret -n ${CHART_NAMESPACE} auth -o=jsonpath='{.data.ldapAdminPassword}' | base64 -d
	@echo ""
	@echo "enter above password to login and to see all users:"
	# if you are not using the default domain name (which you should have changed in production) then this will fail as it is hardcoded to example.org so copy, paste, and modify to use
	@kubectl exec deploy/openldap --stdin --tty -n ${CHART_NAMESPACE} -- ldapsearch -H ldap://127.0.0.1:1389 -x -b "dc=example,dc=org" -D "cn=admin,dc=example,dc=org" -W
	# @kubectl get pods -n ${CAHRT_NAMESPACE} -l=app=authelia -o jsonpath='{.metadata.name}' | xargs -I {} kubectl exec --stdin --tty -n ${CHART_NAMESPACE} {} -- ldapsearch -H ldap://127.0.0.1:1389 -x -b "dc=example,dc=org" -D "cn=admin,dc=example,dc=org" -W

.PHONY: pgadmin
pgadmin:
	@kubectl wait --timeout=600s --for=condition=Available=True -n ${CHART_NAMESPACE} deployment pgadmin-deployment
	@echo "admin username:"
	@kubectl -n ${CHART_NAMESPACE} get deployment pgadmin -o jsonpath="{.spec.template.spec.containers[0].env[0].value}"
	@echo "admin password:"
	@kubectl -n ${CHART_NAMESPACE} get secret auth -o jsonpath="{.data.pgAdminPassword}" | base64 -d && echo
	@xdg-open "http://localhost:${FORWARD_PORT}" &
	@kubectl port-forward svc/pgadmin -n ${CHART_NAMESPACE} ${FORWARD_PORT}:http-port

.PHONY: pla
pla:
	@kubectl wait --timeout=600s --for=condition=Available=True -n ${CHART_NAMESPACE} deployment pla-deployment
	@echo "admin username:"
	@kubectl -n ${CHART_NAMESPACE} get deployment pla-deployment -o jsonpath="{.spec.template.spec.containers[0].env[0].value}"
	@echo ""
	@echo "admin password:"
	@kubectl -n ${CHART_NAMESPACE} get secret auth -o jsonpath="{.data.pgAdminPassword}" | base64 -d && echo
	@xdg-open "http://localhost:${FORWARD_PORT}" &
	@kubectl port-forward svc/pla -n ${CHART_NAMESPACE} ${FORWARD_PORT}:http

.PHONY: upgrade
upgrade:
	helm dependency build ${CHART_DIR_PATH}
	helm upgrade --namespace ${CHART_NAMESPACE} ${CHART_NAME} ${CHART_DIR_PATH}/.


.PHONY: uninstall
uninstall:
	helm uninstall --namespace ${CHART_NAMESPACE} ${CHART_NAME}
	kubectl delete namespace ${CHART_NAMESPACE}

.PHONY: perm
perm:
	sudo usermod -aG docker ${USER}

.PHONY: unperm
unperm:
	sudo gpasswd -d ${USER} docker

.PHONY: clean
clean:
	rm -f login.lock login.creds templates.yaml
	minikube delete
	# podman logout {PRIVATE_REGISTRY}

.PHONY: stuck
stuck:
	kubectl api-resources --verbs=list --namespaced -o name | xargs -n 1 kubectl get --show-kind --ignore-not-found -n ${CHART_NAMESPACE}
