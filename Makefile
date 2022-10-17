IMG ?= gnmi-fake:latest
NAMESPACE ?= nwctl-system
DEVICE_NAME ?= oc01

.PHONY: docker-build
docker-build:
	docker build -t ${IMG} .

.PHONY: docker-push
docker-push:
	docker push ${IMG}

.PHONY: manifests
manifests: kustomize
	cd config && $(KUSTOMIZE) edit set image gnmi-fake=${IMG}
	cd config && $(KUSTOMIZE) edit set namespace ${NAMESPACE}
	cd config && $(KUSTOMIZE) edit set namesuffix -- -${DEVICE_NAME}
	kubectl kustomize config

.PHONY: deploy
deploy: kustomize
	cd config && $(KUSTOMIZE) edit set image gnmi-fake=${IMG}
	cd config && $(KUSTOMIZE) edit set namespace ${NAMESPACE}
	cd config && $(KUSTOMIZE) edit set namesuffix -- -${DEVICE_NAME}
	kubectl apply -k config

.PHONY: undeploy
undeploy: kustomize
	cd config && $(KUSTOMIZE) edit set image gnmi-fake=${IMG}
	cd config && $(KUSTOMIZE) edit set namespace ${NAMESPACE}
	cd config && $(KUSTOMIZE) edit set namesuffix -- -${DEVICE_NAME}
	kubectl delete -k config


##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"

## Tool Versions
KUSTOMIZE_VERSION ?= v4.5.7

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	test -s $(LOCALBIN)/kustomize || { curl -s $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); }
