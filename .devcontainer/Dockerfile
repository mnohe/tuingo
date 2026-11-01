FROM golang:1.27.1-bookworm

RUN apt-get update \
	&& apt-get install -y --no-install-recommends git openssh-client curl ca-certificates make sudo \
	&& mkdir -p -m 755 /etc/apt/keyrings \
	&& curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg -o /etc/apt/keyrings/githubcli-archive-keyring.gpg \
	&& chmod go+r /etc/apt/keyrings/githubcli-archive-keyring.gpg \
	&& echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" > /etc/apt/sources.list.d/github-cli.list \
	&& apt-get update \
	&& apt-get install -y --no-install-recommends gh \
	&& rm -rf /var/lib/apt/lists/*

ENV COLORTERM=truecolor

# Non-root user. The config directories are created here, owned by it, so the
# named volumes mounted over them start out writable.
RUN useradd --create-home --shell /bin/bash --uid 1000 vscode \
	&& echo "vscode ALL=(root) NOPASSWD:ALL" > /etc/sudoers.d/vscode \
	&& chmod 0440 /etc/sudoers.d/vscode \
	&& mkdir -p /home/vscode/.claude /home/vscode/.codex /home/vscode/.copilot /home/vscode/.config/gh \
	&& chown -R vscode:vscode /home/vscode /go

USER vscode
