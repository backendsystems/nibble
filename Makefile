.PHONY: all build demo history vhs update run pip npm goreleaser apt fix signing bench

all: run

build:
	@go build -o nibble .
	@echo "Built nibble binary"

signing:
	@git config --global gpg.format ssh
	@git config --global user.signingkey "$$(ssh-add -L)"
	@git config --global gpg.ssh.allowedSignersFile ~/.ssh/allowed_signers
	@git config --global commit.gpgsign true
	@mkdir -p ~/.ssh
	@echo "$$(git config user.email) $$(ssh-add -L)" > ~/.ssh/allowed_signers
	@echo "SSH commit signing enabled"

nibble: build

run: nibble
	@./nibble

vhs:
	@if ! command -v vhs >/dev/null 2>&1; then \
		echo "vhs not found. Install it from https://github.com/charmbracelet/vhs"; \
		exit 1; \
	fi

demo: vhs nibble
	@TERM=xterm-256color COLORTERM=truecolor VHS_NO_SANDBOX=1 vhs demo.tape
	@echo "Generated demo.gif"

history: vhs nibble
	@TERM=xterm-256color COLORTERM=truecolor VHS_NO_SANDBOX=1 vhs history.tape
	@echo "Generated history.gif"

pip:
	@cd python-package && \
	if python3 -m venv .venv >/dev/null 2>&1; then \
		. .venv/bin/activate && \
		python -m pip install -U pip build twine && \
		python -m build && \
		python -m twine check dist/*; \
	else \
		echo "python3-venv not available; using user-site fallback"; \
		python3 -m pip install --user --break-system-packages -U build twine && \
		python3 -m build && \
		python3 -m twine check dist/*; \
	fi
	@echo "Built Python package in python-package/dist"

npm:
	@cd npm-package && npm pack --silent
	@echo "Built npm package tarball in npm-package/"

goreleaser:
	@if ! command -v goreleaser >/dev/null 2>&1; then \
		echo "Installing goreleaser locally..."; \
		TMP_DIR=$$(mktemp -d); \
		curl -sL https://github.com/goreleaser/goreleaser/releases/latest/download/goreleaser_Linux_x86_64.tar.gz | tar xz -C "$$TMP_DIR"; \
		sudo mv "$$TMP_DIR/goreleaser" /usr/local/bin/; \
		rm -rf "$$TMP_DIR"; \
	fi
	@goreleaser check
	@goreleaser release --snapshot --clean
	@echo "GoReleaser snapshot validation passed"

apt:
	@if ! command -v debuild >/dev/null 2>&1; then \
		echo "debuild not found. Install with: sudo apt install devscripts build-essential"; \
		exit 1; \
	fi
	@if ! command -v dput >/dev/null 2>&1; then \
		echo "dput not found. Install with: sudo apt install dput"; \
		exit 1; \
	fi
	@echo "Vendoring Go modules..."
	@go mod vendor
	@echo "Building source package..."
	@debuild -S -sa
	@echo "Source package built. Upload to PPA with:"
	@echo "  dput ppa:backendsystems/ppa ../nibble_*_source.changes"

fix:
	@go fmt ./...
	@go vet ./...
	@go fix ./...
	@echo "Code formatted, vetted and fixed"

bench: nibble
	@if [ -z "$(ip)" ]; then echo "Usage: make bench ip=<ip>"; exit 1; fi
	@if ! command -v nmap >/dev/null 2>&1; then echo "nmap not found. Install with: sudo apt install nmap"; exit 1; fi
	@echo "=== nibble (all ports, TCP connect) ==="
	@bash -c 'time ./nibble -i $(ip) -p -'
	@echo ""
	@echo "=== nmap (all ports, TCP connect) ==="
	@bash -c 'time nmap -sT -p - $(ip)'
	@echo ""
	@echo "=== nmap (all ports, SYN scan, sudo) ==="
	@bash -c 'time sudo nmap -sS -p - $(ip)'

move: build
	@sudo cp nibble /usr/local/bin/
	@echo "Installed nibble binary to /usr/local/bin/"
