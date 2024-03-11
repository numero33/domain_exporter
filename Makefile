

gentest:
	@echo "Generating test data..."
	./gentest.sh $(domains)
	
build:
	@echo "Building..."
	CGO_ENABLED=0 go build -ldflags="-w -s" -o domain_exporter