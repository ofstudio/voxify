.PHONY: mocks

# Generate all mocks
mocks:
	mockery

# Clean all generated mocks
clean-mocks:
	rm -rf ./internal/mocks

# clean local dev data
clean-dev:
	rm -rf ./playground/data/dev.db
	rm -rf ./playground/data/downloads/*
	rm -rf ./playground/data/public/*