GO_CMD ?= go

OPERATOR_REPOSITORY ?= docker.io/rebellions/rbln-npu-operator

HELM_CHART ?= $(CURDIR)/../../deployments/rbln-npu-operator
KUBE_CONTEXT ?=

.PHONY: e2e-test
e2e-test:
	@if [ -z ${KUBECONFIG} ]; then \
		echo "[ERR] KUBECONFIG is missing, must be set"; \
		exit 1; \
	fi
	@if [ -z ${OPERATOR_VERSION} ]; then \
		echo "[ERR] OPERATOR_VERSION is missing, must be set"; \
		exit 1; \
	fi
	$(GO_CMD) test -v . -args \
		-kubeconfig=$(KUBECONFIG) \
		$(if $(strip $(KUBE_CONTEXT)),-context=$(KUBE_CONTEXT)) \
		-operator-repository=$(OPERATOR_REPOSITORY) \
		-operator-version=$(OPERATOR_VERSION) \
		-helm-chart=$(HELM_CHART)
