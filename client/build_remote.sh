#!/bin/bash
set -e

rsync --human-readable --archive --xattrs --progress --delete --exclude "/node_modules/*" --exclude "/jspm_packages/*" --exclude "app/*.js" --exclude "app/*.js.map" --exclude "app/**/*.js" --exclude "app/**/*.js.map" /home/cloud/git/pritunl-client/client/ $NPM_SERVER:/home/cloud/pritunl-client-www/

ssh cloud@$NPM_SERVER "
set -e
export PATH=\$HOME/.local/bin:\$PATH
cd /home/cloud/pritunl-client-www/
rm -rf node_modules
npm install
flatpak-node-generator npm package-lock.json -o generated-sources.json
"

scp $NPM_SERVER:/home/cloud/pritunl-client-www/package.json /home/cloud/git/pritunl-client/client/package.json
scp $NPM_SERVER:/home/cloud/pritunl-client-www/package-lock.json /home/cloud/git/pritunl-client/client/package-lock.json
scp $NPM_SERVER:/home/cloud/pritunl-client-www/generated-sources.json /home/cloud/git/pritunl-client/flatpak/generated-sources.json
python3 /home/cloud/git/pritunl-client/tools/flatpak_go.py /home/cloud/git/pritunl-client/service --dest-prefix service --out-dir /home/cloud/git/pritunl-client/flatpak --name service
python3 /home/cloud/git/pritunl-client/tools/flatpak_go.py /home/cloud/git/pritunl-client/cli --dest-prefix cli --out-dir /home/cloud/git/pritunl-client/flatpak --name cli
rsync --human-readable --archive --xattrs --progress --delete $NPM_SERVER:/home/cloud/pritunl-client-www/node_modules/ /home/cloud/git/pritunl-client/client/node_modules/
rsync --human-readable --archive --xattrs --progress --delete --exclude "/node_modules/*" --exclude "/jspm_packages/*" --exclude "app/*.js" --exclude "app/*.js.map" --exclude "app/**/*.js" --exclude "app/**/*.js.map" /home/cloud/git/pritunl-client/client/ $NPM_SERVER:/home/cloud/pritunl-client-www/

ssh cloud@$NPM_SERVER "
cd /home/cloud/pritunl-client-www/
sh build.sh
"

rsync --human-readable --archive --xattrs --progress --delete $NPM_SERVER:/home/cloud/pritunl-client-www/dist/ /home/cloud/git/pritunl-client/client/dist/
rsync --human-readable --archive --xattrs --progress --delete $NPM_SERVER:/home/cloud/pritunl-client-www/dist-dev/ /home/cloud/git/pritunl-client/client/dist-dev/
