# install from local source code
```bash
go build -o vidra-cli .

mv vidra-cli "$(go env GOPATH)/bin/"

vidra-cli --help
```

# Generate ZSH completion script for Vidra CLI
```bash
vidra-cli completion zsh > _vidra

mkdir -p ~/.zsh/completions
mv _vidra ~/.zsh/completions/

fpath+=~/.zsh/completions
autoload -Uz compinit
compinit

source ~/.zshrc
```

# Test Vidra CLI

```bash
go test -coverprofile=coverage.out ./internal/...

go tool cover -html=coverage.out
```
