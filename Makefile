TEST?=$$(go list ./... | grep -v 'vendor') 
BINARY=terraform-provider-appstore
VERSION=1.0.0
ARCH?=darwin_arm64
VAR_FILE=manifest.json
TF_LOG=ERROR
INSTALL_LOC=~/.terraform.d/plugins/lifeomic.com/tf/appstore/${VERSION}/${ARCH}

default: install

build:
	go build -o ${BINARY}

clean:
	rm -f .terraform.lock.hcl
	rm -rf .terraform
	rm -rf ${INSTALL_LOC}

install: build clean
	install -d ${INSTALL_LOC}
	install ${BINARY} ${INSTALL_LOC}/${BINARY}_v${VERSION}

init: install
	terraform init

plan: init
	TF_LOG=${TF_LOG} terraform plan -var-file=${VAR_FILE}

apply: init
	TF_LOG=${TF_LOG} terraform apply -var-file=${VAR_FILE}

test: 
	go test -i $(TEST) || exit 1                                                   
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4                    

testacc: 
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m   
