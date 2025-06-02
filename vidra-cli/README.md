# install from local source code
```bash
go build -o vidra .

mv vidra "$(go env GOPATH)/bin/"

vidra --help
```

# Generate ZSH completion script for Vidra CLI
```bash
vidra completion zsh > _vidra

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
