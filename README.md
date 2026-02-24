# pritunl-client-electron: pritunl vpn client

[![package-macOS](https://img.shields.io/badge/package-macOS-cfcfcf.svg?style=flat)](https://github.com/pritunl/pritunl-client-electron/releases)
[![package-windows](https://img.shields.io/badge/package-windows-00adef.svg?style=flat)](https://github.com/pritunl/pritunl-client-electron/releases)
[![github](https://img.shields.io/badge/github-pritunl-11bdc2.svg?style=flat)](https://github.com/pritunl)
[![twitter](https://img.shields.io/badge/twitter-pritunl-55acee.svg?style=flat)](https://twitter.com/pritunl)
[![substack](https://img.shields.io/badge/substack-pritunl-ff6719.svg?style=flat)](https://pritunl.substack.com/)
[![forum](https://img.shields.io/badge/discussion-forum-ffffff.svg?style=flat)](https://forum.pritunl.com)

[Pritunl-client-electron](https://github.com/pritunl/pritunl-client-electron)
is an open source openvpn client. Documentation and more information can be
found at the home page [client.pritunl.com](https://client.pritunl.com)

## Install from Source (macOS)

If the Pritunl package is currently installed run the uninstall command
below. Requires homebrew with git, go and node.

```bash
brew install git go node
bash <(curl -s https://raw.githubusercontent.com/pritunl/pritunl-client-electron/master/tools/install_macos.sh)
```

## Uninstall from Source (macOS)

```bash
bash <(curl -s https://raw.githubusercontent.com/pritunl/pritunl-client-electron/master/tools/uninstall_macos.sh)
```

## Install Linux Client from Source

```bash
sudo dnf -y install git-core wireguard-tools openvpn

sudo rm -rf /usr/local/go
wget https://go.dev/dl/go1.25.5.linux-amd64.tar.gz
echo "9e9b755d63b36acf30c12a9a3fc379243714c1c6d3dd72861da637f336ebb35b go1.25.5.linux-amd64.tar.gz" | sha256sum -c - && sudo tar -C /usr/local -xf go1.25.5.linux-amd64.tar.gz
rm -f go1.25.5.linux-amd64.tar.gz

tee -a ~/.bashrc << EOF
export GOPATH=\$HOME/go
export GOROOT=/usr/local/go
export PATH=/usr/local/go/bin:\$PATH
EOF
source ~/.bashrc

go install github.com/pritunl/pritunl-client-electron/service@latest
go install github.com/pritunl/pritunl-client-electron/cli@latest
sudo cp ~/go/bin/service /usr/bin/pritunl-client-service
sudo cp ~/go/bin/cli /usr/bin/pritunl-client

sudo cp "$(ls -td ~/go/pkg/mod/github.com/pritunl/pritunl-client-electron@*/ | head -n1)/resources_linux/pritunl-client.service" /etc/systemd/system/pritunl-client.service
sudo systemctl daemon-reload
sudo systemctl enable --now pritunl-client.service

sudo pritunl-client add <profile_uri>
sudo pritunl-client list
```
