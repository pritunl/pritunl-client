#!/bin/bash
set -e

rsync --human-readable --archive --xattrs --progress --delete --exclude "/node_modules/*" --exclude "/jspm_packages/*" --exclude "app/*.js" --exclude "app/*.js.map" --exclude "app/**/*.js" --exclude "app/**/*.js.map" /home/cloud/git/pritunl-client/client/ $NPM_SERVER:/home/cloud/pritunl-client-www/

ssh cloud@$NPM_SERVER "
cd /home/cloud/pritunl-client-www/
rm -rf node_modules
npm install
"

scp $NPM_SERVER:/home/cloud/pritunl-client-www/package.json /home/cloud/git/pritunl-client/client/package.json
scp $NPM_SERVER:/home/cloud/pritunl-client-www/package-lock.json /home/cloud/git/pritunl-client/client/package-lock.json
rsync --human-readable --archive --xattrs --progress --delete $NPM_SERVER:/home/cloud/pritunl-client-www/node_modules/ /home/cloud/git/pritunl-client/client/node_modules/
rsync --human-readable --archive --xattrs --progress --delete --exclude "/node_modules/*" --exclude "/jspm_packages/*" --exclude "app/*.js" --exclude "app/*.js.map" --exclude "app/**/*.js" --exclude "app/**/*.js.map" /home/cloud/git/pritunl-client/client/ $NPM_SERVER:/home/cloud/pritunl-client-www/

ssh cloud@$NPM_SERVER "
cd /home/cloud/pritunl-client-www/
sh build.sh
"

rsync --human-readable --archive --xattrs --progress --delete $NPM_SERVER:/home/cloud/pritunl-client-www/dist/ /home/cloud/git/pritunl-client/client/dist/
rsync --human-readable --archive --xattrs --progress --delete $NPM_SERVER:/home/cloud/pritunl-client-www/dist-dev/ /home/cloud/git/pritunl-client/client/dist-dev/
